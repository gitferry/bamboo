package mempool

import (
	"container/list"
	"fmt"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/pq"
	"sync"
)

type Timemem struct {
	mbpq          *pq.PriorityQueue
	txnList       *list.List
	microblockMap map[crypto.Identifier]*blockchain.MicroBlock
	bsize         int // number of microblocks in a proposal
	msize         int // byte size of transactions in a microblock
	memsize       int // number of microblocks in mempool
	mu            sync.Mutex
}

// NewTimeme creates a new naive mempool
func NewTimemem() *Timemem {
	return &Timemem{
		bsize:         config.GetConfig().BSize,
		msize:         config.GetConfig().MSize,
		memsize:       config.GetConfig().MemSize,
		mbpq:          pq.NewPriorityQueue(),
		microblockMap: map[crypto.Identifier]*blockchain.MicroBlock{},
		txnList:       list.New(),
	}
}

// AddTxn adds a transaction and returns a microblock if msize is reached
// then the contained transactions should be deleted
func (tm *Timemem) AddTxn(txn *message.Transaction) (bool, *blockchain.MicroBlock) {
	success := false
	var mb *blockchain.MicroBlock
	return success, mb
}

// AddMicroblock adds a microblock into a priority queue
// return an err if the queue is full (memsize)
func (tm *Timemem) AddMicroblock(mb *blockchain.MicroBlock) error {
	if tm.isFull() {
		return fmt.Errorf("mempool is full")
	}
	tm.mbpq.Insert(mb, mb.FutureTimestamp.UnixNano())

	return nil
}

// GeneratePayload generates a list of microblocks according to bsize
// if the remaining microblocks is less than bsize then return all
func (tm *Timemem) GeneratePayload() *blockchain.Payload {
	var batchSize int
	tm.mu.Lock()
	defer tm.mu.Unlock()
	batchSize = tm.mbpq.Len()
	if batchSize >= tm.bsize {
		batchSize = tm.bsize
	}
	microblockList := make([]*blockchain.MicroBlock, batchSize)
	for i := 0; i < batchSize; i++ {
		mb, _ := tm.mbpq.Pop()
		if mb == nil {
			break
		}
		microblockList = append(microblockList, mb.(*blockchain.MicroBlock))
	}
	return blockchain.NewPayload(microblockList)
}

// CheckExistence checks if the referred microblocks in the proposal exists
// in the mempool and return missing ones if there's any
// return true if there's no missing transactions
func (tm *Timemem) CheckExistence(p *blockchain.Proposal) (bool, []crypto.Identifier) {
	exists := false
	missingList := make([]crypto.Identifier, 0)
	return exists, missingList
}

// RemoveMicroblock removes reffered microblocks from the mempool
func (tm *Timemem) RemoveMicroblock(id crypto.Identifier) error {
	var err error
	return err
}

// FindMicroblock finds a reffered microblock
func (tm *Timemem) FindMicroblock(id crypto.Identifier) (bool, *blockchain.MicroBlock) {
	found := false
	var mb *blockchain.MicroBlock
	return found, mb
}

// FillProposal pulls microblocks from the mempool and build a pending block,
// a pending block should include the proposal, micorblocks that already exist,
// and a missing list if there's any
func (tm *Timemem) FillProposal(p *blockchain.Proposal) *blockchain.PendingBlock {
	var pd *blockchain.PendingBlock
	return pd
}

func (tm *Timemem) isFull() bool {
	return tm.memsize <= tm.mbpq.Len()
}
