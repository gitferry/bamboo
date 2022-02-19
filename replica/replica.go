package replica

import (
	"encoding/gob"
	"fmt"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/limiter"
	"github.com/gitferry/bamboo/utils"
	"github.com/kelindar/bitmap"
	"math/rand"
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
	sm mempool.SharedMempool
	pm *pacemaker.Pacemaker
	//estimator       *Estimator
	start           chan bool // signal to start the node
	isStarted       atomic.Bool
	isByz           bool
	timer           *time.Timer // timeout for each view
	committedBlocks chan *blockchain.Block
	forkedBlocks    chan *blockchain.Block
	eventChan       chan interface{}

	/* for monitoring node statistics */
	thrus                     string
	lastViewTime              time.Time
	startTime                 time.Time
	tmpTime                   time.Time
	voteStart                 time.Time
	totalCreateDuration       time.Duration
	totalProcessDuration      time.Duration
	totalProposeDuration      time.Duration
	totalDisseminationTime    time.Duration
	totalDelay                time.Duration
	totalRoundTime            time.Duration
	totalVoteTime             time.Duration
	totalSlowDisemminationDur time.Duration
	totalSlowMBs              int
	totalBlockSize            int
	totalMicroblocks          int
	totalProposedMBs          int
	missingMicroblocks        int
	receivedNo                int
	roundNo                   int
	voteNo                    int
	totalCommittedTx          int
	latencyNo                 int
	proposedNo                int
	processedNo               int
	committedNo               int
	totalHops                 int
	totalCommittedMBs         int
	totalRedundantMBs         int
	totalReceivedTxs          int
	txNoInMB                  int
	missingCounts             map[identity.NodeID]int
	pendingBlockMap           map[crypto.Identifier]*blockchain.PendingBlock
	missingMBs                map[crypto.Identifier]crypto.Identifier // microblock hash to proposal hash
	receivedMBs               map[crypto.Identifier]struct{}
	selfMBChan                chan blockchain.MicroBlock
	otherMBChan               chan blockchain.MicroBlock
	limiter                   *limiter.Bucket
	mbSentNodes               map[crypto.Identifier]bitmap.Bitmap
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
	//r.estimator = NewEstimator()
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
	r.mbSentNodes = make(map[crypto.Identifier]bitmap.Bitmap)
	r.limiter = limiter.NewBucket(time.Duration(config.Configuration.FillInterval)*time.Millisecond, int64(config.Configuration.Capacity))
	memType := config.GetConfig().MemType
	switch memType {
	case "naive":
		r.sm = mempool.NewNaiveMem()
	//case "time":
	//	r.sm = mempool.NewTimemem()
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
	r.Register(blockchain.Ack{}, r.HandleAck)
	gob.Register(blockchain.Proposal{})
	gob.Register(blockchain.MicroBlock{})
	gob.Register(blockchain.Vote{})
	gob.Register(pacemaker.TC{})
	gob.Register(pacemaker.TMO{})
	gob.Register(message.MissingMBRequest{})
	gob.Register(blockchain.Ack{})

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
	//if config.Configuration.MemType == "time" {
	//	ack := message.Ack{
	//		SentTime: proposal.Timestamp,
	//		AckTime:  time.Now(),
	//		Receiver: r.ID(),
	//		ID:       proposal.ID,
	//		Type:     "p",
	//	}
	//	r.Send(proposal.Proposer, ack)
	//}
	r.totalBlockSize += len(proposal.HashList)
	pendingBlock := r.sm.FillProposal(&proposal)
	if config.Configuration.MemType == "ack" {
		if !r.verifySigs(pendingBlock.Payload.SigMap) {
			log.Warningf("[%v] received an block %x with invalid sigs for microblocks", r.ID(), proposal.ID)
		}
	}
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
	if config.Configuration.MemType == "ack" {
		block = blockchain.BuildBlock(pendingBlock.Proposal, pendingBlock.Payload)
		r.eventChan <- *block
		return
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
	r.startSignal()
	// gossip
	//if a quorum of acks is not reached, gossip the microblock
	go func() {
		if config.Configuration.Gossip == true && !mb.IsRequested && r.ID() != config.Configuration.Master {
			mb.Hops++
			if mb.Hops <= config.Configuration.R {
				r.otherMBChan <- mb
			}
		}
	}()
	if config.Configuration.LoadBalance && mb.IsForward {
		mb.IsForward = false
		r.Broadcast(mb)
		return
	}
	_, ok := r.receivedMBs[mb.Hash]
	if ok {
		r.totalRedundantMBs++
		return
	}
	defer r.kickOff()
	r.totalDisseminationTime += time.Now().Sub(mb.Timestamp)
	if mb.Sender.Node() <= config.Configuration.SlowNo {
		r.totalSlowDisemminationDur += time.Now().Sub(mb.Timestamp)
		r.totalSlowMBs++
	}
	r.receivedMBs[mb.Hash] = struct{}{}
	r.totalMicroblocks++
	mb.FutureTimestamp = time.Now()

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
				if config.Configuration.Gossip == true && config.Configuration.MemType == "ack" {
					return
				}
				r.eventChan <- *block
			}
		}
	} else {
		err := r.sm.AddMicroblock(&mb)
		if err != nil {
			log.Errorf("[%v] can not add a microblock, id: %x", r.ID(), mb.Hash)
		}
		// ack
		if !mb.IsRequested && config.Configuration.MemType == "ack" {
			//if config.Configuration.MemType == "time" {
			//	r.Send(mb.Sender, ack)
			//} else {
			//	r.Broadcast(ack)
			//}
			leader := r.GetCurrentLeader()
			ack := blockchain.MakeAck(r.ID(), mb.Hash)
			if leader != r.ID() {
				r.Send(leader, blockchain.MakeAck(r.ID(), mb.Hash))
			} else {
				r.HandleAck(*ack)
			}
		}
	}
}

