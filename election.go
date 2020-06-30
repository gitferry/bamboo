package zeitgeber

import (
	"crypto/sha1"
	"strconv"
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
	h.Write([]byte(strconv.Itoa(int(view))))
	bs, _ := strconv.Atoi(string(h.Sum(nil)))
	return bs%r.peerNo == id.Node()
}

func (r *rotation) FindLeaderFor(view View) ID {
	id := int(view) % r.peerNo
	return NewID(id)
}
