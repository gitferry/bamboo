package replica

import (
	"encoding/gob"
	"time"

	"go.uber.org/atomic"

	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/election"
	"github.com/gitferry/bamboo/hotstuff"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/mempool"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/node"
	"github.com/gitferry/bamboo/pacemaker"
	"github.com/gitferry/bamboo/streamlet"
	"github.com/gitferry/bamboo/types"
)

type Replica struct {
	node.Node
	Safety
	election.Election
	pd              *mempool.Producer
	pm              *pacemaker.Pacemaker
	start           chan bool
	isStarted       atomic.Bool
	isByz           bool
	timer           *time.Timer
	committedBlocks chan *blockchain.Block
	eventChan       chan interface{}
}

// NewReplica creates a new replica instance
func NewReplica(id identity.NodeID, alg string, isByz bool) *Replica {
	r := new(Replica)
	r.Node = node.NewNode(id, isByz)
	if isByz {
		log.Infof("[%v] is Byzantine", r.ID())
	}
	r.Election = election.NewRotation(config.GetConfig().N())
	r.pd = mempool.NewProducer()
	r.pm = pacemaker.NewPacemaker(config.GetConfig().N())
	r.start = make(chan bool)
	r.eventChan = make(chan interface{})
	r.committedBlocks = make(chan *blockchain.Block)
	r.isByz = isByz
	r.Register(blockchain.Block{}, r.HandleBlock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	r.Register(pacemaker.TMO{}, r.HandleTmo)
	r.Register(blockchain.QC{}, r.HandleQC)
	r.Register(message.Transaction{}, r.handleTxn)
	gob.Register(blockchain.QC{})
	gob.Register(blockchain.Block{})
	gob.Register(blockchain.Vote{})
	gob.Register(pacemaker.TC{})
	gob.Register(pacemaker.TMO{})
	switch alg {
	case "hotstuff":
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks)
	//case "tchs":
	//	r.Safety = tchs.Newtchs(bc)
	case "streamlet":
		r.Safety = streamlet.NewStreamlet(r.Node, r.pm, r.Election, r.committedBlocks)
	default:
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks)
	}
	return r
}

/* Message Handlers */

func (r *Replica) HandleBlock(block blockchain.Block) {
	log.Debugf("[%v] received a block from %v, view is %v", r.ID(), block.Proposer, block.View)
	//err := r.Safety.ProcessBlock(&block)
	//if err != nil {
	//	log.Warningf("[%v] cannot process block %w", r.ID(), err)
	//	return
	//}
	r.eventChan <- block
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	log.Debugf("[%v] received a vote from %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	//r.Safety.ProcessVote(&vote)
	r.eventChan <- vote
}

func (r *Replica) HandleQC(qc blockchain.QC) {
	log.Debugf("[%v] received a qc, blockID is %x", r.ID(), qc.BlockID)
	//r.Safety.ProcessVote(&vote)
	r.eventChan <- qc
}

func (r *Replica) HandleTmo(tmo pacemaker.TMO) {
	log.Debugf("[%v] received a timeout from %v for view %v", r.ID(), tmo.NodeID, tmo.View)
	//r.Safety.ProcessRemoteTmo(&tmo)
	r.eventChan <- tmo
}

func (r *Replica) handleTxn(m message.Transaction) {
	r.pd.CollectTxn(&m)
	if !r.isStarted.Load() {
		log.Debugf("[%v] is boosting", r.ID())
		r.isStarted.Store(true)
		r.start <- true
		// wait for others to get started
		time.Sleep(200 * time.Millisecond)
	}

	//	the last node is to kick-off the protocol
	if r.pm.GetCurView() == 0 && r.IsLeader(r.ID(), 1) {
		log.Debugf("[%v] is going to kick off the protocol", r.ID())
		r.pm.AdvanceView(0)
	}
}

/* Processors */

func (r *Replica) processCommittedBlock(block *blockchain.Block) {
	for _, txn := range block.Payload {
		if r.ID() == txn.NodeID {
			txn.Reply(message.TransactionReply{})
		}
	}
	r.pd.RemoveTxns(block.Payload)
	log.Infof("[%v] the block is committed, No. of transactions: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.Payload), block.View, r.pm.GetCurView(), block.ID)
}

func (r *Replica) processNewView(newView types.View) {
	log.Debugf("[%v] is processing new view: %v", r.ID(), newView)
	if !r.IsLeader(r.ID(), newView) {
		return
	}
	r.proposeBlock(newView)
}

func (r *Replica) proposeBlock(view types.View) {
	block := r.Safety.MakeProposal(r.pd.GeneratePayload())
	log.Infof("[%v] is going to propose block for view: %v, id: %x, prevID: %x", r.ID(), view, block.ID, block.PrevID)
	go r.HandleBlock(*block)
	r.Broadcast(block)
	log.Debugf("[%v] broadcast is done for sending the block", r.ID())
}

func (r *Replica) ListenLocalEvent() {
	for {
		r.timer = time.NewTimer(r.pm.GetTimerForView())
	L:
		for {
			select {
			case view := <-r.pm.EnteringViewEvent():
				r.processNewView(view)
				break L
			case <-r.timer.C:
				r.Safety.ProcessLocalTmo(r.pm.GetCurView())
			case committedBlock := <-r.committedBlocks:
				r.processCommittedBlock(committedBlock)
			}
		}
	}
}

// Start starts event loop
func (r *Replica) Start() {
	// fail-stop case
	if r.isByz {
		return
	}
	go r.Run()
	<-r.start
	go r.ListenLocalEvent()
	for r.isStarted.Load() {
		event := <-r.eventChan
		switch v := event.(type) {
		case blockchain.Block:
			_ = r.Safety.ProcessBlock(&v)
		case blockchain.Vote:
			r.Safety.ProcessVote(&v)
		case pacemaker.TMO:
			r.Safety.ProcessRemoteTmo(&v)
		case blockchain.QC:
			r.Safety.ProcessCertificate(&v)
		}
	}
}
