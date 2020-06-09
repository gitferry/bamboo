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
	mu         sync.Mutex // Lock to protect shared access to this peer's state
	curView    int
	me         int // this peer's index into peers[]
	threshold  int
	dead       int32 // set by Kill()
	isByz      bool  // true if the replica is Byzantine
	state      State
	viewSync   chan int
	timeout    time.Duration
	wishCounts map[int]int
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
		nextView := r.curView + 1
		r.mu.Unlock()
		if r.IsSelfLeader(nextView) {
			DPrintf("[%d] is the leader for the next view %d", r.me, nextView)
			r.viewSync <- nextView
		}
		r.mu.Lock()
	}
	reply.Success = true
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
		reply.VerifiedView = r.curView
		r.viewSync <- r.curView
		r.mu.Unlock()
		return
	}
	if args.WishView > r.curView {
		r.wishCounts[args.WishView]++
		DPrintf("[%d] has %d Wish msg for view %d", r.me, r.wishCounts[args.WishView], args.WishView)
		r.mu.Unlock()
		r.checkAndSync(args.WishView)
		reply.Success = true
		return
	}
	DPrintf("[%d] Wish msg from %d for view %d is stale", r.me, args.NodeID, args.WishView)
	reply.VerifiedView = r.curView
	r.mu.Unlock()
	reply.Success = true
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
	if !r.IsSelfLeader(view) || r.isByz {
		if r.isByz {
			DPrintf("[%d] is Byzantine, abort sending ViewSync", r.me)
		}
		return
	}
	var wg sync.WaitGroup
	DPrintf("[%d] has %d peers", r.me, len(r.peers))
	for index := 0; index < len(r.peers); index++ {
		if index == r.me {
			continue
		}
		wg.Add(1)
		go func(nodeID int) {
			DPrintf("[%d] is going to send ViewSync for view %d to %d", r.me, view, nodeID)
			ok, _ := r.CallViewSync(nodeID, view)
			if ok {
				// TODO: process reply
			} else {
				DPrintf("[%d] did not get ViewSync reply from %d", r.me, nodeID)
			}
			wg.Done()
		}(index)
	}
	wg.Wait()
}

// WishView sends RPCs to the rest of the replicas
func (r *Replica) WishView() {
	if r.isByz {
		DPrintf("[%d] is Byzantine, abort sending Wish", r.me)
		return
	}
	r.mu.Lock()
	view := r.curView + 1
	r.mu.Unlock()
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
		r.mu.Unlock()
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
	time.Sleep(20 * time.Millisecond)
	ok := r.peers[nodeID].Call("Replica.HandleViewSync", args, reply)
	return ok
}

func (r *Replica) sendWish(nodeID int, args *WishArgs, reply *WishReply) bool {
	time.Sleep(100 * time.Millisecond)
	ok := r.peers[nodeID].Call("Replica.HandleWish", args, reply)
	return ok
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
		r.mu.Unlock()
		return false
	}
	isLeader := view%len(r.peers) == r.me
	return isLeader
}

// FindLeaderForView returns the leader ID for the given view
func (r *Replica) FindLeaderForView(view int) int {
	return view % len(r.peers)
}

// Run kicks off the protocol
func (r *Replica) Run() {
	// the leader kicks off
	if r.IsSelfLeader(1) {
		r.curView = 1
		r.ViewSync(r.curView)
	} else {
		r.state = REPLICA
	}
	for !r.killed() {
		select {
		case <-time.After(r.timeout):
			DPrintf("[%d] timeout", r.me)
			r.WishView()
			break
		case view := <-r.viewSync:
			r.ViewSync(view)
			break
		}
	}
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
func Make(peers []*labrpc.ClientEnd, me int, isByz bool, threshold int) *Replica {
	r := &Replica{}
	r.peers = peers
	r.me = me
	r.curView = 0
	r.isByz = isByz
	r.threshold = threshold
	r.timeout = 2 * time.Second
	r.wishCounts = make(map[int]int)
	r.viewSync = make(chan int)
	go r.Run()

	return r
}
