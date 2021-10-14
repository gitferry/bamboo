package fhs

import (
	"fmt"
	"sync"

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

const FORK = "fork"

type Fhs struct {
	node.Node
	election.Election
	pm              *pacemaker.Pacemaker
	lastVotedView   types.View
	preferredView   types.View
	bc              *blockchain.BlockChain
	committedBlocks chan *blockchain.Block
	forkedBlocks    chan *blockchain.Block
	bufferedQCs     map[crypto.Identifier]*blockchain.QC
	bufferedBlocks  map[types.View]*blockchain.Block
	highQC          *blockchain.QC
	mu              sync.Mutex
}

func NewFhs(
	node node.Node,
	pm *pacemaker.Pacemaker,
	elec election.Election,
	committedBlocks chan *blockchain.Block,
	forkedBlocks chan *blockchain.Block) *Fhs {
	f := new(Fhs)
	f.Node = node
	f.Election = elec
	f.pm = pm
	f.bc = blockchain.NewBlockchain(config.GetConfig().N())
	f.bufferedBlocks = make(map[types.View]*blockchain.Block)
	f.bufferedQCs = make(map[crypto.Identifier]*blockchain.QC)
	f.highQC = &blockchain.QC{View: 0}
	f.committedBlocks = committedBlocks
	f.forkedBlocks = forkedBlocks
	return f
}

func (f *Fhs) ProcessBlock(block *blockchain.Block) error {
	log.Debugf("[%v] is processing block, view: %v, id: %x", f.ID(), block.View, block.ID)
	curView := f.pm.GetCurView()
	if block.Proposer != f.ID() {
		blockIsVerified, _ := crypto.PubVerify(block.Sig, crypto.IDToByte(block.ID), block.Proposer)
		if !blockIsVerified {
			log.Warningf("[%v] received a block with an invalid signature", f.ID())
		}
	}
	if block.View > curView+1 {
		//	buffer the block
		f.bufferedBlocks[block.View-1] = block
		log.Debugf("[%v] the block is buffered, view: %v, current view is: %v, id: %x", f.ID(), block.View, curView, block.ID)
		return nil
	}
	if block.QC != nil {
		f.updateHighQC(block.QC)
	} else {
		return fmt.Errorf("the block should contain a QC")
	}
	if block.Proposer != f.ID() {
		f.processCertificate(block.QC)
	}
	curView = f.pm.GetCurView()
	if block.View < curView {
		log.Warningf("[%v] received a stale proposal from %v, block view: %v, current view: %v, block id: %x", f.ID(), block.Proposer, block.View, curView, block.ID)
		return nil
	}
	if !f.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	f.bc.AddBlock(block)

	// check commit rule
	qc := block.QC
	if qc.View >= 2 && qc.View+1 == block.View {
		ok, b, _ := f.commitRule(block)
		if !ok {
			return nil
		}
		committedBlocks, forkedBlocks, err := f.bc.CommitBlock(b.ID, f.pm.GetCurView())
		if err != nil {
			return fmt.Errorf("[%v] cannot commit blocks", f.ID())
		}
		for _, cBlock := range committedBlocks {
			f.committedBlocks <- cBlock
		}
		for _, fBlock := range forkedBlocks {
			f.forkedBlocks <- fBlock
		}
	}

	// process buffered QC
	qc, ok := f.bufferedQCs[block.ID]
	if ok {
		f.processCertificate(qc)
		delete(f.bufferedQCs, block.ID)
	}

	shouldVote, err := f.votingRule(block)
	if err != nil {
		log.Errorf("cannot decide whether to vote the block, %w", err)
		return err
	}
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", f.ID(), block.ID)
		return nil
	}
	vote := blockchain.MakeVote(block.View, f.ID(), block.ID)
	// vote to the next leader
	voteAggregator := f.FindLeaderFor(block.View + 1)
	if voteAggregator == f.ID() {
		f.ProcessVote(vote)
	} else {
		f.Send(voteAggregator, vote)
	}
	log.Debugf("[%v] vote is sent, id: %x", f.ID(), vote.BlockID)

	b, ok := f.bufferedBlocks[block.View]
	if ok {
		err := f.ProcessBlock(b)
		return err
	}

	return nil
}

func (f *Fhs) ProcessVote(vote *blockchain.Vote) {
	log.Debugf("[%v] is processing the vote from %v, block id: %x", f.ID(), vote.Voter, vote.BlockID)
	if f.ID() != vote.Voter {
		voteIsVerified, err := crypto.PubVerify(vote.Signature, crypto.IDToByte(vote.BlockID), vote.Voter)
		if err != nil {
			log.Fatalf("[%v] Error in verifying the signature in vote id: %x", f.ID(), vote.BlockID)
			return
		}
		if !voteIsVerified {
			log.Warningf("[%v] received a vote with unvalid signature. vote id: %x", f.ID(), vote.BlockID)
			return
		}
	}
	isBuilt, qc := f.bc.AddVote(vote)
	if !isBuilt {
		log.Debugf("[%v] not sufficient votes to build a QC, block id: %x", f.ID(), vote.BlockID)
		return
	}
	qc.Leader = f.ID()
	_, err := f.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		f.bufferedQCs[qc.BlockID] = qc
		return
	}
	f.processCertificate(qc)
}

