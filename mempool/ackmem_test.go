package mempool

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/message"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// add Ack before the microblock arrives
func TestAckMem_AddAck1(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	ack1 := NewMockAck(mb1.Hash, "1")
	ack2 := NewMockAck(mb1.Hash, "2")
	am.AddAck(ack1)
	am.AddAck(ack2)
	require.Equal(t, 0, am.stableMicroblocks.Len())
	pl := am.GeneratePayload()
	require.Equal(t, 0, len(pl.MicroblockList))
	_ = am.AddMicroblock(mb1)
	require.Equal(t, 1, am.stableMicroblocks.Len())
	pl = am.GeneratePayload()
	require.Equal(t, 1, len(pl.MicroblockList))
	require.Equal(t, 0, am.stableMicroblocks.Len())
}

// add ack but it is not over the threshold
func TestAckMem_AddAck2(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash, "1")
	am.AddAck(ack1)
	pl := am.GeneratePayload()
	require.Equal(t, 0, len(pl.MicroblockList))
	require.Equal(t, 0, am.stableMicroblocks.Len())
}

//add ack and it is over the threshold
func TestAckMem_AddAck3(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash, "1")
	ack2 := NewMockAck(mb1.Hash, "2")
	am.AddAck(ack1)
	am.AddAck(ack2)
	require.Equal(t, 1, am.stableMicroblocks.Len())
	pl := am.GeneratePayload()
	require.Equal(t, 1, len(pl.MicroblockList))
	require.Equal(t, 0, am.stableMicroblocks.Len())
}

//test ack exists in the buffer
func TestAckMem_AddAck4(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	ack1 := NewMockAck(mb1.Hash, "1")
	ack2 := NewMockAck(mb1.Hash, "2")
	am.AddAck(ack1)
	am.AddAck(ack2)
	require.Equal(t, 2, len(am.ackBuffer[mb1.Hash]))
}

// add two same microblocks
func TestAckMem_AddMicroblock(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash, "1")
	am.AddAck(ack1)
	pl := blockchain.NewPayload([]*blockchain.MicroBlock{mb1})
	p := &blockchain.Proposal{
		BlockHeader: blockchain.BlockHeader{
			ID: mb1.Hash,
		},
		HashList: pl.GenerateHashList(),
	}
	require.Equal(t, 0, am.stableMicroblocks.Len())
	require.Equal(t, 1, len(am.pendingMicroblocks))
	pd := am.FillProposal(p)
	require.Equal(t, 0, len(am.pendingMicroblocks))
	require.Equal(t, 0, pd.MissingCount())
}

// test fill proposal from pending blocks
func TestAckMem_FillProposal(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	mb2 := mb1
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash, "1")
	am.AddAck(ack1)
	pl := blockchain.NewPayload([]*blockchain.MicroBlock{mb1})
	p := &blockchain.Proposal{
		BlockHeader: blockchain.BlockHeader{
			ID: mb1.Hash,
		},
		HashList: pl.GenerateHashList(),
	}
	require.Equal(t, 0, am.stableMicroblocks.Len())
	require.Equal(t, 1, len(am.pendingMicroblocks))
	pd := am.FillProposal(p)
	require.Equal(t, 0, len(am.pendingMicroblocks))
	require.Equal(t, 0, pd.MissingCount())
	err = am.AddMicroblock(mb2)
	require.Nil(t, err)
	require.Equal(t, 0, len(am.pendingMicroblocks))
}

// test fill proposal from stable blocks
func TestAckMem_FillProposal2(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash, "1")
	ack2 := NewMockAck(mb1.Hash, "2")
	am.AddAck(ack1)
	am.AddAck(ack2)
	require.Equal(t, 1, am.stableMicroblocks.Len())
	pl := blockchain.NewPayload([]*blockchain.MicroBlock{mb1})
	p := &blockchain.Proposal{
		BlockHeader: blockchain.BlockHeader{
			ID: mb1.Hash,
		},
		HashList: pl.GenerateHashList(),
	}
	require.Equal(t, 0, len(am.pendingMicroblocks))
	pd := am.FillProposal(p)
	require.Equal(t, 0, am.stableMicroblocks.Len())
	require.Equal(t, 0, pd.MissingCount())
	pl2 := am.GeneratePayload()
	require.Equal(t, 0, len(pl2.MicroblockList))
}

// test fill proposal from a stable block and a pending block
func TestAckMem_FillProposal3(t *testing.T) {
	am := NewMockAckmem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	mb2 := NewMockMicroblock(timestamp.Add(1))
	err = am.AddMicroblock(mb2)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash, "1")
	ack2 := NewMockAck(mb1.Hash, "2")
	require.Equal(t, 2, len(am.pendingMicroblocks))
	am.AddAck(ack1)
	am.AddAck(ack2)
	require.Equal(t, 1, am.stableMicroblocks.Len())
	require.Equal(t, 1, len(am.pendingMicroblocks))
	pl := blockchain.NewPayload([]*blockchain.MicroBlock{mb1, mb2})
	p := &blockchain.Proposal{
		HashList: pl.GenerateHashList(),
	}
	pd := am.FillProposal(p)
	require.Equal(t, 0, am.stableMicroblocks.Len())
	require.Equal(t, 0, len(am.pendingMicroblocks))
	require.Equal(t, 0, pd.MissingCount())
}

func NewMockAck(identifier crypto.Identifier, receiver identity.NodeID) *message.Ack {
	return &message.Ack{
		ID:       identifier,
		Receiver: receiver,
	}
}
func NewMockAckmem() *AckMem {
	config.Configuration.BSize = 1
	config.Configuration.Q = 3
	return NewAckMem()
}
