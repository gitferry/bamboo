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
	totalRoundTime       time.Duration
	totalVoteTime        time.Duration
	totalTransDelay      time.Duration
	totalBlockSize       int
	receivedNo           int
	roundNo              int
	voteNo               int
	totalCommittedTx     int
	latencyNo            int
	proposedNo           int
	processedNo          int
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
	r.committedBlocks = make(chan *blockchain.Block, 100)
	r.forkedBlocks = make(chan *blockchain.Block, 100)
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
	r.totalTransDelay += time.Now().Sub(block.Timestamp)
	r.receivedNo++
	if !r.isStarted.Load() {
		log.Debugf("[%v] is boosting", r.ID())
		r.isStarted.Store(true)
		r.start <- true
	}
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
	aveProcessTime := float64(r.totalProcessDuration.Milliseconds()) / float64(r.processedNo)
	aveVoteProcessTime := float64(r.totalVoteTime.Milliseconds()) / float64(r.voteNo)
	aveBlockSize := float64(r.totalBlockSize) / float64(r.proposedNo)
	aveTransDelay := float64(r.totalTransDelay.Milliseconds()) / float64(r.receivedNo)
	requestRate := float64(r.pd.TotalReceivedTxNo()) / time.Now().Sub(r.startTime).Seconds()
	aveRoundTime := float64(r.totalRoundTime.Milliseconds()) / float64(r.roundNo)
	latency := float64(r.totalDelay.Milliseconds()) / float64(r.latencyNo)
	throughput := float64(r.totalCommittedTx) / time.Now().Sub(r.startTime).Seconds()
	status := fmt.Sprintf("chain status is: %s\nAve. block size is %v.\nAve. trans. delay is %v ms.\nAve. creation time is %f ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nRequest rate is %f txs/s.\nAve. round time is %f ms.\nLatency is %f ms.\nThroughput is %f txs/s.\n", r.Safety.GetChainStatus(), aveBlockSize, aveTransDelay, aveCreateDuration, aveProcessTime, aveVoteProcessTime, requestRate, aveRoundTime, latency, throughput)
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
		//time.Sleep(100 * time.Millisecond)
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
		if block.Proposer == r.ID() {
			delay := time.Now().Sub(txn.Timestamp)
			r.totalDelay += delay
			r.latencyNo++
		}
		r.totalCommittedTx++
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
	r.totalBlockSize += len(block.Payload)
	createEnd := time.Now()
	createDuration := createEnd.Sub(createStart)
	log.Debugf("[%v] spent %v to create the block for view: %v", r.ID(), createDuration, block.View)
	block.Timestamp = time.Now()
	r.Broadcast(block)
	r.totalCreateDuration += createDuration
	r.proposedNo++
	//log.Debugf("[%v] finished creating the block for view: %v, payload size: %v, used: %v microseconds, id: %x, prevID: %x", r.ID(), view, len(block.Payload), createDuration.Microseconds(), block.ID, block.PrevID)
	_ = r.Safety.ProcessBlock(block)
	processedDuration := time.Now().Sub(createEnd)
	r.totalProcessDuration += processedDuration
	r.processedNo++
	log.Debugf("[%v] spent %v to process the block for view: %v", r.ID(), processedDuration, block.View)
	//log.Debugf("[%v] finished processing the block for view: %v, used: %v microseconds, id: %x, prevID: %x", r.ID(), view, processDuration.Microseconds(), block.ID, block.PrevID)
}

func (r *Replica) ListenLocalEvent() {
	r.lastViewTime = time.Now()
	r.timer = time.NewTimer(r.pm.GetTimerForView())
	//roundTimeMeasure := make([]time.Duration, 0, 5000)
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
				//if view >= 10 {
				log.Debugf("[%v] spent %v in the last view %v", r.ID(), lasts, r.pm.GetCurView()-1)
				r.totalRoundTime += lasts
				r.roundNo++
				//}
				//if int(view) == config.GetConfig().MaxRound {
				//	var sumRoundTime time.Duration
				//	for _, t := range roundTimeMeasure {
				//		sumRoundTime = sumRoundTime + t
				//	}
				//	log.Infof("[%v] the average view duration is %v microseconds, measured %v views", r.ID(), float64(sumRoundTime.Microseconds())/float64(len(roundTimeMeasure)))
				//}
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
			startProcessTime := time.Now()
			_ = r.Safety.ProcessBlock(&v)
			r.processedNo++
			processingDuration := time.Now().Sub(startProcessTime)
			r.totalProcessDuration += processingDuration
			log.Debugf("[%v] spent %v to process the block for view: %v", r.ID(), processingDuration, v.View)
		case blockchain.Vote:
			startProcessTime := time.Now()
			r.Safety.ProcessVote(&v)
			r.voteNo++
			processingDuration := time.Now().Sub(startProcessTime)
			r.totalVoteTime += processingDuration
			log.Debugf("[%v] spent %v to process the vote for view: %v", r.ID(), processingDuration, v.View)
		case pacemaker.TMO:
			r.Safety.ProcessRemoteTmo(&v)
		}
	}
}
