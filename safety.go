package zeitgeber

import (
	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/crypto"
)

type Safety interface {
	UpdateState(qc *blockchain.QC)
	CommitRule(qc *blockchain.QC) (bool, crypto.Identifier)
	VotingRule(block *blockchain.Block) bool
}
