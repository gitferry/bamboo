package zeitgeber

import (
	"time"

	"github.com/gitferry/zeitgeber/bcb"
	"github.com/gitferry/zeitgeber/log"
)

type Replica struct {
	Node
	Synchronizer
	Election
	isByz bool
}

// GetState returns currentTerm and whether this server
// believes it is the leader.
func (r *Replica) GetState() (int, bool) {

	var view int
	var isleader bool

	return view, isleader
}

func (r *Replica) HandleProposal(proposal ProposalMsg) {
	tc := proposal.TimeCert
	r.HandleTC(tc)
	curView := r.GetCurView()
	if proposal.View != curView {
		log.Warningf("[%s] received a stale proposal")
		return
	}
	if r.Election.IsLeader(proposal.NodeID, proposal.View) {
		log.Warningf("[%s] received a proposal from an invalid leader")
		return
	}
}

func (r *Replica) StartTimer() {
	for {
		timer := time.NewTimer(GetTimer())
		go func() {
			select {
			case <-timer.C:
				r.handleTimeout()
			case <-r.Synchronizer.ResetTimer():
				return
			}
		}()
	}
}

func (r *Replica) handleTimeout() {
	r.Synchronizer.TimeoutFor(r.GetCurView())
}

func NewReplica(id ID, syncAlg string, isByz bool) *Replica {
	r := new(Replica)
	r.Node = NewNode(id)
	r.isByz = isByz
	elect := NewRotation(GetConfig().N())
	r.Election = elect
	switch syncAlg {
	case "bcb":
		r.Synchronizer = bcb.NewBcb(r.Node, elect)
	default:
		r.Synchronizer = bcb.NewBcb(r.Node, elect)
	}
	r.Register(ProposalMsg{}, r.HandleProposal)
	return r
}
