package zeitgeber

import (
	"time"

	"github.com/gitferry/zeitgeber/log"
)

type Replica struct {
	Node
	Pacemaker
	Election
}

func (r *Replica) HandleProposal(proposal ProposalMsg) {
	//log.Infof("[%v] received a proposal from %v, view is %v", r.ID(), proposal.NodeID, proposal.View)
	r.HandleTC(TCMsg{
		View:   proposal.TimeCert.View,
		NodeID: proposal.NodeID,
	})
	curView := r.GetCurView()
	if proposal.View != curView {
		log.Warningf("[%v] received a stale proposal", r.ID())
		return
	}
	if !r.Election.IsLeader(proposal.NodeID, proposal.View) {
		log.Warningf(
			"[%v] received a proposal (%v) from an invalid leader (%v)",
			r.ID(), proposal.View, proposal.NodeID)
		return
	}
}

func (r *Replica) MakeProposal(view View) {
	curView := r.GetCurView()
	// the replica should propose if it is the leader
	proposal := ProposalMsg{
		NodeID:   r.ID(),
		View:     view,
		TimeCert: NewTC(curView),
	}
	time.Sleep(20 * time.Millisecond)
	//log.Infof("[%v] is proposing for view %v", r.ID(), curView)
	if r.IsByz() {
		r.MulticastQuorum(GetConfig().ByzNo, proposal)
	} else {
		r.Broadcast(proposal)
	}
	r.HandleProposal(proposal)
}

func (r *Replica) ProcessNewView(newView View) {
	//log.Debugf("[%v] is processing new view: %v", r.ID(), newView)
	if !r.IsLeader(r.ID(), newView+1) {
		return
	}
	r.MakeProposal(newView + 1)
}

func (r *Replica) startTimer() {
	duration := GetTimer()
	timer := time.NewTimer(duration)
	for {
		timer.Reset(duration)
		go func() {
			<-timer.C
			r.handleTimeout()
			return
		}()
		newView := <-r.Pacemaker.EnteringViewEvent()
		timer.Stop()
		go r.ProcessNewView(newView)
	}
}

func (r *Replica) handleTimeout() {
	r.Pacemaker.TimeoutFor(r.GetCurView())
}

func NewReplica(id ID, isByz bool) *Replica {
	r := new(Replica)
	r.Node = NewNode(id, isByz)
	if isByz {
		log.Infof("[%v] is Byzantine", r.ID())
	}
	elect := NewRotation(GetConfig().N())
	r.Election = elect
	r.Register(ProposalMsg{}, r.HandleProposal)
	go r.startTimer()
	return r
}
