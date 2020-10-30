package streamlet

import (
	"encoding/gob"
	"fmt"
	"github.com/gitferry/bamboo/node"
	"sync"

	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/election"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
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

func NewHotStuff(
	node node.Node,
	pm *pacemaker.Pacemaker,
	elec election.Election,
	committedBlocks chan *blockchain.Block) *Streamlet {
	hs := new(HotStuff)
	hs.Node = node
	hs.Election = elec
	hs.pm = pm
	hs.bc = blockchain.NewBlockchain(config.GetConfig().N())
	hs.committedBlocks = committedBlocks
	hs.Register(blockchain.QC{}, hs.HandleQC)
	gob.Register(blockchain.QC{})
	return hs
}

func (hs *HotStuff) HandleQC(qc blockchain.QC) {
	log.Debugf("[%v] received a QC", hs.ID())
	hs.processCertificate(&qc)
}

func (hs *HotStuff) ProcessBlock(block *blockchain.Block) error {
	log.Debugf("[%v] is processing block, view: %v, id: %x", hs.ID(), block.View, block.ID)
	hs.processCertificate(block.QC)
	// TODO: should uncomment the following checks
	//curView := r.pm.GetCurView()
	//if block.View != curView {
	//	log.Warningf("[%v] received a stale proposal from %v", r.ID(), block.Proposer)
	//	return
	//}
	if !hs.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	hs.bc.AddBlock(block)

	//shouldVote, err := r.VotingRule(block)
	// TODO: add block buffer
	//if err != nil {
	//	log.Errorf("cannot decide whether to vote the block, %w", err)
	//	return
	//}
	//if !shouldVote {
	//	log.Debugf("[%v] is not going to vote for block, id: %x", r.ID(), block.ID)
	//	return
	//}
	vote := blockchain.MakeVote(block.View, hs.ID(), block.ID)
	//err = r.UpdateStateByView(vote.View)
	//if err != nil {
	//	log.Warningf("cannot update state after voting: %w", err)
	//}
	// TODO: sign the vote
	// vote to the current leader
	voteAggregator := block.Proposer
	if voteAggregator == hs.ID() {
		hs.ProcessVote(vote)
	} else {
		hs.Send(voteAggregator, vote)
	}
	return nil
}

func (hs *HotStuff) ProcessVote(vote *blockchain.Vote) {
	isBuilt, qc := hs.bc.AddVote(vote)
	if !isBuilt {
		return
	}
	// send the QC to the next leader
	log.Debugf("[%v] a qc is built, block id: %x", hs.ID(), qc.BlockID)
	nextLeader := hs.FindLeaderFor(qc.View + 1)
	if nextLeader == hs.ID() {
		if config.Configuration.IsByzantine(nextLeader) {
			hs.preprocessQC(qc)
		} else {
			hs.processCertificate(qc)
		}
	} else {
		hs.Send(nextLeader, qc)
	}

	return
}

func (hs *HotStuff) ProcessRemoteTmo(tmo *pacemaker.TMO) {
	log.Debugf("[%v] is processing tmo from %v", hs.ID(), tmo.NodeID)
	hs.bc.UpdateHighQC(tmo.HighQC)
	isBuilt, tc := hs.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		log.Debugf("[%v] not enough tc for %v", hs.ID(), tmo.View)
		return
	}
	log.Debugf("[%v] a tc is built for view %v", hs.ID(), tc.View)
	hs.processTC(tc)
}

func (hs *HotStuff) ProcessLocalTmo(view types.View) {
	tmo := &pacemaker.TMO{
		View:   view + 1,
		NodeID: hs.ID(),
		HighQC: hs.bc.GetHighQC(),
	}
	hs.Broadcast(tmo)
	hs.ProcessRemoteTmo(tmo)
	log.Debugf("[%v] broadcast is done for sending tmo", hs.ID())
}

func (hs *HotStuff) MakeProposal(payload []*message.Transaction) *blockchain.Block {
	block := blockchain.MakeBlock(hs.pm.GetCurView(), hs.bc.GetHighQC(), payload, hs.ID())
	return block
}

