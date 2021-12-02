package mempool

import (
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/message"
	"github.com/stretchr/testify/require"
	"testing"
)

// add Ack before the microblock arrives
func TestAckMem_AddAck1(t *testing.T) {
	am := NewMockAckmem()
	mb1 := NewMockMicroblock(1)
	ack1 := NewMockAck(mb1.Hash)
	ack2 := NewMockAck(mb1.Hash)
	am.AddAck(ack1)
	am.AddAck(ack2)
	temp1, exist := am.microblockMap[mb1.Hash]
	require.False(t, exist)
	require.Nil(t, temp1)
}

// add ack but it is not over the threshold
func TestAckMem_AddAck2(t *testing.T) {
	am := NewMockAckmem()
	mb1 := NewMockMicroblock(1)
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash)
	am.AddAck(ack1)
	temp1, exist := am.microblockMap[mb1.Hash]
	require.False(t, exist)
	require.Nil(t, temp1)
}

//add ack and it is over the threshold
func TestAckMem_AddAck3(t *testing.T) {
	am := NewMockAckmem()
	mb1 := NewMockMicroblock(1)
	err := am.AddMicroblock(mb1)
	require.Nil(t, err)
	ack1 := NewMockAck(mb1.Hash)
	ack2 := NewMockAck(mb1.Hash)
	am.AddAck(ack1)
	am.AddAck(ack2)
	temp1, exist := am.microblockMap[mb1.Hash]
	require.True(t, exist)
	require.NotNil(t, temp1)
}

//test ack exists in the buffer
func TestAckMem_AddAck4(t *testing.T) {
	am := NewMockAckmem()
	mb1 := NewMockMicroblock(1)
	ack1 := NewMockAck(mb1.Hash)
	ack2 := NewMockAck(mb1.Hash)
	am.AddAck(ack1)
	am.AddAck(ack2)
	require.Equal(t, 2, len(am.ackBuffer))
}

func NewMockAck(identifier crypto.Identifier) *message.Ack {
	return &message.Ack{
		ID: identifier,
	}
}
func NewMockAckmem() *AckMem {
	config.Configuration.EstimateNum = 2
	return NewAckMem()
}
