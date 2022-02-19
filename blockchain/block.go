package blockchain

import (
	"github.com/gitferry/bamboo/crypto/merkle"
	"github.com/kelindar/bitmap"
	"strconv"
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
	MicroblockList []*MicroBlock
	SigMap         map[crypto.Identifier]map[identity.NodeID]crypto.Signature
}

type MicroBlock struct {
	ProposalID      crypto.Identifier
	Hash            crypto.Identifier
	Txns            []*message.Transaction
	Timestamp       time.Time
	FutureTimestamp time.Time
	Sender          identity.NodeID
	IsRequested     bool
	IsForward       bool
	Bitmap          bitmap.Bitmap
	Hops            int
}

type Proposal struct {
	BlockHeader
	HashList []crypto.Identifier
}

type PendingBlock struct {
	Payload    *Payload // microblocks that already exist
	Proposal   *Proposal
	MissingMap map[crypto.Identifier]struct{} // missing list
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

func NewPayload(microblockList []*MicroBlock, sigs map[crypto.Identifier]map[identity.NodeID]crypto.Signature) *Payload {
	return &Payload{
		MicroblockList: microblockList,
		SigMap:         sigs,
	}
}

func (b *Block) MicroblockList() []*MicroBlock {
	return b.payload.MicroblockList
}

func (pl *Payload) GenerateHashList() []crypto.Identifier {
	hashList := make([]crypto.Identifier, 0)
	for _, mb := range pl.MicroblockList {
		if mb == nil {
			continue
		}
		hashList = append(hashList, mb.Hash)
	}
	return hashList
}

func (pl *Payload) addMicroblock(mb *MicroBlock) {
	pl.MicroblockList = append(pl.MicroblockList, mb)
}

func (pl *Payload) LastItem() *MicroBlock {
	if len(pl.MicroblockList) == 0 {
		return nil
	}
	return pl.MicroblockList[len(pl.MicroblockList)-1]
}

func (mb *MicroBlock) FindSentNodes() []identity.NodeID {
	nodes := make([]identity.NodeID, 0)
	mb.Bitmap.Range(func(x uint32) {
		nodes = append(nodes, identity.NodeID(strconv.Itoa(int(x))))
		return
	})

	return nodes
}

func (mb *MicroBlock) AddSentNodes(nodes []identity.NodeID) {
	for _, id := range nodes {
		mb.Bitmap.Set(uint32(id.Node()))
	}
}

// BuildBlock fills microblocks to make a block
func BuildBlock(proposal *Proposal, payload *Payload) *Block {
	return &Block{
		BlockHeader: proposal.BlockHeader,
		payload:     payload,
	}
}

func NewMicroblock(proposalID crypto.Identifier, txnList []*message.Transaction) *MicroBlock {
	mb := new(MicroBlock)
	mb.ProposalID = proposalID
	mb.Txns = txnList
	mb.Timestamp = time.Now()
	mb.Hash = mb.hash()
	return mb
}

func NewPendingBlock(proposal *Proposal, missingMap map[crypto.Identifier]struct{}, microBlocks []*MicroBlock) *PendingBlock {
	return &PendingBlock{
		Proposal:   proposal,
		MissingMap: missingMap,
		Payload:    &Payload{MicroblockList: microBlocks},
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
	hashList := make([][]byte, 0)
	for _, tx := range mb.Txns {
		hashList = append(hashList, crypto.IDToByte(crypto.MakeID(tx)))
	}
	hashList = append(hashList, []byte(mb.Timestamp.String()))
	return crypto.MakeID(merkle.HashFromByteSlices(hashList))
}

func (pd *PendingBlock) AddMicroblock(mb *MicroBlock) *Block {
	_, exists := pd.MissingMap[mb.Hash]
	if exists {
		pd.Payload.addMicroblock(mb)
		delete(pd.MissingMap, mb.Hash)
	}
	if len(pd.MissingMap) == 0 {
		return BuildBlock(pd.Proposal, pd.Payload)
	}
	return nil
}

func (pd *PendingBlock) CompleteBlock() *Block {
	if len(pd.MissingMap) == 0 {
		return BuildBlock(pd.Proposal, pd.Payload)
	}
	return nil
}

func (pd *PendingBlock) MissingCount() int {
	return len(pd.MissingMap)
}

func (pd *PendingBlock) MissingMBList() []crypto.Identifier {
	missingList := make([]crypto.Identifier, 0)
	for k, _ := range pd.MissingMap {
		missingList = append(missingList, k)
	}
	return missingList
}
