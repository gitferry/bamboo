package blockchain

import (
	"fmt"

	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/types"
)

type Vote struct {
	types.View
	Voter   identity.NodeID
	BlockID crypto.Identifier
	crypto.Signature
}

type QC struct {
	View    types.View
	BlockID crypto.Identifier
	crypto.AggSig
	crypto.Signature
}

type Quorum struct {
	total int
	votes map[crypto.Identifier]map[identity.NodeID]*Vote
}

func MakeVote(view types.View, voter identity.NodeID, id crypto.Identifier) *Vote {
	return &Vote{
		View:    view,
		Voter:   voter,
		BlockID: id,
	}
}

func NewQuorum(total int) *Quorum {
	return &Quorum{
		total: total,
		votes: make(map[crypto.Identifier]map[identity.NodeID]*Vote),
	}
}

// Add adds id to quorum ack records
func (q *Quorum) Add(vote *Vote) (bool, *QC) {
	if q.superMajority(vote.BlockID) {
		return false, nil
	}
	_, exist := q.votes[vote.BlockID]
	if !exist {
		//	first time of receiving the vote for this block
		q.votes[vote.BlockID] = make(map[identity.NodeID]*Vote)
	}
	q.votes[vote.BlockID][vote.Voter] = vote
	if q.superMajority(vote.BlockID) {
		qc := &QC{
			View:    vote.View,
			BlockID: vote.BlockID,
		}
		return true, qc
	}
	return false, nil
}

// Super majority quorum satisfied
func (q *Quorum) superMajority(blockID crypto.Identifier) bool {
	return q.size(blockID) > q.total*2/3
}

// Size returns ack size for the block
func (q *Quorum) size(blockID crypto.Identifier) int {
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
