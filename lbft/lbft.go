package lbft

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

type Lbft struct {
	node.Node
	election.Election
	pm                     *pacemaker.Pacemaker
	bc                     *blockchain.BlockChain
	notarizedChain         [][]*blockchain.Block
	bufferedBlocks         map[crypto.Identifier]*blockchain.Block
	bufferedQCs            map[crypto.Identifier]*blockchain.QC
	bufferedNotarizedBlock map[crypto.Identifier]*blockchain.QC
	committedBlocks        chan *blockchain.Block
	forkedBlocks           chan *blockchain.Block
	echoedBlock            map[crypto.Identifier]struct{}
	echoedVote             map[crypto.Identifier]struct{}
}

// NewLbft creates a new Lbft instance
func NewLbft(
	node node.Node,
	pm *pacemaker.Pacemaker,
	elec election.Election,
	committedBlocks chan *blockchain.Block,
	forkedBlocks chan *blockchain.Block) *Lbft {
	lb := new(Lbft)
	lb.Node = node
	lb.Election = elec
	lb.pm = pm
	lb.committedBlocks = committedBlocks
	lb.forkedBlocks = forkedBlocks
	lb.bc = blockchain.NewBlockchain(config.GetConfig().N())
	lb.bufferedBlocks = make(map[crypto.Identifier]*blockchain.Block)
	lb.bufferedQCs = make(map[crypto.Identifier]*blockchain.QC)
	lb.bufferedNotarizedBlock = make(map[crypto.Identifier]*blockchain.QC)
	lb.notarizedChain = make([][]*blockchain.Block, 0)
	lb.echoedBlock = make(map[crypto.Identifier]struct{})
	lb.echoedVote = make(map[crypto.Identifier]struct{})
	lb.pm.AdvanceView(0)
	return lb
}

// ProcessBlock processes an incoming block as follows:
// 1. check if the view of the block matches current view (ignore for now)
// 2. check if the view of the block matches the proposer's view (ignore for now)
// 3. insert the block into the block tree
// 4. if the view of the block is lower than the current view, don't vote
// 5. if the block is extending the longest notarized chain, vote for the block
// 6. if the view of the block is higher than the the current view, buffer the block
// and process it when entering that view
func (lb *Lbft) ProcessBlock(block *blockchain.Block) error {
	if lb.bc.Exists(block.ID) {
		return nil
	}
	log.Debugf("[%v] is processing block, view: %v, id: %x", lb.ID(), block.View, block.ID)
	curView := lb.pm.GetCurView()
	if block.View < curView {
		return fmt.Errorf("received a stale block")
	}
	_, err := lb.bc.GetBlockByID(block.PrevID)
	if err != nil && block.View > 1 {
		// buffer future blocks
		lb.bufferedBlocks[block.PrevID] = block
		log.Debugf("[%v] buffer the block for future processing, view: %v, id: %x", lb.ID(), block.View, block.ID)
		return nil
	}
	if !lb.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	if block.Proposer != lb.ID() {
		blockIsVerified, _ := crypto.PubVerify(block.Sig, crypto.IDToByte(block.ID), block.Proposer)
		if !blockIsVerified {
			log.Warningf("[%v] received a block with an invalid signature", lb.ID())
		}
	}
	_, exists := lb.echoedBlock[block.ID]
	if !exists {
		lb.echoedBlock[block.ID] = struct{}{}
		lb.Broadcast(block)
	}
	lb.bc.AddBlock(block)
	shouldVote := lb.votingRule(block)
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", lb.ID(), block.ID)
		lb.bufferedBlocks[block.PrevID] = block
		log.Debugf("[%v] buffer the block for future processing, view: %v, id: %x", lb.ID(), block.View, block.ID)
		return nil
	}
	vote := blockchain.MakeVote(block.View, lb.ID(), block.ID)
	// vote to the current leader
	lb.ProcessVote(vote)
	lb.Broadcast(vote)

	// process buffers
	qc, ok := lb.bufferedQCs[block.ID]
	if ok {
		lb.processCertificate(qc)
	}
	b, ok := lb.bufferedBlocks[block.ID]
	if ok {
		_ = lb.ProcessBlock(b)
	}
	return nil
}

