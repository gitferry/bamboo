package blockchain

import (
	"github.com/gitferry/zeitgeber/crypto"
	"github.com/gitferry/zeitgeber/identity"
	"github.com/gitferry/zeitgeber/message"
	"github.com/gitferry/zeitgeber/types"
)

type Block struct {
	types.View
	QC       *QC
	Proposer identity.NodeID
	Payload  []*message.Transaction
	PrevID   crypto.Identifier
	Sig      crypto.Signature
	ID       crypto.Identifier
}

// MakeBlock creates an unsigned block
func MakeBlock(view types.View, qc *QC, payload []*message.Transaction) *Block {
	b := new(Block)
	b.View = view
	b.QC = qc
	b.Payload = payload
	b.PrevID = qc.BlockID
	b.ID = crypto.MakeID(b)
	return b
}
