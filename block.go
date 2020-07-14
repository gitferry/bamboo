package zeitgeber

import (
	"github.com/gitferry/zeitgeber/crypto"
)

type Block struct {
	Height     int
	ProposerID int
	ID         crypto.Identifier
	PrevID     crypto.Identifier
	Committee  map[int]int // mapping from the proposer id and its weight
	Ttl        int
}

// NewBlock creates a new block instance from the wire
func NewBlock(id crypto.Identifier, previd crypto.Identifier, height int, committee map[int]int, ttl int) *Block {
	return &Block{
		ID:        id,
		Height:    height,
		PrevID:    previd,
		Committee: committee,
		Ttl:       ttl,
	}
}

// MakeBlock creates a new block instance from mining
func MakeBlock(previd crypto.Identifier, height int, proposerID int) *Block {
	return &Block{
		Height:     height,
		PrevID:     previd,
		ID:         IdentifierFixture(),
		ProposerID: proposerID,
	}
}