func (lb *Lbft) ProcessVote(vote *blockchain.Vote) {
	log.Debugf("[%v] is processing the vote, block id: %x", lb.ID(), vote.BlockID)
	if vote.Voter != lb.ID() {
		voteIsVerified, err := crypto.PubVerify(vote.Signature, crypto.IDToByte(vote.BlockID), vote.Voter)
		if err != nil {
			log.Fatalf("[%v] Error in verifying the signature in vote id: %x", lb.ID(), vote.BlockID)
			return
		}
		if !voteIsVerified {
			log.Warningf("[%v] received a vote with invalid signature. vote id: %x", lb.ID(), vote.BlockID)
			return
		}
	}
	// echo the message
	_, exists := lb.echoedBlock[vote.BlockID]
	if !exists {
		lb.echoedBlock[vote.BlockID] = struct{}{}
		lb.Broadcast(vote)
	}
	isBuilt, qc := lb.bc.AddVote(vote)
	if !isBuilt {
		//log.Debugf("[%v] votes are not sufficient to build a qc, view: %v, block id: %x", lb.ID(), vote.View, vote.BlockID)
		return
	}
	// send the QC to the next leader
	log.Debugf("[%v] a qc is built, view: %v, block id: %x", lb.ID(), qc.View, qc.BlockID)
	lb.processCertificate(qc)

	return
}

func (lb *Lbft) ProcessRemoteTmo(tmo *pacemaker.TMO) {
	log.Debugf("[%v] is processing tmo from %v", lb.ID(), tmo.NodeID)
	isBuilt, tc := lb.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		log.Debugf("[%v] not enough tc for %v", lb.ID(), tmo.View)
		return
	}
	log.Debugf("[%v] a tc is built for view %v", lb.ID(), tc.View)
	lb.processTC(tc)
}

func (lb *Lbft) ProcessLocalTmo(view types.View) {
	tmo := &pacemaker.TMO{
		View:   view,
		NodeID: lb.ID(),
	}
	lb.Broadcast(tmo)
	lb.ProcessRemoteTmo(tmo)
}

func (lb *Lbft) MakeProposal(view types.View, payload []*message.Transaction) *blockchain.Block {
	prevID := lb.forkChoice()
	block := blockchain.BuildProposal(view, &blockchain.QC{
		View:      0,
		BlockID:   prevID,
		AggSig:    nil,
		Signature: nil,
	}, prevID, payload, lb.ID())
	return block
}

func (lb *Lbft) forkChoice() crypto.Identifier {
	var prevID crypto.Identifier
	if lb.GetNotarizedHeight() == 0 {
		prevID = crypto.MakeID("Genesis block")
	} else {
		tailNotarizedBlock := lb.notarizedChain[lb.GetNotarizedHeight()-1][0]
		prevID = tailNotarizedBlock.ID
	}
	return prevID
}

func (lb *Lbft) processTC(tc *pacemaker.TC) {
	if tc.View < lb.pm.GetCurView() {
		return
	}
	go lb.pm.AdvanceView(tc.View)
}

