package streamlet

import (
	"fmt"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/election"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/node"
	"github.com/gitferry/bamboo/pacemaker"
	"github.com/gitferry/bamboo/types"
)

type Streamlet struct {
	node.Node
	election.Election
	pm              *pacemaker.Pacemaker
	bc              *blockchain.BlockChain
	notarizedChain  [][]*blockchain.Block
	bufferedBlocks  map[crypto.Identifier]*blockchain.Block
	bufferedQCs     map[crypto.Identifier]*blockchain.QC
	committedBlocks chan *blockchain.Block
}

// NewStreamlet creates a new Streamlet instance
func NewStreamlet(
	node node.Node,
	pm *pacemaker.Pacemaker,
	elec election.Election,
	committedBlocks chan *blockchain.Block) *Streamlet {
	sl := new(Streamlet)
	sl.Node = node
	sl.Election = elec
	sl.pm = pm
	sl.committedBlocks = committedBlocks
	sl.bc = blockchain.NewBlockchain(config.GetConfig().N())
	sl.bufferedBlocks = make(map[crypto.Identifier]*blockchain.Block)
	sl.bufferedQCs = make(map[crypto.Identifier]*blockchain.QC)
	sl.notarizedChain = make([][]*blockchain.Block, 0)
	sl.pm.AdvanceView(0)
	return sl
}

// ProcessBlock processes an incoming block as follows:
// 1. check if the view of the block matches current view (ignore for now)
// 2. check if the view of the block matches the proposer's view (ignore for now)
// 3. insert the block into the block tree
// 4. if the view of the block is lower than the current view, don't vote
// 5. if the block is extending the longest notarized chain, vote for the block
// 6. if the view of the block is higher than the the current view, buffer the block
// and process it when entering that view
func (sl *Streamlet) ProcessBlock(block *blockchain.Block) error {
	log.Debugf("[%v] is processing block, view: %v, id: %x", sl.ID(), block.View, block.ID)
	curView := sl.pm.GetCurView()
	if block.View < curView {
		return fmt.Errorf("received a stale block")
	}
	if block.View > curView {
		// buffer future blocks
		sl.bufferedBlocks[block.PrevID] = block
		log.Debugf("[%v] buffer the block for future processing")
		return nil
	}
	if !sl.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	sl.bc.AddBlock(block)
	shouldVote := sl.votingRule(block)
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", sl.ID(), block.ID)
		return nil
	}
	vote := blockchain.MakeVote(block.View, sl.ID(), block.ID)
	// TODO: sign the vote
	// vote to the current leader
	sl.ProcessVote(vote)
	sl.Broadcast(vote)

	b, ok := sl.bufferedBlocks[block.ID]
	if ok {
		return sl.ProcessBlock(b)
	}
	return nil
}

func (sl *Streamlet) ProcessVote(vote *blockchain.Vote) {
	isBuilt, qc := sl.bc.AddVote(vote)
	if !isBuilt {
		return
	}
	// send the QC to the next leader
	log.Debugf("[%v] a qc is built, block id: %x", sl.ID(), qc.BlockID)
	sl.processCertificate(qc)

	return
}

func (sl *Streamlet) ProcessRemoteTmo(tmo *pacemaker.TMO) {
	log.Debugf("[%v] is processing tmo from %v", sl.ID(), tmo.NodeID)
	isBuilt, tc := sl.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		log.Debugf("[%v] not enough tc for %v", sl.ID(), tmo.View)
		return
	}
	log.Debugf("[%v] a tc is built for view %v", sl.ID(), tc.View)
	sl.processTC(tc)
}

func (sl *Streamlet) ProcessLocalTmo(view types.View) {
	tmo := &pacemaker.TMO{
		View:   view + 1,
		NodeID: sl.ID(),
	}
	sl.Broadcast(tmo)
	sl.ProcessRemoteTmo(tmo)
}

