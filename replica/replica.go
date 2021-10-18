package replica

import (
	"encoding/gob"
	"fmt"
	"github.com/gitferry/bamboo/crypto"
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
	"github.com/gitferry/bamboo/types"
)

type Replica struct {
	node.Node
	Safety
	election.Election
	sm              mempool.SharedMempool
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
	totalBlockSize       int
	receivedNo           int
	roundNo              int
	voteNo               int
	totalCommittedTx     int
	latencyNo            int
	proposedNo           int
	processedNo          int
	committedNo          int
	pendingBlockMap      map[crypto.Identifier]*blockchain.PendingBlock
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
	r.sm = mempool.NewNaiveMem()
	r.pm = pacemaker.NewPacemaker(config.GetConfig().N())
	r.start = make(chan bool)
	r.eventChan = make(chan interface{})
	r.committedBlocks = make(chan *blockchain.Block, 100)
	r.forkedBlocks = make(chan *blockchain.Block, 100)
	r.pendingBlockMap = make(map[crypto.Identifier]*blockchain.PendingBlock)
	r.Register(blockchain.Proposal{}, r.HandleProposal)
	r.Register(blockchain.MicroBlock{}, r.HandleMicroblock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	r.Register(pacemaker.TMO{}, r.HandleTmo)
	r.Register(message.Transaction{}, r.handleTxn)
	r.Register(message.Query{}, r.handleQuery)
	r.Register(message.MissingMBRequest{}, r.HandleMissingMBRequest)
	gob.Register(blockchain.Proposal{})
	gob.Register(blockchain.MicroBlock{})
	gob.Register(blockchain.Vote{})
	gob.Register(pacemaker.TC{})
	gob.Register(pacemaker.TMO{})
	gob.Register(message.MissingMBRequest{})

	// Is there a better way to reduce the number of parameters?
	switch alg {
	case "hotstuff":
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	//case "tchs":
	//	r.Safety = tchs.NewTchs(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	//case "streamlet":
	//	r.Safety = streamlet.NewStreamlet(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	//case "lbft":
	//	r.Safety = lbft.NewLbft(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	//case "fasthotstuff":
	//	r.Safety = fhs.NewFhs(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	default:
		r.Safety = hotstuff.NewHotStuff(r.Node, r.pm, r.Election, r.committedBlocks, r.forkedBlocks)
	}
	return r
}

/* Message Handlers */

// HandleProposal handles proposals from the leader
// it first checks if the referred microblocks exist in the mempool
// and requests the missing ones
func (r *Replica) HandleProposal(proposal blockchain.Proposal) {
	r.receivedNo++
	r.startSignal()
	log.Debugf("[%v] received a proposal from %v, view is %v, id: %x, prevID: %x", r.ID(), proposal.Proposer, proposal.View, proposal.ID, proposal.PrevID)
	pendingBlock := r.sm.FillProposal(&proposal)
	block := pendingBlock.CompleteBlock()
	if block != nil {
		log.Debugf("[%v] a block is ready, view: %v, id: %x", r.ID(), proposal.View, proposal.ID)
		r.eventChan <- block
	} else {
		r.pendingBlockMap[proposal.ID] = pendingBlock
		log.Debugf("[%v] %v microblocks are missing in a pending block, id: %x", r.ID(), pendingBlock.MissingCount(), proposal.ID)
		missingRequest := message.MissingMBRequest{
			RequesterID:   r.ID(),
			ProposalID:    proposal.ID,
			MissingMBList: pendingBlock.MissingMBList(),
		}
		r.Send(proposal.Proposer, missingRequest)
	}
}

// HandleMicroblock handles microblocks from replicas
// it first checks if the relevant proposal is pending
// if so, tries to complete the block
func (r *Replica) HandleMicroblock(mb blockchain.MicroBlock) {
	pd, exists := r.pendingBlockMap[mb.ProposalID]
	if exists {
		block := pd.AddMicroblock(&mb)
		if block != nil {
			log.Debugf("[%v] a block is ready, view: %v, id: %x", r.ID(), pd.Proposal.View, pd.Proposal.ID)
			delete(r.pendingBlockMap, mb.ProposalID)
			r.eventChan <- block
		}
	} else {
		err := r.sm.AddMicroblock(&mb)
		if err != nil {
			log.Errorf("[%v] can not add a microblock, id: %x", r.ID(), mb.Hash)
		}
	}
}

func (r *Replica) HandleMissingMBRequest(mbr message.MissingMBRequest) {
	log.Debugf("[%v] %d missing microblocks request is received from %v for proposal %x", r.ID(), len(mbr.MissingMBList), mbr.RequesterID, mbr.ProposalID)
	for _, mbid := range mbr.MissingMBList {
		found, mb := r.sm.FindMicroblock(mbid)
		if found {
			log.Debugf("[%v] a microblock is found in mempool for proposal %x", r.ID(), mbr.ProposalID)
			r.Send(mbr.RequesterID, mb)
		} else {
			log.Errorf("[%v] a requested microblock for proposal %x is not found in mempool, id: %x", r.ID(), mbr.ProposalID, mb.Hash)
		}
	}
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
	realAveProposeTime := float64(r.totalProposeDuration.Milliseconds()) / float64(r.processedNo)
	aveProcessTime := float64(r.totalProcessDuration.Milliseconds()) / float64(r.processedNo)
	aveVoteProcessTime := float64(r.totalVoteTime.Milliseconds()) / float64(r.roundNo)
	aveBlockSize := float64(r.totalBlockSize) / float64(r.proposedNo)
	//requestRate := float64(r.sm.TotalReceivedTxNo()) / time.Now().Sub(r.startTime).Seconds()
	//committedRate := float64(r.committedNo) / time.Now().Sub(r.startTime).Seconds()
	aveRoundTime := float64(r.totalRoundTime.Milliseconds()) / float64(r.roundNo)
	aveProposeTime := aveRoundTime - aveProcessTime - aveVoteProcessTime
	latency := float64(r.totalDelay.Milliseconds()) / float64(r.latencyNo)
	//r.thrus += fmt.Sprintf("Time: %v s. Throughput: %v txs/s\n", time.Now().Sub(r.startTime).Seconds(), float64(r.totalCommittedTx)/time.Now().Sub(r.tmpTime).Seconds())
	r.totalCommittedTx = 0
	r.tmpTime = time.Now()
	//status := fmt.Sprintf("chain status is: %s\nCommitted rate is %v.\nAve. block size is %v.\nAve. trans. delay is %v ms.\nAve. creation time is %f ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nRequest rate is %f txs/s.\nAve. round time is %f ms.\nLatency is %f ms.\nThroughput is %f txs/s.\n", r.Safety.GetChainStatus(), committedRate, aveBlockSize, aveTransDelay, aveCreateDuration, aveProcessTime, aveVoteProcessTime, requestRate, aveRoundTime, latency, throughput)
	status := fmt.Sprintf("Ave. actual proposing time is %v ms.\nAve. proposing time is %v ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nAve. block size is %v.\nAve. round time is %v ms.\nLatency is %v ms.\n", realAveProposeTime, aveProposeTime, aveProcessTime, aveVoteProcessTime, aveBlockSize, aveRoundTime, latency)
	m.Reply(message.QueryReply{Info: status})
}

func (r *Replica) handleTxn(m message.Transaction) {
	isbuilt, mb := r.sm.AddTxn(&m)
	if isbuilt {
		r.Broadcast(mb)
	}
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
		for _, mb := range block.MicroblockList() {
			for _, txn := range mb.Txns {
				// only record the delay of transactions from the local memory pool
				delay := time.Now().Sub(txn.Timestamp)
				r.totalDelay += delay
				r.latencyNo++
			}
			err := r.sm.RemoveMicroblock(mb.Hash)
			if err != nil {
				log.Errorf("[%v] failed to remove microblock, id: %x", r.ID(), mb.Hash)
			}
		}
	}
	r.committedNo++
	log.Infof("[%v] the block is committed, No. of microblocks: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.MicroblockList()), block.View, r.pm.GetCurView(), block.ID)
}

func (r *Replica) processForkedBlock(block *blockchain.Block) {
	//if block.Proposer == r.Hash() {
	//	for _, txn := range block.payload {
	//		// collect txn back to mem pool
	//		//r.sm.CollectTxn(txn)
	//	}
	//}
	//log.Infof("[%v] the block is forked, No. of transactions: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.payload), block.View, r.pm.GetCurView(), block.ID)
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
	payload := r.sm.GeneratePayload()
	proposal := r.Safety.MakeProposal(view, payload.GenerateHashList())
	r.totalBlockSize += len(proposal.HashList)
	r.proposedNo++
	createEnd := time.Now()
	createDuration := createEnd.Sub(createStart)
	proposal.Timestamp = time.Now()
	r.totalCreateDuration += createDuration
	r.Broadcast(proposal)
	block := blockchain.BuildBlock(proposal, payload)
	_ = r.Safety.ProcessBlock(block)
	r.voteStart = time.Now()
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
				if view >= 2 {
					r.totalVoteTime += time.Now().Sub(r.voteStart)
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
	for r.isStarted.Load() {
		event := <-r.eventChan
		switch v := event.(type) {
		case types.View:
			r.processNewView(v)
		case blockchain.Block:
			startProcessTime := time.Now()
			r.totalProposeDuration += startProcessTime.Sub(v.Timestamp)
			_ = r.Safety.ProcessBlock(&v)
			r.totalProcessDuration += time.Now().Sub(startProcessTime)
			r.voteStart = time.Now()
			r.processedNo++
		case blockchain.Vote:
			startProcessTime := time.Now()
			r.Safety.ProcessVote(&v)
			processingDuration := time.Now().Sub(startProcessTime)
			r.totalVoteTime += processingDuration
			r.voteNo++
		case pacemaker.TMO:
			r.Safety.ProcessRemoteTmo(&v)
		}
	}
}
