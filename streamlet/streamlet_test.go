package streamlet

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/election"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/node"
	"github.com/gitferry/bamboo/pacemaker"
	"github.com/gitferry/bamboo/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

// one chain
func TestStreamlet_CommitRule1(t *testing.T) {
	sl := MakeStreamlet()
	b1 := blockchain.BuildProposal()
	bc.AddBlock(b1)
	err := hs.UpdateStateByQC(b1.QC)
	require.NoError(t, err)
	canCommit, blockID, err := hs.CommitRule(b1.QC)
	require.Error(t, err)
	require.False(t, canCommit)
	require.Nil(t, blockID)
}

// two chain
func TestStreamlet_CommitRule2(t *testing.T) {
}

// three chain
func TestStreamlet_CommitRule3(t *testing.T) {
}

func TestStreamlet_ForkingForkchoice(t *testing.T) {
}

func MakeStreamlet() *Streamlet {
	id := identity.NewNodeID(1)
	n := node.NewNode(id, false)
	pm := pacemaker.NewPacemaker(4)
	elect := election.NewRotation(4)
	committedBlocks := make(chan *blockchain.Block)
	return NewStreamlet(n, pm, elect, committedBlocks)
}
