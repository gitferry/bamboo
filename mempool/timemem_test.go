package mempool

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

// test GeneratePayload, and FillProposal
// bsize = 2

// add microblocks with future timestamp of 1, 2, 3, 4, should pull 1, 2 into payload
func TestTimemem_GeneratePayload1(t *testing.T) {
	tm := NewMockTimemem()
	mb1 := NewMockMicroblock(1)
	mb2 := NewMockMicroblock(2)
	mb3 := NewMockMicroblock(3)
	mb4 := NewMockMicroblock(4)
	_ = tm.AddMicroblock(mb1)
	_ = tm.AddMicroblock(mb2)
	_ = tm.AddMicroblock(mb3)
	_ = tm.AddMicroblock(mb4)
	require.Equal(t, 4, tm.mbpq.Len())
	pl := tm.GeneratePayload()
	require.Equal(t, 2, tm.mbpq.Len())
	require.Equal(t, 2, len(pl.MicroblockList))
	require.Equal(t, mb1, pl.MicroblockList[0])
	require.Equal(t, mb2, pl.MicroblockList[1])
}

// add microblocks with future timestamp of 3, 1, 2, 4, should pull 1, 2 into payload
func TestTimemem_GeneratePayload2(t *testing.T) {
	tm := NewMockTimemem()
	mb1 := NewMockMicroblock(3)
	mb2 := NewMockMicroblock(1)
	mb3 := NewMockMicroblock(2)
	mb4 := NewMockMicroblock(4)
	_ = tm.AddMicroblock(mb1)
	_ = tm.AddMicroblock(mb2)
	_ = tm.AddMicroblock(mb3)
	_ = tm.AddMicroblock(mb4)
	pl := tm.GeneratePayload()
	require.Equal(t, 2, len(pl.MicroblockList))
	require.Equal(t, mb2, pl.MicroblockList[0])
	require.Equal(t, mb3, pl.MicroblockList[1])
}

// add 2 microblocks in the mempool, fill a proposal containing the two
func TestTimemem_FillProposal1(t *testing.T) {
	tm := NewMockTimemem()
	mb1 := NewMockMicroblock(1)
	mb2 := NewMockMicroblock(2)
	_ = tm.AddMicroblock(mb1)
	_ = tm.AddMicroblock(mb2)
	mbs := make([]*blockchain.MicroBlock, 0)
	mbs = append(mbs, mb1)
	mbs = append(mbs, mb2)
	pl := blockchain.NewPayload(mbs)
	p := &blockchain.Proposal{
		HashList: pl.GenerateHashList(),
	}
	pendingBlock := tm.FillProposal(p)
	block := blockchain.BuildBlock(p, pl)
	require.Equal(t, 0, pendingBlock.MissingCount())
	require.Equal(t, block, pendingBlock.CompleteBlock())
}

// add 2 microblocks in the mempool, fill a proposal that contains another microblock besides the two
func TestTimemem_FillProposal2(t *testing.T) {
	tm := NewMockTimemem()
	mb1 := NewMockMicroblock(1)
	mb2 := NewMockMicroblock(2)
	mb3 := NewMockMicroblock(3)
	_ = tm.AddMicroblock(mb1)
	_ = tm.AddMicroblock(mb2)
	mbs := make([]*blockchain.MicroBlock, 0)
	mbs = append(mbs, mb1)
	mbs = append(mbs, mb2)
	mbs = append(mbs, mb3)
	pl := blockchain.NewPayload(mbs)
	p := &blockchain.Proposal{
		HashList: pl.GenerateHashList(),
	}
	pendingBlock := tm.FillProposal(p)
	block := blockchain.BuildBlock(p, pl)
	require.Equal(t, 1, pendingBlock.MissingCount())
	_, exists := pendingBlock.MissingMap[mb3.Hash]
	require.True(t, exists)
	pendingBlock.AddMicroblock(mb3)
	require.Equal(t, block, pendingBlock.CompleteBlock())
}

// msize = 128, add 1 transaction with payload size of 0, and not mircoblock should be generated
func TestTimemem_AddTxn1(t *testing.T) {
	tm := NewMockTimemem()
	tx1 := NewMockTxn(0)
	isBuilt, mb := tm.AddTxn(tx1)
	require.False(t, isBuilt)
	require.Nil(t, mb)
}

func NewMockTimemem() *Timemem {
	config.Configuration.BSize = 2
	config.Configuration.MSize = 128
	config.Configuration.MemSize = 50000
	return NewTimemem()
}

func NewMockMicroblock(futureTimeStamp int64) *blockchain.MicroBlock {

	txn := &message.Transaction{
		ID: hex.EncodeToString(crypto.IDToByte(utils.IdentifierFixture())),
	}
	txnList := make([]*message.Transaction, 1)
	txnList = append(txnList, txn)
	return &blockchain.MicroBlock{
		ProposalID:      utils.IdentifierFixture(),
		Hash:            utils.IdentifierFixture(),
		Txns:            txnList,
		FutureTimestamp: futureTimeStamp,
	}
}

func NewMockTxn(size int) *message.Transaction {
	var req message.Transaction
	value := make([]byte, size)
	rand.Read(value)
	req.Command.Value = value
	return &req
}
