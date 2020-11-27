package mempool

import (
	"container/list"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
	"sync"
	"time"
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
	return b.txns.Front()
}

func (b *Backend) remove(ele *list.Element) {
	if ele == nil {
		return
	}
	b.txns.Remove(ele)
}

func (b *Backend) some(n int) []*message.Transaction {
	for {
		s := b.size()
		log.Debugf("has %v remaining tx in the mempool", s)
		if s >= n {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	batch := make([]*message.Transaction, 0, n)
	for i := 0; i < n; i++ {
		ele := b.front()
		val, ok := ele.Value.(*message.Transaction)
		if !ok {
			log.Warning("not enough tx to batch, only has %v", len(batch))
			break
		}
		batch = append(batch, val)
		b.remove(ele)
	}
	return batch
}
