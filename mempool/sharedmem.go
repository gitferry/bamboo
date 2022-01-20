package mempool

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/message"
)

type SharedMempool interface {
	// AddTxn adds a new transaction and returns a microblock if sufficient
	// transactions are received
	AddTxn(tx *message.Transaction) (bool, *blockchain.MicroBlock)

	// AddMicroblock pushes a new microblock into a FIFO queue
	AddMicroblock(mb *blockchain.MicroBlock) error

	// GeneratePayload pulls a hash list of microblocks from the queue,
	GeneratePayload() *blockchain.Payload

	// CheckExistence checks if the microblocks contained in the proposal
	// exists, and return a hash list of the missing ones
	CheckExistence(p *blockchain.Proposal) (bool, []crypto.Identifier)

	// RemoveMicroBlock removes the referred microblock
	RemoveMicroblock(id crypto.Identifier) error

	// FindMicroblock finds the referred microblock
	FindMicroblock(id crypto.Identifier) (bool, *blockchain.MicroBlock)

	// FillProposal pulls microblocks from the mempool and build a pending block,
	// return missing list if there's any
	FillProposal(p *blockchain.Proposal) *blockchain.PendingBlock

	AddAck(ack *blockchain.Ack)

	IsStable(id crypto.Identifier) bool

	AckList(id crypto.Identifier) []identity.NodeID

	TotalTx() int64

	RemainingTx() int64

	TotalMB() int64

	RemainingMB() int64
}
