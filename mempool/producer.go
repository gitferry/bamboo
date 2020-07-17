package mempool

import (
	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/blockchain"
)

type Producer struct {
	mempool *MemPool
}

func (pd *Producer) ProduceBlock(view zeitgeber.View, qc *blockchain.QC) *blockchain.Block {
	block := blockchain.MakeBlock(view, qc, pd.mempool.GetPayload())
	return block
}

func (pd *Producer) CollectTxn(request zeitgeber.Request) {
	pd.mempool.StoreTxn(request)
}
