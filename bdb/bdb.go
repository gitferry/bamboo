package bdb

import (
	"sync"

	"github.com/gitferry/zeitgeber/log"

	"github.com/gitferry/zeitgeber"
)

type Bcd struct {
	zeitgeber.Node
	zeitgeber.Election

	curView zeitgeber.View
	quorum  *zeitgeber.Quorum

	mu sync.Mutex
}

func NewBdb(n zeitgeber.Node, election zeitgeber.Election) *Bcd {
	bdb := new(Bcd)
	bdb.Node = n
	bdb.Election = election
	bdb.quorum = zeitgeber.NewQuorum()
	bdb.Register(TCMsg{}, bdb.HandleTC)
	bdb.Register(TmoMsg{}, bdb.HandleTmo)

	return bdb
}

func (b *Bcd) HandleTmo(wish zeitgeber.WishMsg) {
	b.mu.Lock()
	if wish.View <= b.curView {
		log.Warningf("[%s] received with msg with view %d lower thant the current view %d", b.ID(), wish.View, b.curView)
		b.mu.Unlock()
		return
	}
	b.mu.Unlock()
	// store the wish
	b.quorum.ACK(wish.View, wish.NodeID)
	if !b.quorum.SuperMajority(wish.View) {
		return
	}
	// TC is generated
	log.Infof("[%s] a time certificate for view %d is generated", b.ID(), wish.View)
	b.NewView(wish.View)
}

func (b *Bcd) HandleTC(tc *TCMsg) {
	log.Infof("[%s] received tc from %d for view %v", b.ID(), tc.NodeID, tc.View)
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
func (b *Bcd) TimeoutFor(view zeitgeber.View) {
	tmoMsg := TmoMsg{
		View:   view,
		NodeID: b.ID(),
	}
	b.Broadcast(tmoMsg)
}

func (b *Bcd) NewView(view zeitgeber.View) {
	b.mu.Lock()
	if view < b.curView {
		log.Warningf("the view %d is lower than current view %d", view, b.curView)
	}
	// TODO: stop local timer for the view
	b.curView = view + 1
	b.mu.Unlock()
	curView := view + 1
	if !b.IsLeader(b.ID(), curView) {
		log.Warningf("[%s] is not the leader of view %v", b.ID(), curView)
		b.Send(b.FindLeaderFor(curView), TCMsg{View: curView})
		return
	}
	// TODO: start timer for the view + 1
	// TODO: change this part into pub-sub pattern
	proposal := &zeitgeber.ProposalMsg{
		NodeID:   b.ID(),
		View:     curView,
		TimeCert: zeitgeber.NewTC(view),
	}
	log.Infof("[%s] is proposing for view %v", b.ID(), curView)
	b.Broadcast(proposal)
}

func (b *Bcd) GetCurView() zeitgeber.View {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.curView
}
