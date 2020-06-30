package bcb

import (
	"sync"

	"github.com/gitferry/zeitgeber/log"

	"github.com/gitferry/zeitgeber"
)

type bcb struct {
	zeitgeber.Node
	zeitgeber.Election

	curView zeitgeber.View
	quorum  *zeitgeber.Quorum

	Reset chan bool

	mu sync.Mutex
}

func NewBcb(n zeitgeber.Node, election zeitgeber.Election) *bcb {
	bcb := new(bcb)
	bcb.Node = n
	bcb.Election = election
	bcb.Reset = make(chan bool)
	bcb.quorum = zeitgeber.NewQuorum()
	bcb.Register(TCMsg{}, bcb.HandleTC)
	bcb.Register(TmoMsg{}, bcb.HandleTmo)

	return bcb
}

func (b *bcb) HandleTmo(tmo TmoMsg) {
	b.mu.Lock()
	if tmo.View <= b.curView {
		log.Warningf("[%s] received with msg with view %d lower thant the current view %d", b.ID(), tmo.View, b.curView)
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()
	// store the wish
	b.quorum.ACK(tmo.View, tmo.NodeID)
	if !b.quorum.SuperMajority(tmo.View) {
		return
	}
	// TC is generated
	log.Infof("[%s] a time certificate for view %d is generated", b.ID(), tmo.View)
	b.NewView(tmo.View)
}

func (b *bcb) HandleTCMsg(tc *TCMsg) {
	log.Infof("[%s] received tc from %d for view %v", b.ID(), tc.NodeID, tc.View)
	b.HandleTC(zeitgeber.NewTC(tc.View))
}

func (b *bcb) HandleTC(tc *zeitgeber.TC) {
	b.mu.Lock()
	if tc.View < b.curView {
		log.Warningf("[%s] received tc's view %v is lower than current view %v", b.ID(), tc.View, b.curView)
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()
	b.NewView(tc.View)
}

// WishAdvance broadcasts the wish msg for the view when it timeouts
func (b *bcb) TimeoutFor(view zeitgeber.View) {
	tmoMsg := TmoMsg{
		View:   view,
		NodeID: b.ID(),
	}
	b.Broadcast(tmoMsg)
}

func (b *bcb) NewView(view zeitgeber.View) {
	b.mu.Lock()
	if view < b.curView {
		log.Warningf("the view %d is lower than current view %d", view, b.curView)
	}
	b.Reset <- true // reset timer for the next view
	b.curView = view + 1
	b.mu.Unlock()
	curView := view + 1
	if !b.IsLeader(b.ID(), curView) {
		log.Warningf("[%s] is not the leader of view %v", b.ID(), curView)
		b.Send(b.FindLeaderFor(curView), TCMsg{View: curView})
		return
	}
	proposal := &zeitgeber.ProposalMsg{
		NodeID:   b.ID(),
		View:     curView,
		TimeCert: zeitgeber.NewTC(view),
	}
	log.Infof("[%s] is proposing for view %v", b.ID(), curView)
	if b.IsByz() {
		b.MulticastQuorum(zeitgeber.GetConfig().ByzNo, proposal)
		return
	}
	b.Broadcast(proposal)
}

func (b *bcb) ResetTimer() chan bool {
	return b.Reset
}

func (b *bcb) GetCurView() zeitgeber.View {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.curView
}
