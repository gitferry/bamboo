package socket

import (
	"github.com/gitferry/bamboo/config"
	"math/rand"
	"sync"
	"time"

	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/transport"
	"github.com/gitferry/bamboo/utils"
)

// Socket integrates all networking interface and fault injections
type Socket interface {

	// Send put message to outbound queue
	Send(to identity.NodeID, m interface{})

	// MulticastQuorum sends msg to a set of nodes
	MulticastQuorum(nodes []identity.NodeID, m interface{})

	// Broadcast send to all peers
	Broadcast(m interface{})

	// Recv receives a message
	Recv() interface{}

	Close()

	// Fault injection
	Drop(id identity.NodeID, t int)             // drops every message send to NodeID last for t seconds
	Slow(id identity.NodeID, d int, t int)      // delays every message send to NodeID for d ms and last for t seconds
	Flaky(id identity.NodeID, p float64, t int) // drop message by chance p for t seconds
	Crash(t int)                                // node crash for t seconds

	SendRate() float64
	RecvRate() float64
}

type socket struct {
	id        identity.NodeID
	addresses map[identity.NodeID]string
	nodes     map[identity.NodeID]transport.Transport

	crash bool
	drop  map[identity.NodeID]bool
	slow  map[identity.NodeID]int
	flaky map[identity.NodeID]float64

	lock sync.RWMutex // locking map nodes
}

// NewSocket return Socket interface instance given self NodeID, node list, transport and codec name
func NewSocket(id identity.NodeID, addrs map[identity.NodeID]string) Socket {
	socket := &socket{
		id:        id,
		addresses: addrs,
		nodes:     make(map[identity.NodeID]transport.Transport),
		crash:     false,
		drop:      make(map[identity.NodeID]bool),
		slow:      make(map[identity.NodeID]int),
		flaky:     make(map[identity.NodeID]float64),
	}

	socket.nodes[id] = transport.NewTransport(addrs[id])
	socket.nodes[id].Listen()

	return socket
}

// in Mbps
func (s *socket) SendRate() float64 {
	var totalRate int64
	for _, t := range s.nodes {
		totalRate += t.SendBitsCount()
	}
	return float64(totalRate) / 1024 / 1024
}

// in Mbps
func (s *socket) RecvRate() float64 {
	return float64(s.nodes[s.id].RecvBitsCount()) / 1024 / 1024
}

func (s *socket) Send(to identity.NodeID, m interface{}) {
	//log.Debugf("node %s send message %+v to %v", s.id, m, to)

	if s.crash {
		return
	}

	if s.drop[to] {
		return
	}

	if p, ok := s.flaky[to]; ok && p > 0 {
		if rand.Float64() < p {
			return
		}
	}

	s.lock.RLock()
	t, exists := s.nodes[to]
	s.lock.RUnlock()
	if !exists {
		s.lock.RLock()
		address, ok := s.addresses[to]
		s.lock.RUnlock()
		if !ok {
			log.Errorf("socket does not have address of node %s", to)
			return
		}
		t = transport.NewTransport(address)
		err := utils.Retry(t.Dial, 100, time.Duration(50)*time.Millisecond)
		if err != nil {
			panic(err)
		}
		s.lock.Lock()
		s.nodes[to] = t
		s.lock.Unlock()
	}

	// add simulated transmission delay
	if config.GetConfig().Delay != 0 {
		delay := config.GetConfig().Delay
		err := config.GetConfig().DErr
		rand.Seed(time.Now().UnixNano())
		max := delay + err
		min := delay - err
		randDelay := time.Duration(rand.Intn(max-min+1)+min) * time.Millisecond
		timer := time.NewTimer(randDelay)
		go func() {
			<-timer.C
			t.Send(m)
		}()
		return

	}
	if delay, ok := s.slow[to]; ok && delay > 0 {
		timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
		go func() {
			<-timer.C
			t.Send(m)
		}()
		return
	}
	t.Send(m)
	//log.Debugf("[%v] message %v is sent to %v", s.id, m, to)
}

func (s *socket) Recv() interface{} {
	s.lock.RLock()
	t := s.nodes[s.id]
	s.lock.RUnlock()
	for {
		m := t.Recv()
		if !s.crash {
			return m
		}
	}
}

func (s *socket) MulticastQuorum(nodes []identity.NodeID, m interface{}) {
	//log.Debugf("node %s multicasting message %+v for %d nodes", s.id, m, quorum)
	//a := make([]int, len(s.addresses))
	//for i := range a {
	//	a[i] = i + 1
	//}
	//a = append(a[:s.id.Node()-1], a[s.id.Node():]...)
	//rand.Seed(time.Now().UnixNano())
	//rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	//for i := 0; i < quorum; i++ {
	//	s.Send(identity.NewNodeID(a[i]), m)
	//}
	if nodes == nil {
		return
	}

	for _, id := range nodes {
		if id == s.id {
			continue
		}
		s.Send(id, m)
	}
}

func (s *socket) Broadcast(m interface{}) {
	//log.Debugf("node %s broadcasting message %+v", s.id, m)
	for id := range s.addresses {
		if id == s.id {
			continue
		}
		s.Send(id, m)
	}
	//log.Debugf("node %s done  broadcasting message %+v", s.id, m)
}

func (s *socket) Close() {
	for _, t := range s.nodes {
		t.Close()
	}
}

func (s *socket) Drop(id identity.NodeID, t int) {
	s.drop[id] = true
	timer := time.NewTimer(time.Duration(t) * time.Second)
	go func() {
		<-timer.C
		s.drop[id] = false
	}()
}

func (s *socket) Slow(id identity.NodeID, delay int, t int) {
	s.slow[id] = delay
	timer := time.NewTimer(time.Duration(t) * time.Second)
	go func() {
		<-timer.C
		s.slow[id] = 0
	}()
}

func (s *socket) Flaky(id identity.NodeID, p float64, t int) {
	s.flaky[id] = p
	timer := time.NewTimer(time.Duration(t) * time.Second)
	go func() {
		<-timer.C
		s.flaky[id] = 0
	}()
}

func (s *socket) Crash(t int) {
	s.crash = true
	if t > 0 {
		timer := time.NewTimer(time.Duration(t) * time.Second)
		go func() {
			<-timer.C
			s.crash = false
		}()
	}
}
