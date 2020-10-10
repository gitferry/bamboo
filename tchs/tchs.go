package tchs

import (
	"fmt"
	"sync"

	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/types"
)

const (
	FORKING = "forking"
	HIGHEST = "highest"
	LONGEST = "longest"
)

type Tchs struct {
	lastVotedView  types.View
	preferredView  types.View
	forkchoiceType string
	bc             *blockchain.BlockChain
	mu             sync.Mutex
}

func Newtchs(blockchain *blockchain.BlockChain, forkchoice string) *Tchs {
	hs := new(Tchs)
	hs.bc = blockchain
	hs.forkchoiceType = forkchoice
	return hs
}

func (hs *Tchs) VotingRule(block *blockchain.Block) (bool, error) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
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

func (hs *Tchs) CommitRule(qc *blockchain.QC) (bool, *blockchain.Block, error) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	parentBlock, err := hs.bc.GetParentBlock(qc.BlockID)
	if err != nil {
		return false, nil, fmt.Errorf("cannot commit any block: %w", err)
	}
	if (parentBlock.View + 1) == qc.View {
		return true, parentBlock, nil
	}
	return false, nil, nil
}

func (hs *Tchs) UpdateStateByView(view types.View) error {
	return hs.updateLastVotedView(view)
}

func (hs *Tchs) UpdateStateByQC(qc *blockchain.QC) error {
	if qc.View <= 2 {
		return nil
	}
	return hs.updatePreferredView(qc)
}

func (hs *Tchs) updateLastVotedView(targetView types.View) error {
	if targetView < hs.lastVotedView {
		return fmt.Errorf("target view is lower than the last voted view")
	}
	hs.lastVotedView = targetView
	return nil
}

func (hs *Tchs) updatePreferredView(qc *blockchain.QC) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	if qc.View > hs.preferredView {
		hs.preferredView = qc.View
	}
	return nil
}

func (hs *Tchs) Forkchoice() *blockchain.QC {
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
func (hs *Tchs) forkingForkchoice() *blockchain.QC {
	//if hs.preferredView <= 1 {
	//	return hs.bc.GetHighQC()
	//}
	//highBlock, _ := hs.bc.GetBlockByID(hs.bc.GetHighQC().BlockID)
	//if config.Configuration.IsByzantine(highBlock.Proposer) {
	//	return hs.bc.GetHighQC()
	//}
	//preferredBlock := hs.bc.GetBlockByView(hs.preferredView)
	//block := hs.bc.GetChildrenBlocks(preferredBlock.ID)[len(hs.bc.GetChildrenBlocks(preferredBlock.ID))-1]
	//if !config.Configuration.IsByzantine(block.Proposer) {
	//	log.Debugf("create a fork, id: %x", block.QC.BlockID)
	//	return block.QC
	//}

	//grandChildrenBlocks := hs.bc.GetChildrenBlocks(block.ID)
	//for _, b := range grandChildrenBlocks {
	//	if !config.Configuration.IsByzantine(b.Proposer) {
	//		log.Debugf("create a fork, id: %x", b.QC.BlockID)
	//		return b.QC
	//	}
	//}
	return hs.bc.GetHighQC()
}

// highestForkchoice returns the high QC
func (hs *Tchs) highestForkchoice() *blockchain.QC {
	return hs.bc.GetHighQC()
}

// higestForkchoice returns the highest QC from the longest chain
func (hs *Tchs) longestForkchoice() *blockchain.QC {
	var qc *blockchain.QC
	return qc
}
