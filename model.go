package zeitgeber

import (
	"fmt"
	"time"
)

type View int

type TMO struct {
	View   View
	NodeID NodeID
	HighTC *TC
}

type TC struct {
	View
}

func NewTC(view View) *TC {
	return &TC{View: view}
}

// Quorum records each acknowledgement and check for different types of quorum satisfied
type Quorum struct {
	acks       map[View]map[NodeID]bool
	timestamps map[View]time.Time // keeps track of the time of first receiving the wish for each view
}

// NewQuorum returns a new Quorum
func NewQuorum() *Quorum {
	q := &Quorum{
		acks:       make(map[View]map[NodeID]bool),
		timestamps: make(map[View]time.Time),
	}
	return q
}

// Reset resets the quorum to empty
func (q *Quorum) Reset() {
	q.acks = make(map[View]map[NodeID]bool)
}

func (q *Quorum) All(view View) bool {
	return q.Size(view) == config.n
}

// Majority quorum satisfied
func (q *Quorum) Majority(view View) bool {
	return q.Size(view) > config.n/2
}

func (q *Quorum) GenerateQC(view View) (*QC, error) {
	if !q.SuperMajority(view) {
		return nil, fmt.Errorf("votes are not sufficient")
	}
	return NewQC(view), nil
}
