package replica

import (
	"encoding/gob"
	"time"

	"go.uber.org/atomic"

	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/election"
	"github.com/gitferry/bamboo/hotstuff"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/mempool"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/node"
	"github.com/gitferry/bamboo/pacemaker"
	"github.com/gitferry/bamboo/streamlet"
	"github.com/gitferry/bamboo/tchs"
	"github.com/gitferry/bamboo/types"
)

const SILENCE = "silence"
const FORK = "fork"

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
	forkedBlocks    chan *blockchain.Block
	eventChan       chan interface{}
	hasher          crypto.Hasher
	signer          string
	lastViewTime    time.Time
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
	r.forkedBlocks = make(chan *blockchain.Block)
	r.isByz = isByz
	r.Register(blockchain.Block{}, r.HandleBlock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	r.Register(pacemaker.TMO{}, r.HandleTmo)
	r.Register(message.Transaction{}, r.handleTxn)
	gob.Register(blockchain.Block{})
	gob.Register(blockchain.Vote{})
	gob.Register(pacemaker.TC{})
	gob.Register(pacemaker.TMO{})
	switch alg {
	case "hotstuff":
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	case "tchs":
		r.Safety = tchs.NewTchs(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	case "streamlet":
		r.Safety = streamlet.NewStreamlet(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	default:
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	}
	return r
}

/* Message Handlers */

func (r *Replica) HandleBlock(block blockchain.Block) {
	log.Debugf("[%v] received a block from %v, view is %v, id: %x", r.ID(), block.Proposer, block.View, block.ID)
	r.eventChan <- block
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	if vote.View < r.pm.GetCurView() {
		log.Debugf("[%v] received a stale vote, view: %v, block id: %x", r.ID(), vote.View, vote.BlockID)
		return
	}

	log.Debugf("[%v] received a vote from %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	r.eventChan <- vote
}

func (r *Replica) HandleTmo(tmo pacemaker.TMO) {
	log.Debugf("[%v] received a timeout from %v for view %v", r.ID(), tmo.NodeID, tmo.View)
	r.eventChan <- tmo
}

func (r *Replica) handleTxn(m message.Transaction) {
	r.pd.AddTxn(&m)
	if !r.isStarted.Load() {
		log.Debugf("[%v] is boosting", r.ID())
		r.isStarted.Store(true)
		r.start <- true
		// wait for others to get started
		time.Sleep(100 * time.Millisecond)
	}

	// kick-off the protocol
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
	log.Infof("[%v] the block is committed, No. of transactions: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.Payload), block.View, r.pm.GetCurView(), block.ID)
}

func (r *Replica) processForkedBlock(block *blockchain.Block) {
	for _, txn := range block.Payload {
		if r.ID() == txn.NodeID {
			// collect txn back to mem pool
			r.pd.CollectTxn(txn)
		}
	}
	log.Infof("[%v] the block is forked, No. of transactions: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.Payload), block.View, r.pm.GetCurView(), block.ID)
}

func (r *Replica) processNewView(newView types.View) {
	log.Debugf("[%v] is processing new view: %v, leader is %v", r.ID(), newView, r.FindLeaderFor(newView))
	if !r.IsLeader(r.ID(), newView) {
		return
	}
	r.proposeBlock(newView)
}

func (r *Replica) proposeBlock(view types.View) {
	createStart := time.Now()
	block := r.Safety.MakeProposal(r.pd.GeneratePayload())
	createEnd := time.Now()
	createDuration := createEnd.Sub(createStart)
	log.Infof("[%v] finished creating the block for view: %v, used: %v microseconds, id: %x, prevID: %x", r.ID(), view, createDuration.Microseconds(), block.ID, block.PrevID)
	_ = r.Safety.ProcessBlock(block)
	processDuration := time.Now().Sub(createEnd)
	log.Infof("[%v] finished processing the block for view: %v, used: %v microseconds, id: %x, prevID: %x", r.ID(), view, processDuration.Microseconds(), block.ID, block.PrevID)
	r.Broadcast(block)
	log.Debugf("[%v] broadcast is done for sending the block", r.ID())
}

func (r *Replica) ListenLocalEvent() {
	r.lastViewTime = time.Now()
	r.timer = time.NewTimer(r.pm.GetTimerForView())
	roundTimeMeasure := make([]time.Duration, 0, 5000)
	for {
		r.timer.Reset(r.pm.GetTimerForView())
	L:
		for {
			select {
			case view := <-r.pm.EnteringViewEvent():
				// measure round time
				now := time.Now()
				lasts := now.Sub(r.lastViewTime)
				if view >= 10 && int(view) < config.GetConfig().MaxRound {
					roundTimeMeasure = append(roundTimeMeasure, lasts)
				}
				if int(view) == config.GetConfig().MaxRound {
					var sumRoundTime time.Duration
					for _, t := range roundTimeMeasure {
						sumRoundTime = sumRoundTime + t
					}
					log.Infof("[%v] the average view duration is %v microseconds, measured %v views", r.ID(), float64(sumRoundTime.Microseconds())/float64(len(roundTimeMeasure)))
				}
				r.lastViewTime = now
				//log.Infof("[%v] the last view lasts %v milliseconds, current view: %v", r.ID(), lasts.Milliseconds(), view)
				r.eventChan <- view
				break L
			case <-r.timer.C:
				r.Safety.ProcessLocalTmo(r.pm.GetCurView())
			}
		}
	}
}

func (r *Replica) ListenCommittedBlocks() {
	for {
		select {
		case committedBlock := <-r.committedBlocks:
			r.processCommittedBlock(committedBlock)
		case forkedBlock := <-r.forkedBlocks:
			r.processForkedBlock(forkedBlock)
		}
	}
}

// Start starts event loop
func (r *Replica) Start() {
	if r.isByz && config.GetConfig().Strategy == SILENCE {
		// perform silence attack
		log.Debugf("[%v] is performing silence attack", r.ID())
		return
	}
	go r.Run()
	<-r.start
	go r.ListenLocalEvent()
	go r.ListenCommittedBlocks()
	for r.isStarted.Load() {
		event := <-r.eventChan
		switch v := event.(type) {
		case types.View:
			r.processNewView(v)
		case blockchain.Block:
			_ = r.Safety.ProcessBlock(&v)
		case blockchain.Vote:
			r.Safety.ProcessVote(&v)
		case pacemaker.TMO:
			r.Safety.ProcessRemoteTmo(&v)
		}
	}
}
