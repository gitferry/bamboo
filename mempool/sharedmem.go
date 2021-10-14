package mempool

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/message"
)

type SharedMempool interface {
	AddTxn(tx *message.Transaction)
	AddMicroblock(mb *blockchain.MicroBlock)
	GeneratePayload() [][]byte
	CheckExistence(p *blockchain.Proposal) bool
}
