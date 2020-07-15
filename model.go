package zeitgeber

import (
	"fmt"
	"math/rand"
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

func NewQC(view View) *QC {
	qc := new(QC)
	qc.View = view
	// placeholder
	qc.AggSig = make([]byte, 65)
	rand.Read(qc.AggSig)
	return qc
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

// ACK adds id to quorum ack records
func (q *Quorum) ACK(view View, id NodeID) {
	_, exist := q.acks[view]
	if !exist {
		//	first time of receiving the wish for this view
		q.acks[view] = make(map[NodeID]bool)
		q.timestamps[view] = time.Now()
	}
	q.acks[view][id] = true
}

// Size returns current ack size
func (q *Quorum) Size(view View) int {
	return len(q.acks[view])
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

// Super majority quorum satisfied
func (q *Quorum) SuperMajority(view View) bool {
	return q.Size(view) > config.n*2/3
}
