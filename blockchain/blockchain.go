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

func (bc *BlockChain) GetParentBlock(id crypto.Identifier) (*Block, error) {
	vertex, exists := bc.forrest.GetVertex(id)
	if !exists {
		return nil, fmt.Errorf("the block does not exist, id: %x", id)
	}
	parentID, _ := vertex.Parent()
	parentVertex, exists := bc.forrest.GetVertex(parentID)
	if !exists {
		return nil, fmt.Errorf("parent block does not exist, id: %x", parentID)
	}
	return parentVertex.GetBlock(), nil
}

func (bc *BlockChain) GetGrandParentBlock(id crypto.Identifier) (*Block, error) {
	parentBlock, err := bc.GetParentBlock(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get parent block: %w", err)
	}
	return bc.GetParentBlock(parentBlock.ID)
}

// CommitBlock prunes blocks and returns committed blocks up to the last committed one
func (bc *BlockChain) CommitBlock(id crypto.Identifier) ([]*Block, error) {
	vertex, ok := bc.forrest.GetVertex(id)
	if !ok {
		return nil, fmt.Errorf("cannot find the block, id: %x", id)
	}
	committedNo := vertex.Level() - bc.forrest.LowestLevel
	committedBlocks := make([]*Block, committedNo)
	for i := uint64(0); i < committedNo; i++ {
		committedBlocks = append(committedBlocks, vertex.GetBlock())
		parentID, _ := vertex.Parent()
		parentVertex, exists := bc.forrest.GetVertex(parentID)
		if !exists {
			return nil, fmt.Errorf("cannot find the parent block, id: %x", parentID)
		}
		vertex = parentVertex
	}
	err := bc.forrest.PruneUpToLevel(vertex.Level())
	if err != nil {
		return nil, fmt.Errorf("cannot prune the blockchain to the committed block, id: %w", err)
	}
	return committedBlocks, nil
}
