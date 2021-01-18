package mempool

import (
	"container/list"
	"github.com/gitferry/bamboo/message"
	"sync"
)

type Backend struct {
	txns          *list.List
	limit         int
	totalReceived int64
	*BloomFilter
	mu *sync.Mutex
}

func NewBackend(limit int) *Backend {
	var mu sync.Mutex
	return &Backend{
		txns:        list.New(),
		BloomFilter: NewBloomFilter(),
		mu:          &mu,
		limit:       limit,
	}
}

func (b *Backend) insertBack(txn *message.Transaction) {
	if txn == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.size() > b.limit {
		return
	}
	b.totalReceived++
	b.txns.PushBack(txn)
}

func (b *Backend) insertFront(txn *message.Transaction) {
	if txn == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.txns.PushFront(txn)
}

func (b *Backend) size() int {
	return b.txns.Len()
}

func (b *Backend) front() *message.Transaction {
	if b.size() == 0 {
		return nil
	}
	ele := b.txns.Front()
	val, ok := ele.Value.(*message.Transaction)
	if !ok {
		return nil
	}
	b.txns.Remove(ele)
	return val
}

func (b *Backend) some(n int) []*message.Transaction {
	var batchSize int
	b.mu.Lock()
	defer b.mu.Unlock()
	batchSize = b.size()
	if batchSize >= n {
		batchSize = n
	}
	batch := make([]*message.Transaction, 0, batchSize)
	for i := 0; i < batchSize; i++ {
		tx := b.front()
		batch = append(batch, tx)
	}
	return batch
}

func (b *Backend) addToBloom(id string) {
	b.Add(id)
}
