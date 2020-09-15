package mempool

import (
	"time"

	"github.com/gitferry/zeitgeber/blockchain"
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
	payload := pd.mempool.GetPayload()
	for len(payload) == 0 {
		payload = pd.mempool.GetPayload()
		time.Sleep(10 * time.Millisecond)
	}
	block := blockchain.MakeBlock(view, qc, payload, proposer, ts)
	return block
}

func (pd *Producer) CollectTxn(txn *message.Transaction) {
	pd.mempool.Add(txn)
}

func (pd *Producer) RemoveTxn(id string) {
	pd.mempool.Rem(id)
}
