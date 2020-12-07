package tchs

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

type Tchs struct {
	node.Node
	election.Election
	pm              *pacemaker.Pacemaker
	lastVotedView   types.View
	preferredView   types.View
	bc              *blockchain.BlockChain
	committedBlocks chan *blockchain.Block
	forkedBlocks    chan *blockchain.Block
	bufferedQCs     map[crypto.Identifier]*blockchain.QC
	bufferedBlocks  map[crypto.Identifier]*blockchain.Block
	highQC          *blockchain.QC
	mu              sync.Mutex
}

func NewTchs(
	node node.Node,
	pm *pacemaker.Pacemaker,
	elec election.Election,
	committedBlocks chan *blockchain.Block,
	forkedBlocks chan *blockchain.Block) *Tchs {
	th := new(Tchs)
	th.Node = node
	th.Election = elec
	th.pm = pm
	th.bc = blockchain.NewBlockchain(config.GetConfig().N())
	th.bufferedBlocks = make(map[crypto.Identifier]*blockchain.Block)
	th.bufferedQCs = make(map[crypto.Identifier]*blockchain.QC)
	th.highQC = &blockchain.QC{View: 0}
	th.committedBlocks = committedBlocks
	th.forkedBlocks = forkedBlocks
	return th
}

func (th *Tchs) ProcessBlock(block *blockchain.Block) error {
	log.Debugf("[%v] is processing block, view: %v, id: %x", th.ID(), block.View, block.ID)
	curView := th.pm.GetCurView()
	if block.View > curView+1 {
		//	buffer the block
		th.bufferedBlocks[block.PrevID] = block
		log.Debugf("[%v] the block is buffered, id: %x", th.ID(), block.ID)
		return nil
	}
	if block.Proposer != th.ID() {
		blockIsVerified, _ := crypto.PubVerify(block.Sig, crypto.IDToByte(block.ID), block.Proposer)
		if !blockIsVerified {
			log.Warningf("[%v] received a block with an invalid signature", th.ID())
		}
	}
	if block.QC != nil {
		th.updateHighQC(block.QC)
	} else {
		return fmt.Errorf("the block should contain a QC")
	}
	th.processCertificate(block.QC)
	curView = th.pm.GetCurView()
	if block.View < curView {
		log.Warningf("[%v] received a stale proposal from %v, block view: %v, current view: %v, block id: %x", th.ID(), block.Proposer, block.View, curView, block.ID)
		return nil
	}
	if !th.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	th.bc.AddBlock(block)

	// process buffered QC
	qc, ok := th.bufferedQCs[block.ID]
	if ok {
		th.processCertificate(qc)
		delete(th.bufferedQCs, block.ID)
	}

	shouldVote, err := th.votingRule(block)
	if err != nil {
		log.Errorf("cannot decide whether to vote the block, %w", err)
		return err
	}
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", th.ID(), block.ID)
		return nil
	}
	vote := blockchain.MakeVote(block.View, th.ID(), block.ID)
	// vote to the next leader
	voteAggregator := th.FindLeaderFor(block.View + 1)
	if voteAggregator == th.ID() {
		th.ProcessVote(vote)
	} else {
		th.Send(voteAggregator, vote)
	}
	log.Debugf("[%v] vote is sent, id: %x", th.ID(), vote.BlockID)

	b, ok := th.bufferedBlocks[block.ID]
	if ok {
		return th.ProcessBlock(b)
	}
	return nil
}

func (th *Tchs) ProcessVote(vote *blockchain.Vote) {
	log.Debugf("[%v] is processing the vote, block id: %x", th.ID(), vote.BlockID)
	if th.ID() != vote.Voter {
		voteIsVerified, err := crypto.PubVerify(vote.Signature, crypto.IDToByte(vote.BlockID), vote.Voter)
		if err != nil {
			log.Fatalf("[%v] Error in verifying the signature in vote id: %x", th.ID(), vote.BlockID)
			return
		}
		if !voteIsVerified {
			log.Warningf("[%v] received a vote with unvalid signature. vote id: %x", th.ID(), vote.BlockID)
			return
		}
	}
	isBuilt, qc := th.bc.AddVote(vote)
	if !isBuilt {
		log.Debugf("[%v] not sufficient votes to build a QC, block id: %x", th.ID(), vote.BlockID)
		return
	}
	qc.Leader = th.ID()
	_, err := th.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		th.bufferedQCs[qc.BlockID] = qc
		return
	}
	th.processCertificate(qc)
}

func (th *Tchs) ProcessRemoteTmo(tmo *pacemaker.TMO) {
	log.Debugf("[%v] is processing tmo from %v", th.ID(), tmo.NodeID)
	th.updateHighQC(tmo.HighQC)
	isBuilt, tc := th.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		log.Debugf("[%v] not enough tc for %v", th.ID(), tmo.View)
		return
	}
	log.Debugf("[%v] a tc is built for view %v", th.ID(), tc.View)
	th.processTC(tc)
}

func (th *Tchs) ProcessLocalTmo(view types.View) {
	tmo := &pacemaker.TMO{
		View:   view + 1,
		NodeID: th.ID(),
		HighQC: th.GetHighQC(),
	}
	th.Broadcast(tmo)
	th.ProcessRemoteTmo(tmo)
	log.Debugf("[%v] broadcast is done for sending tmo", th.ID())
}

