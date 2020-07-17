package zeitgeber

import (
	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/crypto"
)

type Safety interface {
	UpdateStateByQC(qc *blockchain.QC) error
	UpdateStateByView(view View) error
	CommitRule(qc *blockchain.QC) (bool, crypto.Identifier)
	VotingRule(block *blockchain.Block) bool
}
