package zeitgeber

type Election interface {
	IsLeader(id ID, view View)bool
	FindLeaderFor(view View)ID
}

type rotation struct {
	peerNo int
}

func NewRotation(peerNo int) *rotation{
	return &rotation{
		peerNo:peerNo,
	}
}

func (r *rotation) IsLeader(id ID, view View)bool {
	return int(view)%r.peerNo == id.Node()
}

func (r *rotation) FindLeaderFor(view View) ID {
	id := int(view) % r.peerNo
	return NewID(id)
}
