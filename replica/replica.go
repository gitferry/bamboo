package replica

import (
	"encoding/gob"
	"fmt"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/limiter"
	"github.com/gitferry/bamboo/utils"
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
	estimator       *Estimator
	start           chan bool // signal to start the node
	isStarted       atomic.Bool
	isByz           bool
	timer           *time.Timer // timeout for each view
	committedBlocks chan *blockchain.Block
	forkedBlocks    chan *blockchain.Block
	eventChan       chan interface{}

	/* for monitoring node statistics */
	thrus                  string
	lastViewTime           time.Time
	startTime              time.Time
	tmpTime                time.Time
	voteStart              time.Time
	totalCreateDuration    time.Duration
	totalProcessDuration   time.Duration
	totalProposeDuration   time.Duration
	totalDisseminationTime time.Duration
	totalDelay             time.Duration
	totalRoundTime         time.Duration
	totalVoteTime          time.Duration
	totalBlockSize         int
	totalMicroblocks       int
	totalProposedMBs       int
	missingMicroblocks     int
	receivedNo             int
	roundNo                int
	voteNo                 int
	totalCommittedTx       int
	latencyNo              int
	proposedNo             int
	processedNo            int
	committedNo            int
	totalHops              int
	totalCommittedMBs      int
	totalRedundantMBs      int
	totalReceivedTxs       int
	missingCounts          map[identity.NodeID]int
	pendingBlockMap        map[crypto.Identifier]*blockchain.PendingBlock
	missingMBs             map[crypto.Identifier]crypto.Identifier // microblock hash to proposal hash
	receivedMBs            map[crypto.Identifier]struct{}
	selfMBChan             chan blockchain.MicroBlock
	otherMBChan            chan blockchain.MicroBlock
	limiter                *limiter.Bucket
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
	r.pm = pacemaker.NewPacemaker(config.GetConfig().N())
	r.estimator = NewEstimator()
	r.start = make(chan bool)
	r.eventChan = make(chan interface{})
	r.committedBlocks = make(chan *blockchain.Block, 100)
	r.forkedBlocks = make(chan *blockchain.Block, 100)
	r.pendingBlockMap = make(map[crypto.Identifier]*blockchain.PendingBlock)
	r.missingMBs = make(map[crypto.Identifier]crypto.Identifier)
	r.receivedMBs = make(map[crypto.Identifier]struct{})
	r.missingCounts = make(map[identity.NodeID]int)
	r.selfMBChan = make(chan blockchain.MicroBlock, 1024)
	r.otherMBChan = make(chan blockchain.MicroBlock, 1024)
	r.limiter = limiter.NewBucket(time.Duration(config.Configuration.FillInterval)*time.Millisecond, int64(config.Configuration.Capacity))
	memType := config.GetConfig().MemType
	switch memType {
	case "naive":
		r.sm = mempool.NewNaiveMem()
	case "time":
		r.sm = mempool.NewTimemem()
	case "ack":
		r.sm = mempool.NewAckMem()
	}
	r.Register(blockchain.Proposal{}, r.HandleProposal)
	r.Register(blockchain.MicroBlock{}, r.HandleMicroblock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	r.Register(pacemaker.TMO{}, r.HandleTmo)
	r.Register(message.Transaction{}, r.handleTxn)
	r.Register(message.Query{}, r.handleQuery)
	r.Register(message.MissingMBRequest{}, r.HandleMissingMBRequest)
	r.Register(message.Ack{}, r.HandleAck)
	gob.Register(blockchain.Proposal{})
	gob.Register(blockchain.MicroBlock{})
	gob.Register(blockchain.Vote{})
	gob.Register(pacemaker.TC{})
	gob.Register(pacemaker.TMO{})
	gob.Register(message.MissingMBRequest{})
	gob.Register(message.Ack{})

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
	r.totalProposeDuration += time.Now().Sub(proposal.Timestamp)
	log.Debugf("[%v] received a proposal from %v, containing %v microblocks, view is %v, id: %x, prevID: %x", r.ID(), proposal.Proposer, len(proposal.HashList), proposal.View, proposal.ID, proposal.PrevID)
	if config.Configuration.MemType == "time" {
		ack := message.Ack{
			SentTime: proposal.Timestamp,
			AckTime:  time.Now(),
			Receiver: r.ID(),
			ID:       proposal.ID,
			Type:     "p",
		}
		r.Send(proposal.Proposer, ack)
	}
	pendingBlock := r.sm.FillProposal(&proposal)
	block := pendingBlock.CompleteBlock()
	if block != nil {
		log.Debugf("[%v] a block is ready, view: %v, id: %x", r.ID(), proposal.View, proposal.ID)
		r.eventChan <- *block
		return
	}
	r.pendingBlockMap[proposal.ID] = pendingBlock
	log.Debugf("[%v] %v microblocks are missing in id: %x", r.ID(), pendingBlock.MissingCount(), proposal.ID)
	for _, mbid := range pendingBlock.MissingMBList() {
		r.missingMBs[mbid] = proposal.ID
		log.Debugf("[%v] a microblock is missing, id: %x", r.ID(), mbid)
	}
	missingRequest := message.MissingMBRequest{
		RequesterID:   r.ID(),
		ProposalID:    proposal.ID,
		MissingMBList: pendingBlock.MissingMBList(),
	}
	r.Send(proposal.Proposer, missingRequest)
}

// HandleMicroblock handles microblocks from replicas
// it first checks if the relevant proposal is pending
// if so, tries to complete the block
func (r *Replica) HandleMicroblock(mb blockchain.MicroBlock) {
	//r.startSignal()
	_, ok := r.receivedMBs[mb.Hash]
	if ok {
		r.totalRedundantMBs++
		return
	}
	r.totalDisseminationTime += time.Now().Sub(mb.Timestamp)
	r.receivedMBs[mb.Hash] = struct{}{}
	r.totalMicroblocks++

	// gossip
	//if a quorum of acks is not reached, gossip the microblock
	if config.Configuration.Gossip == true {
		mb.Hops++
		r.otherMBChan <- mb
		log.Debugf("[%v] an other mb is added", r.ID())
		//if config.Configuration.MemType == "naive" {
		//	if mb.Hops <= config.Configuration.R {
		//		go r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
		//	}
		//} else if config.Configuration.MemType == "ack" {
		//	if !r.sm.IsStable(mb.Hash) && mb.Hops <= config.Configuration.R {
		//		go r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
		//	}
		//}
	}
	//log.Debugf("[%v] received a microblock, id: %x", r.ID(), mb.Hash)
	proposalID, exists := r.missingMBs[mb.Hash]
	if exists {
		log.Debugf("[%v] a missing mb for proposal is found", r.ID())
		pd, exists := r.pendingBlockMap[proposalID]
		if exists {
			log.Debugf("[%v] received a microblock %x for pending proposal %x", r.ID(), mb.Hash, proposalID)
			r.missingMicroblocks++
			block := pd.AddMicroblock(&mb)
			if block != nil {
				log.Debugf("[%v] a block is ready, view: %v, id: %x", r.ID(), pd.Proposal.View, pd.Proposal.ID)
				delete(r.pendingBlockMap, mb.ProposalID)
				delete(r.missingMBs, mb.Hash)
				r.eventChan <- *block
			}
		}
	} else {
		err := r.sm.AddMicroblock(&mb)
		if err != nil {
			log.Errorf("[%v] can not add a microblock, id: %x", r.ID(), mb.Hash)
		}
		// ack
		if !mb.IsRequested && (config.Configuration.MemType == "time" || config.Configuration.MemType == "ack") {
			ack := message.Ack{
				SentTime: mb.Timestamp,
				AckTime:  time.Now(),
				Receiver: r.ID(),
				ID:       mb.Hash,
				Type:     "mb",
			}
			if config.Configuration.MemType == "time" {
				r.Send(mb.Sender, ack)
			} else {
				r.Broadcast(ack)
			}
		}
	}
}

func (r *Replica) HandleMissingMBRequest(mbr message.MissingMBRequest) {
	log.Debugf("[%v] %d missing microblocks request is received from %v", r.ID(), len(mbr.MissingMBList), mbr.RequesterID)
	r.missingCounts[mbr.RequesterID] += len(mbr.MissingMBList)
	for _, mbid := range mbr.MissingMBList {
		found, mb := r.sm.FindMicroblock(mbid)
		if found {
			mb.IsRequested = true
			r.Send(mbr.RequesterID, mb)
		} else {
			log.Errorf("[%v] a requested microblock is not found in mempool, id: %x", r.ID(), mbid)
		}
	}
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	if vote.View < r.pm.GetCurView() {
		return
	}
	r.totalVoteTime += time.Now().Sub(vote.Timestamp)
	r.voteNo++
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

func (r *Replica) HandleAck(ack message.Ack) {
	//log.Debugf("[%v] received an ack message, type: %v, id: %x", r.ID(), ack.Type, ack.ID)
	//r.eventChan <- ack
	r.processAcks(&ack)
}

// handleQuery replies a query with the statistics of the node
func (r *Replica) handleQuery(m message.Query) {
	//realAveProposeTime := float64(r.totalProposeDuration.Milliseconds()) / float64(r.processedNo)
	//aveProcessTime := float64(r.totalProcessDuration.Milliseconds()) / float64(r.processedNo)
	//aveVoteProcessTime := float64(r.totalVoteTime.Milliseconds()) / float64(r.roundNo)
	//aveBlockSize := float64(r.totalBlockSize) / float64(r.proposedNo)
	//requestRate := float64(r.sm.TotalReceivedTxNo()) / time.Now().Sub(r.startTime).Seconds()
	//committedRate := float64(r.committedNo) / time.Now().Sub(r.startTime).Seconds()
	//aveRoundTime := float64(r.totalRoundTime.Milliseconds()) / float64(r.roundNo)
	//aveProposeTime := aveRoundTime - aveProcessTime - aveVoteProcessTime
	latency := float64(r.totalDelay.Milliseconds()) / float64(r.latencyNo)
	r.thrus = fmt.Sprintf("Time: %v s. Throughput: %v txs/s\n",
		time.Now().Sub(r.startTime).Seconds(), float64(r.totalCommittedTx)/time.Now().Sub(r.startTime).Seconds())
	//r.totalCommittedTx = 0
	//r.tmpTime = time.Now()
	aveTxRate := float64(r.totalReceivedTxs) / time.Now().Sub(r.startTime).Seconds()
	aveRoundTime := float64(r.totalRoundTime.Milliseconds()) / float64(r.roundNo)
	aveHops := float64(r.totalHops) / float64(r.totalCommittedMBs)
	aveProposeTime := float64(r.totalProposeDuration.Milliseconds()) / float64(r.receivedNo)
	aveDisseminationTime := float64(r.totalDisseminationTime.Milliseconds()) / float64(r.totalMicroblocks)
	aveVoteTime := float64(r.totalVoteTime.Milliseconds()) / float64(r.voteNo)
	var missingCounts string
	for k, v := range r.missingCounts {
		missingCounts += fmt.Sprintf("%v: %v\n", k, v)
	}
	//status := fmt.Sprintf("chain status is: %s\nCommitted rate is %v.\nAve. block size is %v.\nAve. trans. delay is %v ms.\nAve. creation time is %f ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nRequest rate is %f txs/s.\nAve. round time is %f ms.\nLatency is %f ms.\nThroughput is %f txs/s.\n", r.Safety.GetChainStatus(), committedRate, aveBlockSize, aveTransDelay, aveCreateDuration, aveProcessTime, aveVoteProcessTime, requestRate, aveRoundTime, latency, throughput)
	//status := fmt.Sprintf("Ave. actual proposing time is %v ms.\nAve. proposing time is %v ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nAve. block size is %v.\nAve. round time is %v ms.\nLatency is %v ms.\n", realAveProposeTime, aveProposeTime, aveProcessTime, aveVoteProcessTime, aveBlockSize, aveRoundTime, latency)
	status := fmt.Sprintf("Ave. View Time: %vms\nAve. Propose Time: %vms\nAve. Dissemination Time: %vms\nAve. Vote Time: %vms\nAve. Tx Rate: %v\nLatency: %v ms\nRedundant microblocks:%v\nTotal microblocks: %v\nTotal missing microblocks: %v\nTotoal proposed microblocks:%v\nAve. hops:%v\n%sMissing counts:\n%s\n",
		aveRoundTime, aveProposeTime, aveDisseminationTime, aveVoteTime, aveTxRate, latency, r.totalRedundantMBs, r.totalMicroblocks, r.missingMicroblocks, r.totalProposedMBs, aveHops, r.thrus, missingCounts)
	m.Reply(message.QueryReply{Info: status})
}

func (r *Replica) handleTxn(m message.Transaction) {
	r.startSignal()
	isbuilt, mb := r.sm.AddTxn(&m)
	r.totalReceivedTxs++
	if isbuilt {
		if config.Configuration.MemType == "time" {
			stableTime := r.estimator.PredictStableTime("mb")
			//stableTime := time.Duration(0)
			//log.Debugf("[%v] stable time for a microblock is %v", r.ID(), stableTime)
			mb.FutureTimestamp = time.Now().Add(stableTime)
		}
		mb.Sender = r.ID()
		r.sm.AddMicroblock(mb)
		mb.Timestamp = time.Now()
		r.totalMicroblocks++
		if config.Configuration.Gossip == false {
			r.Broadcast(mb)
		} else {
			mb.Hops++
			r.selfMBChan <- *mb
			log.Debugf("a self mb is added")
			//r.MulticastQuorum(r.pickFanoutNodes(mb), mb)
		}
	}
	// the first leader kicks off the protocol
	if r.pm.GetCurView() == 0 && r.IsLeader(r.ID(), 1) {
		log.Debugf("[%v] is going to kick off the protocol", r.ID())
		r.pm.AdvanceView(0)
	}
}

func (r *Replica) gossip() {
	for {
		tt := r.limiter.Take(int64(config.Configuration.Fanout))
		log.Debugf("[%v] wait for %vms to do the next gossip", r.ID(), tt.Milliseconds())
		time.Sleep(tt)
	L:
		for {
			select {
			case mb := <-r.selfMBChan:
				log.Debugf("[%v] is going to gossip a self mb", r.ID())
				r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
				break L
			default:
				select {
				case mb := <-r.selfMBChan:
					log.Debugf("[%v] is going to gossip a self mb", r.ID())
					r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
					break L
				case mb := <-r.otherMBChan:
					if !r.sm.IsStable(mb.Hash) {
						log.Debugf("[%v] is going to gossip an other mb", r.ID())
						r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
						break L
					} else {
						continue
					}
				}
			}
		}

	}
}

func (r *Replica) pickFanoutNodes(mb *blockchain.MicroBlock) []identity.NodeID {
	sentNodes := mb.FindSentNodes()
	nodes := utils.PickRandomNodes(sentNodes)
	mb.AddSentNodes(nodes)
	mb.AddSentNodes(r.sm.AckList(mb.Hash))
	mb.AddSentNodes([]identity.NodeID{r.ID()})
	return nodes
}

/* Processors */

func (r *Replica) processCommittedBlock(block *blockchain.Block) {
	var txCount int
	r.totalCommittedMBs += len(block.MicroblockList())
	for _, mb := range block.MicroblockList() {
		txCount += len(mb.Txns)
		for _, txn := range mb.Txns {
			// only record the delay of transactions from the local memory pool
			delay := time.Now().Sub(txn.Timestamp)
			r.totalDelay += delay
			r.latencyNo++
			r.totalCommittedTx++
		}
		r.totalHops += mb.Hops
		//err := r.sm.RemoveMicroblock(mb.Hash)
		//if err != nil {
		//	log.Debugf("[%v] processing committed block err: %w", r.ID(), err)
		//}
	}
	r.committedNo++
	log.Infof("[%v] the block is committed, No. of microblocks: %v, No. of tx: %v, view: %v, current view: %v, id: %x",
		r.ID(), len(block.MicroblockList()), txCount, block.View, r.pm.GetCurView(), block.ID)
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

func (r *Replica) processAcks(ack *message.Ack) {
	if config.Configuration.MemType == "time" {
		r.estimator.AddAck(ack)
	} else if config.Configuration.MemType == "ack" {
		r.sm.AddAck(ack)
		found, _ := r.sm.FindMicroblock(ack.ID)
		if !found && r.sm.IsStable(ack.ID) {
			missingRequest := message.MissingMBRequest{
				RequesterID:   r.ID(),
				MissingMBList: []crypto.Identifier{ack.ID},
			}
			r.Send(ack.Receiver, missingRequest)
			log.Debugf("[%v] has received enough acks, but not received the microblock id: %x, fetch from %v",
				r.ID(), ack.ID, ack.Receiver)
		}
	}
}

func (r *Replica) proposeBlock(view types.View) {
	createStart := time.Now()
	payload := r.sm.GeneratePayload()
	// if we are using time-based shared mempool, wait until all the microblocks are stable
	if config.Configuration.MemType == "time" {
		r.waitUntilStable(payload)
	}
	r.totalProposedMBs += len(payload.MicroblockList)
	proposal := r.Safety.MakeProposal(view, payload.GenerateHashList())
	log.Debugf("[%v] is making a proposal for view %v, containing %v microblocks, id:%x", proposal.Proposer, proposal.View, len(proposal.HashList), proposal.ID)
	log.Debugf("[%v] contained microblocks are", r.ID())
	for _, id := range proposal.HashList {
		log.Debugf("[%v] id: %x", r.ID(), id)
	}
	r.totalBlockSize += len(proposal.HashList)
	r.proposedNo++
	createEnd := time.Now()
	createDuration := createEnd.Sub(createStart)
	r.totalCreateDuration += createDuration
	proposal.Timestamp = time.Now()
	r.Broadcast(proposal)
	block := blockchain.BuildBlock(proposal, payload)
	_ = r.Safety.ProcessBlock(block)
	r.voteStart = time.Now()
}

func (r *Replica) waitUntilStable(payload *blockchain.Payload) {
	lastItem := payload.LastItem()
	if lastItem == nil {
		return
	}
	stableTime := r.estimator.PredictStableTime("p")
	//stableTime := time.Duration(0)
	wait := lastItem.FutureTimestamp.Sub(time.Now()) - stableTime
	//log.Debugf("[%v] stable time for a proposal is %v", r.ID(), stableTime)
	if stableTime < 0 {
		log.Errorf("[%v] stable time for proposal is less than 0")
	}
	if wait > 0 {
		log.Debugf("[%v] wait for %v until the contained microblocks are stable", r.ID(), wait)
		time.Sleep(wait)
	}
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
					//r.totalVoteTime += time.Now().Sub(r.voteStart)
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
	go r.gossip()
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
			_ = r.Safety.ProcessBlock(&v)
			r.totalProcessDuration += time.Now().Sub(startProcessTime)
			r.voteStart = time.Now()
			r.processedNo++
		case blockchain.Vote:
			//startProcessTime := time.Now()
			r.Safety.ProcessVote(&v)
			//processingDuration := time.Now().Sub(startProcessTime)
			//r.totalVoteTime += processingDuration
			//r.voteNo++
		case pacemaker.TMO:
			r.Safety.ProcessRemoteTmo(&v)
		default:
			log.Errorf("[%v] received an unknown event %v", r.ID(), v)
		}
	}
}
