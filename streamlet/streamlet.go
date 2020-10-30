package streamlet

import (
	"fmt"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
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
	sl.bc = blockchain.NewBlockchain(config.GetConfig().N())
	sl.committedBlocks = committedBlocks
	return sl
}

func (sl *Streamlet) ProcessBlock(block *blockchain.Block) error {
	log.Debugf("[%v] is processing block, view: %v, id: %x", sl.ID(), block.View, block.ID)
	// TODO: should uncomment the following checks
	if !sl.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	sl.bc.AddBlock(block)

	shouldVote, err := sl.votingRule(block)
	// TODO: add block buffer
	if err != nil {
		return fmt.Errorf("cannot decide whether to vote: %w", err)
	}
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", sl.ID(), block.ID)
		return nil
	}
	vote := blockchain.MakeVote(block.View, sl.ID(), block.ID)
	// TODO: sign the vote
	// vote to the current leader
	sl.ProcessVote(vote)
	sl.Broadcast(vote)
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
	// TODO: choose the tail of the longest notarized chain
	block := blockchain.MakeBlock(sl.pm.GetCurView(), sl.bc.GetHighQC(), payload, sl.ID())
	return block
}

func (sl *Streamlet) processTC(tc *pacemaker.TC) {
	if tc.View < sl.pm.GetCurView() {
		return
	}
	sl.pm.UpdateTC(tc)
	go sl.pm.AdvanceView(tc.View)
}

func (sl *Streamlet) processCertificate(qc *blockchain.QC) {
	if qc.View < sl.pm.GetCurView() {
		return
	}
	go sl.pm.AdvanceView(qc.View)
	sl.bc.UpdateHighQC(qc)
	log.Debugf("[%v] has advanced to view %v", sl.ID(), sl.pm.GetCurView())
	sl.updateStateByQC(qc)
	log.Debugf("[%v] has updated state by qc: %v", sl.ID(), qc.View)
	if qc.View < 3 {
		return
	}
	ok, block, _ := sl.commitRule(qc)
	if !ok {
		return
	}
	committedBlocks, err := sl.bc.CommitBlock(block.ID)
	if err != nil {
		log.Errorf("[%v] cannot commit blocks", sl.ID())
		return
	}
	for _, block := range committedBlocks {
		sl.committedBlocks <- block
	}
}

func (sl *Streamlet) votingRule(block *blockchain.Block) (bool, error) {
	if block.View <= 2 {
		return true, nil
	}
	// TODO: check if the block is extending the longest notarized chain

	return true, nil
}

func (sl *Streamlet) commitRule(qc *blockchain.QC) (bool, *blockchain.Block, error) {
	// TODO: chop off the tail block of the longest notarized chain and
	// commit the remaining blocks that have not been committed

	return false, nil, nil
}

func (sl *Streamlet) updateStateByQC(qc *blockchain.QC) error {
	if qc.View <= 2 {
		return nil
	}
	// TODO: update the notarized chain
	return nil
}