func (th *Tchs) MakeProposal(payload []*message.Transaction) *blockchain.Block {
	qc := th.forkChoice()
	block := blockchain.MakeBlock(th.pm.GetCurView(), qc, qc.BlockID, payload, th.ID())
	return block
}

func (th *Tchs) forkChoice() *blockchain.QC {
	choice := th.GetHighQC()
	choice.View = th.pm.GetCurView() - 1
	return choice
}

func (th *Tchs) processTC(tc *pacemaker.TC) {
	if tc.View < th.pm.GetCurView() {
		return
	}
	th.pm.UpdateTC(tc)
	th.pm.AdvanceView(tc.View)
}

func (th *Tchs) preprocessQC(qc *blockchain.QC) {
	isThreeChain, _, err := th.commitRule(qc)
	if err != nil {
		log.Warningf("[%v] cannot check commit rule", th.ID())
		return
	}
	if isThreeChain {
		th.pm.AdvanceView(qc.View)
		return
	}
	for i := qc.View; ; i++ {
		nextLeader := th.FindLeaderFor(i + 1)
		if !config.Configuration.IsByzantine(nextLeader) {
			qc.View = i
			log.Debugf("[%v] is going to send a stale qc to %v, view: %v, id: %x", th.ID(), nextLeader, qc.View, qc.BlockID)
			th.Send(nextLeader, qc)
			return
		}
	}
}

func (th *Tchs) GetChainStatus() string {
	chainGrowthRate := th.bc.GetChainGrowth()
	blockIntervals := th.bc.GetBlockIntervals()
	return fmt.Sprintf("[%v] The current view is: %v, chain growth rate is: %v, ave block interval is: %v", th.ID(), th.pm.GetCurView(), chainGrowthRate, blockIntervals)
}

func (th *Tchs) GetHighQC() *blockchain.QC {
	th.mu.Lock()
	defer th.mu.Unlock()
	return th.highQC
}

func (th *Tchs) updateHighQC(qc *blockchain.QC) {
	th.mu.Lock()
	defer th.mu.Unlock()
	if qc.View > th.highQC.View {
		th.highQC = qc
	}
}

func (th *Tchs) processCertificate(qc *blockchain.QC) {
	log.Debugf("[%v] is processing a QC, block id: %x", th.ID(), qc.BlockID)
	if qc.View < th.pm.GetCurView() {
		return
	}
	if qc.Leader != th.ID() {
		quorumIsVerified, _ := crypto.VerifyQuorumSignature(qc.AggSig, qc.BlockID, qc.Signers)
		if quorumIsVerified == false {
			log.Warningf("[%v] received a quorum with invalid signatures", th.ID())
			return
		}
	}
	if th.IsByz() && config.GetConfig().Strategy == FORK && th.IsLeader(th.ID(), qc.View+1) {
		th.pm.AdvanceView(qc.View)
		return
	}
	err := th.updatePreferredView(qc)
	if err != nil {
		th.bufferedQCs[qc.BlockID] = qc
		log.Debugf("[%v] a qc is buffered, view: %v, id: %x", th.ID(), qc.View, qc.BlockID)
		return
	}
	th.pm.AdvanceView(qc.View)
	th.updateHighQC(qc)
	if qc.View < 2 {
		return
	}
	ok, block, _ := th.commitRule(qc)
	if !ok {
		return
	}
	committedBlocks, forkedBlocks, err := th.bc.CommitBlock(block.ID, th.pm.GetCurView())
	if err != nil {
		log.Errorf("[%v] cannot commit blocks", th.ID())
		return
	}
	go func() {
		for _, cBlock := range committedBlocks {
			th.committedBlocks <- cBlock
		}
		for _, fBlock := range forkedBlocks {
			th.forkedBlocks <- fBlock
		}
	}()
}

func (th *Tchs) votingRule(block *blockchain.Block) (bool, error) {
	if block.View <= 2 {
		return true, nil
	}
	//parentBlock, err := th.bc.GetParentBlock(block.ID)
	//if err != nil {
	//	return false, fmt.Errorf("cannot vote for block: %w", err)
	//}
	//if (block.View <= th.lastVotedView) || (parentBlock.View < th.preferredView) {
	//	return false, nil
	//}
	if block.View <= th.lastVotedView {
		return false, nil
	}
	return true, nil
}

func (th *Tchs) commitRule(qc *blockchain.QC) (bool, *blockchain.Block, error) {
	parentBlock, err := th.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return false, nil, fmt.Errorf("cannot commit any block: %w", err)
	}
	if (parentBlock.View + 1) == qc.View {
		return true, parentBlock, nil
	}
	return false, nil, nil
}

func (th *Tchs) updateLastVotedView(targetView types.View) error {
	if targetView < th.lastVotedView {
		return fmt.Errorf("target view is lower than the last voted view")
	}
	th.lastVotedView = targetView
	return nil
}

func (th *Tchs) updatePreferredView(qc *blockchain.QC) error {
	if qc.View < 2 {
		return nil
	}
	_, err := th.bc.GetBlockByID(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot update preferred view: %w", err)
	}
	if qc.View > th.preferredView {
		th.preferredView = qc.View
	}
	return nil
}
