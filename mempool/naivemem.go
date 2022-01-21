package mempool

import (
	"container/list"
	"fmt"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
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
	totalTx       int64
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
	// mempool is full
	if nm.RemainingTx() >= int64(nm.memsize) {
		//log.Warningf("mempool's tx list is full")
		return false, nil
	}
	if nm.RemainingMB() >= int64(nm.memsize) {
		//log.Warningf("mempool's mb list is full")
		return false, nil
	}

	// get the size of the structure. txn is the pointer.
	tranSize := utils.SizeOf(txn)
	totalSize := tranSize + nm.currSize

	if tranSize > nm.msize {
		return false, nil
	}
	nm.totalTx++

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
	nm.mu.Lock()
	defer nm.mu.Unlock()
	//if nm.microblocks.Len() >= nm.memsize {
	//	return errors.New("the memory queue is full")
	//}
	_, exists := nm.microblockMap[mb.Hash]
	if exists {
		return nil
	}
	nm.microblocks.PushBack(mb)
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
	microblockList := make([]*blockchain.MicroBlock, 0)

	for i := 0; i < batchSize; i++ {
		mb := nm.front()
		if mb == nil {
			break
		}
		microblockList = append(microblockList, mb)
	}

	return blockchain.NewPayload(microblockList, nil)
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
	nm.mu.Lock()
	defer nm.mu.Unlock()
	_, exists := nm.microblockMap[id]
	if exists {
		delete(nm.microblockMap, id)
		return nil
	}
	return fmt.Errorf("the microblock does not exist, id: %x", id)
}

// FindMicroblock finds a reffered microblock
func (nm *NaiveMem) FindMicroblock(id crypto.Identifier) (bool, *blockchain.MicroBlock) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	mb, found := nm.microblockMap[id]
	return found, mb
}

// FillProposal pulls microblocks from the mempool and build a pending block,
// a pending block should include the proposal, micorblocks that already exist,
// and a missing list if there's any
func (nm *NaiveMem) FillProposal(p *blockchain.Proposal) *blockchain.PendingBlock {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	existingBlocks := make([]*blockchain.MicroBlock, 0)
	missingBlocks := make(map[crypto.Identifier]struct{}, 0)
	for _, id := range p.HashList {
		block, found := nm.microblockMap[id]
		if found {
			existingBlocks = append(existingBlocks, block)
			for e := nm.microblocks.Front(); e != nil; e = e.Next() {
				// do something with e.Value
				mb := e.Value.(*blockchain.MicroBlock)
				if mb == block {
					nm.microblocks.Remove(e)
					break
				}
			}
		} else {
			missingBlocks[id] = struct{}{}
		}
	}
	return blockchain.NewPendingBlock(p, missingBlocks, existingBlocks)
}

func (nm *NaiveMem) TotalTx() int64 {
	return nm.totalTx
}

func (nm *NaiveMem) RemainingTx() int64 {
	return int64(nm.txnList.Len())
}

func (nm *NaiveMem) TotalMB() int64 {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	return int64(len(nm.microblockMap))
}

func (nm *NaiveMem) RemainingMB() int64 {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	return int64(nm.microblocks.Len())
}

func (nm *NaiveMem) AddAck(ack *blockchain.Ack) {
}

func (nm *NaiveMem) AckList(id crypto.Identifier) []identity.NodeID {
	return nil
}

func (nm *NaiveMem) IsStable(id crypto.Identifier) bool {
	return false
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
