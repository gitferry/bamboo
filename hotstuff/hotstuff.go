package hotstuff

import (
	"fmt"
	"sync"
	"time"

	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/election"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/node"
	"github.com/gitferry/bamboo/pacemaker"
	"github.com/gitferry/bamboo/types"
)

const FORK = "fork"

type HotStuff struct {
	node.Node
	election.Election
	pm              *pacemaker.Pacemaker
	lastVotedView   types.View
	preferredView   types.View
	highQC          *blockchain.QC
	bc              *blockchain.BlockChain
	committedBlocks chan *blockchain.Block
	forkedBlocks    chan *blockchain.Block
	bufferedQCs     map[crypto.Identifier]*blockchain.QC
	bufferedBlocks  map[types.View]*blockchain.Block
	mu              sync.Mutex
}

func NewHotStuff(
	node node.Node,
	pm *pacemaker.Pacemaker,
	elec election.Election,
	committedBlocks chan *blockchain.Block,
	forkedBlocks chan *blockchain.Block) *HotStuff {
	hs := new(HotStuff)
	hs.Node = node
	hs.Election = elec
	hs.pm = pm
	hs.bc = blockchain.NewBlockchain(config.GetConfig().N())
	hs.bufferedBlocks = make(map[types.View]*blockchain.Block)
	hs.bufferedQCs = make(map[crypto.Identifier]*blockchain.QC)
	hs.highQC = &blockchain.QC{View: 0}
	hs.committedBlocks = committedBlocks
	hs.forkedBlocks = forkedBlocks
	return hs
}

func (hs *HotStuff) ProcessBlock(block *blockchain.Block) error {
	log.Debugf("[%v] is processing block from %v, view: %v, id: %x", hs.ID(), block.Proposer.Node(), block.View, block.ID)
	curView := hs.pm.GetCurView()
	if block.Proposer != hs.ID() {
		blockIsVerified, _ := crypto.PubVerify(block.Sig, crypto.IDToByte(block.ID), block.Proposer)
		if !blockIsVerified {
			log.Warningf("[%v] received a block with an invalid signature", hs.ID())
		}
	}
	if block.View > curView+1 {
		//	buffer the block
		hs.bufferedBlocks[block.View-1] = block
		log.Debugf("[%v] the block is buffered, id: %x", hs.ID(), block.ID)
		return nil
	}
	if block.QC != nil {
		hs.updateHighQC(block.QC)
	} else {
		return fmt.Errorf("the block should contain a QC")
	}
	// does not have to process the QC if the replica is the proposer
	if block.Proposer != hs.ID() {
		hs.processCertificate(block.QC)
	}
	curView = hs.pm.GetCurView()
	if block.View < curView {
		log.Warningf("[%v] received a stale proposal from %v, id: %x", hs.ID(), block.Proposer, block.ID)
		return nil
	}
	if !hs.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	hs.bc.AddBlock(block)
	// process buffered QC
	qc, ok := hs.bufferedQCs[block.ID]
	if ok {
		hs.processCertificate(qc)
		delete(hs.bufferedQCs, block.ID)
	}

	shouldVote, err := hs.votingRule(block)
	if err != nil {
		log.Errorf("[%v] cannot decide whether to vote the block, %w", hs.ID(), err)
		return err
	}
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", hs.ID(), block.ID)
		return nil
	}
	vote := blockchain.MakeVote(block.View, hs.ID(), block.ID)
	// vote is sent to the next leader
	voteAggregator := hs.FindLeaderFor(block.View + 1)
	if voteAggregator == hs.ID() {
		log.Debugf("[%v] vote is sent to itself, id: %x", hs.ID(), vote.BlockID)
		hs.ProcessVote(vote)
	} else {
		log.Debugf("[%v] vote is sent to %v, id: %x", hs.ID(), voteAggregator, vote.BlockID)
		vote.Timestamp = time.Now()
		hs.Send(voteAggregator, vote)
	}
	b, ok := hs.bufferedBlocks[block.View]
	if ok {
		_ = hs.ProcessBlock(b)
		delete(hs.bufferedBlocks, block.View)
	}
	return nil
}

func (hs *HotStuff) ProcessVote(vote *blockchain.Vote) {
	log.Debugf("[%v] is processing the vote, block id: %x", hs.ID(), vote.BlockID)
	if vote.Voter != hs.ID() {
		voteIsVerified, err := crypto.PubVerify(vote.Signature, crypto.IDToByte(vote.BlockID), vote.Voter)
		if err != nil {
			log.Warningf("[%v] Error in verifying the signature in vote id: %x", hs.ID(), vote.BlockID)
			return
		}
		if !voteIsVerified {
			log.Warningf("[%v] received a vote with invalid signature. vote id: %x", hs.ID(), vote.BlockID)
			return
		}
	}
	isBuilt, qc := hs.bc.AddVote(vote)
	if !isBuilt {
		//log.Debugf("[%v] not sufficient votes to build a QC, block id: %x", hs.ID(), vote.BlockID)
		return
	}
	qc.Leader = hs.ID()
	// buffer the QC if the block has not been received
	_, err := hs.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		hs.bufferedQCs[qc.BlockID] = qc
		return
	}
	hs.processCertificate(qc)
}

