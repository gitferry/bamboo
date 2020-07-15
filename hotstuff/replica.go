package hotstuff

import (
	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/log"
)

type Replica struct {
	zeitgeber.Node
	zeitgeber.Pacemaker
	zeitgeber.Election
	*HotStuff
}

func (r *Replica) handleRequest(m zeitgeber.Request) {
	log.Debugf("[%v] received txn %v\n", r.ID(), m)
	go r.Broadcast(m)
	r.StoreTxn(m)
}

func NewReplica(id zeitgeber.NodeID, isByz bool) *Replica {
	r := new(Replica)
	r.Node = zeitgeber.NewNode(id, isByz)
	if isByz {
		log.Infof("[%v] is Byzantine", r.ID())
	}
	elect := zeitgeber.NewRotation(zeitgeber.GetConfig().N())
	r.Election = elect
	r.Register(zeitgeber.Request{}, r.handleRequest)
	//TODO:
	//1. register hotstuff handlers
	//2. first leader kicks off
	return r
}
