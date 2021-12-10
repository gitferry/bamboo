package blockchain

import (
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMicroBlock_AddSentNodes(t *testing.T) {
	mb := NewMicroblock(utils.IdentifierFixture(), nil)
	nodes := []identity.NodeID{"5", "6"}
	mb.AddSentNodes(nodes)
	require.Equal(t, nodes, mb.FindSentNodes())
}
