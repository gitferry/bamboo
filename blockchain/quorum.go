package blockchain

import (
	"fmt"

	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/crypto"
)

type Vote struct {
	zeitgeber.View
	Voter   zeitgeber.NodeID
	BlockID crypto.Identifier
	crypto.Signature
}

type QC struct {
	View    zeitgeber.View
	BlockID crypto.Identifier
	crypto.AggSig
	crypto.Signature
}

type Quorum struct {
	total int
	votes map[crypto.Identifier]map[zeitgeber.NodeID]*Vote
}

func MakeVote(view zeitgeber.View, voter zeitgeber.NodeID, id crypto.Identifier) *Vote {
	return &Vote{
		View:    view,
		Voter:   voter,
		BlockID: id,
	}
}

func NewQuorum(total int) *Quorum {
	votes := make(map[crypto.Identifier]map[zeitgeber.NodeID]*Vote)
	return &Quorum{
		total: total,
		votes: votes,
	}
}

// Add adds id to quorum ack records
func (q *Quorum) Add(vote *Vote) {
	_, exist := q.votes[vote.BlockID]
	if !exist {
		//	first time of receiving the vote for this block
		q.votes[vote.BlockID] = make(map[zeitgeber.NodeID]*Vote)
	}
	q.votes[vote.BlockID][vote.Voter] = vote
}

// Super majority quorum satisfied
func (q *Quorum) SuperMajority(blockID crypto.Identifier) bool {
	return q.Size(blockID) > q.total*2/3
}

// Size returns ack size for the block
func (q *Quorum) Size(blockID crypto.Identifier) int {
	return len(q.votes[blockID])
}

func (q *Quorum) GetSigs(blockID crypto.Identifier) (crypto.AggSig, error) {
	var sigs crypto.AggSig
	_, exists := q.votes[blockID]
	if !exists {
		return nil, fmt.Errorf("sigs does not exist, id: %x", blockID)
	}
	for _, vote := range q.votes[blockID] {
		sigs = append(sigs, vote.Signature)
	}

	return sigs, nil
}
