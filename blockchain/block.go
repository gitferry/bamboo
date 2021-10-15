package blockchain

import (
	"time"

	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/message"
	"github.com/gitferry/bamboo/types"
)

type BlockHeader struct {
	types.View
	QC        *QC
	Proposer  identity.NodeID
	Timestamp time.Time
	PrevID    crypto.Identifier
	Sig       crypto.Signature
	ID        crypto.Identifier
	Ts        time.Duration
}

type Block struct {
	BlockHeader
	Payload []*MicroBlock
}

type MicroBlock struct {
	Txns []*message.Transaction
}

type Proposal struct {
	BlockHeader
	HashList []crypto.Identifier
}

type rawProposal struct {
	types.View
	QC       *QC
	Proposer identity.NodeID
	Payload  []crypto.Identifier
	PrevID   crypto.Identifier
}

// BuildProposal creates a signed proposal
func BuildProposal(view types.View, qc *QC, prevID crypto.Identifier, payload []crypto.Identifier, proposer identity.NodeID) *Proposal {
	p := new(Proposal)
	p.View = view
	p.Proposer = proposer
	p.QC = qc
	p.HashList = payload
	p.PrevID = prevID
	p.makeID(proposer)
	return p
}

func (p *Proposal) makeID(nodeID identity.NodeID) {
	raw := &rawProposal{
		View:     p.View,
		QC:       p.QC,
		Proposer: p.Proposer,
		Payload:  p.HashList,
		PrevID:   p.PrevID,
	}
	p.ID = crypto.MakeID(raw)
	p.Sig, _ = crypto.PrivSign(crypto.IDToByte(p.ID), nodeID, nil)
}

func (mb *MicroBlock) ID() crypto.Identifier {
	return crypto.MakeID(mb.Txns)
}
