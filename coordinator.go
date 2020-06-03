package zeitgeber

import (
	"math/rand"
	"time"
)

type Coordinator struct {
	highView     uint64 // the highest view seen among replicas
	replicas     []*Replica
	quorumCounts map[uint64]uint
	threshold    uint
}

// OnWishToAdvance updates the quorum map
// func (c *Coordinator) OnWishToAdvance(msg *WishMsg) {
// view := msg.View
// if c.quorumCounts[view] >= c.threshold {
// 	return
// }
// c.quorumCounts[view]++
// if c.quorumCounts[view] >= c.threshold {
// 	replicas := c.selectReplicas(view)
// 	c.EnterNewView(replicas, view)
// }
// }

func (c *Coordinator) selectReplicas(view uint64) []*Replica {
	if !c.isByzantine(view) {
		return c.replicas
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	randIndex := r.Intn(len(c.replicas))
	return c.replicas[randIndex:]
	// TODO: return a subset
}

func (c *Coordinator) isByzantine(view uint64) bool {
	return false
}

// EnterNewView instructs a subset of replicas to enter the new view
func (c *Coordinator) EnterNewView(replicas []*Replica, view int) {

}

func InitCoordinator() *Coordinator {
	return &Coordinator{}
}