// 1. advance view
// 2. update notarized chain
// 3. check commit rule
// 4. commit blocks
func (lb *Lbft) processCertificate(qc *blockchain.QC) {
	log.Debugf("[%v] is processing a qc, view: %v, block id: %x", lb.ID(), qc.View, qc.BlockID)
	if qc.View < lb.pm.GetCurView() {
		return
	}
	_, err := lb.bc.GetBlockByID(qc.BlockID)
	if err != nil && qc.View > 1 {
		log.Debugf("[%v] buffered the QC, view: %v, id: %x", lb.ID(), qc.View, qc.BlockID)
		lb.bufferedQCs[qc.BlockID] = qc
		return
	}
	if qc.Leader != lb.ID() {
		quorumIsVerified, _ := crypto.VerifyQuorumSignature(qc.AggSig, qc.BlockID, qc.Signers)
		if quorumIsVerified == false {
			log.Warningf("[%v] received a quorum with invalid signatures", lb.ID())
			return
		}
	}
	err = lb.updateNotarizedChain(qc)
	if err != nil {
		// the corresponding block does not exist
		log.Debugf("[%v] cannot notarize the block, %x: %w", lb.ID(), qc.BlockID, err)
		return
	}
	lb.pm.AdvanceView(qc.View)
	if qc.View < 3 {
		return
	}
	ok, block := lb.commitRule()
	if !ok {
		return
	}
	committedBlocks, forkedBlocks, err := lb.bc.CommitBlock(block.ID, lb.pm.GetCurView())
	if err != nil {
		log.Errorf("[%v] cannot commit blocks", lb.ID())
		return
	}
	for _, cBlock := range committedBlocks {
		lb.committedBlocks <- cBlock
		delete(lb.echoedBlock, cBlock.ID)
		delete(lb.echoedVote, cBlock.ID)
		log.Debugf("[%v] is going to commit block, view: %v, id: %x", lb.ID(), cBlock.View, cBlock.ID)
	}
	for _, fBlock := range forkedBlocks {
		lb.forkedBlocks <- fBlock
		log.Debugf("[%v] is going to collect forked block, view: %v, id: %x", lb.ID(), fBlock.View, fBlock.ID)
	}
	b, ok := lb.bufferedBlocks[qc.BlockID]
	if ok {
		log.Debugf("[%v] found a buffered block by qc, qc.BlockID: %x", lb.ID(), qc.BlockID)
		_ = lb.ProcessBlock(b)
		delete(lb.bufferedBlocks, qc.BlockID)
	}
	qc, ok = lb.bufferedNotarizedBlock[qc.BlockID]
	if ok {
		log.Debugf("[%v] found a bufferred qc, view: %v, block id: %x", lb.ID(), qc.View, qc.BlockID)
		lb.processCertificate(qc)
		delete(lb.bufferedQCs, qc.BlockID)
	}
}

func (lb *Lbft) updateNotarizedChain(qc *blockchain.QC) error {
	block, err := lb.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot find the block")
	}
	// check the last block in the notarized chain
	// could be improved by checking view
	if lb.GetNotarizedHeight() == 0 {
		log.Debugf("[%v] is processing the first notarized block, view: %v, id: %x", lb.ID(), qc.View, qc.BlockID)
		newArray := make([]*blockchain.Block, 0)
		newArray = append(newArray, block)
		lb.notarizedChain = append(lb.notarizedChain, newArray)
		return nil
	}
	for i := lb.GetNotarizedHeight() - 1; i >= 0 || i >= lb.GetNotarizedHeight()-3; i-- {
		lastBlocks := lb.notarizedChain[i]
		for _, b := range lastBlocks {
			if b.ID == block.PrevID {
				var blocks []*blockchain.Block
				if i < lb.GetNotarizedHeight()-1 {
					blocks = make([]*blockchain.Block, 0)
				}
				blocks = append(blocks, block)
				lb.notarizedChain = append(lb.notarizedChain, blocks)
				return nil
			}
		}
	}
	lb.bufferedNotarizedBlock[block.PrevID] = qc
	log.Debugf("[%v] the parent block is not notarized, buffered for now, view: %v, block id: %x", lb.ID(), qc.View, qc.BlockID)
	return fmt.Errorf("the block is not extending the notarized chain")
}

func (lb *Lbft) GetChainStatus() string {
	chainGrowthRate := lb.bc.GetChainGrowth()
	blockIntervals := lb.bc.GetBlockIntervals()
	return fmt.Sprintf("[%v] The current view is: %v, chain growth rate is: %v, ave block interval is: %v", lb.ID(), lb.pm.GetCurView(), chainGrowthRate, blockIntervals)
}

func (lb *Lbft) GetNotarizedHeight() int {
	return len(lb.notarizedChain)
}

// 1. get the tail of the longest notarized chain (could be more than one)
// 2. check if the block is extending one of them
func (lb *Lbft) votingRule(block *blockchain.Block) bool {
	if block.View <= 2 {
		return true
	}
	lastBlocks := lb.notarizedChain[lb.GetNotarizedHeight()-1]
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
func (lb *Lbft) commitRule() (bool, *blockchain.Block) {
	height := lb.GetNotarizedHeight()
	if height < 3 {
		return false, nil
	}
	lastBlocks := lb.notarizedChain[height-1]
	if len(lastBlocks) != 1 {
		return false, nil
	}
	lastBlock := lastBlocks[0]
	secondBlocks := lb.notarizedChain[height-2]
	if len(secondBlocks) != 1 {
		return false, nil
	}
	secondBlock := secondBlocks[0]
	firstBlocks := lb.notarizedChain[height-3]
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
