package pacemaker

import (
	"github.com/gitferry/bamboo/config"
	"time"

	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/types"
)

type Pacemaker struct {
	curView           types.View
	newViewChan       chan types.View
	highTC            *TC
	timeoutController *TimeoutController
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
	p.newViewChan <- view + 1 // reset timer for the next view
}

func (b *Pacemaker) EnteringViewEvent() chan types.View {
	return b.newViewChan
}

func (b *Pacemaker) GetCurView() types.View {
	return b.curView
}

func (b *Pacemaker) UpdateTC(tc *TC) {
	if b.highTC == nil || tc.View > b.highTC.View {
		b.highTC = tc
	}
}

func (b *Pacemaker) GetHighTC() *TC {
	return b.highTC
}

func (b *Pacemaker) GetTimerForView() time.Duration {
	if b.curView == 0 {
		return 2000 * time.Millisecond
	}
	return time.Duration(config.GetConfig().Timeout) * time.Millisecond
}
