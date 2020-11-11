package mempool

import (
	"fmt"
	"sync"
	"time"

	"github.com/gitferry/bamboo/message"
)

type TxnRecord struct {
	txn          *message.Transaction
	receivedTime time.Time
}

type Backdata struct {
	txns map[string]*TxnRecord
}

func NewBackdata() *Backdata {
	return &Backdata{
		txns: make(map[string]*TxnRecord),
	}
}

// Has checks if we already contain the item with the given hash.
func (b *Backdata) Has(id string) bool {
	_, ok := b.txns[id]
	return ok
}

// Add adds the given item to the pool.
func (b *Backdata) Add(txn *message.Transaction) {
	_, ok := b.txns[txn.ID]
	if ok {
		return
	}
	b.txns[txn.ID] = &TxnRecord{
		txn:          txn,
		receivedTime: time.Now(),
	}
}

// Rem will remove the item with the given hash.
func (b *Backdata) Rem(id string) bool {
	_, ok := b.txns[id]
	if !ok {
		return false
	}
	delete(b.txns, id)
	return true
}

// ByID returns the given item from the pool.
func (b *Backdata) ByID(id string) (*message.Transaction, error) {
	_, ok := b.txns[id]
	if !ok {
		return nil, fmt.Errorf("transaction does not exist, id: %x", id)
	}
	coll := b.txns[id]
	return coll.txn, nil
}

// Size will return the size of the backend.
func (b *Backdata) Size() uint {
	return uint(len(b.txns))
}

// All returns all entities from the pool.
func (b *Backdata) All() []*message.Transaction {
	entities := make([]*message.Transaction, 0, len(b.txns))
	for _, item := range b.txns {
		entities = append(entities, item.txn)
	}
	return entities
}

// Some returns a certain amount of entities from the pool.
func (b *Backdata) Some(size int) []*message.Transaction {
	if len(b.txns) < size {
		return b.All()
	}
	entities := make([]*message.Transaction, 0, size)
	for _, item := range b.txns {
		if len(entities) == size {
			break
		}
		entities = append(entities, item.txn)
	}
	return entities
}

// Backend provides synchronized access to a backend
type Backend struct {
	*Backdata
	mu   *sync.RWMutex
	cond *sync.Cond
}

// NewBackend creates a new memory pool backend.
func NewBackend() *Backend {
	mu := &sync.RWMutex{}
	b := &Backend{
		Backdata: NewBackdata(),
		mu:       mu,
		cond:     sync.NewCond(mu),
	}
	return b
}

// Has checks if we already contain the item with the given hash.
func (b *Backend) Has(id string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Backdata.Has(id)
}

// Add adds the given item to the pool.
func (b *Backend) Add(txn *message.Transaction) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Backdata.Add(txn)
	b.cond.Broadcast()
}

// Rem will remove the item with the given hash.
func (b *Backend) RemTxns(txns []*message.Transaction) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	ok := true
	for _, item := range txns {
		if !b.Backdata.Rem(item.ID) {
			ok = false
		}
	}
	return ok
}

// ByID returns the given item from the pool.
func (b *Backend) ByID(id string) (*message.Transaction, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Backdata.ByID(id)
}

func (b *Backend) GetTimestamp(id string) time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Backdata.txns[id].receivedTime
}

// Run fetches the given item from the pool and runs given function on it, returning the entity after
func (b *Backend) Run(f func(backdata *Backdata) error) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	err := f(b.Backdata)
	if err != nil {
		return err
	}
	return nil
}

// Size will return the size of the backend.
func (b *Backend) Size() uint {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Backdata.Size()
}

// All returns all transactions from the pool.
func (b *Backend) All() []*message.Transaction {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Backdata.All()
}

// Some returns a certain amount of transactions from the pool.
func (b *Backend) Some(size int) []*message.Transaction {
	//for i := 0; i < 10; i++ {
	//	if b.Size() > uint(size) {
	//		b.mu.RLock()
	//		defer b.mu.RUnlock()
	//		return b.Backdata.Some(size)
	//	}
	//	time.Sleep(1 * time.Millisecond)
	//}
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Backdata.Some(size)
}
