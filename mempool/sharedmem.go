package mempool

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/message"
)

type SharedMempool interface {
	// AddTxn adds a new transaction and returns a microblock if sufficient
	// transactions are received
	AddTxn(tx *message.Transaction) (*blockchain.MicroBlock, error)

	// AddMicroblock pushes a new microblock into a FIFO queue
	AddMicroblock(mb *blockchain.MicroBlock) error

	// GeneratePayload pulls a hash list of microblocks from the queue,
	GeneratePayload() (bool, []crypto.Identifier)

	// CheckExistence checks if the microblocks contained in the proposal
	// exists, and return a hash list of the missing ones
	CheckExistence(p *blockchain.Proposal) []crypto.Identifier

	// RemoveMicroBlock removes the referred microblock
	RemoveMicroBlock(id crypto.Identifier) error
}
