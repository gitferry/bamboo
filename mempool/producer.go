package mempool

import (
	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/message"
	"github.com/gitferry/zeitgeber/types"
)

type Producer struct {
	mempool *MemPool
}

func (pd *Producer) ProduceBlock(view types.View, qc *blockchain.QC) *blockchain.Block {
	block := blockchain.MakeBlock(view, qc, pd.mempool.GetPayload())
	return block
}

func (pd *Producer) CollectTxn(request message.Request) {
	pd.mempool.StoreTxn(request)
}