func (sl *Streamlet) MakeProposal(payload []*message.Transaction) *blockchain.Block {
	var prevID crypto.Identifier
	if sl.pm.GetCurView() == 1 {
		prevID = crypto.MakeID("Genesis block")
	} else {
		tailNotarizedBlock := sl.notarizedChain[sl.GetNotarizedHeight()][0]
		prevID = tailNotarizedBlock.ID
	}
	block := blockchain.MakeBlock(sl.pm.GetCurView(), &blockchain.QC{
		View:      0,
		BlockID:   prevID,
		AggSig:    nil,
		Signature: nil,
	}, prevID, payload, sl.ID())
	return block
}

func (sl *Streamlet) processTC(tc *pacemaker.TC) {
	if tc.View < sl.pm.GetCurView() {
		return
	}
	sl.pm.UpdateTC(tc)
	go sl.pm.AdvanceView(tc.View)
}

// 1. advance view
// 2. update notarized chain
// 3. check commit rule
// 4. commit blocks
func (sl *Streamlet) processCertificate(qc *blockchain.QC) {
	if qc.View < sl.pm.GetCurView() {
		return
	}
	sl.pm.AdvanceView(qc.View)
	err := sl.updateNotarizedChain(qc)
	if err != nil {
		// the corresponding block does not exist, buffer the qc
		sl.bufferedQCs[qc.BlockID] = qc
		log.Warningf("[%v] cannot notarize the block, %x: %w", sl.ID(), qc.BlockID, err)
		return
	}
	if qc.View < 3 {
		return
	}
	ok, block := sl.commitRule()
	if !ok {
		return
	}
	committedBlocks, err := sl.bc.CommitBlock(block.ID)
	if err != nil {
		log.Errorf("[%v] cannot commit blocks", sl.ID())
		return
	}
	go func() {
		for _, block := range committedBlocks {
			sl.committedBlocks <- block
		}
	}()
}

func (sl *Streamlet) updateNotarizedChain(qc *blockchain.QC) error {
	block, err := sl.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot find the block")
	}
	// check the last block in the notarized chain
	// could be improved by checking view
	for i := sl.GetNotarizedHeight() - 1; i >= 0; i-- {
		lastBlocks := sl.notarizedChain[i]
		for _, b := range lastBlocks {
			if b.ID == block.PrevID {
				newArray := make([]*blockchain.Block, 0)
				newArray = append(newArray, block)
				sl.notarizedChain = append(sl.notarizedChain, newArray)
				return nil
			}
		}
	}
	return fmt.Errorf("the block is not extending the notarized chain")
}

func (sl *Streamlet) GetNotarizedHeight() int {
	return len(sl.notarizedChain) + 1
}

// 1. get the tail of the longest notarized chain (could be more than one)
// 2. check if the block is extending one of them
func (sl *Streamlet) votingRule(block *blockchain.Block) bool {
	if block.View <= 2 {
		return true
	}
	lastBlocks := sl.notarizedChain[sl.GetNotarizedHeight()-1]
	for _, b := range lastBlocks {
		if block.PrevID == b.ID {
			return true
		}
	}

	return false
}

// 1. get the last three blocks in the notarized chain
// 2. check if they are consecutive
// 3. if so, return the second block to commit
func (sl *Streamlet) commitRule() (bool, *blockchain.Block) {
	height := sl.GetNotarizedHeight()
	if height < 3 {
		return false, nil
	}
	lastBlocks := sl.notarizedChain[height-1]
	if len(lastBlocks) != 1 {
		return false, nil
	}
	lastBlock := lastBlocks[0]
	secondBlocks := sl.notarizedChain[height-2]
	if len(secondBlocks) != 1 {
		return false, nil
	}
	secondBlock := secondBlocks[0]
	firstBlocks := sl.notarizedChain[height-3]
	if len(firstBlocks) != 1 {
		return false, nil
	}
	firstBlock := firstBlocks[0]
	// check three-chain
	if ((firstBlock.View + 1) == secondBlock.View) && ((secondBlock.View + 1) == lastBlock.View) {
		return true, secondBlock
	}
	return false, nil
}