func (r *Replica) HandleMissingMBRequest(mbr message.MissingMBRequest) {
	log.Debugf("[%v] %d missing microblocks request is received from %v, missing mbs are:", r.ID(), len(mbr.MissingMBList), mbr.RequesterID)
	r.missingCounts[mbr.RequesterID] += len(mbr.MissingMBList)
	for _, mbid := range mbr.MissingMBList {
		found, mb := r.sm.FindMicroblock(mbid)
		log.Debugf("[%v] id: %x", r.ID(), mbid)
		if found {
			mb.IsRequested = true
			r.Send(mbr.RequesterID, mb)
		} else {
			log.Errorf("[%v] a requested microblock is not found in mempool, id: %x", r.ID(), mbid)
		}
	}
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	log.Debugf("[%v] received a vote frm %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	if vote.View < r.pm.GetCurView() {
		//log.Warningf("[%v] received a vote has lower view", r.ID())
		return
	}
	r.totalVoteTime += time.Now().Sub(vote.Timestamp)
	r.voteNo++
	//r.startSignal()
	//log.Debugf("[%v] received a vote frm %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	r.eventChan <- vote
}

func (r *Replica) HandleTmo(tmo pacemaker.TMO) {
	log.Debugf("[%v] received a timeout from %v for view %v", r.ID(), tmo.NodeID, tmo.View)
	if tmo.View < r.pm.GetCurView() {
		return
	}
	r.eventChan <- tmo
}

func (r *Replica) HandleAck(ack blockchain.Ack) {
	log.Debugf("[%v] received an ack message form %v, id: %x", r.ID(), ack.Receiver, ack.MicroblockID)
	r.processAcks(&ack)
}

