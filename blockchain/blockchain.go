package blockchain

import (
	"fmt"

	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/crypto"
	"github.com/gitferry/zeitgeber/log"
)

type BlockChain struct {
	highQC  *QC
	forrest *LevelledForest
	quorum  *Quorum
}

func NewBlkTree(n int) *BlockChain {
	bt := new(BlockChain)
	bt.forrest = NewLevelledForest()
	bt.quorum = NewQuorum(n)
	return bt
}

func (bt *BlockChain) AddBlock(block *Block) {
	blockContainer := &BlockContainer{block}
	// TODO: add checks
	bt.forrest.AddVertex(blockContainer)
	err := bt.UpdateHighQC(block.QC)
	if err != nil {
		log.Warningf("found stale qc, view: %v", block.QC.View)
	}
}

func (bt *BlockChain) AddVote(vote *Vote) (bool, *QC) {
	bt.quorum.Add(vote)
	return bt.GenerateQC(vote.View, vote.BlockID)
}

func (bt *BlockChain) GetHighQC() *QC {
	return bt.highQC
}

func (bt *BlockChain) UpdateHighQC(qc *QC) error {
	if qc.View <= bt.highQC.View {
		return fmt.Errorf("cannot update high QC")
	}
	bt.highQC = qc
	return nil
}

func (bt *BlockChain) GenerateQC(view zeitgeber.View, blockID crypto.Identifier) (bool, *QC) {
	if !bt.quorum.SuperMajority(blockID) {
		return false, nil
	}
	sigs, err := bt.quorum.GetSigs(blockID)
	if err != nil {
		log.Warningf("cannot get signatures, %w", err)
		return false, nil
	}
	qc := &QC{
		View:    view,
		BlockID: blockID,
		AggSig:  sigs,
		// TODO: add real sig
		Signature: nil,
	}

	err = bt.UpdateHighQC(qc)
	if err != nil {
		log.Warningf("generated a stale qc, view: %v", qc.View)
	}

	return true, qc
}

func (bt *BlockChain) GenerateProposal(view zeitgeber.View, payload []zeitgeber.Request) *Block {
	return MakeBlock(view, bt.highQC, payload)
}

func (bc *BlockChain) CalForkingRate() float32 {
	var forkingRate float32
	//if bc.Height == 0 {
	//	return 0
	//}
	//total := 0
	//for i := 1; i <= bc.Height; i++ {
	//	total += len(bc.Blocks[i])
	//}
	//
	//forkingrate := float32(bc.Height) / float32(total)
	return forkingRate
}

// MakeForkChoice returns a random highest block hash
func (bc *BlockChain) MakeForkChoice() crypto.Identifier {
	var id crypto.Identifier
	return id
}

// GetLatestBlock randomly returns the highest block
func (bc *BlockChain) GetLatestBlock() *Block {
	var block *Block
	//latestBlockID := zeitgeber.MapRandomKeyGet(bc.Blocks[bc.Height]).(crypto.Identifier)
	//latestBlock := bc.Blocks[bc.Height][latestBlockID]
	return block
}
