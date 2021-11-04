package mempool

import (
	"container/list"
	"errors"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/utils"
	"sync"
)

type NaiveMem struct {
	microblocks   *list.List
	txnList       *list.List
	microblockMap map[crypto.Identifier]*blockchain.MicroBlock
	bsize         int // number of microblocks in a proposal
	msize         int // byte size of transactions in a microblock
	memsize       int // number of microblocks in mempool
	currSize      int
	mu            sync.Mutex
}

// NewNaiveMem creates a new naive mempool
func NewNaiveMem() *NaiveMem {
	return &NaiveMem{
		bsize:         config.GetConfig().BSize,
		msize:         config.GetConfig().MSize,
		memsize:       config.GetConfig().MemSize,
		microblocks:   list.New(),
		microblockMap: map[crypto.Identifier]*blockchain.MicroBlock{},
		currSize:      0,
		txnList:       list.New(),
	}
}

// AddTxn adds a transaction and returns a microblock if msize is reached
// then the contained transactions should be deleted
func (nm *NaiveMem) AddTxn(txn *message.Transaction) (bool, *blockchain.MicroBlock) {
	// get the size of the structure. txn is the pointer.
	tranSize := utils.SizeOf(txn)
	totalSize := tranSize + nm.currSize

	if tranSize > nm.msize {
		return false, nil
	}

	if totalSize > nm.msize {
		//do not add the curr trans, and generate a microBlock
		//set the currSize to curr trans, since it is the only one does not add to the microblock
		var id crypto.Identifier
		nm.currSize = tranSize
		newBlock := blockchain.NewMicroblock(id, nm.makeTxnSlice())
		nm.txnList.PushBack(txn)
		return true, newBlock

	} else if totalSize == nm.msize {
		//add the curr trans, and generate a microBlock
		var id crypto.Identifier
		allTxn := append(nm.makeTxnSlice(), txn)
		nm.currSize = 0
		return true, blockchain.NewMicroblock(id, allTxn)

	} else {
		//
		nm.txnList.PushBack(txn)
		nm.currSize = totalSize
		return false, nil
	}
}

// AddMicroblock adds a microblock into a FIFO queue
// return an err if the queue is full (memsize)
func (nm *NaiveMem) AddMicroblock(mb *blockchain.MicroBlock) error {
	if nm.microblocks.Len() >= nm.memsize {
		return errors.New("the memory queue is full")
	}
	nm.microblocks.PushBack(mb)

	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.microblockMap[mb.Hash] = mb
	return nil
}

// GeneratePayload generates a list of microblocks according to bsize
// if the remaining microblocks is less than bsize then return all
func (nm *NaiveMem) GeneratePayload() *blockchain.Payload {
	var batchSize int
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.microblocks.Len() >= nm.bsize {
		batchSize = nm.bsize
	} else {
		batchSize = nm.microblocks.Len()
	}
	microblockList := make([]*blockchain.MicroBlock, batchSize)

	for i := 0; i < batchSize; i++ {
		mb := nm.front()
		microblockList = append(microblockList, mb)
	}

	return blockchain.NewPayload(microblockList)
}

// CheckExistence checks if the referred microblocks in the proposal exists
// in the mempool and return missing ones if there's any
// return true if there's no missing transactions
func (nm *NaiveMem) CheckExistence(p *blockchain.Proposal) (bool, []crypto.Identifier) {
	id := make([]crypto.Identifier, 0)
	return false, id
}

// RemoveMicroblock removes reffered microblocks from the mempool
func (nm *NaiveMem) RemoveMicroblock(id crypto.Identifier) error {
	var err error
	return err
}

// FindMicroblock finds a reffered microblock
func (nm *NaiveMem) FindMicroblock(id crypto.Identifier) (bool, *blockchain.MicroBlock) {
	var mb *blockchain.MicroBlock
	return false, mb
}

// FillProposal pulls microblocks from the mempool and build a pending block,
// a pending block should include the proposal, micorblocks that already exist,
// and a missing list if there's any
func (nm *NaiveMem) FillProposal(p *blockchain.Proposal) *blockchain.PendingBlock {
	existingBlocks := make([]*blockchain.MicroBlock, 0)
	missingBlocks := make(map[crypto.Identifier]struct{}, 0)
	for _, id := range p.HashList {
		block, found := nm.microblockMap[id]
		if found {
			existingBlocks = append(existingBlocks, block)
		} else {
			missingBlocks[id] = struct{}{}
		}
	}
	return blockchain.NewPendingBlock(p, missingBlocks, existingBlocks)
}

func (nm *NaiveMem) front() *blockchain.MicroBlock {
	if nm.microblocks.Len() == 0 {
		return nil
	}
	ele := nm.microblocks.Front()
	val, ok := ele.Value.(*blockchain.MicroBlock)
	if !ok {
		return nil
	}
	nm.microblocks.Remove(ele)
	return val
}

func (nm *NaiveMem) makeTxnSlice() []*message.Transaction {
	allTxn := make([]*message.Transaction, 0)
	for nm.txnList.Len() > 0 {
		e := nm.txnList.Front()
		allTxn = append(allTxn, e.Value.(*message.Transaction))
		nm.txnList.Remove(e)
	}
	return allTxn
}
