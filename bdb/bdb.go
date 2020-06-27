package bdb

import (
	"fmt"
	"sync"

	"github.com/gitferry/zeitgeber"
)

type Bdb struct {
	zeitgeber.Node

	CurView zeitgeber.View
	quorum *zeitgeber.Quorum

	mu sync.Mutex
}

func NewBdb() *Bdb {
	return &Bdb{
		quorum:zeitgeber.NewQuorum(),
	}
}

func (b *Bdb) handleWish(wish zeitgeber.WishMsg) (bool, *zeitgeber.TC) {
	b.quorum.ACK(wish.View, wish.NodeID)

	if b.quorum.SuperMajority(wish.View) {
		return true, zeitgeber.NewTC(wish.View)
	}

	return false, nil
}

func (b *Bdb) WishAdvance(highQC zeitgeber.QC) {
	b.mu.Lock()
	wishView := b.CurView+1
	b.mu.Unlock()
	wishMsh := WishMsg{
		View:   wishView,
		NodeID: b.ID(),
		HighQC: highQC,
	}
	b.Broadcast(wishMsh)
}

func (b *Bdb) NewView (view zeitgeber.View) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if view < b.CurView {
		return fmt.Errorf("the view %d is lower than current view %d", view, b.CurView)
	}
	b.CurView = view

	return nil
}


