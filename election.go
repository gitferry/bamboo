package zeitgeber

import (
	"crypto/sha1"
	"encoding/binary"
)

type Election interface {
	IsLeader(id ID, view View) bool
	FindLeaderFor(view View) ID
}

type rotation struct {
	peerNo int
}

func NewRotation(peerNo int) *rotation {
	return &rotation{
		peerNo: peerNo,
	}
}

func (r *rotation) IsLeader(id ID, view View) bool {
	h := sha1.New()
	h.Write([]byte(string(view)))
	bs := h.Sum(nil)
	data := binary.BigEndian.Uint64(bs)
	return data%uint64(r.peerNo) == uint64(id.Node()-1)
}

func (r *rotation) FindLeaderFor(view View) ID {
	h := sha1.New()
	h.Write([]byte(string(view)))
	bs := h.Sum(nil)
	data := binary.BigEndian.Uint64(bs)
	id := data%uint64(r.peerNo) + 1
	return NewID(int(id))
}
