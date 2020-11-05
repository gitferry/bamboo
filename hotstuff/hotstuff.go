package hotstuff

import (
	"fmt"
	"github.com/gitferry/bamboo/crypto"
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

type BufferQC struct {
	qcs map[crypto.Identifier]*blockchain.QC
	mu  sync.Mutex
}

type HotStuff struct {
	node.Node
	election.Election
	pm              *pacemaker.Pacemaker
	lastVotedView   types.View
	preferredView   types.View
	bc              *blockchain.BlockChain
	committedBlocks chan *blockchain.Block
	bufferedQCs     *BufferQC
	bufferedBlocks  map[crypto.Identifier]*blockchain.Block
	bufferedVotes   map[crypto.Identifier][]*blockchain.Vote
	highQC          *blockchain.QC
	mu              sync.Mutex
}

func NewBufferQC() *BufferQC {
	bqc := new(BufferQC)
	bqc.qcs = make(map[crypto.Identifier]*blockchain.QC)
	return bqc
}

func NewHotStuff(
	node node.Node,
	pm *pacemaker.Pacemaker,
	elec election.Election,
	committedBlocks chan *blockchain.Block) *HotStuff {
	hs := new(HotStuff)
	hs.Node = node
	hs.Election = elec
	hs.pm = pm
	hs.bc = blockchain.NewBlockchain(config.GetConfig().N())
	hs.bufferedBlocks = make(map[crypto.Identifier]*blockchain.Block)
	hs.bufferedQCs = NewBufferQC()
	hs.highQC = &blockchain.QC{View: 0}
	hs.committedBlocks = committedBlocks
	return hs
}

func (bqc *BufferQC) Add(qc *blockchain.QC) {
	//bqc.mu.Lock()
	//defer bqc.mu.Unlock()
	bqc.qcs[qc.BlockID] = qc
}

func (bqc *BufferQC) Get(id crypto.Identifier) (*blockchain.QC, bool) {
	//bqc.mu.Lock()
	//defer bqc.mu.Unlock()
	qc, ok := bqc.qcs[id]
	return qc, ok
}

func (hs *HotStuff) ProcessBlock(block *blockchain.Block) error {
	log.Debugf("[%v] is processing block, view: %v, id: %x", hs.ID(), block.View, block.ID)
	curView := hs.pm.GetCurView()
	if block.View > curView+1 {
		//	buffer the block
		hs.bufferedBlocks[block.PrevID] = block
		log.Debugf("[%v] the block is buffered, id: %x", hs.ID(), block.ID)
		return nil
	}
	if block.QC != nil {
		hs.updateHighQC(block.QC)
	} else {
		return fmt.Errorf("the block should contain a QC")
	}
	hs.ProcessCertificate(block.QC)
	curView = hs.pm.GetCurView()
	if block.View < curView {
		log.Warningf("[%v] received a stale proposal from %v", hs.ID(), block.Proposer)
		return nil
	}
	if !hs.Election.IsLeader(block.Proposer, block.View) {
		return fmt.Errorf("received a proposal (%v) from an invalid leader (%v)", block.View, block.Proposer)
	}
	hs.bc.AddBlock(block)

	// process buffered QC
	qc, ok := hs.bufferedQCs.Get(block.ID)
	if ok {
		hs.ProcessCertificate(qc)
		// TODO: garbage collection
		//delete(hs.bufferedQCs, block.ID)
	}

	shouldVote, err := hs.votingRule(block)
	if err != nil {
		log.Errorf("cannot decide whether to vote the block, %w", err)
		return err
	}
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", hs.ID(), block.ID)
		return nil
	}
	vote := blockchain.MakeVote(block.View, hs.ID(), block.ID)
	// TODO: sign the vote
	// vote to the next leader
	voteAggregator := hs.FindLeaderFor(block.View + 1)
	if voteAggregator == hs.ID() {
		hs.ProcessVote(vote)
	} else {
		hs.Send(voteAggregator, vote)
	}
	log.Debugf("[%v] vote is sent, id: %x", hs.ID(), vote.BlockID)

	b, ok := hs.bufferedBlocks[block.ID]
	if ok {
		return hs.ProcessBlock(b)
	}

	return nil
}

func (hs *HotStuff) ProcessVote(vote *blockchain.Vote) {
	log.Debugf("[%v] is processing the vote, block id: %x", hs.ID(), vote.BlockID)
	isBuilt, qc := hs.bc.AddVote(vote)
	if !isBuilt {
		log.Debugf("[%v] not sufficient votes to build a QC, block id: %x", hs.ID(), vote.BlockID)
		return
	}
	// send the QC to the next leader
	//log.Debugf("[%v] a qc is built, block id: %x", hs.ID(), qc.BlockID)
	//nextLeader := hs.FindLeaderFor(qc.View + 1)
	//if nextLeader == hs.ID() {
	//	if config.Configuration.IsByzantine(nextLeader) {
	//		hs.preprocessQC(qc)
	//	} else {
	hs.ProcessCertificate(qc)
	//	}
	//} else {
	//	hs.Send(nextLeader, qc)
	//}

	return
}

func (hs *HotStuff) ProcessRemoteTmo(tmo *pacemaker.TMO) {
	log.Debugf("[%v] is processing tmo from %v", hs.ID(), tmo.NodeID)
	hs.updateHighQC(tmo.HighQC)
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
		HighQC: hs.GetHighQC(),
	}
	hs.Broadcast(tmo)
	hs.ProcessRemoteTmo(tmo)
	log.Debugf("[%v] broadcast is done for sending tmo", hs.ID())
}

func (hs *HotStuff) MakeProposal(payload []*message.Transaction) *blockchain.Block {
	block := blockchain.MakeBlock(hs.pm.GetCurView(), hs.GetHighQC(), payload, hs.ID())
	return block
}

func (hs *HotStuff) processTC(tc *pacemaker.TC) {
	if tc.View < hs.pm.GetCurView() {
		return
	}
	hs.pm.UpdateTC(tc)
	hs.pm.AdvanceView(tc.View)
}

func (hs *HotStuff) preprocessQC(qc *blockchain.QC) {
	isThreeChain, _, err := hs.commitRule(qc)
	if err != nil {
		log.Warningf("[%v] cannot check commit rule", hs.ID())
		return
	}
	if isThreeChain {
		hs.pm.AdvanceView(qc.View)
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

func (hs *HotStuff) GetHighQC() *blockchain.QC {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	return hs.highQC
}

func (hs *HotStuff) updateHighQC(qc *blockchain.QC) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	if qc.View > hs.highQC.View {
		hs.highQC = qc
	}
}

func (hs *HotStuff) ProcessCertificate(qc *blockchain.QC) {
	log.Debugf("[%v] is processing a QC, block id: %x", hs.ID(), qc.BlockID)
	if qc.View < hs.pm.GetCurView() {
		return
	}
	err := hs.updatePreferredView(qc)
	if err != nil {
		hs.bufferedQCs.Add(qc)
		log.Debugf("[%v] a qc is buffered, view: %v, id: %x", hs.ID(), qc.View, qc.BlockID)
		return
	}
	hs.pm.AdvanceView(qc.View)
	hs.updateHighQC(qc)
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
	go func() {
		for _, block := range committedBlocks {
			hs.committedBlocks <- block
		}
	}()
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
	//hs.mu.Lock()
	//defer hs.mu.Unlock()
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
	//hs.mu.Lock()
	//defer hs.mu.Unlock()
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
	parentBlock, err := hs.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot update preferred view: %w", err)
	}
	if parentBlock.View > hs.preferredView {
		hs.preferredView = parentBlock.View
	}
	return nil
}
