package mempool

import (
	"container/list"
	"github.com/gitferry/bamboo/message"
	"sync"
)

type Backend struct {
	txns *list.List
	mu   sync.Mutex
}

func NewBackend() *Backend {
	return &Backend{
		txns: list.New(),
	}
}

func (b *Backend) insertBack(txn *message.Transaction) {
	if txn == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
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
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.txns.Len()
}

func (b *Backend) front() *list.Element {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.txns.Front()
}

func (b *Backend) remove(ele *list.Element) {
	if ele == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.txns.Remove(ele)
}

func (b *Backend) some(n int) []*message.Transaction {
	batch := make([]*message.Transaction, 0)
	var batchSize int
	if n <= b.size() {
		batchSize = n
	} else {
		batchSize = b.size()
	}
	for i := 0; i < batchSize; i++ {
		ele := b.front()
		val, ok := ele.Value.(*message.Transaction)
		if !ok {
			break
		}
		batch = append(batch, val)
		b.remove(ele)
	}
	return batch
}
