package bamboo

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
	eventChan chan interface{}

	hasher crypto.Hasher
	signer string
	//privateKey *crypto.PrivateKey
	//publicKeys []crypto.PublicKey
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
	r.eventChan = make(chan interface{}, 1)
	r.hasher, _ = crypto.NewHasher(config.GetConfig().GetHashScheme())
	r.signer = config.GetConfig().GetSignatureScheme()
	//r.privateKey, r.publicKeys = config.GetConfig().GetKeys(id.Node())
	r.start = make(chan bool)
	r.isByz = isByz
	r.Register(blockchain.QC{}, r.HandleQC)
	r.Register(blockchain.Block{}, r.HandleBlock)
	r.Register(blockchain.Vote{}, r.HandleVote)
	r.Register(pacemaker.TMO{}, r.HandleTmo)
	r.Register(pacemaker.TC{}, r.HandleTC)
	r.Register(message.Transaction{}, r.handleTxn)
	gob.Register(blockchain.Block{})
	gob.Register(blockchain.QC{})
	gob.Register(blockchain.Vote{})
	gob.Register(pacemaker.TC{})
	gob.Register(pacemaker.TMO{})
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
	log.Debugf("[%v] received a block from %v, view is %v", r.ID(), block.Proposer, block.View)
	//if block.View < r.pm.GetCurView() {
	//	return
	//}
	r.eventChan <- block
}

func (r *Replica) HandleVote(vote blockchain.Vote) {
	log.Debugf("[%v] received a vote from %v, blockID is %x", r.ID(), vote.Voter, vote.BlockID)
	//if vote.View < r.pm.GetCurView() {
	//	return
	//}
	r.eventChan <- vote
}

func (r *Replica) HandleTmo(tmo pacemaker.TMO) {
	log.Debugf("[%v] received a timeout from %v for view %v", r.ID(), tmo.NodeID, tmo.View)
	//if tmo.View < r.pm.GetCurView() {
	//	return
	//}
	r.eventChan <- tmo
}

func (r *Replica) HandleTC(tc pacemaker.TC) {
	//if tc.View < r.pm.GetCurView() {
	//	return
	//}
	r.eventChan <- tc
}

