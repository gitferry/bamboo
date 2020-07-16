package zeitgeber

import (
	"sync"
	"time"

	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/log"
)

type Pacemaker interface {
	NewView(view View)
	TimeoutFor(view View)
	ProcessQC(qc *blockchain.QC)
	ProcessTmo(tmo *TMO)
	GetCurView() View
	GetHighCert() *TC
	EnteringViewEvent() chan View
}

type pacemaker struct {
	Node
	Election

	curView      View
	quorum       *Quorum
	highCert     *TC
	lastViewTime time.Time
	viewDuration map[View]time.Duration

	newViewChan chan View

	mu sync.Mutex
}

func NewBcb(n Node, election Election) Pacemaker {
	bcb := new(pacemaker)
	bcb.Node = n
	bcb.Election = election
	bcb.newViewChan = make(chan View)
	bcb.quorum = NewQuorum()
	bcb.Register(TCMsg{}, bcb.HandleTC)
	bcb.Register(TmoMsg{}, bcb.HandleTmo)
	bcb.highCert = NewTC(0)
	bcb.lastViewTime = time.Now()
	bcb.viewDuration = make(map[View]time.Duration)
	return bcb
}

func (b *pacemaker) HandleTmo(tmo TmoMsg) {
	b.mu.Lock()
	if tmo.View < b.curView {
		//log.Warningf("[%v] received timeout msg with view %v lower than the current view %v", b.NodeID(), tmo.View, b.curView)
		b.mu.Unlock()
		return
	}
	b.quorum.ACK(tmo.View, tmo.NodeID)
	if b.quorum.SuperMajority(tmo.View) {
		//log.Infof("[%v] a time certificate for view %v is generated", b.NodeID(), tmo.View)
		b.Send(b.FindLeaderFor(tmo.View), TCMsg{View: tmo.View})
		b.mu.Unlock()
		b.NewView(tmo.View)
		return
	}
	if tmo.HighTC.View >= b.curView {
		b.mu.Unlock()
		b.NewView(tmo.HighTC.View)
		return
	}
	b.mu.Unlock()
}

func (b *pacemaker) HandleTC(tc TCMsg) {
	//log.Infof("[%v] is processing tc for view %v", b.NodeID(), tc.View)
	b.mu.Lock()
	if tc.View < b.curView {
		//log.Warningf("[%s] received tc's view %v is lower than current view %v", b.NodeID(), tc.View, b.curView)
		b.mu.Unlock()
		return
	}
	if tc.View > b.highCert.View {
		b.highCert = NewTC(tc.View)
	}
	b.mu.Unlock()
	b.NewView(tc.View)
}

// TimeoutFor broadcasts the timeout msg for the view when it timeouts
func (b *pacemaker) TimeoutFor(view View) {
	tmoMsg := TmoMsg{
		View:   view,
		NodeID: b.ID(),
		HighTC: NewTC(view - 1),
	}
	//log.Debugf("[%s] is timeout for view %v", b.NodeID(), view)
	if b.IsByz() {
		b.MulticastQuorum(GetConfig().ByzNo, tmoMsg)
		return
	}
	b.Broadcast(tmoMsg)
	b.HandleTmo(tmoMsg)
}

func (b *pacemaker) NewView(view View) {
	b.mu.Lock()
	if view < b.curView {
		//log.Warningf("the view %v is lower than current view %v", view, b.curView)
		b.mu.Unlock()
		return
	}
	b.viewDuration[b.curView] = time.Now().Sub(b.lastViewTime)
	b.curView = view + 1
	b.lastViewTime = time.Now()
	b.mu.Unlock()
	if view == 100 {
		b.printViewTime()
	}
	b.newViewChan <- view + 1 // reset timer for the next view
}

func (b *pacemaker) printViewTime() {
	//log.Infof("[%v] is printing view duration", b.NodeID())
	for view, duration := range b.viewDuration {
		log.Infof("view %v duration: %v seconds", view, duration.Seconds())
	}
}

func (b *pacemaker) EnteringViewEvent() chan View {
	return b.newViewChan
}

func (b *pacemaker) GetCurView() View {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.curView
}

func (b *pacemaker) GetHighCert() *TC {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.highCert
}
