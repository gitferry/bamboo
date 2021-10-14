package mempool

import (
	"container/list"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/message"
	"sync"
)

type NaiveMem struct {
	microblocks   *list.List
	txnList       *list.List
	microblockMap map[crypto.Identifier]*blockchain.MicroBlock
	bsize         int // number of microblocks in a proposal
	msize         int // byte size of transactions in a microblock
	memsize       int // number of microblocks in mempool
	mu            sync.Mutex
}

// NewNaiveMem creates a new naive mempool
func NewNaiveMem() *NaiveMem {
	return &NaiveMem{
		bsize:       config.GetConfig().BSize,
		msize:       config.GetConfig().MSize,
		memsize:     config.GetConfig().MemSize,
		microblocks: list.New(),
		txnList:     list.New(),
	}
}

// AddTxn adds a transaction and returns a microblock if msize is reached
// then the contained transactions should be deleted
func (nm *NaiveMem) AddTxn(txn *message.Transaction) (bool, *blockchain.MicroBlock) {
}

// AddMicroblock adds a microblock into a FIFO queue
// return an err if the queue is full (memsize)
func (nm *NaiveMem) AddMicroblock(mb *blockchain.MicroBlock) error {
	var err error
	return err
}

// GeneratePayload generates a hash list of microblocks according to bsize
// if the remaining microblocks is less than bsize then return all
func (nm *NaiveMem) GeneratePayload() []crypto.Identifier {
	var batchSize int
	nm.mu.Lock()
	defer nm.mu.Unlock()
	batchSize = nm.microblocks.Len()
	if batchSize >= nm.bsize {
		batchSize = nm.bsize
	}
	batch := make([]crypto.Identifier, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		mb := nm.front()
		batch = append(batch, mb.ID())
	}
	return batch
}

// CheckExistence checks if the referred microblocks in the proposal exists
// in the mempool and return missing ones if there's any
func (nm *NaiveMem) CheckExistence(p *blockchain.Proposal) (bool, []crypto.Identifier) {

}

// RemoveMicroblock removes reffered microblocks from the mempool
func (nm *NaiveMem) RemoveMicroblock(id crypto.Identifier) error {
	var err error
	return err
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