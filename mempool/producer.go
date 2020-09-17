package mempool

import (
	"time"

	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/config"
	"github.com/gitferry/zeitgeber/identity"
	"github.com/gitferry/zeitgeber/message"
	"github.com/gitferry/zeitgeber/types"
)

type Producer struct {
	mempool *MemPool
}

func NewProducer() *Producer {
	return &Producer{mempool: NewMemPool()}
}

func (pd *Producer) ProduceBlock(view types.View, qc *blockchain.QC, proposer identity.NodeID, ts time.Duration) *blockchain.Block {
	payload := pd.mempool.Some(config.Configuration.BSize)
	block := blockchain.MakeBlock(view, qc, payload, proposer, ts)
	pd.mempool.Backend.RemTxns(payload)
	return block
}

func (pd *Producer) CollectTxn(txn *message.Transaction) {
	pd.mempool.Add(txn)
}

func (pd *Producer) RemoveTxns(txns []*message.Transaction) {
	pd.mempool.Backend.RemTxns(txns)
}
