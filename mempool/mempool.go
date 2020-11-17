package mempool

import (
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

func (mp *MemPool) addNew(tx *message.Transaction) {
	//mp.Backend.Add(tx)
	mp.Backend.insertBack(tx)
}
