package zeitgeber

import (
	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/types"
)

type Safety interface {
	UpdateStateByQC(qc *blockchain.QC) error
	UpdateStateByView(view types.View) error
	CommitRule(qc *blockchain.QC) (bool, *blockchain.Block, error)
	VotingRule(block *blockchain.Block) (bool, error)
	Forkchoice() *blockchain.QC
}
