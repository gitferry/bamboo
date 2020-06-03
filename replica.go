package zeitgeber

import (
	"fmt"
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
	wishCounts map[int]int
	me         int // this peer's index into peers[]
	threshold  int
	dead       int32 // set by Kill()
	timer      *time.Timer
	timout     time.Duration
	wakeup     chan bool
	viewSync   chan int

	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state

	votedFor int  // candidateID that received vote in current term (or nil if none)
	isByz    bool // true if the replica is Byzantine
}

// return currentTerm and whether this server
// believes it is the leader.
func (r *Replica) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A). Done.

	return term, isleader
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

func (r *Replica) onTimeout() {
	r.sendWish(r.curView + 1)
}

func (r *Replica) sendWish(view int) {
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
	} else {
		reply.Success = false
	}
	reply.VerifiedView = r.curView
}

// HandleWish handles Wish msg
func (r *Replica) HandleWish(args *WishArgs, reply *WishReply) {
	DPrintf("[%d] received Wish msg from %d", r.me, args.NodeID)
	r.mu.Lock()
	defer r.mu.Unlock()
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
		if r.wishCounts[args.WishView] >= r.threshold {
			r.curView = args.WishView
			r.viewSync <- r.curView
			DPrintf("[%d] collect enough Wish msg for view %d", r.me, args.WishView)
		}
		reply.Success = true
	}
	DPrintf("[%d] Wish msg from %d for view %d is stale", r.me, args.WishView)
	reply.VerifiedView = r.curView
	reply.Success = false
}

// CallViewSync issues ViewSync RPC
func (r *Replica) CallViewSync(nodeID int, view int) (bool, *ViewSyncReply) {
	args := ViewSyncArgs{
		VerifiedView: r.curView,
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
				fmt.Errorf("network error")
			}
			wg.Done()
		}(index)
		wg.Wait()
	}
}

func (r *Replica) sendViewSync(nodeID int, args *ViewSyncArgs, reply *ViewSyncReply) bool {
	ok := r.peers[nodeID].Call("Zeitgeber.HandleViewSync", args, reply)
	return ok
}

func Init() *Replica {
	timer := time.NewTimer(2 * time.Second)
	return &Replica{
		timer: timer,
	}
}

func (r *Replica) Run() {
	r.resettimer()
	for !r.killed() {
		select {
		case <-r.timer.C:
			r.onTimeout()
		case <-r.wakeup:
			r.resettimer()
		case view := <-r.viewSync:
			r.ViewSync(view)
		}
	}
}

func (r *Replica) resettimer() {
	r.timer.Reset(r.timout)
}

//
// restore previously persisted state.
//
func (rf *Replica) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg, isByz bool) *Replica {
	rf := &Replica{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me
	rf.curView = 0
	rf.isByz = isByz

	go rf.Run()

	rf.readPersist(persister.ReadRaftState())

	return rf
}
