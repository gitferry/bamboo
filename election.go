package zeitgeber

import (
	"crypto/sha1"
	"encoding/binary"
)

type Election interface {
	IsLeader(id NodeID, view View) bool
	FindLeaderFor(view View) NodeID
}

type rotation struct {
	peerNo int
}

func NewRotation(peerNo int) *rotation {
	return &rotation{
		peerNo: peerNo,
	}
}

func (r *rotation) IsLeader(id NodeID, view View) bool {
	h := sha1.New()
	h.Write([]byte(string(view)))
	bs := h.Sum(nil)
	data := binary.BigEndian.Uint64(bs)
	return data%uint64(r.peerNo) == uint64(id.Node()-1)
}

func (r *rotation) FindLeaderFor(view View) NodeID {
	h := sha1.New()
	h.Write([]byte(string(view)))
	bs := h.Sum(nil)
	data := binary.BigEndian.Uint64(bs)
	id := data%uint64(r.peerNo) + 1
	return NewNodeID(int(id))
}
