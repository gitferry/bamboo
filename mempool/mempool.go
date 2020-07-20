package mempool

import (
	"github.com/gitferry/zeitgeber/crypto"
	"github.com/gitferry/zeitgeber/message"
)

type MemPool struct {
	*Backend
}

// NewTransactions creates a new memory pool for transactions.
func NewMemPool() (*MemPool, error) {
	mp := &MemPool{
		Backend: NewBackend(),
	}

	return mp, nil
}

// Add adds a transaction to the mempool.
func (mp *MemPool) Add(tx *message.Transaction) {
	mp.Backend.Add(tx)
}

// ByID returns the transaction with the given ID from the mempool.
func (mp *MemPool) ByID(txID crypto.Identifier) (*message.Transaction, error) {
	txn, err := mp.Backend.ByID(txID)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

// All returns all transactions from the mempool.
func (mp *MemPool) GetPayload() []*message.Transaction {
	txns := mp.Backend.All()
	return txns
}
