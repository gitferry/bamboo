package blockchain

import (
	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/crypto"
)

type PendingBlkTree struct {
	blocks []map[crypto.Identifier]*Block
}

type PendingVotes struct {
	votes map[crypto.Identifier]map[zeitgeber.NodeID]*Vote
}

func NewPendingBlkTree() *PendingBlkTree {
	blocks := make([]map[crypto.Identifier]*Block, 0)
	return &PendingBlkTree{
		blocks: blocks,
	}
}

func NewPendingVotes() *PendingVotes {
	votes := make(map[crypto.Identifier]map[zeitgeber.NodeID]*Vote)
	return &PendingVotes{
		votes: votes,
	}
}

func (pbt *PendingBlkTree) AddBlock() error {

}
