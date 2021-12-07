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
	"time"
)

// test GeneratePayload, and FillProposal
// bsize = 2

// add microblocks with future timestamp of 1, 2, 3, 4, should pull 1, 2 into payload
func TestTimemem_GeneratePayload1(t *testing.T) {
	tm := NewMockTimemem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	mb2 := NewMockMicroblock(timestamp.Add(2))
	mb3 := NewMockMicroblock(timestamp.Add(3))
	mb4 := NewMockMicroblock(timestamp.Add(4))
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
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(3))
	mb2 := NewMockMicroblock(timestamp.Add(1))
	mb3 := NewMockMicroblock(timestamp.Add(2))
	mb4 := NewMockMicroblock(timestamp.Add(4))
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
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	mb2 := NewMockMicroblock(timestamp.Add(2))
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
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	mb2 := NewMockMicroblock(timestamp.Add(2))
	mb3 := NewMockMicroblock(timestamp.Add(3))
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

// msize = 128, add 1 transaction with payload size of 0, and no mircoblock should be generated
func TestTimemem_AddTxn1(t *testing.T) {
	tm := NewMockTimemem()
	tm.msize = 128
	tx1 := NewMockTxn(0)
	// actual byte size of a transaction is 168
	isBuilt, mb := tm.AddTxn(tx1)
	require.False(t, isBuilt)
	require.Nil(t, mb)
}

// msize = 256, add 1 transaction with payload size of 0, and a mircoblock should be generated with one transaction
func TestTimemem_AddTxn2(t *testing.T) {
	tm := NewMockTimemem()
	tm.msize = 256
	tx1 := NewMockTxn(0)
	// actual byte size of a transaction is 168
	isBuilt, mb := tm.AddTxn(tx1)
	require.True(t, isBuilt)
	require.NotNil(t, mb)
	require.Equal(t, 1, len(mb.Txns))
}

// msize = 256, add 2 transaction with payload size of 0, and two microblocks will be generated, each with one transaction, respectively
func TestTimemem_AddTxn3(t *testing.T) {
	tm := NewMockTimemem()
	tm.msize = 256
	tx1 := NewMockTxn(0)
	tx2 := NewMockTxn(0)
	// actual byte size of a transaction is 168
	_, _ = tm.AddTxn(tx1)
	isBuilt, mb := tm.AddTxn(tx2)
	require.True(t, isBuilt)
	require.NotNil(t, mb)
	require.Equal(t, 1, len(mb.Txns))
}

// msize = 512, add 3 transaction with payload size of 0, and only one microblock will be generated with 2 transaction
func TestTimemem_AddTxn4(t *testing.T) {
	tm := NewMockTimemem()
	tm.msize = 512
	tx1 := NewMockTxn(0)
	tx2 := NewMockTxn(0)
	tx3 := NewMockTxn(0)
	tx4 := NewMockTxn(0)
	// actual byte size of a transaction is 168
	isBuilt, mb := tm.AddTxn(tx1)
	require.False(t, isBuilt)
	require.Nil(t, mb)
	isBuilt, mb = tm.AddTxn(tx2)
	require.False(t, isBuilt)
	require.Nil(t, mb)
	isBuilt, mb = tm.AddTxn(tx3)
	require.False(t, isBuilt)
	require.Nil(t, mb)
	isBuilt, mb = tm.AddTxn(tx4)
	require.True(t, isBuilt)
	require.NotNil(t, mb)
	require.Equal(t, 3, len(mb.Txns))
}

func NewMockTimemem() *Timemem {
	config.Configuration.BSize = 2
	config.Configuration.MSize = 128
	config.Configuration.MemSize = 50000
	return NewTimemem()
}

func NewMockMicroblock(futureTimeStamp time.Time) *blockchain.MicroBlock {

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
		Sender:          "0",
	}
}

func NewMockTxn(size int) *message.Transaction {
	var req message.Transaction
	value := make([]byte, size)
	rand.Read(value)
	req.Command.Value = value
	return &req
}
