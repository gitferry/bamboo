package pacemaker

import (
	"github.com/gitferry/bamboo/config"
	"sync"
	"time"

	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/types"
)

type Pacemaker struct {
	curView           types.View
	newViewChan       chan types.View
	highTC            *TC
	timeoutController *TimeoutController
	mu                sync.Mutex
}

func NewPacemaker(n int) *Pacemaker {
	pm := new(Pacemaker)
	pm.newViewChan = make(chan types.View)
	pm.timeoutController = NewTimeoutController(n)
	return pm
}

func (p *Pacemaker) ProcessRemoteTmo(tmo *TMO) (bool, *TC) {
	if tmo.View < p.curView {
		log.Warningf("stale timeout msg")
		return false, nil
	}
	return p.timeoutController.AddTmo(tmo)
}

func (p *Pacemaker) AdvanceView(view types.View) {
	if view < p.curView {
		return
	}
	p.curView = view + 1
	go func() {
		p.newViewChan <- view + 1 // reset timer for the next view
	}()
}

func (p *Pacemaker) EnteringViewEvent() chan types.View {
	return p.newViewChan
}

func (p *Pacemaker) GetCurView() types.View {
	return p.curView
}

func (p *Pacemaker) UpdateTC(tc *TC) {
	if p.highTC == nil || tc.View > p.highTC.View {
		p.highTC = tc
	}
}

func (p *Pacemaker) GetHighTC() *TC {
	return p.highTC
}

func (p *Pacemaker) GetTimerForView() time.Duration {
	//if p.curView == 0 {
	//	return 2000 * time.Millisecond
	//}
	return time.Duration(config.GetConfig().Timeout) * time.Millisecond
}
