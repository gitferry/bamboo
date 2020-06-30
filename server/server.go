package main

import (
	"flag"
	"sync"

	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/log"
)

var algorithm = flag.String("algorithm", "bcb", "Distributed algorithm")
var id = flag.String("id", "", "ID of the node")
var simulation = flag.Bool("sim", false, "simulation mode")

func replica(id zeitgeber.ID) {
	log.Infof("node %v starting...", id)
	zeitgeber.NewReplica(id, *algorithm).Run()
}

func main() {
	zeitgeber.Init()

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		zeitgeber.Simulation()
		for id := range zeitgeber.GetConfig().Addrs {
			n := id
			go replica(n)
		}
		wg.Wait()
	} else {
		replica(zeitgeber.ID(*id))
	}
}
