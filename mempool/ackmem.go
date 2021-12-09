package mempool

import (
	"container/list"
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/utils"
	"sync"
)

type AckMem struct {
	stableMicroblocks  *list.List
	txnList            *list.List
	microblockMap      map[crypto.Identifier]*blockchain.MicroBlock
	pendingMicroblocks map[crypto.Identifier]*PendingMicroblock
	ackBuffer          map[crypto.Identifier][]identity.NodeID //number of the ack received before mb arrived
	stableMBs          map[crypto.Identifier]struct{}          //keeps track of stable microblocks
	bsize              int                                     // number of microblocks in a proposal
	msize              int                                     // byte size of transactions in a microblock
	memsize            int                                     // number of microblocks in mempool
	currSize           int
	threshhold         int // number of acks needed for a stable microblock
	mu                 sync.Mutex
}

type PendingMicroblock struct {
	microblock *blockchain.MicroBlock
	ackMap     map[identity.NodeID]struct{} // who has sent acks
}

// NewAckMem creates a new naive mempool
func NewAckMem() *AckMem {
	return &AckMem{
		bsize:              config.GetConfig().BSize,
		msize:              config.GetConfig().MSize,
		memsize:            config.GetConfig().MemSize,
		threshhold:         config.GetConfig().Q,
		stableMicroblocks:  list.New(),
		microblockMap:      make(map[crypto.Identifier]*blockchain.MicroBlock),
		pendingMicroblocks: make(map[crypto.Identifier]*PendingMicroblock),
		ackBuffer:          make(map[crypto.Identifier][]identity.NodeID),
		stableMBs:          make(map[crypto.Identifier]struct{}),
		currSize:           0,
		txnList:            list.New(),
	}
}

// AddTxn adds a transaction and returns a microblock if msize is reached
// then the contained transactions should be deleted
func (am *AckMem) AddTxn(txn *message.Transaction) (bool, *blockchain.MicroBlock) {
	// mempool is full
	if am.stableMicroblocks.Len()+len(am.pendingMicroblocks) >= am.memsize {
		return false, nil
	}

	// get the size of the structure. txn is the pointer.
	tranSize := utils.SizeOf(txn)
	totalSize := tranSize + am.currSize

	if tranSize > am.msize {
		return false, nil
	}

	if totalSize > am.msize {
		//do not add the curr trans, and generate a microBlock
		//set the currSize to curr trans, since it is the only one does not add to the microblock
		var id crypto.Identifier
		am.currSize = tranSize
		newBlock := blockchain.NewMicroblock(id, am.makeTxnSlice())
		am.txnList.PushBack(txn)
		return true, newBlock

	} else if totalSize == am.msize {
		//add the curr trans, and generate a microBlock
		var id crypto.Identifier
		allTxn := append(am.makeTxnSlice(), txn)
		am.currSize = 0
		return true, blockchain.NewMicroblock(id, allTxn)

	} else {
		//
		am.txnList.PushBack(txn)
		am.currSize = totalSize
		return false, nil
	}
}

// AddMicroblock adds a microblock into a FIFO queue
// return an err if the queue is full (memsize)
func (am *AckMem) AddMicroblock(mb *blockchain.MicroBlock) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	//if am.microblocks.Len() >= am.memsize {
	//	return errors.New("the memory queue is full")
	//}
	_, exists := am.microblockMap[mb.Hash]
	if exists {
		return nil
	}
	pm := &PendingMicroblock{
		microblock: mb,
		ackMap:     make(map[identity.NodeID]struct{}),
	}
	pm.ackMap[mb.Sender] = struct{}{}
	am.pendingMicroblocks[mb.Hash] = pm
	am.microblockMap[mb.Hash] = mb

	//check if there are some acks of this microblock arrived before
	buffer, received := am.ackBuffer[mb.Hash]
	if received {
		// if so, add these ack to the pendingblocks
		for _, ack := range buffer {
			//am.pendingMicroblocks[mb.Hash].ackMap[ack] = struct{}{}
			pm.ackMap[ack] = struct{}{}
			if len(pm.ackMap) >= am.threshhold {
				am.stableMicroblocks.PushBack(mb)
				am.stableMBs[mb.Hash] = struct{}{}
				delete(am.pendingMicroblocks, mb.Hash)
			}
		}
	}
	return nil
}

