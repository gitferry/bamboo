package blockchain

import (
	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/crypto"
)

type BlockChain struct {
	highQC  *QC
	forrest *LevelledForest
	votes   *PendingVotes
}

func NewBlkTree() *BlockChain {
	bt := new(BlockChain)
	bt.forrest = NewLevelledForest()
	bt.votes = NewPendingVotes()
	return bt
}

func (bt *BlockChain) AddBlock(block *Block) error {
	blockContainer := &BlockContainer{block}
	// TODO: add checks
	bt.forrest.AddVertex(blockContainer)
	// TODO: update consensus state
	return nil
}

func (bt *BlockChain) ProcessVote(vote *Vote) error {
	return nil
}

func (bt *BlockChain) GenerateProposal(view zeitgeber.View, payload []zeitgeber.Request) *Block {
	return MakeBlock(view, bt.highQC, payload)
}

// func (bc *BlockChain) CheckOrphans() []*Block {
// 	var unvotedBlocks []*Block
// 	if len(bc.orphans) == 0 {
// 		return unvotedBlocks
// 	}
// 	finished := false
// 	for !finished {
// 		finished = true
// 		for id, block := range bc.orphans {
// 			shouldVote := bc.AddBlock(block)
// 			if shouldVote {
// 				unvotedBlocks = append(unvotedBlocks, block)
// 				delete(bc.orphans, id)
// 				finished = false
// 				break
// 			}
// 		}
// 	}
// 	return unvotedBlocks
// }

func (bc *BlockChain) CalForkingRate() float32 {
	if bc.Height == 0 {
		return 0
	}
	total := 0
	for i := 1; i <= bc.Height; i++ {
		total += len(bc.Blocks[i])
	}

	forkingrate := float32(bc.Height) / float32(total)
	return forkingrate
}

// MakeForkChoice returns a random highest block hash
func (bc *BlockChain) MakeForkChoice(isCertified bool) crypto.Identifier {
	if !isCertified {
		return zeitgeber.MapRandomKeyGet(bc.Blocks[bc.Height]).(crypto.Identifier)
	}
	return bc.CertifiedBlocks[len(bc.CertifiedBlocks)-1].ID
}

// GetLatestBlock randomly returns the highest block
func (bc *BlockChain) GetLatestBlock() *Block {
	latestBlockID := zeitgeber.MapRandomKeyGet(bc.Blocks[bc.Height]).(crypto.Identifier)
	latestBlock := bc.Blocks[bc.Height][latestBlockID]
	return latestBlock
}
