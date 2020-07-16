package zeitgeber

import (
	"time"

	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/log"
)

type Replica struct {
	Node
	pacemaker.Pacemaker
	Election
	Safety
	Producer
	bc         *blockchain.BlockChain
	blockMsg   chan *blockchain.Block
	voteMsg    chan *blockchain.Vote
	timeoutMsg chan *pacemaker.Pacemaker.TMO
	newView    chan View
}

func (r *Replica) HandleBlock(block blockchain.Block) {
	log.Debugf("[%v] received a block from %v, view is %v", r.ID(), block.Proposer, block.View)
	r.blockMsg <- &block
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	log.Debugf("[%v] received a vote from %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	r.voteMsg <- &vote
}

func (r *Replica) processBlock(block *blockchain.Block) {
	r.ProcessQC(block.QC)
	curView := r.GetCurView()
	if block.View != curView {
		log.Warningf("[%v] received a stale proposal", r.ID())
		return
	}
	if !r.Election.IsLeader(block.Proposer, block.View) {
		log.Warningf(
			"[%v] received a proposal (%v) from an invalid leader (%v)",
			r.ID(), block.View, block.Proposer)
		return
	}
	r.bc.AddBlock(block)
	r.updateConsensusState()
	//	TODO: check safety and liveness rules to vote
}

func (r *Replica) processVote(vote *blockchain.Vote) {
	isBuilt, qc := r.bc.AddVote(vote)
	if !isBuilt {
		return
	}
	r.ProcessQC(qc)
}

func (r *Replica) updateConsensusState() {
	//	TODO: update locked QC and commit
}

func (r *Replica) HandleRequest(request Request) {
	//	store the request into the transaction pool
}

//func (r *Replica) MakeProposal(view View) {
//	curView := r.GetCurView()
//	// the replica should propose if it is the leader
//	proposal := ProposalMsg{
//		NodeID:   r.ID(),
//		View:     view,
//		TimeCert: NewTC(curView),
//	}
//	time.Sleep(20 * time.Millisecond)
//	//log.Infof("[%v] is proposing for view %v", r.NodeID(), curView)
//	if r.IsByz() {
//		r.MulticastQuorum(GetConfig().ByzNo, proposal)
//	} else {
//		r.Broadcast(proposal)
//	}
//	r.HandleProposal(proposal)
//}

func (r *Replica) processNewView(newView View) {
	//log.Debugf("[%v] is processing new view: %v", r.NodeID(), newView)
	if !r.IsLeader(r.ID(), newView+1) {
		return
	}
	r.producer.MakeProposal(newView+1, r.bc.GetHighQC(), r.bc.MakeForkChoice())
}

func (r *Replica) startTimer() {
	duration := GetTimer()
	timer := time.NewTimer(duration)
	for {
		timer.Reset(duration)
		go func() {
			<-timer.C
			r.handleTimeout()
			return
		}()
		select {
		case newView := <-r.Pacemaker.EnteringViewEvent():
			timer.Stop()
			go r.processNewView(newView)
		case newBlock := <-r.blockMsg:
			go r.processBlock(newBlock)
		case newVote := <-r.voteMsg:
			go r.processVote(newVote)
		}
	}
}

func (r *Replica) handleTimeout() {
	r.Pacemaker.TimeoutFor(r.GetCurView())
}

func (r *Replica) handleRequest(m Request) {
	log.Debugf("[%v] received txn %v\n", r.ID(), m)
	go r.Broadcast(m)
	r.StoreTxn(m)
}

func NewReplica(id NodeID, isByz bool) *Replica {
	r := new(Replica)
	r.Node = NewNode(id, isByz)
	if isByz {
		log.Infof("[%v] is Byzantine", r.ID())
	}
	elect := NewRotation(GetConfig().N())
	r.Election = elect
	r.Register(Request{}, r.handleRequest)
	//TODO:
	//1. register hotstuff handlers
	//2. first leader kicks off
	return r
}