// AddAck adds an ack and push a microblock into the stableMicroblocks queue if it receives enough acks
func (am *AckMem) AddAck(ack *message.Ack) {
	am.mu.Lock()
	defer am.mu.Unlock()
	target, received := am.pendingMicroblocks[ack.ID]
	//check if the ack arrives before the microblock
	if received {
		target.ackMap[ack.Receiver] = struct{}{}
		if len(target.ackMap) >= am.threshhold {
			am.stableMicroblocks.PushBack(target.microblock)
			am.stableMBs[target.microblock.Hash] = struct{}{}
			delete(am.pendingMicroblocks, ack.ID)
		}
	} else {
		//ack arraives before microblock, record the number of ack received before microblock
		//let the addMicroblock do the rest.
		buffer, exist := am.ackBuffer[ack.ID]
		if exist {
			am.ackBuffer[ack.ID] = append(buffer, ack.Receiver)
		} else {
			temp := make([]identity.NodeID, 0)
			am.ackBuffer[ack.ID] = append(temp, ack.Receiver)
		}
	}
}

// GeneratePayload generates a list of microblocks according to bsize
// if the remaining microblocks is less than bsize then return all
func (am *AckMem) GeneratePayload() *blockchain.Payload {
	var batchSize int
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.stableMicroblocks.Len() >= am.bsize {
		batchSize = am.bsize
	} else {
		batchSize = am.stableMicroblocks.Len()
	}
	microblockList := make([]*blockchain.MicroBlock, 0)

	for i := 0; i < batchSize; i++ {
		mb := am.front()
		if mb == nil {
			break
		}
		microblockList = append(microblockList, mb)
	}

	return blockchain.NewPayload(microblockList)
}

// CheckExistence checks if the referred microblocks in the proposal exists
// in the mempool and return missing ones if there's any
// return true if there's no missing transactions
func (am *AckMem) CheckExistence(p *blockchain.Proposal) (bool, []crypto.Identifier) {
	id := make([]crypto.Identifier, 0)
	return false, id
}

// RemoveMicroblock removes reffered microblocks from the mempool
func (am *AckMem) RemoveMicroblock(id crypto.Identifier) error {
	am.mu.Lock()
	defer am.mu.Unlock()
	_, exists := am.microblockMap[id]
	if exists {
		delete(am.microblockMap, id)
	}
	_, exists = am.stableMBs[id]
	if exists {
		delete(am.microblockMap, id)
	}
	return nil
}

// FindMicroblock finds a reffered microblock
func (am *AckMem) FindMicroblock(id crypto.Identifier) (bool, *blockchain.MicroBlock) {
	am.mu.Lock()
	defer am.mu.Unlock()
	mb, found := am.microblockMap[id]
	return found, mb
}

// FillProposal pulls microblocks from the mempool and build a pending block,
// a pending block should include the proposal, micorblocks that already exist,
// and a missing list if there's any
func (am *AckMem) FillProposal(p *blockchain.Proposal) *blockchain.PendingBlock {
	am.mu.Lock()
	defer am.mu.Unlock()
	existingBlocks := make([]*blockchain.MicroBlock, 0)
	missingBlocks := make(map[crypto.Identifier]struct{}, 0)
	for _, id := range p.HashList {
		block, found := am.microblockMap[id]
		if found {
			existingBlocks = append(existingBlocks, block)
			_, found = am.pendingMicroblocks[id]
			if found {
				delete(am.pendingMicroblocks, id)
				break
			}
			for e := am.stableMicroblocks.Front(); e != nil; e = e.Next() {
				// do something with e.Value
				mb := e.Value.(*blockchain.MicroBlock)
				if mb == block {
					am.stableMicroblocks.Remove(e)
					break
				}
			}
		} else {
			missingBlocks[id] = struct{}{}
		}
	}
	return blockchain.NewPendingBlock(p, missingBlocks, existingBlocks)
}

func (am *AckMem) IsStable(id crypto.Identifier) bool {
	am.mu.Lock()
	defer am.mu.Unlock()
	_, exists := am.stableMBs[id]
	if exists {
		return true
	}
	return false
}

func (am *AckMem) front() *blockchain.MicroBlock {
	if am.stableMicroblocks.Len() == 0 {
		return nil
	}
	ele := am.stableMicroblocks.Front()
	val, ok := ele.Value.(*blockchain.MicroBlock)
	if !ok {
		return nil
	}
	am.stableMicroblocks.Remove(ele)
	return val
}

func (am *AckMem) makeTxnSlice() []*message.Transaction {
	allTxn := make([]*message.Transaction, 0)
	for am.txnList.Len() > 0 {
		e := am.txnList.Front()
		allTxn = append(allTxn, e.Value.(*message.Transaction))
		am.txnList.Remove(e)
	}
	return allTxn
}
