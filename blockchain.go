package zeitgeber

import (
	"github.com/gitferry/zeitgeber/crypto"
)

type Blockchain struct {
	Height          int
	GenesisHash     crypto.Identifier
	Blocks          map[int]map[crypto.Identifier]*Block
	CertifiedBlocks []*Block
	CertifiedMap    map[crypto.Identifier]struct{}
	VoteCount       map[crypto.Identifier]int
	VoteMap         map[crypto.Identifier]map[int]struct{}
	orphans         map[crypto.Identifier]*Block
}

func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		Height:  0,
		Blocks:  make(map[int]map[crypto.Identifier]*Block),
		orphans: make(map[crypto.Identifier]*Block),
	}
	genesisBlock := &Block{
		Height: 0,
	}
	var genesisID crypto.Identifier
	copy(genesisID[:], "Go Crystal")
	genesisBlock.ID = genesisID
	bc.GenesisHash = genesisID
	m := make(map[crypto.Identifier]*Block)
	m[genesisBlock.ID] = genesisBlock
	bc.Blocks[0] = m
	bc.VoteCount = make(map[crypto.Identifier]int)
	bc.VoteMap = make(map[crypto.Identifier]map[int]struct{})
	bc.CertifiedBlocks = make([]*Block, 0)
	bc.CertifiedMap = make(map[crypto.Identifier]struct{})
	return bc
}

func (bc *Blockchain) AddBlock(block *Block) {
	// check if it is orphan
	// if block.Height > bc.Height+1 {
	// 	bc.orphans[block.ID] = block
	// 	return false
	// }
	// if block.Height == bc.Height+1 {
	// check orphan block
	// _, exists := bc.Blocks[bc.Height][block.PrevID]
	// if !exists {
	// 	bc.orphans[block.ID] = block
	// 	return false
	// }
	_, exists := bc.Blocks[block.Height][block.ID]
	if !exists {
		bc.Blocks[block.Height] = make(map[crypto.Identifier]*Block)
	}
	bc.Blocks[block.Height][block.ID] = block
	if bc.Height < block.Height {
		bc.Height = block.Height
	}
	// }
	// check orphan block
	// _, exists := bc.Blocks[block.Height-1][block.PrevID]
	// if !exists {
	// 	bc.orphans[block.ID] = block
	// 	return false
	// }
	// store the block
}

// func (bc *Blockchain) CheckOrphans() []*Block {
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

func (bc *Blockchain) CalForkingRate() float32 {
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
func (bc *Blockchain) MakeForkChoice(isCertified bool) crypto.Identifier {
	if !isCertified {
		return MapRandomKeyGet(bc.Blocks[bc.Height]).(crypto.Identifier)
	}
	return bc.CertifiedBlocks[len(bc.CertifiedBlocks)-1].ID
}

// GetLatestBlock randomly returns the highest block
func (bc *Blockchain) GetLatestBlock() *Block {
	latestBlockID := MapRandomKeyGet(bc.Blocks[bc.Height]).(crypto.Identifier)
	latestBlock := bc.Blocks[bc.Height][latestBlockID]
	return latestBlock
}
