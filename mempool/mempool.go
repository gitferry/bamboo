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
	mp.Backend.insertBack(tx)
}

func (mp *MemPool) addOld(tx *message.Transaction) {
	mp.Backend.insertFront(tx)
}

func (mp *MemPool) removeTxn(tx *message.Transaction) {
	s := mp.size()
	for i := 0; i < s; i++ {
		e := mp.front()
		val, _ := e.Value.(*message.Transaction)
		if val.ID == tx.ID {
			mp.remove(e)
		}
	}
}
