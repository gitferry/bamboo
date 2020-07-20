package mempool

import (
	"fmt"
	"sync"

	"github.com/gitferry/zeitgeber/crypto"
	"github.com/gitferry/zeitgeber/message"
)

// Backdata implements a generic memory pool backed by a Go map.
type Backdata struct {
	txns map[crypto.Identifier]*message.Transaction
}

func NewBackdata() *Backdata {
	return &Backdata{
		txns: make(map[crypto.Identifier]*message.Transaction),
	}
}

// Has checks if we already contain the item with the given hash.
func (b *Backdata) Has(id crypto.Identifier) bool {
	_, ok := b.txns[id]
	return ok
}

// Add adds the given item to the pool.
func (b *Backdata) Add(txn *message.Transaction) {
	id := crypto.MakeID(txn)
	_, ok := b.txns[id]
	if ok {
		return
	}
	b.txns[id] = txn
}

// Rem will remove the item with the given hash.
func (b *Backdata) Rem(id crypto.Identifier) bool {
	_, ok := b.txns[id]
	if !ok {
		return false
	}
	delete(b.txns, id)
	return true
}

// ByID returns the given item from the pool.
func (b *Backdata) ByID(id crypto.Identifier) (*message.Transaction, error) {
	_, ok := b.txns[id]
	if !ok {
		return nil, fmt.Errorf("transaction does not exist, id: %x", id)
	}
	coll := b.txns[id]
	return coll, nil
}

// Size will return the size of the backend.
func (b *Backdata) Size() uint {
	return uint(len(b.txns))
}

// All returns all entities from the pool.
func (b *Backdata) All() []*message.Transaction {
	entities := make([]*message.Transaction, 0, len(b.txns))
	for _, item := range b.txns {
		entities = append(entities, item)
	}
	return entities
}

// Backend provides synchronized access to a backend
type Backend struct {
	*Backdata
	sync.RWMutex
}

// NewBackend creates a new memory pool backend.
func NewBackend() *Backend {
	b := &Backend{
		Backdata: NewBackdata(),
	}
	return b
}

// Has checks if we already contain the item with the given hash.
func (b *Backend) Has(id crypto.Identifier) bool {
	b.RLock()
	defer b.RUnlock()
	return b.Backdata.Has(id)
}

// Add adds the given item to the pool.
func (b *Backend) Add(txn *message.Transaction) {
	b.Lock()
	defer b.Unlock()
	b.Backdata.Add(txn)
}

// Rem will remove the item with the given hash.
func (b *Backend) Rem(id crypto.Identifier) bool {
	b.Lock()
	defer b.Unlock()
	return b.Backdata.Rem(id)
}

// ByID returns the given item from the pool.
func (b *Backend) ByID(id crypto.Identifier) (*message.Transaction, error) {
	b.RLock()
	defer b.RUnlock()
	return b.Backdata.ByID(id)
}

// Run fetches the given item from the pool and runs given function on it, returning the entity after
func (b *Backend) Run(f func(backdata *Backdata) error) error {
	b.RLock()
	defer b.RUnlock()

	err := f(b.Backdata)
	if err != nil {
		return err
	}
	return nil
}

// Size will return the size of the backend.
func (b *Backend) Size() uint {
	b.RLock()
	defer b.RUnlock()
	return b.Backdata.Size()
}

// All returns all transactions from the pool.
func (b *Backend) All() []*message.Transaction {
	b.RLock()
	defer b.RUnlock()
	return b.Backdata.All()
}
