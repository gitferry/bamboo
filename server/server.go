package main

import (
	"flag"
	"sync"

	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/bcb"
	"github.com/gitferry/zeitgeber/log"
)

var algorithm = flag.String("algorithm", "bcb", "synchronizer algorithm")
var id = flag.String("id", "", "ID of the node")
var simulation = flag.Bool("sim", false, "simulation mode")
var isByz = flag.Bool("isByz", false, "this is a Byzantine node")

func replica(id zeitgeber.ID, isByz bool) {
	log.Infof("node %v starting...", id)

	r := zeitgeber.NewReplica(id, isByz)

	switch *algorithm {
	case "bcb":
		r.Synchronizer = bcb.NewBcb(r.Node, r.Election)
	default:
		r.Synchronizer = bcb.NewBcb(r.Node, r.Election)
	}
	if r.IsLeader(id, 1) {
		log.Debugf("[%v] should kick off", id)
		go r.MakeProposal(1)
	}
	r.Run()
}

func main() {
	zeitgeber.Init()

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		zeitgeber.Simulation()
		for id := range zeitgeber.GetConfig().Addrs {
			isByz := false
			if id.Node() <= zeitgeber.GetConfig().ByzNo {
				isByz = true
			}
			n := id
			go replica(n, isByz)
		}
		wg.Wait()
	} else {
		replica(zeitgeber.ID(*id), *isByz)
	}
}
