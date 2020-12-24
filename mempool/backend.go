package mempool

import (
	"container/list"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
	"sync"
)

type Backend struct {
	txns          *list.List
	totalReceived int64
	*BloomFilter
	mu *sync.Mutex
	//cond *sync.Cond
}

func NewBackend() *Backend {
	var mu sync.Mutex
	return &Backend{
		txns:        list.New(),
		BloomFilter: NewBloomFilter(),
		mu:          &mu,
		//cond:        sync.NewCond(&mu),
	}
}

func (b *Backend) insertBack(txn *message.Transaction) {
	if txn == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.totalReceived++
	b.txns.PushBack(txn)
	//b.cond.Broadcast()
}

func (b *Backend) insertFront(txn *message.Transaction) {
	if txn == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.txns.PushFront(txn)
	//b.cond.Broadcast()
}

func (b *Backend) size() int {
	return b.txns.Len()
}

func (b *Backend) front() *message.Transaction {
	if b.size() == 0 {
		return nil
	}
	ele := b.txns.Front()
	if ele == nil {
		return nil
	}
	val, ok := ele.Value.(*message.Transaction)
	if !ok {
		return nil
	}
	b.txns.Remove(ele)
	return val
}

func (b *Backend) remove(ele *list.Element) {
	//b.mu.Lock()
	//defer b.mu.Unlock()
	if ele == nil {
		return
	}
	b.txns.Remove(ele)
}

func (b *Backend) some(n int) []*message.Transaction {
	var batchSize int
	b.mu.Lock()
	defer b.mu.Unlock()
	// trying to get full size
	for i := 0; i < 1; i++ {
		batchSize = b.size()
		if !config.GetConfig().Fixed && batchSize < n {
			break
		}
		log.Debugf("has %v remaining tx in the mempool", batchSize)
		if batchSize >= n {
			batchSize = n
			break
		}
		//b.cond.Wait()
	}
	batch := make([]*message.Transaction, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		tx := b.front()
		//if tx == nil || b.Contains(tx.ID) {
		//	continue
		//}
		//i++
		batch = append(batch, tx)
	}
	return batch
}
