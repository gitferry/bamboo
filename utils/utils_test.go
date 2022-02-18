package utils

import (
	"fmt"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/identity"
	"github.com/stretchr/testify/require"
	"math/rand"
	"strconv"
	"testing"
)

func TestPickRandomNodes(t *testing.T) {
	n := 4
	fanout := 2
	config.Configuration.Fanout = fanout
	config.Configuration.Addrs = MockAddresses(n)
	sentNodes := []identity.NodeID{"1", "2"}
	for i := 0; i < n; i++ {
		nodes := PickRandomNodes(sentNodes)
		require.Equal(t, fanout, len(nodes))
		for _, node := range nodes {
			require.NotEqual(t, identity.NewNodeID(2), node)
			require.NotEqual(t, identity.NewNodeID(1), node)
		}
	}
}

func MockAddresses(n int) map[identity.NodeID]string {
	address := make(map[identity.NodeID]string, 0)
	for i := 0; i < n; i++ {
		port := 3737 + i
		address[identity.NewNodeID(i+1)] = "tcp://127.0.0.1:" + strconv.Itoa(port)
	}

	return address
}

func TestRandomPick(t *testing.T) {
	n := 100
	f := n/3 + 1
	pick := RandomPick(n, f)
	require.Equal(t, 34, len(pick))
	fmt.Printf("%v", RandomPick(n, f))
}

func TestZipf(t *testing.T) {
	rounds := 1000000
	n := 200
	ids := make([]identity.NodeID, n)
	for i := 0; i < n; i++ {
		ids[i] = identity.NewNodeID(i + 1)
	}
	masterIndex := 0
	ids = append(ids[:masterIndex], ids[masterIndex+1:]...)
	r := rand.New(rand.NewSource(1))
	zipf := rand.NewZipf(r, 1.01, 10, uint64(n-1))
	zipfMap := make(map[uint64]int)
	for i := 0; i < rounds; i++ {
		num := zipf.Uint64()
		zipfMap[num]++
	}
	//data := make([]int, n)
	for i := 0; i < n-1; i++ {
		fmt.Printf("%v,", zipfMap[uint64(i)])
		//data[i] = zipfMap[uint64(i)]
	}
	//fmt.Printf("%v", data)
}
