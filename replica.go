package zeitgeber

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gitferry/zeitgeber/labrpc"
)

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type Replica struct {
	curView    int
	me         int // this peer's index into peers[]
	threshold  int
	dead       int32 // set by Kill()
	isByz      bool  // true if the replica is Byzantine
	state      State
	wakeup     chan bool
	viewSync   chan int
	timer      *time.Timer
	timout     time.Duration
	mu         sync.Mutex // Lock to protect shared access to this peer's state
	wishCounts map[int]int
	persister  *Persister          // Object to hold this peer's persisted state
	peers      []*labrpc.ClientEnd // RPC end points of all peers
}

// State indicates the curretn state of the replica
type State string

const (
	REPLICA = "replica"
	LEADER  = "leader"
)

// GetState returns currentTerm and whether this server
// believes it is the leader.
func (r *Replica) GetState() (int, bool) {

	var view int
	var isleader bool

	r.mu.Lock()
	view = r.curView
	isleader = r.state == LEADER
	defer r.mu.Unlock()

	return view, isleader
}

// Kill kills a replica instance
func (r *Replica) Kill() {
	atomic.StoreInt32(&r.dead, 1)
	// Your code here, if desired.
}

func (r *Replica) killed() bool {
	z := atomic.LoadInt32(&r.dead)
	return z == 1
}

// HandleViewSync handles ViewSync msg
func (r *Replica) HandleViewSync(args *ViewSyncArgs, reply *ViewSyncReply) {
	DPrintf("[%d] received ViewSync msg from %d", r.me, args.NodeID)
	r.mu.Lock()
	defer r.mu.Unlock()
	if args.VerifiedView > r.curView {
		r.curView = args.VerifiedView
		reply.Success = true
		// wake up sleep gorutine to reset timer
		r.wakeup <- true
		r.viewSync <- r.curView
	} else {
		reply.Success = false
	}
	reply.VerifiedView = r.curView
}

// HandleWish handles Wish msg
func (r *Replica) HandleWish(args *WishArgs, reply *WishReply) {
	DPrintf("[%d] received Wish msg from %d", r.me, args.NodeID)
	r.mu.Lock()
	if args.VerifiedView > r.curView {
		DPrintf("[%d] update curView from %d to %d via Wish msg from %d", r.me, r.curView, args.VerifiedView, args.NodeID)
		r.curView = args.VerifiedView
		reply.Success = true
		r.wakeup <- true
		reply.VerifiedView = r.curView
		return
	}
	if args.WishView > r.curView {
		r.wishCounts[args.WishView]++
		DPrintf("[%d] has %d Wish msg for view %d", r.me, r.wishCounts[args.WishView], args.WishView)
		r.mu.Unlock()
		r.checkAndSync(args.WishView)
		reply.Success = true
	}
	r.mu.Unlock()
	DPrintf("[%d] Wish msg from %d for view %d is stale", r.me, args.NodeID, args.WishView)
	reply.VerifiedView = r.curView
	reply.Success = false
}

// CallWish issues Wish RPC
func (r *Replica) CallWish(nodeID int, view int) (bool, *WishReply) {
	r.mu.Lock()
	cv := r.curView
	r.mu.Unlock()
	args := WishArgs{
		WishView:     view,
		VerifiedView: cv,
		NodeID:       r.me,
	}
	var reply WishReply
	ok := r.sendWish(nodeID, &args, &reply)
	if ok {
		return true, &reply
	}
	return false, nil
}

// CallViewSync issues ViewSync RPC
func (r *Replica) CallViewSync(nodeID int, view int) (bool, *ViewSyncReply) {
	args := ViewSyncArgs{
		VerifiedView: view,
		NodeID:       r.me,
	}
	var reply ViewSyncReply
	ok := r.sendViewSync(nodeID, &args, &reply)
	if ok {
		return true, &reply
	}
	return false, nil
}