func (hs *HotStuff) processTC(tc *pacemaker.TC) {
	if tc.View < hs.pm.GetCurView() {
		return
	}
	hs.pm.UpdateTC(tc)
	go hs.pm.AdvanceView(tc.View)
}

func (hs *HotStuff) preprocessQC(qc *blockchain.QC) {
	isThreeChain, _, err := hs.commitRule(qc)
	if err != nil {
		log.Warningf("[%v] cannot check commit rule", hs.ID())
		return
	}
	if isThreeChain {
		go hs.pm.AdvanceView(qc.View)
		return
	}
	for i := qc.View; ; i++ {
		nextLeader := hs.FindLeaderFor(i + 1)
		if !config.Configuration.IsByzantine(nextLeader) {
			qc.View = i
			log.Debugf("[%v] is going to send a stale qc to %v, view: %v, id: %x", hs.ID(), nextLeader, qc.View, qc.BlockID)
			hs.Send(nextLeader, qc)
			return
		}
	}
}

func (hs *HotStuff) processCertificate(qc *blockchain.QC) {
	if qc.View < hs.pm.GetCurView() {
		return
	}
	go hs.pm.AdvanceView(qc.View)
	hs.bc.UpdateHighQC(qc)
	log.Debugf("[%v] has advanced to view %v", hs.ID(), hs.pm.GetCurView())
	hs.updateStateByQC(qc)
	log.Debugf("[%v] has updated state by qc: %v", hs.ID(), qc.View)
	// TODO: send the qc to next leader
	//if !r.IsLeader(r.ID(), r.pm.GetCurView()) {
	//	go r.Send(r.FindLeaderFor(r.pm.GetCurView()), qc)
	//}
	if qc.View < 3 {
		return
	}
	ok, block, _ := hs.commitRule(qc)
	if !ok {
		return
	}
	committedBlocks, err := hs.bc.CommitBlock(block.ID)
	if err != nil {
		log.Errorf("[%v] cannot commit blocks", hs.ID())
		return
	}
	for _, block := range committedBlocks {
		hs.committedBlocks <- block
	}
}

func (hs *HotStuff) votingRule(block *blockchain.Block) (bool, error) {
	if block.View <= 2 {
		return true, nil
	}
	parentBlock, err := hs.bc.GetParentBlock(block.ID)
	if err != nil {
		return false, fmt.Errorf("cannot vote for block: %w", err)
	}
	if (block.View <= hs.lastVotedView) || (parentBlock.View < hs.preferredView) {
		return false, nil
	}
	return true, nil
}

func (hs *HotStuff) commitRule(qc *blockchain.QC) (bool, *blockchain.Block, error) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	parentBlock, err := hs.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return false, nil, fmt.Errorf("cannot commit any block: %w", err)
	}
	grandParentBlock, err := hs.bc.GetParentBlock(parentBlock.ID)
	if err != nil {
		return false, nil, fmt.Errorf("cannot commit any block: %w", err)
	}
	if ((grandParentBlock.View + 1) == parentBlock.View) && ((parentBlock.View + 1) == qc.View) {
		return true, grandParentBlock, nil
	}
	return false, nil, nil
}

func (hs *HotStuff) updateStateByQC(qc *blockchain.QC) error {
	if qc.View <= 2 {
		return nil
	}
	return hs.updatePreferredView(qc)
}

func (hs *HotStuff) updateLastVotedView(targetView types.View) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	if targetView < hs.lastVotedView {
		return fmt.Errorf("target view is lower than the last voted view")
	}
	hs.lastVotedView = targetView
	return nil
}

func (hs *HotStuff) updatePreferredView(qc *blockchain.QC) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	parentBlock, err := hs.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot update preferred view: %w", err)
	}
	if parentBlock.View > hs.preferredView {
		hs.preferredView = parentBlock.View
	}
	return nil
}
