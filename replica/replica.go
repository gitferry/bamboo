package replica

import (
	"time"

	"github.com/gitferry/zeitgeber/messages"
)

type Replica struct {
	curView      uint64
	timer        *time.Timer
	timout       time.Duration
	WishMsgs     chan *messages.WishMsg
	proposalMsgs chan *messages.ProposalMsg
}

func (r *Replica) onTimeout() {
	r.sendWish(r.curView + 1)
}

func (r *Replica) sendWish(view uint64) {
	msg := &messages.WishMsg{
		View: view,
	}
	r.WishMsgs <- msg
}

func (r *Replica) onProposalMsg(msg *messages.ProposalMsg) {
	if r.curView >= msg.View {
		return
	}
	r.curView = msg.View
}

func Init() *Replica {
	timer := time.NewTimer(2 * time.Second)
	return &Replica{
		timer: timer,
	}
}

func (r *Replica) Run() {
	r.resettimer()
	for {
		select {
		case <-r.timer.C:
			r.onTimeout()
		case msg := <-r.proposalMsgs:
			r.resettimer()
			r.onProposalMsg(msg)
		}
	}
}

func (r *Replica) resettimer() {
	r.timer.Reset(r.timout)
}
