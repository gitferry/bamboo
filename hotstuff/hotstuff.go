package hotstuff

import (
	"fmt"

	"github.com/gitferry/zeitgeber/blockchain"
	"github.com/gitferry/zeitgeber/config"
	"github.com/gitferry/zeitgeber/log"
	"github.com/gitferry/zeitgeber/types"
)

const (
	FORKING = "forking"
	HIGHEST = "highest"
	LONGEST = "longest"
)

type HotStuff struct {
	lastVotedView  types.View
	preferredView  types.View
	forkchoiceType string
	bc             *blockchain.BlockChain
}

func NewHotStuff(blockchain *blockchain.BlockChain, forkchoice string) *HotStuff {
	hs := new(HotStuff)
	hs.bc = blockchain
	hs.forkchoiceType = forkchoice
	return hs
}

func (hs *HotStuff) VotingRule(block *blockchain.Block) (bool, error) {
	if block.View <= 2 {
		return true, nil
	}
	parentBlock, err := hs.bc.GetParentBlock(block.ID)
	if err != nil {
		return false, fmt.Errorf("cannot vote for block: %w", err)
	}
	if (block.View <= hs.lastVotedView) || (parentBlock.View < hs.preferredView) {
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
	if qc.View <= 2 {
		return nil
	}
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
	if parentBlock.View > hs.preferredView {
		hs.preferredView = parentBlock.View
	}
	return nil
}

func (hs *HotStuff) Forkchoice() *blockchain.QC {
	switch hs.forkchoiceType {
	case FORKING:
		// Byzantine choice
		return hs.forkingForkchoice()
	case HIGHEST:
		return hs.highestForkchoice()
	case LONGEST:
		return hs.longestForkchoice()
	default:
		return hs.highestForkchoice()
	}
}

// forkingForkchoice returns the QC contained in the first honest block after the locked block
func (hs *HotStuff) forkingForkchoice() *blockchain.QC {
	if hs.preferredView <= 1 {
		return hs.bc.GetHighQC()
	}
	highBlock, _ := hs.bc.GetBlockByID(hs.bc.GetHighQC().BlockID)
	if config.Configuration.IsByzantine(highBlock.Proposer) {
		return hs.bc.GetHighQC()
	}
	preferredBlock := hs.bc.GetBlockByView(hs.preferredView)
	block := hs.bc.GetChildrenBlocks(preferredBlock.ID)[len(hs.bc.GetChildrenBlocks(preferredBlock.ID))-1]
	if !config.Configuration.IsByzantine(block.Proposer) {
		log.Debugf("create a fork, id: %x", block.QC.BlockID)
		return block.QC
	}

	grandChildrenBlocks := hs.bc.GetChildrenBlocks(block.ID)
	for _, b := range grandChildrenBlocks {
		if !config.Configuration.IsByzantine(b.Proposer) {
			log.Debugf("create a fork, id: %x", b.QC.BlockID)
			return b.QC
		}
	}
	return hs.bc.GetHighQC()
}

// highestForkchoice returns the high QC
func (hs *HotStuff) highestForkchoice() *blockchain.QC {
	return hs.bc.GetHighQC()
}

// higestForkchoice returns the highest QC from the longest chain
func (hs *HotStuff) longestForkchoice() *blockchain.QC {
	var qc *blockchain.QC
	return qc
}
