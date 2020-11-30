package mempool

import (
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/message"
	"sync"
)

type Producer struct {
	mempool    *MemPool
	receivedTx map[string]*message.Transaction
	mu         sync.Mutex
}

func NewProducer() *Producer {
	return &Producer{
		mempool:    NewMemPool(),
		receivedTx: make(map[string]*message.Transaction),
	}
}

func (pd *Producer) GeneratePayload() []*message.Transaction {
	return pd.mempool.some(config.Configuration.BSize)
}

func (pd *Producer) AddTxn(txn *message.Transaction) {
	pd.mempool.addNew(txn)
}

func (pd *Producer) CollectTxn(txn *message.Transaction) {
	pd.mempool.addOld(txn)
}

func (pd *Producer) RemoveTxn(txn *message.Transaction) {
	pd.mempool.removeTxn(txn)
}

func (pd *Producer) ReceiveTxFromClient(txn *message.Transaction) {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	pd.receivedTx[txn.ID] = txn
}

func (pd *Producer) GetAndRmTxByID(id string) (*message.Transaction, bool) {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	tx, ok := pd.receivedTx[id]
	if ok {
		delete(pd.receivedTx, id)
	}
	return tx, ok
}
