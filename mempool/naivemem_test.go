package mempool

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// test GeneratePayload, and FillProposal
// bsize = 2

// add microblocks with future timestamp of 1, 2, 3, 4, should pull 1, 2 into payload
func TestNaiveMem_GeneratePayload1(t *testing.T) {
	nm := NewMockNaiveMem()
	ts := time.Now()
	mb1 := NewMockMicroblock(ts)
	mb2 := NewMockMicroblock(ts)
	mb3 := NewMockMicroblock(ts)
	mb4 := NewMockMicroblock(ts)
	_ = nm.AddMicroblock(mb1)
	_ = nm.AddMicroblock(mb2)
	_ = nm.AddMicroblock(mb3)
	_ = nm.AddMicroblock(mb4)
	require.Equal(t, 4, nm.microblocks.Len())
	require.Equal(t, 4, len(nm.microblockMap))
	pl := nm.GeneratePayload()
	require.Equal(t, 2, nm.microblocks.Len())
	require.Equal(t, mb1, pl.MicroblockList[0])
	require.Equal(t, mb2, pl.MicroblockList[1])
}

// add 2 microblocks in the mempool, fill a proposal containing the two
func TestNaiveMem_FillProposal1(t *testing.T) {
	nm := NewMockNaiveMem()
	ts := time.Now()
	mb1 := NewMockMicroblock(ts)
	mb2 := NewMockMicroblock(ts)
	_ = nm.AddMicroblock(mb1)
	_ = nm.AddMicroblock(mb2)
	mbs := make([]*blockchain.MicroBlock, 0)
	mbs = append(mbs, mb1)
	mbs = append(mbs, mb2)
	pl := blockchain.NewPayload(mbs)
	p := &blockchain.Proposal{
		HashList: pl.GenerateHashList(),
	}
	require.Equal(t, 2, nm.microblocks.Len())
	pendingBlock := nm.FillProposal(p)
	require.Equal(t, 0, nm.microblocks.Len())
	block := blockchain.BuildBlock(p, pl)
	require.Equal(t, 0, pendingBlock.MissingCount())
	require.Equal(t, block, pendingBlock.CompleteBlock())
	require.Equal(t, 0, nm.microblocks.Len())
}

// add 2 microblocks in the mempool, fill a proposal that contains another microblock besides the two
func TestNaiveMem_FillProposal2(t *testing.T) {
	nm := NewMockTimemem()
	timestamp := time.Now()
	mb1 := NewMockMicroblock(timestamp.Add(1))
	mb2 := NewMockMicroblock(timestamp.Add(2))
	mb3 := NewMockMicroblock(timestamp.Add(3))
	_ = nm.AddMicroblock(mb1)
	_ = nm.AddMicroblock(mb2)
	mbs := make([]*blockchain.MicroBlock, 0)
	mbs = append(mbs, mb1)
	mbs = append(mbs, mb2)
	mbs = append(mbs, mb3)
	pl := blockchain.NewPayload(mbs)
	p := &blockchain.Proposal{
		HashList: pl.GenerateHashList(),
	}
	pendingBlock := nm.FillProposal(p)
	block := blockchain.BuildBlock(p, pl)
	require.Equal(t, 1, pendingBlock.MissingCount())
	_, exists := pendingBlock.MissingMap[mb3.Hash]
	require.True(t, exists)
	pendingBlock.AddMicroblock(mb3)
	require.Equal(t, block, pendingBlock.CompleteBlock())
}

// msize = 128, add 1 transaction with payload size of 0, and no mircoblock should be generated
func TestNaiveMem_AddTxn1(t *testing.T) {
	nm := NewMockNaiveMem()
	nm.msize = 128
	tx1 := NewMockTxn(0)
	// actual byte size of a transaction is 168
	isBuilt, mb := nm.AddTxn(tx1)
	require.False(t, isBuilt)
	require.Nil(t, mb)
}

// msize = 256, add 1 transaction with payload size of 0, and a mircoblock should be generated with one transaction
func TestNaiveMem_AddTxn2(t *testing.T) {
	nm := NewMockNaiveMem()
	nm.msize = 256
	tx1 := NewMockTxn(0)
	tx2 := NewMockTxn(0)
	// actual byte size of a transaction is 168
	_, _ = nm.AddTxn(tx1)
	isBuilt, mb := nm.AddTxn(tx2)
	require.True(t, isBuilt)
	require.Equal(t, 1, len(mb.Txns))
}

// msize = 512, add 3 transaction with payload size of 0, and only one microblock will be generated with 2 transaction
func TestNaiveMem_AddTxn4(t *testing.T) {
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

func NewMockNaiveMem() *NaiveMem {
	config.Configuration.BSize = 2
	config.Configuration.MSize = 128
	config.Configuration.MemSize = 50000
	return NewNaiveMem()
}
