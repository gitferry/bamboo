package hotstuff

import (
	"fmt"

	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/types"
)

type HotStuff struct {
	lastVotedView types.View
	preferredView types.View
	lockedQC      blockchain.QC
	bc            *blockchain.BlockChain
}

func NewHotStuff(blockchain *blockchain.BlockChain) *HotStuff {
	hs := new(HotStuff)
	hs.bc = blockchain
	return hs
}

func (hs *HotStuff) VotingRule(block *blockchain.Block) (bool, error) {
	parentQC, err := hs.bc.GetParentBlock(block.QC.BlockID)
	if err != nil {
		return false, fmt.Errorf("cannot vote for block: %w", err)
	}
	if (block.View <= hs.lastVotedView) || (parentQC.View < hs.preferredView) {
		return false, nil
	}
	return true, nil
}

func (hs *HotStuff) CommitRule(qc *blockchain.QC) (bool, *blockchain.Block, error) {
	grandParentBlock, err := hs.bc.GetGrandParentBlock(qc.BlockID)
	if err != nil {
		return false, nil, fmt.Errorf("cannot commit any block: %w", err)
	}
	parentBlock, err := hs.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return false, nil, fmt.Errorf("cannot commit any block: %w", err)
	}
	if ((grandParentBlock.View + 1) == parentBlock.View) && ((parentBlock.View + 1) == qc.View) {
		return true, grandParentBlock, nil
	}
	return false, nil, nil
}

func (hs *HotStuff) UpdateStateByView(view types.View) error {
	return hs.updateLastVotedView(view)
}

func (hs *HotStuff) UpdateStateByQC(qc *blockchain.QC) error {
	return hs.updatePreferredView(qc)
}

func (hs *HotStuff) updateLastVotedView(targetView types.View) error {
	if targetView < hs.lastVotedView {
		return fmt.Errorf("target view is lower than the last voted view")
	}
	hs.lastVotedView = targetView
	return nil
}

func (hs *HotStuff) updatePreferredView(qc *blockchain.QC) error {
	parentBlock, err := hs.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return fmt.Errorf("cannot update preferred view: %w", err)
	}
	if parentBlock.View < hs.preferredView {
		return fmt.Errorf("qc's parenview is lower than current preferred view")
	}
	hs.preferredView = parentBlock.View
	return nil
}