// ViewSync sends RPCs to the rest of the replicas
func (r *Replica) ViewSync(view int) {
	var wg sync.WaitGroup
	if !r.IsSelfLeader(view) {
		return
	}
	DPrintf("[%d] is going to send ViewSync msg for view %d", r.me, view)
	for index := 0; index < len(r.peers); index++ {
		if index == r.me {
			continue
		}
		wg.Add(1)
		go func(nodeID int) {
			ok, _ := r.CallViewSync(nodeID, view)
			if ok {
				// TODO: process reply
			} else {
				DPrintf("[%d] did not get ViewSync reply from %d", r.me, nodeID)
			}
			wg.Done()
		}(index)
		wg.Wait()
	}
}

// WishView sends RPCs to the rest of the replicas
func (r *Replica) WishView(view int) {
	var wg sync.WaitGroup
	for index := 0; index < len(r.peers); index++ {
		if index == r.me {
			r.wishCounts[view]++
			r.checkAndSync(view)
			DPrintf("[%d] wishes to enter view %d", r.me, view)
			continue
		}
		wg.Add(1)
		go func(nodeID int) {
			ok, _ := r.CallWish(nodeID, view)
			if ok {
				// TODO: process reply
			} else {
				DPrintf("[%d] did not get Wish reply from %d", r.me, nodeID)
			}
			wg.Done()
		}(index)
		wg.Wait()
	}
}

func (r *Replica) checkAndSync(view int) {
	r.mu.Lock()
	if r.wishCounts[view] < r.threshold {
		return
	}
	r.curView = view
	r.mu.Unlock()
	// send ViewSync to the leader of the view
	DPrintf("[%d] collect enough Wish msg for view %d", r.me, view)
	nextLeader := r.FindLeaderForView(view)
	r.CallViewSync(nextLeader, view)
	DPrintf("[%d] is going to send ViewSync msg to the leader %d for view %d", r.me, nextLeader, view)
}

func (r *Replica) sendViewSync(nodeID int, args *ViewSyncArgs, reply *ViewSyncReply) bool {
	if r.isByz {
		DPrintf("[%d] is Byzantine, abort sending ViewSync", r.me)
		return false
	}
	ok := r.peers[nodeID].Call("Zeitgeber.HandleViewSync", args, reply)
	return ok
}

func (r *Replica) sendWish(nodeID int, args *WishArgs, reply *WishReply) bool {
	if r.isByz {
		DPrintf("[%d] is Byzantine, abort sending Wish", r.me)
		return false
	}
	ok := r.peers[nodeID].Call("Zeitgeber.HandleWish", args, reply)
	return ok
}

// Run kicks off the protocol
func (r *Replica) Run() {
	r.ResetTimer()
	for !r.killed() {
		select {
		case <-r.timer.C:
			r.mu.Lock()
			view := r.curView + 1
			r.mu.Unlock()
			r.WishView(view)
		case <-r.wakeup:
			r.ResetTimer()
		case view := <-r.viewSync:
			r.ViewSync(view)
		}
	}
}

// ResetTimer resets the timer. There could be more
// advanced strategies to adjust the timer according
// to the network status
func (r *Replica) ResetTimer() {
	r.timer.Reset(r.timout)
}

// restore previously persisted state.
func (r *Replica) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}

// IsSelfLeader checks if the replica is leader based on current view
func (r *Replica) IsSelfLeader(view int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if view < r.curView {
		return false
	}
	return view%len(r.peers) == r.me
}

// FindLeaderForView returns the leader ID for the given view
func (r *Replica) FindLeaderForView(view int) int {
	return view % len(r.peers)
}

// Make makes an instance of the Replica.
// the service or tester wants to create a replica server. the ports
// of all the replica servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects replica to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg, isByz bool, threshold int) *Replica {
	r := &Replica{}
	r.peers = peers
	r.persister = persister
	r.me = me
	r.curView = 0
	r.isByz = isByz
	r.threshold = threshold
	r.timer = time.NewTimer(2 * time.Second)
	if r.IsSelfLeader(r.curView) {
		r.state = LEADER
		r.curView++
		go r.ViewSync(r.curView)
	} else {
		r.state = REPLICA
	}

	go r.Run()

	r.readPersist(persister.ReadRaftState())

	return r
}
