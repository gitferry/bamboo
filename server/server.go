package main

import (
	"flag"
	"sync"

	"github.com/gitferry/bamboo"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/replica"
)

var algorithm = flag.String("algorithm", "hotstuff", "BFT consensus algorithm")
var id = flag.String("id", "", "NodeID of the node")
var simulation = flag.Bool("sim", false, "simulation mode")
var isByz = flag.Bool("isByz", false, "this is a Byzantine node")

func initReplica(id identity.NodeID, isByz bool) {
	log.Infof("node %v starting...", id)
	if isByz {
		log.Infof("node %v is Byzantine", id)
	}

	r := replica.NewReplica(id, *algorithm, isByz)
	r.Start()
}

func main() {
	bamboo.Init()

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		config.Simulation()
		for id := range config.GetConfig().Addrs {
			isByz := false
			if id.Node() <= config.GetConfig().ByzNo {
				isByz = true
			}
			go initReplica(id, isByz)
		}
		wg.Wait()
	} else {
		initReplica(identity.NodeID(*id), *isByz)
	}
}
