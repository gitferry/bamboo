package mempool

import (
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/message"
)

type Producer struct {
	mempool *MemPool
}

func NewProducer() *Producer {
	return &Producer{
		mempool: NewMemPool(),
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

func (pd *Producer) TotalReceivedTxNo() int64 {
	return pd.mempool.totalReceived
}
