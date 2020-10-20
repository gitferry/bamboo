package pacemaker

import (
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
	p.mu.Unlock()
	if tmo.View < p.GetCurView() {
		log.Warningf("stale timeout msg")
		return false, nil
	}
	return p.timeoutController.AddTmo(tmo)
}

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

	if tc.View > b.highTC.View {
		b.highTC = tc
	}
}

func (b *Pacemaker) GetHighTC() *TC {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.highTC
}

func (b *Pacemaker) GetTimerForView() time.Duration {
	return 1000 * time.Millisecond
}