func (hs *HotStuff) ProcessRemoteTmo(tmo *pacemaker.TMO) {
	log.Debugf("[%v] is processing tmo from %v", hs.ID(), tmo.NodeID)
	hs.processCertificate(tmo.HighQC)
	isBuilt, tc := hs.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		return
	}
	log.Debugf("[%v] a tc is built for view %v", hs.ID(), tc.View)
	hs.processTC(tc)
}

func (hs *HotStuff) ProcessLocalTmo(view types.View) {
	hs.pm.AdvanceView(view)
	tmo := &pacemaker.TMO{
		View:   view + 1,
		NodeID: hs.ID(),
		HighQC: hs.GetHighQC(),
	}
	hs.Broadcast(tmo)
	hs.ProcessRemoteTmo(tmo)
}

func (hs *HotStuff) MakeProposal(view types.View, payload []crypto.Identifier) *blockchain.Proposal {
	qc := hs.forkChoice()
	proposal := blockchain.BuildProposal(view, qc, qc.BlockID, payload, hs.ID())
	return proposal
}

func (hs *HotStuff) forkChoice() *blockchain.QC {
	var choice *blockchain.QC
	if !hs.IsByz() || config.GetConfig().Strategy != FORK {
		return hs.GetHighQC()
	}
	//	create a fork by returning highQC's parent's QC
	parBlockID := hs.GetHighQC().BlockID
	parBlock, err := hs.bc.GetBlockByID(parBlockID)
	if err != nil {
		log.Warningf("cannot get parent block of block id: %x: %w", parBlockID, err)
	}
	if parBlock.QC.View < hs.preferredView {
		choice = hs.GetHighQC()
	} else {
		choice = parBlock.QC
	}
	// to simulate TC's view
	choice.View = hs.pm.GetCurView() - 1
	return choice
}

func (hs *HotStuff) processTC(tc *pacemaker.TC) {
	if tc.View < hs.pm.GetCurView() {
		return
	}
	hs.pm.AdvanceView(tc.View)
}

func (hs *HotStuff) GetHighQC() *blockchain.QC {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	return hs.highQC
}

func (hs *HotStuff) GetChainStatus() string {
	chainGrowthRate := hs.bc.GetChainGrowth()
	blockIntervals := hs.bc.GetBlockIntervals()
	return fmt.Sprintf("[%v] The current view is: %v, chain growth rate is: %v, ave block interval is: %v", hs.ID(), hs.pm.GetCurView(), chainGrowthRate, blockIntervals)
}

func (hs *HotStuff) updateHighQC(qc *blockchain.QC) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	if qc.View > hs.highQC.View {
		hs.highQC = qc
	}
}

func (hs *HotStuff) processCertificate(qc *blockchain.QC) {
	log.Debugf("[%v] is processing a QC, block id: %x", hs.ID(), qc.BlockID)
	if qc.View < hs.pm.GetCurView() {
		return
	}
	if qc.Leader != hs.ID() {
		quorumIsVerified, _ := crypto.VerifyQuorumSignature(qc.AggSig, qc.BlockID, qc.Signers)
		if quorumIsVerified == false {
			log.Warningf("[%v] received a quorum with invalid signatures", hs.ID())
			return
		}
	}
	if hs.IsByz() && config.GetConfig().Strategy == FORK && hs.IsLeader(hs.ID(), qc.View+1) {
		hs.pm.AdvanceView(qc.View)
		return
	}
	err := hs.updatePreferredView(qc)
	if err != nil {
		hs.bufferedQCs[qc.BlockID] = qc
		log.Debugf("[%v] a qc is buffered, view: %v, id: %x", hs.ID(), qc.View, qc.BlockID)
		return
	}
	hs.pm.AdvanceView(qc.View)
	hs.updateHighQC(qc)
	if qc.View < 3 {
		return
	}
	ok, block, _ := hs.commitRule(qc)
	if !ok {
		return
	}
	// forked blocks are found when pruning
	committedBlocks, forkedBlocks, err := hs.bc.CommitBlock(block.ID, hs.pm.GetCurView())
	if err != nil {
		log.Errorf("[%v] cannot commit blocks, %w", hs.ID(), err)
		return
	}
	for _, cBlock := range committedBlocks {
		hs.committedBlocks <- cBlock
	}
	for _, fBlock := range forkedBlocks {
		hs.forkedBlocks <- fBlock
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

func (hs *HotStuff) updateLastVotedView(targetView types.View) error {
	if targetView < hs.lastVotedView {
		return fmt.Errorf("target view is lower than the last voted view")
	}
	hs.lastVotedView = targetView
	return nil
}

func (hs *HotStuff) updatePreferredView(qc *blockchain.QC) error {
	if qc.View <= 2 {
		return nil
	}
	_, err := hs.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot update preferred view: %w", err)
	}
	grandParentBlock, err := hs.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot update preferred view: %w", err)
	}
	if grandParentBlock.View > hs.preferredView {
		hs.preferredView = grandParentBlock.View
	}
	return nil
}
