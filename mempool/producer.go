package mempool

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/types"
)

type Producer struct {
	mempool *MemPool
}

func NewProducer() *Producer {
	return &Producer{mempool: NewMemPool()}
}

func (pd *Producer) ProduceBlock(view types.View, qc *blockchain.QC, proposer identity.NodeID) *blockchain.Block {
	var payload []*message.Transaction
	payload = pd.mempool.Some(config.Configuration.BSize)
	block := blockchain.MakeBlock(view, qc, qc.BlockID, payload, proposer)
	pd.mempool.Backend.RemTxns(payload)
	return block
}

func (pd *Producer) GeneratePayload() []*message.Transaction {
	var payload []*message.Transaction
	payload = pd.mempool.Some(config.Configuration.BSize)
	//payload = pd.mempool.All()
	pd.mempool.Backend.RemTxns(payload)
	return payload
}

func (pd *Producer) CollectTxn(txn *message.Transaction) {
	pd.mempool.Add(txn)
}

func (pd *Producer) RemoveTxns(txns []*message.Transaction) {
	pd.mempool.Backend.RemTxns(txns)
}