func (r *Replica) HandleQC(qc blockchain.QC) {
	log.Debugf("[%v] received a qc, blockID is %x", r.ID(), qc.BlockID)
	//if qc.View < r.pm.GetCurView() {
	//	return
	//}
	r.eventChan <- qc
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

func (r *Replica) processBlock(block *blockchain.Block) {
	log.Debugf("[%v] is processing block, view: %v, id: %x", r.ID(), block.View, block.ID)
	blockIsVerified, _ := crypto.PubVerify(block.Sig, crypto.IDToByte(block.ID), block.Proposer)
	//blockIsVerified, _ := r.publicKeys[block.Proposer.Node()].Verify(block.Sig, nil)
	if blockIsVerified == false {
		log.Warningf("[%v] received a block with an invalid signature", r.ID())
	}
	r.processCertificate(block.QC)
	// TODO: should uncomment the following checks
	//curView := r.pm.GetCurView()
	//if block.View != curView {
	//	log.Warningf("[%v] received a stale proposal from %v", r.ID(), block.Proposer)
	//	return
	//}
	if !r.Election.IsLeader(block.Proposer, block.View) {
		log.Warningf(
			"[%v] received a proposal (%v) from an invalid leader (%v)",
			r.ID(), block.View, block.Proposer)
		return
	}
	r.bc.AddBlock(block)

	//shouldVote, err := r.VotingRule(block)
	// TODO: add block buffer
	//if err != nil {
	//	log.Errorf("cannot decide whether to vote the block, %w", err)
	//	return
	//}
	//if !shouldVote {
	//	log.Debugf("[%v] is not going to vote for block, id: %x", r.ID(), block.ID)
	//	return
	//}
	log.Debugf("[%v] is going to vote for block, id: %x", r.ID(), block.ID)
	vote := blockchain.MakeVote(block.View, r.ID(), block.ID)
	//err = r.UpdateStateByView(vote.View)
	//if err != nil {
	//	log.Errorf("cannot update state after voting: %w", err)
	//}
	// TODO: sign the vote -----> I've signed it in blockchain.MakeVote
	// vote to the current leader
	voteAggregator := block.Proposer
	if voteAggregator == r.ID() {
		r.processVote(vote)
	} else {
		r.Send(voteAggregator, vote)
	}
	//log.Debugf("[%v] voted for %v", r.ID(), voteAggregator)
}

func (r *Replica) preprocessQC(qc *blockchain.QC) {
	isThreeChain, _, err := r.Safety.CommitRule(qc)
	if err != nil {
		log.Warningf("[%v] cannot check commit rule", r.ID())
		return
	}
	if isThreeChain {
		go r.pm.AdvanceView(qc.View)
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

//func (r *Replica) verifyQuorumSignature(qc *blockchain.QC) (bool, error) {
//	return true, nil
//}

func (r *Replica) processCertificate(qc *blockchain.QC) {
	if qc.View < r.pm.GetCurView() {
		return
	}
	// ALI: Here is where you can use the crypto package to verify the signatures
	// of a QC.
	//quorumIsVerified, _ := crypto.verifyQuorumSignature()
	//if quorumIsVerified == false {
	//	log.Warningf("[%v] received a quorum with unvalid signatures", r.ID())
	//	return
	//}
	go r.pm.AdvanceView(qc.View)
	r.bc.UpdateHighQC(qc)
	log.Debugf("[%v] has advanced to view %v", r.ID(), r.pm.GetCurView())
	r.UpdateStateByQC(qc)
	log.Debugf("[%v] has updated state by qc: %v", r.ID(), qc.View)
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
	committedBlocks, err := r.bc.CommitBlock(block.ID)
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
		}
		r.pd.RemoveTxns(block.Payload)
		//delay := int(r.pm.GetCurView() - block.View)
		//if r.ID().Node() == config.Configuration.N() {
		log.Infof("[%v] the block is committed, No. of transactions: %v, view: %v, current view: %v, id: %x", r.ID(), len(block.Payload), block.View, r.pm.GetCurView(), block.ID)
		//}
	}
	//	print measurement
	//if r.ID().Node() == config.Configuration.N() {
	//log.Warningf("[%v] Honest committed blocks: %v, total blocks: %v, chain growth: %v", r.ID(), r.bc.GetHonestCommittedBlocks(), r.bc.GetHighestComitted(), r.bc.GetChainGrowth())
	//log.Warningf("[%v] Honest committed blocks: %v, committed blocks: %v, chain quality: %v", r.ID(), r.bc.GetHonestCommittedBlocks(), r.bc.GetCommittedBlocks(), r.bc.GetChainQuality())
	//log.Warningf("[%v] Ave. delay is %v, total committed block number: %v", r.ID(), r.totalDelay.Seconds()/float64(r.bc.GetHonestCommittedBlocks()), r.bc.GetHonestCommittedBlocks())
	//}
}

//func (r *Replica) verifyVoteSignature(vote *blockchain.Vote) (bool, error) {
//	return r.publicKeys[vote.Voter.Node()].Verify(vote.Signature, crypto.IDToByte(vote.BlockID))
//}

func (r *Replica) processVote(vote *blockchain.Vote) {
	voteIsVerified, err := crypto.PubVerify(vote.Signature, crypto.IDToByte(vote.BlockID), vote.Voter)
	if err != nil {
		log.Fatalf("[%v] Error in verifying the signature in vote id: %x", r.ID(), vote.BlockID)
		return
	}
	if voteIsVerified == false {
		log.Warningf("[%v] received a vote with unvalid signature. vote id: %x", r.ID(), vote.BlockID)
		return
	}
	isBuilt, qc := r.bc.AddVote(vote)
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
	log.Debugf("[%v] is processing tmo from %v", r.ID(), tmo.NodeID)
	r.bc.UpdateHighQC(tmo.HighQC)
	isBuilt, tc := r.pm.ProcessRemoteTmo(tmo)
	if !isBuilt {
		log.Debugf("[%v] not enough tc for %v", r.ID(), tmo.View)
		return
	}
	log.Debugf("[%v] a tc is built for view %v", r.ID(), tc.View)
	r.processTC(tc)
	//nextLeader := r.FindLeaderFor(tc.View + 1)
	//if nextLeader != r.ID() {
	//	r.Send(nextLeader, tc)
	//}
}

func (r *Replica) processTC(tc *pacemaker.TC) {
	if tc.View < r.pm.GetCurView() {
		return
	}
	r.pm.UpdateTC(tc)
	go r.pm.AdvanceView(tc.View)
}

func (r *Replica) processNewView(newView types.View) {
	log.Debugf("[%v] is processing new view: %v", r.ID(), newView)
	if !r.IsLeader(r.ID(), newView) {
		return
	}

	r.proposeBlock(newView)
}

func (r *Replica) processLocalTmo() {
	view := r.pm.GetCurView()
	log.Debugf("[%v] timeout for view %v", r.ID(), view+1)
	//	TODO: send tmo msg
	tmo := &pacemaker.TMO{
		View:   view + 1,
		NodeID: r.ID(),
		HighQC: r.bc.GetHighQC(),
	}
	r.Broadcast(tmo)
	r.processTmoMsg(tmo)
	log.Debugf("[%v] broadcast is done for sending tmo", r.ID())
}

func (r *Replica) proposeBlock(view types.View) {
	log.Infof("[%v] is trying to propose a block", r.ID())
	block := r.pd.ProduceBlock(view, r.Safety.Forkchoice(), r.ID())
	tc := r.pm.GetHighTC()
	if tc != nil && tc.View > block.QC.View {
		block.QC.View = tc.View
	}
	//if len(block.Payload) == 0 {
	//	log.Debugf("[%v] is stalled because no txns left in the mempool", r.ID())
	//	return
	//}
	log.Infof("[%v] is going to propose block for view: %v, id: %x, prevID: %x", r.ID(), view, block.ID, block.PrevID)
	r.processBlock(block)
	r.Broadcast(block)
	log.Debugf("[%v] broadcast is done for sending the block", r.ID())
}

func (r *Replica) LocalListen() {
	for {
		r.timer = time.NewTimer(r.pm.GetTimerForView())
	L:
		for {
			select {
			case view := <-r.pm.EnteringViewEvent():
				r.eventChan <- view
				break L
			case timeout := <-r.timer.C:
				r.eventChan <- timeout
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
	go r.LocalListen()
	for r.isStarted.Load() {
		event := <-r.eventChan
		switch v := event.(type) {
		case types.View:
			r.processNewView(v)
		case blockchain.Block:
			r.processBlock(&v)
		case blockchain.Vote:
			r.processVote(&v)
		case pacemaker.TMO:
			r.processTmoMsg(&v)
		case time.Time:
			r.processLocalTmo()
		case pacemaker.TC:
			r.processTC(&v)
		case blockchain.QC:
			r.processCertificate(&v)
		default:
			log.Debugf("[%v] unknown event %v", r.ID(), v)
		}
	}
}