// handleQuery replies a query with the statistics of the node
func (r *Replica) handleQuery(m message.Query) {
	//realAveProposeTime := float64(r.totalProposeDuration.Milliseconds()) / float64(r.processedNo)
	//aveProcessTime := float64(r.totalProcessDuration.Milliseconds()) / float64(r.processedNo)
	//aveVoteProcessTime := float64(r.totalVoteTime.Milliseconds()) / float64(r.roundNo)
	aveBlockSize := float64(r.totalBlockSize) / float64(r.proposedNo+r.receivedNo)
	//requestRate := float64(r.sm.TotalReceivedTxNo()) / time.Now().Sub(r.startTime).Seconds()
	//committedRate := float64(r.committedNo) / time.Now().Sub(r.startTime).Seconds()
	//aveRoundTime := float64(r.totalRoundTime.Milliseconds()) / float64(r.roundNo)
	//aveProposeTime := aveRoundTime - aveProcessTime - aveVoteProcessTime
	//latency := float64(r.totalDelay.Milliseconds()) / float64(r.latencyNo)
	r.thrus += fmt.Sprintf("Time: %v s. Throughput: %v txs/s. Delay: %v ms\n",
		time.Now().Sub(r.startTime).Seconds(), float64(r.totalCommittedTx)/time.Now().Sub(r.tmpTime).Seconds(), float64(r.totalDelay.Milliseconds())/float64(r.latencyNo))
	r.totalCommittedTx = 0
	r.tmpTime = time.Now()
	r.totalDelay = 0
	r.latencyNo = 0
	aveCreationTime := float64(r.totalCreateDuration.Milliseconds()) / float64(r.proposedNo)
	aveTxRate := float64(r.sm.TotalTx()) / time.Now().Sub(r.startTime).Seconds()
	aveRoundTime := float64(r.totalRoundTime.Milliseconds()) / float64(r.roundNo)
	aveHops := float64(r.totalHops) / float64(r.totalCommittedMBs)
	aveProposeTime := float64(r.totalProposeDuration.Milliseconds()) / float64(r.receivedNo)
	aveDisseminationTime := float64(r.totalDisseminationTime.Milliseconds()) / float64(r.totalMicroblocks)
	aveSlowDisseminationTime := float64(r.totalSlowDisemminationDur.Milliseconds()) / float64(r.totalSlowMBs)
	r.totalSlowDisemminationDur = 0
	r.totalSlowMBs = 0
	aveVoteTime := float64(r.totalVoteTime.Milliseconds()) / float64(r.voteNo)
	mbRate := float64(r.sm.TotalMB()) / time.Now().Sub(r.startTime).Seconds()
	//status := fmt.Sprintf("chain status is: %s\nCommitted rate is %v.\nAve. block size is %v.\nAve. trans. delay is %v ms.\nAve. creation time is %f ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nRequest rate is %f txs/s.\nAve. round time is %f ms.\nLatency is %f ms.\nThroughput is %f txs/s.\n", r.Safety.GetChainStatus(), committedRate, aveBlockSize, aveTransDelay, aveCreateDuration, aveProcessTime, aveVoteProcessTime, requestRate, aveRoundTime, latency, throughput)
	//status := fmt.Sprintf("Ave. actual proposing time is %v ms.\nAve. proposing time is %v ms.\nAve. processing time is %v ms.\nAve. vote time is %v ms.\nAve. block size is %v.\nAve. round time is %v ms.\nLatency is %v ms.\n", realAveProposeTime, aveProposeTime, aveProcessTime, aveVoteProcessTime, aveBlockSize, aveRoundTime, latency)
	status := fmt.Sprintf("Ave. View Time: %vms\nAve. Propose Time: %vms\nAve. Dissemination Time: %vms, slow dissemination time: %v\nAve. Creation Time: %v, a proposal contains %v microblocks\nAve. Vote Time: %vms\nAve. Tx Rate: %v\nAve. MB Rate: %v, an MB contains %v txs\nRedundant microblocks:%v\nTotal microblocks: %v, Remaining microblocks: %v\nTotal missing microblocks: %v\nTotoal proposed microblocks:%v\nAve. hops:%v\nSend Rate: %v Mbps\nRecv Rate: %v Mbps\nTotal txs: %v, Remaining txs: %v\n%s\n",
		aveRoundTime, aveProposeTime, aveDisseminationTime, aveSlowDisseminationTime, aveCreationTime, aveBlockSize, aveVoteTime, aveTxRate, mbRate, r.txNoInMB, r.totalRedundantMBs, r.sm.TotalMB(), r.sm.RemainingMB(), r.missingMicroblocks, r.totalProposedMBs, aveHops, r.SendRate(), r.RecvRate(), r.sm.TotalTx(), r.sm.RemainingTx(), r.thrus)
	m.Reply(message.QueryReply{Info: status})
}

