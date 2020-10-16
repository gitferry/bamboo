package pacemaker

import (
	"sync"

	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/types"
)

type Pacemaker struct {
	curView           types.View
	newViewChan       chan types.View
	timeoutController *TimeoutController
	mu                sync.Mutex
}

func NewPacemaker() *Pacemaker {
	pm := new(Pacemaker)
	pm.newViewChan = make(chan types.View)
	pm.timeoutController = NewTimeoutController()
	//bcb.Node = n
	//bcb.Election = election
	//bcb.newViewChan = make(chan View)
	//bcb.quorum = NewQuorum()
	//bcb.Register(TCMsg{}, bcb.HandleTC)
	//bcb.Register(TmoMsg{}, bcb.HandleTmo)
	//bcb.highCert = NewTC(0)
	//bcb.lastViewTime = time.Now()
	//bcb.viewDuration = make(map[View]time.Duration)
	return pm
}

func (p *Pacemaker) ProcessRemoteTmo(tmo *TMO) (bool, *TC) {
	p.mu.Lock()
	p.mu.Unlock()
	if tmo.View < p.GetCurView() {
		log.Warningf("stale timeout msg")
		return false, nil
	}
	return p.timeoutController.AddTmo(tmo)
}

func (p *Pacemaker) ProcessLocalTmo() {

}

//
//func (b *Pacemaker) HandleTC(tc TCMsg) {
//	//log.Infof("[%v] is processing tc for view %v", b.NodeID(), tc.View)
//	b.mu.Lock()
//	if tc.View < b.curView {
//		//log.Warningf("[%s] received tc's view %v is lower than current view %v", b.NodeID(), tc.View, b.curView)
//		b.mu.Unlock()
//		return
//	}
//	if tc.View > b.highCert.View {
//		b.highCert = NewTC(tc.View)
//	}
//	b.mu.Unlock()
//	b.AdvanceView(tc.View)
//}
//
//// TimeoutFor broadcasts the timeout msg for the view when it timeouts
//func (b *Pacemaker) TimeoutFor(view View) {
//	tmoMsg := TmoMsg{
//		View:   view,
//		NodeID: b.ID(),
//		HighTC: NewTC(view - 1),
//	}
//	//log.Debugf("[%s] is timeout for view %v", b.NodeID(), view)
//	if b.IsByz() {
//		b.MulticastQuorum(GetConfig().ByzNo, tmoMsg)
//		return
//	}
//	b.Broadcast(tmoMsg)
//	b.HandleTmo(tmoMsg)
//}

func (b *Pacemaker) AdvanceView(view types.View) {
	b.mu.Lock()
	if view < b.curView {
		b.mu.Unlock()
		return
	}
	timeouts := view - b.curView
	if timeouts < 0 {
		timeouts = 0
	}
	b.curView = view + 1
	b.mu.Unlock()
	b.newViewChan <- view + 1 // reset timer for the next view
}

//func (b *Pacemaker) printViewTime() {
//	//log.Infof("[%v] is printing view duration", b.NodeID())
//	for view, duration := range b.viewDuration {
//		log.Infof("view %v duration: %v seconds", view, duration.Seconds())
//	}
//}

func (b *Pacemaker) EnteringViewEvent() chan types.View {
	return b.newViewChan
}

func (b *Pacemaker) GetCurView() types.View {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.curView
}
