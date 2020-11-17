package election

import (
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/types"
)

type Election interface {
	IsLeader(id identity.NodeID, view types.View) bool
	FindLeaderFor(view types.View) identity.NodeID
}

type rotation struct {
	peerNo int
}

func NewRotation(peerNo int) *rotation {
	return &rotation{
		peerNo: peerNo,
	}
}

func (r *rotation) IsLeader(id identity.NodeID, view types.View) bool {
	//if view <= 3 {
	//	if id.Node() < r.peerNo {
	//		return false
	//	}
	//	return true
	//}
	//h := sha1.New()
	//h.Write([]byte(strconv.Itoa(int(view) + 1)))
	//bs := h.Sum(nil)
	//data := binary.BigEndian.Uint64(bs)
	//return data%uint64(r.peerNo) == uint64(id.Node()-1)
	if id.Node() != 4 {
		return false
	}
	return true
}

func (r *rotation) FindLeaderFor(view types.View) identity.NodeID {
	//if view <= 3 {
	//	return identity.NewNodeID(r.peerNo)
	//}
	//h := sha1.New()
	//h.Write([]byte(strconv.Itoa(int(view + 1))))
	//bs := h.Sum(nil)
	//data := binary.BigEndian.Uint64(bs)
	//id := data%uint64(r.peerNo) + 1
	//return identity.NewNodeID(int(id))
	return identity.NewNodeID(4)
}