func (f *Fhs) ProcessRemoteTmo(tmo *pacemaker.TMO) {
	log.Debugf("[%v] is processing tmo from %v", f.ID(), tmo.NodeID)
	if tmo.View < f.pm.GetCurView() {
		return
	}
	isBuilt, tc := f.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		log.Debugf("[%v] not enough tc for %v", f.ID(), tmo.View)
		return
	}
	log.Debugf("[%v] a tc is built for view %v", f.ID(), tc.View)
	f.processTC(tc)
}

func (f *Fhs) ProcessLocalTmo(view types.View) {
	f.pm.AdvanceView(view + 1)
	tmo := &pacemaker.TMO{
		View:   view + 1,
		NodeID: f.ID(),
		HighQC: f.GetHighQC(),
	}
	f.Broadcast(tmo)
	f.ProcessRemoteTmo(tmo)
	log.Debugf("[%v] broadcast is done for sending tmo", f.ID())
}

func (f *Fhs) MakeProposal(view types.View, payload []*message.Transaction) *blockchain.Block {
	qc := f.forkChoice()
	block := blockchain.BuildProposal(view, qc, qc.BlockID, payload, f.ID())
	return block
}

func (f *Fhs) forkChoice() *blockchain.QC {
	choice := f.GetHighQC()
	// to simulate TC under forking attack
	choice.View = f.pm.GetCurView() - 1
	return choice
}

func (f *Fhs) processTC(tc *pacemaker.TC) {
	if tc.View < f.pm.GetCurView() {
		return
	}
	f.pm.AdvanceView(tc.View)
}

func (f *Fhs) GetChainStatus() string {
	chainGrowthRate := f.bc.GetChainGrowth()
	blockIntervals := f.bc.GetBlockIntervals()
	return fmt.Sprintf("[%v] The current view is: %v, chain growth rate is: %v, ave block interval is: %v", f.ID(), f.pm.GetCurView(), chainGrowthRate, blockIntervals)
}

func (f *Fhs) GetHighQC() *blockchain.QC {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.highQC
}

func (f *Fhs) updateHighQC(qc *blockchain.QC) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if qc.View > f.highQC.View {
		f.highQC = qc
	}
}

func (f *Fhs) processCertificate(qc *blockchain.QC) {
	log.Debugf("[%v] is processing a QC, block id: %x", f.ID(), qc.BlockID)
	if qc.View < f.pm.GetCurView() {
		return
	}
	if qc.Leader != f.ID() {
		quorumIsVerified, _ := crypto.VerifyQuorumSignature(qc.AggSig, qc.BlockID, qc.Signers)
		if quorumIsVerified == false {
			log.Warningf("[%v] received a quorum with invalid signatures", f.ID())
			return
		}
	}
	if f.IsByz() && config.GetConfig().Strategy == FORK && f.IsLeader(f.ID(), qc.View+1) {
		f.pm.AdvanceView(qc.View)
		return
	}
	err := f.updatePreferredView(qc)
	if err != nil {
		f.bufferedQCs[qc.BlockID] = qc
		log.Debugf("[%v] a qc is buffered, view: %v, id: %x", f.ID(), qc.View, qc.BlockID)
		return
	}
	f.updateHighQC(qc)
	f.pm.AdvanceView(qc.View)
}

func (f *Fhs) votingRule(block *blockchain.Block) (bool, error) {
	if block.View <= 2 {
		return true, nil
	}
	parentBlock, err := f.bc.GetParentBlock(block.ID)
	if err != nil {
		return false, fmt.Errorf("cannot vote for block: %w", err)
	}
	if (block.View <= f.lastVotedView) || (parentBlock.View < f.preferredView) {
		if parentBlock.View < f.preferredView {
			log.Debugf("[%v] parent block view is: %v and preferred view is: %v", f.ID(), parentBlock.View, f.preferredView)
		}
		return false, nil
	}
	return true, nil
}

func (f *Fhs) commitRule(block *blockchain.Block) (bool, *blockchain.Block, error) {
	qc := block.QC
	parentBlock, err := f.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return false, nil, fmt.Errorf("cannot commit any block: %w", err)
	}
	if (parentBlock.View + 1) == qc.View {
		return true, parentBlock, nil
	}
	return false, nil, nil
}

func (f *Fhs) updateLastVotedView(targetView types.View) error {
	if targetView < f.lastVotedView {
		return fmt.Errorf("target view is lower than the last voted view")
	}
	f.lastVotedView = targetView
	return nil
}

func (f *Fhs) updatePreferredView(qc *blockchain.QC) error {
	if qc.View < 2 {
		return nil
	}
	_, err := f.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot update preferred view: %w", err)
	}
	if qc.View > f.preferredView {
		log.Debugf("[%v] preferred view has been updated to %v", f.ID(), qc.View)
		f.preferredView = qc.View
	}
	return nil
}
