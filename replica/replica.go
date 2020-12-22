package replica

import (
	"encoding/gob"
	"fmt"
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

type Replica struct {
	node.Node
	Safety
	election.Election
	pd                   *mempool.Producer
	pm                   *pacemaker.Pacemaker
	start                chan bool
	isStarted            atomic.Bool
	isByz                bool
	timer                *time.Timer
	committedBlocks      chan *blockchain.Block
	forkedBlocks         chan *blockchain.Block
	eventChan            chan interface{}
	hasher               crypto.Hasher
	signer               string
	lastViewTime         time.Time
	startTime            time.Time
	totalCreateDuration  time.Duration
	totalProcessDuration time.Duration
	totalDelay           time.Duration
	totalCommittedTx     int
	proposedNo           int
}

// NewReplica creates a new replica instance
func NewReplica(id identity.NodeID, alg string, isByz bool) *Replica {
	r := new(Replica)
	r.Node = node.NewNode(id, isByz)
	if isByz {
		log.Infof("[%v] is Byzantine", r.ID())
	}
	if config.GetConfig().Master == "0" {
		r.Election = election.NewRotation(config.GetConfig().N())
	} else {
		r.Election = election.NewStatic(config.GetConfig().Master)
	}
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
	r.Register(message.Query{}, r.handleQuery)
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
	if !r.isStarted.Load() {
		log.Debugf("[%v] is boosting", r.ID())
		r.isStarted.Store(true)
		r.start <- true
	}
	log.Debugf("The block is printed as %v", block)
	r.eventChan <- block
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	if vote.View < r.pm.GetCurView() {
		log.Debugf("[%v] received a stale vote, view: %v, block id: %x", r.ID(), vote.View, vote.BlockID)
		return
	}
	if !r.isStarted.Load() {
		log.Debugf("[%v] is boosting", r.ID())
		r.isStarted.Store(true)
		r.start <- true
	}

	log.Debugf("[%v] received a vote from %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	r.eventChan <- vote
}

func (r *Replica) HandleTmo(tmo pacemaker.TMO) {
	log.Debugf("[%v] received a timeout from %v for view %v", r.ID(), tmo.NodeID, tmo.View)
	r.eventChan <- tmo
}

func (r *Replica) handleQuery(m message.Query) {
	aveCreateDuration := float64(r.totalCreateDuration.Milliseconds()) / float64(r.proposedNo)
	aveProcessDuration := float64(r.totalProcessDuration.Milliseconds()) / float64(r.proposedNo)
	requestRate := float64(r.pd.TotalReceivedTxNo()) / time.Now().Sub(r.startTime).Seconds()
	latency := float64(r.totalDelay.Milliseconds()) / float64(r.totalCommittedTx)
	throughput := float64(r.totalCommittedTx) / time.Now().Sub(r.startTime).Seconds()
	status := fmt.Sprintf("chain status is: %s\nAve. creation time is %f ms.\nAve. processing time is %f ms.\nRequest rate is %f txs/s.\nLatency is %f ms.\nThroughput is %f txs/s.", r.Safety.GetChainStatus(), aveCreateDuration, aveProcessDuration, requestRate, latency, throughput)
	m.Reply(message.QueryReply{Info: status})
}

func (r *Replica) handleTxn(m message.Transaction) {
	r.pd.AddTxn(&m)
	if !r.isStarted.Load() {
		r.startTime = time.Now()
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
	if block.Proposer == r.ID() {
		for _, txn := range block.Payload {
			delay := time.Now().Sub(txn.Timestamp)
			txn.Reply(message.NewReply(delay))
			r.totalDelay += delay
			r.totalCommittedTx++
		}
	}
	log.Infof("[%v] the block is committed, No. of transactions: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.Payload), block.View, r.pm.GetCurView(), block.ID)
}

func (r *Replica) processForkedBlock(block *blockchain.Block) {
	if block.Proposer == r.ID() {
		for _, txn := range block.Payload {
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
	r.totalCreateDuration += createDuration
	//log.Debugf("[%v] finished creating the block for view: %v, payload size: %v, used: %v microseconds, id: %x, prevID: %x", r.ID(), view, len(block.Payload), createDuration.Microseconds(), block.ID, block.PrevID)
	_ = r.Safety.ProcessBlock(block)
	processDuration := time.Now().Sub(createEnd)
	r.totalProcessDuration += processDuration
	r.proposedNo++
	//log.Debugf("[%v] finished processing the block for view: %v, used: %v microseconds, id: %x, prevID: %x", r.ID(), view, processDuration.Microseconds(), block.ID, block.PrevID)
	r.Broadcast(block)
}

func (r *Replica) ListenLocalEvent() {
	r.lastViewTime = time.Now()
	r.timer = time.NewTimer(r.pm.GetTimerForView())
	roundTimeMeasure := make([]time.Duration, 0, 5000)
	for {
		r.timer.Reset(r.pm.GetTimerForView())
		currentLeader := r.FindLeaderFor(r.pm.GetCurView())
		if config.GetConfig().IsByzantine(currentLeader) && config.GetConfig().Strategy == "silence" {
			r.pm.AdvanceView(r.pm.GetCurView())
		}
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
