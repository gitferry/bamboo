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
	p.mu.Lock()
	defer p.mu.Unlock()
	if tmo.View < p.curView {
		log.Warningf("stale timeout msg")
		return false, nil
	}
	return p.timeoutController.AddTmo(tmo)
}

func (b *Pacemaker) AdvanceView(view types.View) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if view < b.curView {
		return
	}
	timeouts := view - b.curView
	if timeouts < 0 {
		timeouts = 0
	}
	b.curView = view + 1
	b.newViewChan <- view + 1 // reset timer for the next view
}

func (b *Pacemaker) EnteringViewEvent() chan types.View {
	return b.newViewChan
}

func (b *Pacemaker) GetCurView() types.View {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.curView
}

func (b *Pacemaker) UpdateTC(tc *TC) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.highTC == nil || tc.View > b.highTC.View {
		b.highTC = tc
	}
}

func (b *Pacemaker) GetHighTC() *TC {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.highTC
}

func (b *Pacemaker) GetTimerForView() time.Duration {
	b.mu.Lock()
	if b.curView == 0 {
		b.mu.Unlock()
		return 2000 * time.Millisecond
	}
	b.mu.Unlock()
	return time.Duration(config.GetConfig().Timeout) * time.Millisecond
}
