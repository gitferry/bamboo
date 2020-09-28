package blockchain

import (
	"fmt"
	"sync"

	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/types"
)

type BlockChain struct {
	highQC  *QC
	forrest *LevelledForest
	quorum  *Quorum
	// measurement
	highestComitted       int
	committedBlocks       int
	honestCommittedBlocks int
	mu                    sync.Mutex
}

func NewBlockchain(n int) *BlockChain {
	bc := new(BlockChain)
	bc.forrest = NewLevelledForest()
	bc.quorum = NewQuorum(n)
	bc.highQC = &QC{View: 0}
	return bc
}

func (bc *BlockChain) AddBlock(block *Block) {
	blockContainer := &BlockContainer{block}
	// TODO: add checks
	//bc.mu.Lock()
	bc.forrest.AddVertex(blockContainer)
	//bc.mu.Unlock()
	err := bc.UpdateHighQC(block.QC)
	if err != nil {
		log.Debugf("found stale qc, view: %v", block.QC.View)
	}
}

func (bc *BlockChain) AddVote(vote *Vote) (bool, *QC) {
	bc.quorum.Add(vote)
	return bc.GenerateQC(vote.View, vote.BlockID)
}

func (bc *BlockChain) GetHighQC() *QC {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.highQC
}

func (bc *BlockChain) UpdateHighQC(qc *QC) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if qc.View < bc.highQC.View {
		return fmt.Errorf("cannot update high QC")
	}
	bc.highQC = qc
	return nil
}

func (bc *BlockChain) GenerateQC(view types.View, blockID crypto.Identifier) (bool, *QC) {
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

	//err = bc.UpdateHighQC(qc)
	//if err != nil {
	//	log.Warningf("generated a stale qc, view: %v", qc.View)
	//}

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

func (bc *BlockChain) GetBlockByID(id crypto.Identifier) (*Block, error) {
	//bc.mu.Lock()
	//defer bc.mu.Unlock()
	vertex, exists := bc.forrest.GetVertex(id)
	if !exists {
		return nil, fmt.Errorf("the block does not exist, id: %x", id)
	}
	return vertex.GetBlock(), nil
}

func (bc *BlockChain) GetParentBlock(id crypto.Identifier) (*Block, error) {
	//bc.mu.Lock()
	//defer bc.mu.Unlock()
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
	committedView := vertex.GetBlock().View
	bc.highestComitted = int(vertex.GetBlock().View)
	var committedBlocks []*Block
	for block := vertex.GetBlock(); uint64(block.View) > bc.forrest.LowestLevel; {
		committedBlocks = append(committedBlocks, block)
		bc.committedBlocks++
		if !config.Configuration.IsByzantine(block.Proposer) {
			bc.honestCommittedBlocks++
		}
		vertex, exists := bc.forrest.GetVertex(block.PrevID)
		if !exists {
			break
		}
		block = vertex.GetBlock()
	}
	err := bc.forrest.PruneUpToLevel(uint64(committedView))
	if err != nil {
		return nil, fmt.Errorf("cannot prune the blockchain to the committed block, id: %w", err)
	}
	return committedBlocks, nil
}

func (bc *BlockChain) GetChildrenBlocks(id crypto.Identifier) []*Block {
	//bc.mu.Lock()
	//defer bc.mu.Unlock()
	var blocks []*Block
	iterator := bc.forrest.GetChildren(id)
	for I := iterator; I.HasNext(); {
		blocks = append(blocks, I.NextVertex().GetBlock())
	}
	return blocks
}

func (bc *BlockChain) GetChainGrowth() float64 {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return float64(bc.honestCommittedBlocks) / float64(bc.highestComitted)
}

func (bc *BlockChain) GetChainQuality() float64 {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return float64(bc.honestCommittedBlocks) / float64(bc.committedBlocks)
}

func (bc *BlockChain) GetHighestComitted() int {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.highestComitted
}

func (bc *BlockChain) GetHonestCommittedBlocks() int {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.honestCommittedBlocks
}

func (bc *BlockChain) GetCommittedBlocks() int {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.committedBlocks
}

func (bc *BlockChain) GetBlockByView(view types.View) *Block {
	//bc.mu.Lock()
	//defer bc.mu.Unlock()
	iterator := bc.forrest.GetVerticesAtLevel(uint64(view))
	return iterator.next.GetBlock()
}
