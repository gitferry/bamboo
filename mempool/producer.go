package mempool

import (
	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/crypto"
	"github.com/gitferry/zeitgeber/message"
	"github.com/gitferry/zeitgeber/types"
)

type Producer struct {
	mempool *MemPool
}

func NewProducer() *Producer {
	return &Producer{mempool: NewMemPool()}
}

func (pd *Producer) ProduceBlock(view types.View, qc *blockchain.QC) *blockchain.Block {
	payload := pd.mempool.GetPayload()
	block := blockchain.MakeBlock(view, qc, payload)
	return block
}

func (pd *Producer) CollectTxn(txn *message.Transaction) {
	pd.mempool.Add(txn)
}

func (pd *Producer) RemoveTxn(id crypto.Identifier) {
	pd.mempool.Rem(id)
}
