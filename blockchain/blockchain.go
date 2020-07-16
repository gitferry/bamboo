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
	bc := new(BlockChain)
	bc.forrest = NewLevelledForest()
	bc.quorum = NewQuorum(n)
	return bc
}

func (bc *BlockChain) AddBlock(block *Block) {
	blockContainer := &BlockContainer{block}
	// TODO: add checks
	bc.forrest.AddVertex(blockContainer)
	err := bc.UpdateHighQC(block.QC)
	if err != nil {
		log.Warningf("found stale qc, view: %v", block.QC.View)
	}
}

func (bc *BlockChain) AddVote(vote *Vote) (bool, *QC) {
	bc.quorum.Add(vote)
	return bc.GenerateQC(vote.View, vote.BlockID)
}

func (bc *BlockChain) GetHighQC() *QC {
	return bc.highQC
}

func (bc *BlockChain) UpdateHighQC(qc *QC) error {
	if qc.View <= bc.highQC.View {
		return fmt.Errorf("cannot update high QC")
	}
	bc.highQC = qc
	return nil
}

func (bc *BlockChain) GenerateQC(view zeitgeber.View, blockID crypto.Identifier) (bool, *QC) {
	if !bc.quorum.SuperMajority(blockID) {
		return false, nil
	}
	sigs, err := bc.quorum.GetSigs(blockID)
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

	err = bc.UpdateHighQC(qc)
	if err != nil {
		log.Warningf("generated a stale qc, view: %v", qc.View)
	}

	return true, qc
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
