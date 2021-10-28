package mempool

import (
	"bytes"
	"container/list"
	"errors"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/message"
	"sync"
	"unsafe"
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
	tranSize := unsafe.Sizeof(*txn)
	totalSize := int(tranSize) + nm.currSize
	if totalSize > nm.msize {
		//do not add the curr trans, and generate a microBlock
		//set the currSize to curr trans, since it is the only one does not add to the microblock
		nm.currSize = int(tranSize)
	} else if totalSize == nm.msize {
		//add the curr trans, and generate a microBlock
		nm.currSize = 0
	} else {
		//
		nm.currSize = totalSize
	}
	var mb *blockchain.MicroBlock
	return true, mb
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
		mb := nm.microblocks.Front().Value.(*blockchain.MicroBlock)
		microblockList = append(microblockList, mb)
	}

	return blockchain.NewPayload(microblockList)
}

// CheckExistence checks if the referred microblocks in the proposal exists
// in the mempool and return missing ones if there's any
// return true if there's no missing transactions
func (nm *NaiveMem) CheckExistence(p *blockchain.Proposal) (bool, []crypto.Identifier) {
	checkList := p.HashList
	missingList := make([]crypto.Identifier, 0)
	allExist := true

	nm.mu.Lock()
	defer nm.mu.Unlock()

	for _, id := range checkList {
		_, exist := nm.microblockMap[id]
		if !exist {
			allExist = true
			missingList = append(missingList, id)
		}
	}

	if allExist {
		return allExist, nil
	} else {
		return allExist, missingList
	}
}

// RemoveMicroblock removes reffered microblocks from the mempool
func (nm *NaiveMem) RemoveMicroblock(id crypto.Identifier) error {
	if nm.microblockMap == nil || nm.microblocks == nil {
		return errors.New("The mempool is not initialized yet")
	}

	var next *list.Element
	targetByte := crypto.IDToByte(id)

	for e := nm.microblocks.Front(); e != nil; e = next {
		next = e.Next()
		currId := e.Value.(*blockchain.MicroBlock).Hash
		currByte := crypto.IDToByte(currId)

		if bytes.Equal(currByte, targetByte) {
			nm.microblocks.Remove(e)

			nm.mu.Lock()
			defer nm.mu.Unlock()

			delete(nm.microblockMap, id)
		}
	}

	return errors.New("Cannot find this MicroBlock")
}

// FindMicroblock finds a reffered microblock
func (nm *NaiveMem) FindMicroblock(id crypto.Identifier) (bool, *blockchain.MicroBlock) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	block, exist := nm.microblockMap[id]
	if exist {
		return true, block
	} else {
		return false, nil
	}
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

//func (nm *NaiveMem) front() *blockchain.MicroBlock {
//	if nm.microblocks.Len() == 0 {
//		return nil
//	}
//	ele := nm.microblocks.Front()
//	val, ok := ele.Value.(*blockchain.MicroBlock)
//	if !ok {
//		return nil
//	}
//	nm.microblocks.Remove(ele)
//	return val
//}
