package election

import (
	"crypto/sha1"
	"encoding/binary"
	"strconv"

	"github.com/gitferry/zeitgeber/identity"
	"github.com/gitferry/zeitgeber/types"
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
	h := sha1.New()
	h.Write([]byte(strconv.Itoa(int(view) + 2)))
	bs := h.Sum(nil)
	data := binary.BigEndian.Uint64(bs)
	return data%uint64(r.peerNo) == uint64(id.Node()-1)
}

func (r *rotation) FindLeaderFor(view types.View) identity.NodeID {
	h := sha1.New()
	h.Write([]byte(strconv.Itoa(int(view + 2))))
	bs := h.Sum(nil)
	data := binary.BigEndian.Uint64(bs)
	id := data%uint64(r.peerNo) + 1
	return identity.NewNodeID(int(id))
}
