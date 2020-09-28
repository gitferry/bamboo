package bamboo

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/types"
)

type Safety interface {
	UpdateStateByQC(qc *blockchain.QC) error
	UpdateStateByView(view types.View) error
	CommitRule(qc *blockchain.QC) (bool, *blockchain.Block, error)
	VotingRule(block *blockchain.Block) (bool, error)
	Forkchoice() *blockchain.QC
}
