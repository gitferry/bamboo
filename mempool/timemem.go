package mempool

import (
	"container/list"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/pq"
	"github.com/gitferry/bamboo/utils"
	"sync"
)

type Timemem struct {
	mbpq          *pq.PriorityQueue
	txnList       *list.List
	microblockMap map[crypto.Identifier]*blockchain.MicroBlock
	bsize         int // number of microblocks in a proposal
	msize         int // byte size of transactions in a microblock
	memsize       int // number of microblocks in mempool
	currSize      int // current byte size of txns
	mu            sync.Mutex
}

// NewTimeme creates a new naive mempool
func NewTimemem() *Timemem {
	return &Timemem{
		bsize:         config.GetConfig().BSize,
		msize:         config.GetConfig().MSize,
		memsize:       config.GetConfig().MemSize,
		mbpq:          pq.NewPriorityQueue(),
		microblockMap: make(map[crypto.Identifier]*blockchain.MicroBlock),
		txnList:       list.New(),
	}
}

// AddTxn adds a transaction and returns a microblock if msize is reached
// then the contained transactions should be deleted
func (tm *Timemem) AddTxn(txn *message.Transaction) (bool, *blockchain.MicroBlock) {
	// mempool is full
	if tm.mbpq.Len() >= tm.memsize {
		return false, nil
	}

	// get the size of the structure. txn is the pointer.
	tranSize := utils.SizeOf(txn)
	totalSize := tranSize + tm.currSize

	if tranSize > tm.msize {
		return false, nil
	}
	if totalSize > tm.msize {
		//do not add the curr trans, and generate a microBlock
		//set the currSize to curr trans, since it is the only one does not add to the microblock
		var id crypto.Identifier
		tm.currSize = tranSize
		newBlock := blockchain.NewMicroblock(id, tm.makeTxnSlice())
		tm.txnList.PushBack(txn)
		return true, newBlock

	} else if totalSize == tm.msize {
		//add the curr trans, and generate a microBlock
		var id crypto.Identifier
		allTxn := append(tm.makeTxnSlice(), txn)
		tm.currSize = 0
		return true, blockchain.NewMicroblock(id, allTxn)

	} else {
		//
		tm.txnList.PushBack(txn)
		tm.currSize = totalSize
		return false, nil
	}
}

// AddMicroblock adds a microblock into a priority queue
// return an err if the queue is full (memsize)
func (tm *Timemem) AddMicroblock(mb *blockchain.MicroBlock) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	//if tm.mbpq.Len() >= tm.memsize {
	//	return fmt.Errorf("mempool is full")
	//}
	_, exists := tm.microblockMap[mb.Hash]
	if exists {
		return nil
	}
	tm.mbpq.Insert(mb, mb.FutureTimestamp.UnixNano())
	tm.microblockMap[mb.Hash] = mb

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
	microblockList := make([]*blockchain.MicroBlock, 0)
	for i := 0; i < batchSize; i++ {
		mb, _ := tm.mbpq.Pop()
		if mb == nil {
			break
		}
		microblockList = append(microblockList, mb.(*blockchain.MicroBlock))
	}
	return blockchain.NewPayload(microblockList, nil)
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
	tm.mu.Lock()
	defer tm.mu.Unlock()
	mb, exists := tm.microblockMap[id]
	return exists, mb
}

// FillProposal pulls microblocks from the mempool and build a pending block,
// a pending block should include the proposal, micorblocks that already exist,
// and a missing list if there's any
func (tm *Timemem) FillProposal(p *blockchain.Proposal) *blockchain.PendingBlock {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	existingBlocks := make([]*blockchain.MicroBlock, 0)
	missingBlocks := make(map[crypto.Identifier]struct{}, 0)
	for _, id := range p.HashList {
		mb, found := tm.microblockMap[id]
		if found {
			existingBlocks = append(existingBlocks, mb)
			err := tm.mbpq.Remove(mb)
			if err != nil {
				log.Errorf("a microblock does not exist in pq")
			}
			delete(tm.microblockMap, mb.Hash)
		} else {
			missingBlocks[id] = struct{}{}
		}
	}
	return blockchain.NewPendingBlock(p, missingBlocks, existingBlocks)
}

func (tm *Timemem) IsStable(id crypto.Identifier) bool {
	return false
}

func (tm *Timemem) AddAck(ack *blockchain.Ack) {
}

func (tm *Timemem) AckList(id crypto.Identifier) []identity.NodeID {
	return nil
}

func (tm *Timemem) makeTxnSlice() []*message.Transaction {
	allTxn := make([]*message.Transaction, 0)
	for tm.txnList.Len() > 0 {
		e := tm.txnList.Front()
		allTxn = append(allTxn, e.Value.(*message.Transaction))
		tm.txnList.Remove(e)
	}
	return allTxn
}
