package replica

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/pacemaker"
	"github.com/gitferry/bamboo/types"
)

type Safety interface {
	ProcessBlock(block *blockchain.Block) error
	ProcessVote(vote *blockchain.Vote)
	ProcessRemoteTmo(tmo *pacemaker.TMO)
	ProcessLocalTmo(view types.View)
	MakeProposal(view types.View, payload []crypto.Identifier) *blockchain.Proposal
	GetChainStatus() string
}