func (r *Replica) handleTxn(m message.Transaction) {
	r.startSignal()
	m.Timestamp = time.Now()
	isbuilt, mb := r.sm.AddTxn(&m)
	if isbuilt {
		//if config.Configuration.MemType == "time" {
		//	stableTime := r.estimator.PredictStableTime("mb")
		//stableTime := time.Duration(0)
		//log.Debugf("[%v] stable time for a microblock is %v", r.ID(), stableTime)
		//	mb.FutureTimestamp = time.Now().Add(stableTime)
		//}
		r.txNoInMB = len(mb.Txns)
		mb.Sender = r.ID()
		r.sm.AddMicroblock(mb)
		mb.Timestamp = time.Now()
		r.totalMicroblocks++
		r.totalProposedMBs++
		//if config.Configuration.Gossip == false {
		//	if r.isByz && config.Configuration.Strategy == "missing" {
		//		if config.Configuration.MemType == "naive" {
		//			r.Send(r.GetCurrentLeader(), mb)
		//		} else if config.Configuration.MemType == "ack" {
		//			r.MulticastQuorum(r.randomPick(), mb)
		//		}
		//	} else {
		//		r.Broadcast(mb)
		//	}
		//} else {
		//	mb.Hops++
		//	r.selfMBChan <- *mb
		//}
		if config.Configuration.LoadBalance == false {
			if r.isByz && config.Configuration.Strategy == "missing" {
				if config.Configuration.MemType == "naive" {
					r.Send(r.GetCurrentLeader(), mb)
				} else if config.Configuration.MemType == "ack" {
					r.MulticastQuorum(r.randomPick(), mb)
				}
			} else {
				r.Broadcast(mb)
			}
		} else {
			mb.Hops++
			r.selfMBChan <- *mb
		}
	}
	r.kickOff()
}

func (r *Replica) kickOff() {
	// the first leader kicks off the protocol
	if r.pm.GetCurView() == 0 && r.IsLeader(r.ID(), 1) {
		log.Debugf("[%v] is going to kick off the protocol", r.ID())
		r.pm.AdvanceView(0)
	}
}

func (r *Replica) loadbalance() {
	for {
		select {
		case mb := <-r.selfMBChan:
			if rand.Intn(100) < config.Configuration.ForwardP && r.ID().Node() <= config.Configuration.LoadedIndex {
				mb.IsForward = true
				pick := pickRandomNodes(config.Configuration.N()-1, 1, config.Configuration.LoadedIndex)[0]
				log.Debugf("[%v] is going to forward a mb to %v", r.ID(), pick)
				r.Send(pick, mb)
			} else {
				r.Broadcast(mb)
			}
			//log.Debugf("[%v] is going to gossip a self mb", r.ID())
			//r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
		default:
		}
	}
}

func (r *Replica) gossip() {
	for {
		tt := r.limiter.Take(int64(config.Configuration.Fanout))
		time.Sleep(tt)
	L:
		for {
			select {
			case mb := <-r.selfMBChan:
				//log.Debugf("[%v] is going to gossip a self mb", r.ID())
				r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
				break L
			default:
				select {
				case mb := <-r.selfMBChan:
					//log.Debugf("[%v] is going to gossip a self mb", r.ID())
					r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
					break L
				case mb := <-r.otherMBChan:
					if !r.sm.IsStable(mb.Hash) && mb.Hops <= config.Configuration.R {
						if r.ID().Node() > config.Configuration.SlowNo {
							r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
							break L
						} else if rand.Intn(100) < config.Configuration.P {
							r.MulticastQuorum(r.pickFanoutNodes(&mb), mb)
							break L
						} else {
							continue
						}
					} else {
						continue
					}
				}
			}
		}
	}
}

func (r *Replica) randomPick() []identity.NodeID {
	n := config.GetConfig().N() - 1 // exluding the master
	f := n/3 + 1
	pick := utils.RandomPick(n, f)
	pickedNode := make([]identity.NodeID, f)
	for i, item := range pick {
		pickedNode[i] = identity.NewNodeID(item + 2)
	}
	return pickedNode
}

func pickRandomNodes(n, d, index int) []identity.NodeID {
	pick := utils.RandomPick(n-index, d)
	pickedNode := make([]identity.NodeID, d)
	for i, item := range pick {
		pickedNode[i] = identity.NewNodeID(item + 1 + index)
	}
	return pickedNode
}

