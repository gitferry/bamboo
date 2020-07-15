package blockchain

import (
	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/crypto"
)

type Block struct {
	zeitgeber.View
	QC       *QC
	Proposer zeitgeber.NodeID
	Payload  []zeitgeber.Request
	PrevID   crypto.Identifier
	Sig      crypto.Signature
	ID       crypto.Identifier
}

// MakeBlock creates an unsigned block
func MakeBlock(view zeitgeber.View, qc *QC, payload []zeitgeber.Request) *Block {
	b := new(Block)
	b.View = view
	b.QC = qc
	b.Payload = payload
	b.PrevID = qc.BlockID
	b.ID = crypto.MakeID(b)
	return b
}
