package utils

import (
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/identity"
	"github.com/stretchr/testify/require"
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
