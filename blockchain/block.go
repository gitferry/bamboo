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
	payload *Payload
}

type Payload struct {
	microblockList []*MicroBlock
}

type MicroBlock struct {
	ProposalID      crypto.Identifier
	Hash            crypto.Identifier
	Txns            []*message.Transaction
	Timestamp       time.Time
	FutureTimestamp time.Time
	Sender          identity.NodeID
	IsRequested     bool
}

type Proposal struct {
	BlockHeader
	HashList []crypto.Identifier
}

type PendingBlock struct {
	payload    *Payload // microblocks that already exist
	Proposal   *Proposal
	missingMap map[crypto.Identifier]struct{} // missing list
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

func NewPayload(microblockList []*MicroBlock) *Payload {
	return &Payload{microblockList: microblockList}
}

func (b *Block) MicroblockList() []*MicroBlock {
	return b.payload.microblockList
}

func (pl *Payload) GenerateHashList() []crypto.Identifier {
	hashList := make([]crypto.Identifier, len(pl.microblockList))
	for _, mb := range pl.microblockList {
		hashList = append(hashList, mb.Hash)
	}
	return hashList
}

func (pl *Payload) addMicroblock(mb *MicroBlock) {
	pl.microblockList = append(pl.microblockList, mb)
}

func (pl *Payload) LastItem() *MicroBlock {
	return pl.microblockList[len(pl.microblockList)-1]
}

// BuildBlock fills microblocks to make a block
func BuildBlock(proposal *Proposal, payload *Payload) *Block {
	return &Block{
		BlockHeader: proposal.BlockHeader,
		payload:     payload,
	}
}

func NewMicroblock(proposalID crypto.Identifier, txnList []*message.Transaction) {
	mb := new(MicroBlock)
	mb.ProposalID = proposalID
	mb.Txns = txnList
	mb.Hash = mb.hash()
}

func NewPendingBlock(proposal *Proposal, missingMap map[crypto.Identifier]struct{}, microblockList []*MicroBlock) *PendingBlock {
	return &PendingBlock{
		Proposal:   proposal,
		missingMap: missingMap,
		payload:    &Payload{microblockList: microblockList},
	}
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

func (mb *MicroBlock) hash() crypto.Identifier {
	return crypto.MakeID(mb.Txns)
}

func (pd *PendingBlock) AddMicroblock(mb *MicroBlock) *Block {
	_, exists := pd.missingMap[mb.Hash]
	if exists {
		pd.payload.addMicroblock(mb)
		delete(pd.missingMap, mb.Hash)
	}
	if len(pd.missingMap) == 0 {
		return BuildBlock(pd.Proposal, pd.payload)
	}
	return nil
}

func (pd *PendingBlock) CompleteBlock() *Block {
	if len(pd.missingMap) == 0 {
		return BuildBlock(pd.Proposal, pd.payload)
	}
	return nil
}

func (pd *PendingBlock) MissingCount() int {
	return len(pd.missingMap)
}

func (pd *PendingBlock) MissingMBList() []crypto.Identifier {
	missingList := make([]crypto.Identifier, pd.MissingCount())
	for k, _ := range pd.missingMap {
		missingList = append(missingList, k)
	}
	return missingList
}
