package blockchain

import "github.com/gitferry/bamboo/crypto"

// BlockContainer wraps a block to implement forest.Vertex
// In addition, it holds some additional properties for efficient processing of blocks
// by the Finalizer
type BlockContainer struct {
	Block *Block
}

// functions implementing forest.vertex
func (b *BlockContainer) VertexID() crypto.Identifier { return b.Block.ID }
func (b *BlockContainer) Level() uint64               { return uint64(b.Block.View) }
func (b *BlockContainer) Parent() (crypto.Identifier, uint64) {
	return b.Block.PrevID, uint64(b.Block.QC.View)
}
func (b *BlockContainer) GetBlock() *Block { return b.Block }
