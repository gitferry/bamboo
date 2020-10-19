package bamboo

import (
	"encoding/gob"
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
	"github.com/gitferry/bamboo/pacemaker"
	"github.com/gitferry/bamboo/tchs"
	"github.com/gitferry/bamboo/types"
)

type Replica struct {
	Node
	election.Election
	Safety
	pd        *mempool.Producer
	bc        *blockchain.BlockChain
	pm        *pacemaker.Pacemaker
	start     chan bool
	isStarted atomic.Bool
	isByz     bool
	bElectNo  int
	totalView int
	timer     *time.Timer
	blockMsg  chan *blockchain.Block
	voteMsg   chan *blockchain.Vote
	qcMsg     chan *blockchain.QC
	timeouts  chan *pacemaker.TMO
	tcs       chan *pacemaker.TC
	newView   chan types.View
	mu        sync.Mutex
}

// NewReplica creates a new replica instance
func NewReplica(id identity.NodeID, alg string, isByz bool) *Replica {
	r := new(Replica)
	r.Node = NewNode(id, isByz)
	if isByz {
		log.Infof("[%v] is Byzantine", r.ID())
	}
	r.Election = election.NewRotation(config.GetConfig().N())
	bc := blockchain.NewBlockchain(config.GetConfig().N())
	r.bc = bc
	r.pd = mempool.NewProducer()
	r.pm = pacemaker.NewPacemaker(config.GetConfig().N())
	r.blockMsg = make(chan *blockchain.Block, 1)
	r.voteMsg = make(chan *blockchain.Vote, 1)
	r.qcMsg = make(chan *blockchain.QC, 1)
	r.timeouts = make(chan *pacemaker.TMO, 1)
	r.start = make(chan bool)
	r.isByz = isByz
	r.Register(blockchain.QC{}, r.HandleQC)
	r.Register(blockchain.Block{}, r.HandleBlock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	r.Register(pacemaker.TMO{}, r.HandleTmo)
	r.Register(message.Transaction{}, r.handleTxn)
	gob.Register(blockchain.Block{})
	gob.Register(blockchain.QC{})
	gob.Register(blockchain.Vote{})
	switch alg {
	case "hotstuff":
		forkchoice := "highest"
		if isByz {
			forkchoice = "forking"
		}
		r.Safety = hotstuff.NewHotStuff(bc, forkchoice)
	case "tchs":
		forkchoice := "highest"
		if isByz {
			forkchoice = "forking"
		}
		r.Safety = tchs.Newtchs(bc, forkchoice)
	default:
		r.Safety = hotstuff.NewHotStuff(bc, "default")
	}
	return r
}

/* Message Handlers */

func (r *Replica) HandleBlock(block blockchain.Block) {
	//log.Debugf("[%v] received a block from %v, view is %v", r.ID(), block.Proposer, block.View)
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

func (r *Replica) HandleTmo(tmo pacemaker.TMO) {
	log.Debugf("[%v] received a timeout from %v for view %v", r.ID(), tmo.NodeID, tmo.View)
	if tmo.View < r.pm.GetCurView() {
		return
	}
	r.timeouts <- &tmo
}

func (r *Replica) HandleTC(tc pacemaker.TC) {
	if tc.View < r.pm.GetCurView() {
		return
	}
	r.tcs <- &tc
}

func (r *Replica) HandleQC(qc blockchain.QC) {
	//log.Debugf("[%v] received a qc from, blockID is %x", r.ID(), qc.BlockID)
	if qc.View < r.pm.GetCurView() {
		return
	}
	r.qcMsg <- &qc
}

func (r *Replica) handleTxn(m message.Transaction) {
	r.pd.CollectTxn(&m)
	if !r.isStarted.Load() {
		log.Debugf("[%v] is boosting", r.ID())
		r.isStarted.Store(true)
		r.start <- true
		time.Sleep(500 * time.Millisecond)
	}

	//	kick-off the protocol
	if r.pm.GetCurView() == 0 && r.IsLeader(r.ID(), 1) {
		log.Debugf("[%v] is going to kick off the protocol", r.ID())
		r.pm.AdvanceView(0)
	}
}

/* Processors */

func (r *Replica) processBlock(block *blockchain.Block) {
	log.Debugf("[%v] is processing block, view: %v, id: %x", r.ID(), block.View, block.ID)
	// TODO: process TC
	// to simulate forking attack without a tc, create a qc with view set to block.view-1
	tc := &blockchain.QC{
		View:    block.View - 1,
		BlockID: block.QC.BlockID,
	}
	if r.ID().Node() == config.Configuration.N() {
		log.Infof("[%v] block view: %v", r.ID(), block.View)
	}
	r.processCertificate(tc)
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
	r.mu.Lock()
	r.bc.AddBlock(block)
	r.mu.Unlock()

	shouldVote, err := r.VotingRule(block)
	if err != nil {
		log.Errorf("cannot decide whether to vote the block, %w", err)
		return
	}
	if !shouldVote {
		log.Debugf("[%v] is not going to vote for block, id: %x", r.ID(), block.ID)
		return
	}
	log.Debugf("[%v] is going to vote for block, id: %x", r.ID(), block.ID)
	vote := blockchain.MakeVote(block.View, r.ID(), block.ID)
	err = r.UpdateStateByView(vote.View)
	if err != nil {
		log.Errorf("cannot update state after voting: %w", err)
	}
	// TODO: sign the vote
	time.Sleep(10 * time.Millisecond)
	// vote to the current leader
	voteAggregator := block.Proposer
	if voteAggregator == r.ID() {
		r.processVote(vote)
	} else {
		r.Send(voteAggregator, vote)
	}
}

func (r *Replica) preprocessQC(qc *blockchain.QC) {
	isThreeChain, _, err := r.Safety.CommitRule(qc)
	if err != nil {
		log.Warningf("[%v] cannot check commit rule", r.ID())
		return
	}
	if isThreeChain {
		r.pm.AdvanceView(qc.View)
		return
	}
	for i := qc.View; ; i++ {
		nextLeader := r.FindLeaderFor(i + 1)
		if !config.Configuration.IsByzantine(nextLeader) {
			qc.View = i
			log.Debugf("[%v] is going to send a stale qc to %v, view: %v, id: %x", r.ID(), nextLeader, qc.View, qc.BlockID)
			r.Send(nextLeader, qc)
			return
		}
	}
}

func (r *Replica) processCertificate(qc *blockchain.QC) {
	if qc.View < r.pm.GetCurView() {
		return
	}
	r.pm.AdvanceView(qc.View)
	r.bc.UpdateHighQC(qc)
	log.Debugf("[%v] has advanced to view %v", r.ID(), r.pm.GetCurView())
	r.UpdateStateByQC(qc)
	// TODO: send the qc to next leader
	//if !r.IsLeader(r.ID(), r.pm.GetCurView()) {
	//	go r.Send(r.FindLeaderFor(r.pm.GetCurView()), qc)
	//}
	if qc.View < 3 {
		return
	}
	ok, block, _ := r.CommitRule(qc)
	if !ok {
		return
	}
	r.mu.Lock()
	committedBlocks, err := r.bc.CommitBlock(block.ID)
	r.mu.Unlock()
	if err != nil {
		log.Errorf("[%v] cannot commit blocks", r.ID())
		return
	}
	r.processCommittedBlocks(committedBlocks)
}

func (r *Replica) processCommittedBlocks(blocks []*blockchain.Block) {
	for _, block := range blocks {
		if config.Configuration.IsByzantine(block.Proposer) {
			continue
		}
		for _, txn := range block.Payload {
			if r.ID() == txn.NodeID {
				txn.Reply(message.TransactionReply{})
			}
			//if r.ID() != block.Proposer { // txns are removed when being proposed
			//	r.pd.RemoveTxn(txn.ID)
			//}
		}
		r.pd.RemoveTxns(block.Payload)
		//delay := int(r.pm.GetCurView() - block.View)
		if r.ID().Node() == config.Configuration.N() {
			log.Infof("[%v] the block is committed, No. of transactions: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.Payload), block.View, r.pm.GetCurView(), block.ID)
		}
		//r.totalDelayRounds += int(r.pm.GetCurView() - block.View)
	}
	//	print measurement
	//if r.ID().Node() == config.Configuration.N() {
	//log.Warningf("[%v] Honest committed blocks: %v, total blocks: %v, chain growth: %v", r.ID(), r.bc.GetHonestCommittedBlocks(), r.bc.GetHighestComitted(), r.bc.GetChainGrowth())
	//log.Warningf("[%v] Honest committed blocks: %v, committed blocks: %v, chain quality: %v", r.ID(), r.bc.GetHonestCommittedBlocks(), r.bc.GetCommittedBlocks(), r.bc.GetChainQuality())
	//log.Warningf("[%v] Ave. delay is %v, total committed block number: %v", r.ID(), r.totalDelay.Seconds()/float64(r.bc.GetHonestCommittedBlocks()), r.bc.GetHonestCommittedBlocks())
	//}
}

func (r *Replica) processVote(vote *blockchain.Vote) {
	r.mu.Lock()
	isBuilt, qc := r.bc.AddVote(vote)
	r.mu.Unlock()
	if !isBuilt {
		return
	}
	// send the QC to the next leader
	log.Debugf("[%v] a qc is built, block id: %x", r.ID(), qc.BlockID)
	nextLeader := r.FindLeaderFor(qc.View + 1)
	if nextLeader == r.ID() {
		if config.Configuration.IsByzantine(nextLeader) {
			r.preprocessQC(qc)
		} else {
			r.processCertificate(qc)
		}
	} else {
		r.Send(nextLeader, qc)
	}
}

func (r *Replica) processTmoMsg(tmo *pacemaker.TMO) {
	r.bc.UpdateHighQC(tmo.HighQC)
	isBuilt, tc := r.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		return
	}
	nextLeader := r.FindLeaderFor(tc.View + 1)
	r.Send(nextLeader, tc)
}

func (r *Replica) processTC(tc *pacemaker.TC) {
	if tc.View <= r.pm.GetCurView() {
		return
	}
	r.pm.AdvanceView(tc.View)
}

func (r *Replica) processNewView(newView types.View) {
	log.Debugf("[%v] is processing new view: %v", r.ID(), newView)
	if !r.IsLeader(r.ID(), newView) {
		return
	}
	r.totalView = int(newView)
	if newView == 1 {
		r.proposeBlock(newView)
		return
	}
	if r.isByz {
		r.bElectNo++
		log.Warningf("[%v] the number of Byzantine election is %v, total election number is %v", r.ID(), r.bElectNo, r.totalView)
	}

	r.proposeBlock(newView)
}

func (r *Replica) processLocalTmo(view types.View) {
	log.Debugf("[%v] timeout for view %v", r.ID(), view)
}

func (r *Replica) proposeBlock(view types.View) {
	log.Infof("[%v] is trying to propose a block", r.ID())
	block := r.pd.ProduceBlock(view, r.Safety.Forkchoice(), r.ID())
	if len(block.Payload) == 0 {
		log.Debugf("[%v] is stalled because no txns left in the mempool", r.ID())
		return
	}
	log.Infof("[%v] is going to propose block for view: %v, id: %x, prevID: %x", r.ID(), view, block.ID, block.PrevID)
	time.Sleep(10 * time.Millisecond)
	r.Broadcast(block)
	r.processBlock(block)
}

// Start starts event loop
func (r *Replica) Start() {
	go r.Run()
	<-r.start
	for r.isStarted.Load() {
		r.timer = time.NewTimer(r.pm.GetTimerForView())
		log.Debugf("[%v] timer is reset", r.ID())
		for {
			select {
			case newView := <-r.pm.EnteringViewEvent():
				go r.processNewView(newView)
				break
			case newBlock := <-r.blockMsg:
				go r.processBlock(newBlock)
			case newVote := <-r.voteMsg:
				go r.processVote(newVote)
			case newTimeout := <-r.timeouts:
				go r.processTmoMsg(newTimeout)
			case <-r.timer.C:
				go r.processLocalTmo(r.pm.GetCurView())
			case newTC := <-r.tcs:
				go r.processTC(newTC)
			case newQC := <-r.qcMsg:
				if r.isByz {
					go r.preprocessQC(newQC)
				} else {
					go r.processCertificate(newQC)
				}
			}
		}
	}
}
