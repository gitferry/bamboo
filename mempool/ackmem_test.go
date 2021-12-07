package mempool

import (
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

func NewMockAck(identifier crypto.Identifier, receiver identity.NodeID) *message.Ack {
	return &message.Ack{
		ID:       identifier,
		Receiver: receiver,
	}
}
func NewMockAckmem() *AckMem {
	config.Configuration.BSize = 1
	config.Configuration.EstimateNum = 3
	return NewAckMem()
}
