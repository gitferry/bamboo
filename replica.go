package zeitgeber

import (
	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/config"
	"github.com/gitferry/zeitgeber/election"
	"github.com/gitferry/zeitgeber/hotstuff"
	"github.com/gitferry/zeitgeber/identity"
	"github.com/gitferry/zeitgeber/log"
	"github.com/gitferry/zeitgeber/mempool"
	"github.com/gitferry/zeitgeber/message"
	"github.com/gitferry/zeitgeber/pacemaker"
	"github.com/gitferry/zeitgeber/types"
)

type Replica struct {
	Node
	election.Election
	Safety
	pd         *mempool.Producer
	bc         *blockchain.BlockChain
	pm         *pacemaker.Pacemaker
	isMaster   bool
	blockMsg   chan *blockchain.Block
	voteMsg    chan *blockchain.Vote
	qcMsg      chan *blockchain.QC
	timeoutMsg chan *pacemaker.TMO
	newView    chan types.View
}

func NewReplica(id identity.NodeID, isByz bool) *Replica {
	r := new(Replica)
	r.Node = NewNode(id, isByz)
	if isByz {
		log.Infof("[%v] is Byzantine", r.ID())
	}
	r.Election = election.NewRotation(config.GetConfig().N())
	bc := blockchain.NewBlockchain(config.GetConfig().N())
	r.Safety = hotstuff.NewHotStuff(bc)
	r.pd = mempool.NewProducer()
	r.bc = bc
	r.blockMsg = make(chan *blockchain.Block, 1)
	r.voteMsg = make(chan *blockchain.Vote, 1)
	r.qcMsg = make(chan *blockchain.QC, 1)
	r.timeoutMsg = make(chan *pacemaker.TMO, 1)
	r.Register(message.Transaction{}, r.handleTxn)
	r.Register(blockchain.QC{}, r.HandleQC)
	r.Register(blockchain.Block{}, r.HandleBlock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	//TODO: first leader kicks off
	return r
}

/* Message Handlers */

func (r *Replica) HandleBlock(block blockchain.Block) {
	log.Debugf("[%v] received a block from %v, view is %v", r.ID(), block.Proposer, block.View)
	if block.View < r.pm.GetCurView() {
		return
	}
	r.blockMsg <- &block
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	log.Debugf("[%v] received a vote from %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	if vote.View < r.pm.GetCurView() {
		return
	}
	r.voteMsg <- &vote
}

func (r *Replica) HandleQC(qc blockchain.QC) {
	log.Debugf("[%v] received a qc from, blockID is %x", r.ID(), qc.BlockID)
	if qc.View < r.pm.GetCurView() {
		return
	}
	r.qcMsg <- &qc
}

func (r *Replica) handleTxn(m message.Transaction) {
	log.Debugf("[%v] received txn %v\n", r.ID(), m)
	r.pd.CollectTxn(&m)
}

/* Processors */

func (r *Replica) processBlock(block *blockchain.Block) {
	// TODO: process TC
	r.processCertificate(block.QC)
	curView := r.pm.GetCurView()
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
	shouldVote, err := r.VotingRule(block)
	if err != nil {
		log.Errorf("cannot decide whether to vote the block, %w", err)
		return
	}
	if shouldVote {
		vote := blockchain.MakeVote(block.View, r.ID(), block.ID)
		err := r.UpdateStateByView(vote.View)
		if err != nil {
			log.Errorf("cannot update state after voting: %w", err)
		}
		// TODO: sign the vote
		go r.Send(r.FindLeaderFor(curView+1), vote)
		r.processVote(vote)
	}
}

func (r *Replica) processCertificate(qc *blockchain.QC) {
	r.pm.AdvanceView(qc.View)
	err := r.UpdateStateByQC(qc)
	if err != nil {
		log.Errorf("cannot update state when processing qc: %w", err)
		return
	}
	if !r.IsLeader(r.ID(), r.pm.GetCurView()) {
		go r.Send(r.FindLeaderFor(r.pm.GetCurView()), qc)
	}
	ok, block, err := r.CommitRule(qc)
	if err != nil {
		log.Errorf("cannot process the qc %w", err)
		return
	}
	if !ok {
		return
	}
	committedBlocks, err := r.bc.CommitBlock(block.ID)
	if err != nil {
		log.Errorf("[%v] cannot commit blocks", r.ID())
		return
	}
	r.processCommittedBlocks(committedBlocks)
}

func (r *Replica) processCommittedBlocks(blocks []*blockchain.Block) {
	for _, block := range blocks {
		for _, txn := range block.Payload {
			txn.Reply(message.TransactionReply{})
			if r.ID() != block.Proposer { // txns are removed when being proposed
				r.pd.RemoveTxn(txn.ID)
			}
		}
	}
}

func (r *Replica) processVote(vote *blockchain.Vote) {
	isBuilt, qc := r.bc.AddVote(vote)
	if !isBuilt {
		return
	}
	r.processCertificate(qc)
}

func (r *Replica) processNewView(newView types.View) {
	log.Debugf("[%v] is processing new view: %v", r.ID(), newView)
	if !r.IsLeader(r.ID(), newView+1) {
		return
	}
	block := r.pd.ProduceBlock(newView+1, r.bc.GetHighQC())
	//	TODO: sign the block
	r.Broadcast(block)
	for _, txn := range block.Payload {
		r.pd.RemoveTxn(txn.ID)
	}
}

func (r *Replica) startTimer() {
	for {
		// TODO: add timeout handler
		select {
		case newView := <-r.pm.EnteringViewEvent():
			go r.processNewView(newView)
		case newBlock := <-r.blockMsg:
			go r.processBlock(newBlock)
		case newVote := <-r.voteMsg:
			go r.processVote(newVote)
		case newQC := <-r.qcMsg:
			go r.processCertificate(newQC)
		}
	}
}