func (r *Replica) pickFanoutNodes(mb *blockchain.MicroBlock) []identity.NodeID {
	if !config.Configuration.Opt {
		return utils.PickRandomNodes(nil)
	}
	if bm, exists := r.mbSentNodes[mb.Hash]; exists {
		mb.AddSentNodes(utils.BitmapToNodes(bm))
	}
	mb.AddSentNodes(r.sm.AckList(mb.Hash))
	mb.AddSentNodes([]identity.NodeID{r.ID()})
	sentNodes := mb.FindSentNodes()
	nodes := utils.PickRandomNodes(sentNodes)
	mb.AddSentNodes(nodes)
	r.mbSentNodes[mb.Hash] = mb.Bitmap
	//mb.AddSentNodes(nodes)
	//log.Debugf("[%v] mb %x has received %v, is going to send to %v, %v hops", r.ID(), mb.Hash, sentNodes, nodes, mb.Hops)
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

func (r *Replica) processAcks(ack *blockchain.Ack) {
	//if config.Configuration.MemType == "time" {
	//	r.estimator.AddAck(ack)
	if config.Configuration.MemType == "ack" {
		if r.sm.IsStable(ack.MicroblockID) {
			return
		}
		if ack.Receiver != r.ID() {
			voteIsVerified, err := crypto.PubVerify(ack.Signature, crypto.IDToByte(ack.MicroblockID), ack.Receiver)
			if err != nil {
				log.Warningf("[%v] Error in verifying the signature in ack id: %x", r.ID(), ack.MicroblockID)
				return
			}
			if !voteIsVerified {
				log.Warningf("[%v] received an ack with invalid signature. vote id: %x", r.ID(), ack.MicroblockID)
				return
			}
		}
		r.sm.AddAck(ack)
		found, _ := r.sm.FindMicroblock(ack.MicroblockID)
		if !found && r.sm.IsStable(ack.MicroblockID) {
			missingRequest := message.MissingMBRequest{
				RequesterID:   r.ID(),
				MissingMBList: []crypto.Identifier{ack.MicroblockID},
			}
			r.Send(ack.Receiver, missingRequest)
			log.Debugf("[%v] has received enough acks, but not received the microblock id: %x, fetch from %v",
				r.ID(), ack.MicroblockID, ack.Receiver)
		}
	}
}

func (r *Replica) proposeBlock(view types.View) {
	createStart := time.Now()
	//time.Sleep(time.Duration(config.Configuration.ProposeTime) * time.Millisecond)
	payload := r.sm.GeneratePayload()
	// if we are using time-based shared mempool, wait until all the microblocks are stable
	//if config.Configuration.MemType == "time" {
	//	r.waitUntilStable(payload)
	//}
	proposal := r.Safety.MakeProposal(view, payload.GenerateHashList())
	log.Debugf("[%v] is making a proposal for view %v, containing %v microblocks, %v left,id:%x", proposal.Proposer, proposal.View, len(proposal.HashList), r.sm.RemainingMB(), proposal.ID)
	//log.Debugf("[%v] contained microblocks are", r.ID())
	//for _, id := range proposal.HashList {
	//	log.Debugf("[%v] id: %x", r.ID(), id)
	//}
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

func (r *Replica) verifySigs(sigMap map[crypto.Identifier]map[identity.NodeID]crypto.Signature) bool {
	for mbID, sigs := range sigMap {
		for id, sig := range sigs {
			voteIsVerified, err := crypto.PubVerify(sig, crypto.IDToByte(mbID), id)
			if err != nil {
				log.Warningf("[%v] Error in verifying the signature in ack id: %x", r.ID(), mbID)
				return false
			}
			if !voteIsVerified {
				log.Warningf("[%v] received an ack with invalid signature. vote id: %x", r.ID(), mbID)
				return false
			}
		}
	}
	return true
}

//func (r *Replica) waitUntilStable(payload *blockchain.Payload) {
//	lastItem := payload.LastItem()
//	if lastItem == nil {
//		return
//	}
//	stableTime := r.estimator.PredictStableTime("p")
//	//stableTime := time.Duration(0)
//	wait := lastItem.FutureTimestamp.Sub(time.Now()) - stableTime
//	//log.Debugf("[%v] stable time for a proposal is %v", r.ID(), stableTime)
//	if stableTime < 0 {
//		log.Errorf("[%v] stable time for proposal is less than 0")
//	}
//	if wait > 0 {
//		log.Debugf("[%v] wait for %v until the contained microblocks are stable", r.ID(), wait)
//		time.Sleep(wait)
//	}
//}

func (r *Replica) GetCurrentLeader() identity.NodeID {
	return r.Election.FindLeaderFor(r.pm.GetCurView())
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
	//go r.gossip()
	go r.loadbalance()
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
