package main

import (
	"flag"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"

	"github.com/gitferry/bamboo"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/log"
)

var algorithm = flag.String("algorithm", "hotstuff", "BFT consensus algorithm")
var id = flag.String("id", "", "NodeID of the node")
var simulation = flag.Bool("sim", false, "simulation mode")
var isByz = flag.Bool("isByz", false, "this is a Byzantine node")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func replica(id identity.NodeID, isByz bool) {
	log.Infof("node %v starting...", id)
	if isByz {
		log.Infof("node %v is Byzantine", id)
	}

	r := bamboo.NewReplica(id, *algorithm, isByz)
	r.Start()
}

func main() {
	bamboo.Init()
	// The private and public keys are generated here. Check the crypto package.
	errCrypto := crypto.SetKeys()
	if errCrypto != nil {
		log.Fatal("Could not generate keys:", errCrypto)
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if *simulation {
		var wg sync.WaitGroup
		wg.Add(1)
		config.Simulation()
		for id := range config.GetConfig().Addrs {
			isByz := false
			if id.Node() <= config.GetConfig().ByzNo {
				isByz = true
			}
			go replica(id, isByz)
		}
		wg.Wait()
	} else {
		replica(identity.NodeID(*id), *isByz)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
