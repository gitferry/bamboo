package emulator

import (
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/gitferry/zeitgeber/messages"

	"github.com/gitferry/zeitgeber/replica"
)

type Coordinator struct {
	highView     uint64 // the highest view seen among replicas
	replicas     []*replica.Replica
	quorumCounts map[uint64]uint
	threshold    uint
}

// OnWishToAdvance updates the quorum map
func (c *Coordinator) OnWishToAdvance(msg *messages.WishMsg) {
	view := msg.View
	if c.quorumCounts[view] >= c.threshold {
		return
	}
	c.quorumCounts[view]++
	if c.quorumCounts[view] >= c.threshold {
		replicas := c.selectReplicas(view)
		c.EnterNewView(replicas, view)
	}
}

func (c *Coordinator) selectReplicas(view uint64) []*replica.Replica {
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
func (c *Coordinator) EnterNewView(replicas []*replica.Replica, view uint64) {

}

func Init() *Coordinator {
	return &Coordinator{}
}

func (c *Coordinator) Run() {
	var wg sync.WaitGroup

	for _, rep := range c.replicas {
		wg.Add(1)
		go rep.Run()
	}
	msgs := c.mergeRecieveChan()
	for {
		msg := <-msgs
		c.OnWishToAdvance(msg)
	}
	wg.Wait()
}

func (c *Coordinator) mergeRecieveChan() chan *messages.WishMsg {
	chans := make([]chan *messages.WishMsg, len(c.replicas))
	for _, rep := range c.replicas {
		chans = append(chans, rep.WishMsgs)
	}
	out := make(chan *messages.WishMsg)
	var cases []reflect.SelectCase
	for _, c := range chans {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(c),
		})
		i, v, ok := reflect.Select(cases)
		if !ok {
			cases = append(cases[:i], cases[i+1:]...)
			continue
		}
		out <- v.Interface().(*messages.WishMsg)
	}

	return out
}
