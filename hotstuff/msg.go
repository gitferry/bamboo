package hotstuff

import "github.com/gitferry/zeitgeber"

type ProposalMsg struct {
	NodeID zeitgeber.ID
	block  *Block
}
