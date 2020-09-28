package identity

import (
	"strconv"

	"github.com/gitferry/bamboo/log"
)

// NodeID represents a generic identifier in format of Zone.Node
type NodeID string

// NewNodeID returns a new NodeID type given two int number of zone and node
func NewNodeID(node int) NodeID {
	if node < 0 {
		node = -node
	}
	// return NodeID(fmt.Sprintf("%d.%d", zone, node))
	return NodeID(strconv.Itoa(node))
}

// Zone returns Zond NodeID component
//func (i NodeID) Zone() int {
//	if !strings.Contains(string(i), ".") {
//		log.Warningf("id %s does not contain \".\"\n", i)
//		return 0
//	}
//	s := strings.Split(string(i), ".")[0]
//	zone, err := strconv.ParseUint(s, 10, 64)
//	if err != nil {
//		log.Errorf("Failed to convert Zone %s to int\n", s)
//		return 0
//	}
//	return int(zone)
//}

// Node returns Node NodeID component
func (i NodeID) Node() int {
	var s string
	s = string(i)
	node, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Errorf("Failed to convert Node %s to int\n", s)
		return 0
	}
	return int(node)
}

type IDs []NodeID

func (a IDs) Len() int      { return len(a) }
func (a IDs) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

//func (a IDs) Less(i, j int) bool {
//	if a[i].Zone() < a[j].Zone() {
//		return true
//	} else if a[i].Zone() > a[j].Zone() {
//		return false
//	} else {
//		return a[i].Node() < a[j].Node()
//	}
//}
