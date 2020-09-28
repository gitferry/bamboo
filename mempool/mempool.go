package mempool

import (
	"time"

	"github.com/gitferry/bamboo/message"
)

type MemPool struct {
	*Backend
}

// NewTransactions creates a new memory pool for transactions.
func NewMemPool() *MemPool {
	mp := &MemPool{
		Backend: NewBackend(),
	}

	return mp
}

// Add adds a transaction to the mempool.
func (mp *MemPool) Add(tx *message.Transaction) {
	mp.Backend.Add(tx)
}

// ByID returns the transaction with the given ID from the mempool.
func (mp *MemPool) ByID(txID string) (*message.Transaction, error) {
	txn, err := mp.Backend.ByID(txID)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func (mp *MemPool) GetTimestamp(txID string) time.Time {
	t := mp.Backend.GetTimestamp(txID)
	return t
}

// All returns all transactions from the mempool.
func (mp *MemPool) GetPayload() []*message.Transaction {
	txns := mp.Backend.All()
	return txns
}
