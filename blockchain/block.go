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

type rawBlock struct {
	types.View
	QC       *QC
	Proposer identity.NodeID
	Payload  []string
	PrevID   crypto.Identifier
	Sig      crypto.Signature
	ID       crypto.Identifier
}

// MakeBlock creates an unsigned block
func MakeBlock(view types.View, qc *QC, payload []*message.Transaction, proposer identity.NodeID) *Block {
	b := new(Block)
	b.View = view
	b.Proposer = proposer
	b.QC = qc
	b.Payload = payload
	b.PrevID = qc.BlockID
	b.makeID()
	return b
}

func (b *Block) makeID() {
	raw := &rawBlock{
		View:     b.View,
		QC:       b.QC,
		Proposer: b.Proposer,
		PrevID:   b.PrevID,
		Sig:      b.Sig,
	}
	var payloadIDs []string
	for _, txn := range b.Payload {
		payloadIDs = append(payloadIDs, txn.ID)
	}
	raw.Payload = payloadIDs
	b.ID = crypto.MakeID(raw)
}
