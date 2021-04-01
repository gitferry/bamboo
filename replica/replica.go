package replica

import (
	"encoding/gob"
	"fmt"
	fhs "github.com/gitferry/bamboo/fasthostuff"
	"github.com/gitferry/bamboo/lbft"
	"sync"
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
	"github.com/gitferry/bamboo/tchs"
	"github.com/gitferry/bamboo/types"
)

type TimeTrack struct {
	startTime        time.Time
	startProcessTime time.Time
	endProcessTime   time.Time
	endTime          time.Time
}

type Replica struct {
	node.Node
	Safety
	election.Election
	pd              *mempool.Producer
	pm              *pacemaker.Pacemaker
	start           chan bool // signal to start the node
	isStarted       atomic.Bool
	isByz           bool
	timer           *time.Timer // timeout for each view
	committedBlocks chan *blockchain.Block
	forkedBlocks    chan *blockchain.Block
	eventChan       chan interface{}

	/* for monitoring node statistics */
	thrus                string
	lastViewTime         time.Time
	startTime            time.Time
	tmpTime              time.Time
	voteStart            time.Time
	totalCreateDuration  time.Duration
	totalProcessDuration time.Duration
	totalProposeDuration time.Duration
	totalDelay           time.Duration
	totalRoundTime       time.Duration
	totalVoteTime        time.Duration
	viewTimeTrack        map[types.View]*TimeTrack
	aveProposeTime       float64
	aveProcessTime       float64
	aveVotingTime        float64

	totalBlockSize   int
	receivedNo       int
	roundNo          int
	voteNo           int
	totalCommittedTx int
	latencyNo        int
	proposedNo       int
	processedNo      int
	committedNo      int
	mu               sync.Mutex
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
	r.isByz = isByz
	r.pd = mempool.NewProducer()
	r.pm = pacemaker.NewPacemaker(config.GetConfig().N())
	r.start = make(chan bool)
	r.eventChan = make(chan interface{})
	r.committedBlocks = make(chan *blockchain.Block, 100)
	r.viewTimeTrack = make(map[types.View]*TimeTrack)
	r.forkedBlocks = make(chan *blockchain.Block, 100)
	r.Register(blockchain.Block{}, r.HandleBlock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	r.Register(pacemaker.TMO{}, r.HandleTmo)
	r.Register(message.Transaction{}, r.handleTxn)
	r.Register(message.Query{}, r.handleQuery)
	gob.Register(blockchain.Block{})
	gob.Register(blockchain.Vote{})
	gob.Register(pacemaker.TC{})
	gob.Register(pacemaker.TMO{})

	// Is there a better way to reduce the number of parameters?
	switch alg {
	case "hotstuff":
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	case "tchs":
		r.Safety = tchs.NewTchs(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	case "streamlet":
		r.Safety = streamlet.NewStreamlet(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	case "lbft":
		r.Safety = lbft.NewLbft(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	case "fasthotstuff":
		r.Safety = fhs.NewFhs(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	default:
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	}
	return r
}

/* Message Handlers */

func (r *Replica) HandleBlock(block blockchain.Block) {
	r.receivedNo++
	r.startSignal()
	log.Debugf("[%v] received a block from %v, view is %v, id: %x, prevID: %x", r.ID(), block.Proposer, block.View, block.ID, block.PrevID)
	r.eventChan <- block
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	if vote.View < r.pm.GetCurView() {
		return
	}
	r.startSignal()
	log.Debugf("[%v] received a vote frm %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	r.eventChan <- vote
}

func (r *Replica) HandleTmo(tmo pacemaker.TMO) {
	if tmo.View < r.pm.GetCurView() {
		return
	}
	log.Debugf("[%v] received a timeout from %v for view %v", r.ID(), tmo.NodeID, tmo.View)
	r.eventChan <- tmo
}

// handleQuery replies a query with the statistics of the node
func (r *Replica) handleQuery(m message.Query) {
	aveBlockSize := float64(r.totalBlockSize) / float64(r.proposedNo)
	//for _, tc := range r.viewTimeTrack {
	//	if tc.startProcessTime.IsZero() || tc.endProcessTime.IsZero() {
	//		r.totalProposeDuration += tc.startProcessTime.Sub(tc.startTime)
	//		r.totalProcessDuration += tc.endProcessTime.Sub(tc.startProcessTime)
	//		r.totalVoteTime += tc.endTime.Sub(tc.endProcessTime)
	//	}
	//}
	//aveProposeTime := float64(r.totalProcessDuration) / float64(len(r.viewTimeTrack))
	//aveProcessTime := float64(r.totalProcessDuration) / float64(len(r.viewTimeTrack))
	//aveVotingTime := float64(r.totalVoteTime) / float64(len(r.viewTimeTrack))
	latency := float64(r.totalDelay.Milliseconds()) / float64(r.latencyNo)
	r.totalCommittedTx = 0
	r.tmpTime = time.Now()
	status := fmt.Sprintf("No. of views: %v.\nAve. proposing time is %v ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nAve. block size is %v.\nLatency is %v ms.\n", len(r.viewTimeTrack), r.aveProposeTime, r.aveProcessTime, r.aveVotingTime, aveBlockSize, latency)
	m.Reply(message.QueryReply{Info: status})
}

func (r *Replica) handleTxn(m message.Transaction) {
	r.pd.AddTxn(&m)
	r.startSignal()
	// the first leader kicks off the protocol
	if r.pm.GetCurView() == 0 && r.IsLeader(r.ID(), 1) {
		log.Debugf("[%v] is going to kick off the protocol", r.ID())
		r.pm.AdvanceView(0)
	}
}

/* Processors */

func (r *Replica) processCommittedBlock(block *blockchain.Block) {
	if block.Proposer == r.ID() {
		for _, txn := range block.Payload {
			// only record the delay of transactions from the local memory pool
			delay := time.Now().Sub(txn.Timestamp)
			r.totalDelay += delay
			r.latencyNo++
		}
	}
	r.committedNo++
	r.totalCommittedTx += len(block.Payload)
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
	block := r.Safety.MakeProposal(view, r.pd.GeneratePayload())
	r.totalBlockSize += len(block.Payload)
	r.proposedNo++
	r.Broadcast(block)
	if _, exists := r.viewTimeTrack[block.View]; !exists {
		r.viewTimeTrack[block.View] = &TimeTrack{startTime: time.Now()}
	}
	r.viewTimeTrack[block.View].startProcessTime = time.Now()
	_ = r.Safety.ProcessBlock(block)
	r.viewTimeTrack[block.View].endProcessTime = time.Now()
}

// ListenLocalEvent listens new view and timeout events
func (r *Replica) ListenLocalEvent() {
	r.lastViewTime = time.Now()
	r.timer = time.NewTimer(r.pm.GetTimerForView())
	for {
		r.timer.Reset(r.pm.GetTimerForView())
	L:
		for {
			select {
			case view := <-r.pm.EnteringViewEvent():
				r.mu.Lock()
				r.viewTimeTrack[view] = &TimeTrack{
					startTime: time.Now(),
				}
				r.mu.Unlock()
				if view >= 1 {
					r.mu.Lock()
					if tc, exists := r.viewTimeTrack[view-1]; exists {
						tc.endTime = time.Now()
						//log.Debugf("time track for view %d, start: %v, startP: %v, endP: %v, end: %v", view-1, tc.startTime.Second(), tc.startProcessTime.Second(), tc.endProcessTime.Second(), tc.endTime.Second())
						//r.totalProposeDuration += tc.startProcessTime.Sub(tc.startTime)
						r.totalProcessDuration += tc.endProcessTime.Sub(tc.startProcessTime)
						//r.totalVoteTime += tc.endTime.Sub(tc.endProcessTime)
						//r.aveProposeTime = float64(r.totalProposeDuration.Milliseconds()) / float64(len(r.viewTimeTrack))
						r.aveProcessTime = float64(r.totalProcessDuration.Milliseconds()) / float64(len(r.viewTimeTrack))
						//r.aveVotingTime = float64(r.totalVoteTime.Milliseconds()) / float64(len(r.viewTimeTrack))
						//log.Debugf("Ave. proposing time is %v ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\n", r.totalProposeDuration, r.totalProcessDuration, r.totalVoteTime)
					}
					r.mu.Unlock()
				}
				// measure round time
				now := time.Now()
				lasts := now.Sub(r.lastViewTime)
				r.totalRoundTime += lasts
				r.roundNo++
				r.lastViewTime = now
				r.eventChan <- view
				log.Debugf("[%v] the last view lasts %v milliseconds, current view: %v", r.ID(), lasts.Milliseconds(), view)
				break L
			case <-r.timer.C:
				r.Safety.ProcessLocalTmo(r.pm.GetCurView())
				break L
			}
		}
	}
}

// ListenCommittedBlocks listens committed blocks and forked blocks from the protocols
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

func (r *Replica) startSignal() {
	if !r.isStarted.Load() {
		r.startTime = time.Now()
		r.tmpTime = time.Now()
		log.Debugf("[%v] is boosting", r.ID())
		r.isStarted.Store(true)
		r.start <- true
	}
}

// Start starts event loop
func (r *Replica) Start() {
	go r.Run()
	// wait for the start signal
	<-r.start
	go r.ListenLocalEvent()
	go r.ListenCommittedBlocks()
	r.viewTimeTrack[0] = &TimeTrack{
		startTime: time.Now(),
	}
	for r.isStarted.Load() {
		event := <-r.eventChan
		switch v := event.(type) {
		case types.View:
			r.processNewView(v)
		case blockchain.Block:
			r.mu.Lock()
			if _, exists := r.viewTimeTrack[v.View]; !exists {
				r.viewTimeTrack[v.View] = &TimeTrack{startTime: time.Now()}
			}
			r.viewTimeTrack[v.View].startProcessTime = time.Now()
			r.mu.Unlock()
			_ = r.Safety.ProcessBlock(&v)
			r.mu.Lock()
			r.viewTimeTrack[v.View].endProcessTime = time.Now()
			r.mu.Unlock()
		case blockchain.Vote:
			r.Safety.ProcessVote(&v)
		case pacemaker.TMO:
			r.Safety.ProcessRemoteTmo(&v)
		}
	}
}
