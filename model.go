package zeitgeber

type View int

type QC struct {
	view View
}

type TC struct {
	view View
}

func NewTC(view View) *TC  {
	return &TC{view:view}
}

// Quorum records each acknowledgement and check for different types of quorum satisfied
type Quorum struct {
	acks  map[View]map[ID]bool
}

// NewQuorum returns a new Quorum
func NewQuorum() *Quorum {
	q := &Quorum{
		acks:  make(map[View]map[ID]bool),
	}
	return q
}

// ACK adds id to quorum ack records
func (q *Quorum) ACK(view View, id ID) {
	_, exist := q.acks[view][id]
	if exist {
		return
	}
	q.acks[view] = make(map[ID]bool)
	q.acks[view][id] = true
}

// Size returns current ack size
func (q *Quorum) Size(view View) int {
	return len(q.acks[view])
}

// Reset resets the quorum to empty
func (q *Quorum) Reset() {
	q.acks = make(map[View]map[ID]bool)
}

func (q *Quorum) All(view View) bool {
	return q.Size(view) == config.n
}

// Majority quorum satisfied
func (q *Quorum) Majority(view View) bool {
	return q.Size(view) > config.n/2
}

// Super majority quorum satisfied
func (q *Quorum) SuperMajority(view View) bool {
	return q.size > config.n*2/3
}

// FastQuorum from fast paxos
func (q *Quorum) FastQuorum() bool {
	return q.size >= config.n*3/4
}

// AllZones returns true if there is at one ack from each zone
func (q *Quorum) AllZones() bool {
	return len(q.zones) == config.z
}

// ZoneMajority returns true if majority quorum satisfied in any zone
func (q *Quorum) ZoneMajority() bool {
	for z, n := range q.zones {
		if n > config.npz[z]/2 {
			return true
		}
	}
	return false
}

// GridRow == AllZones
func (q *Quorum) GridRow() bool {
	return q.AllZones()
}

// GridColumn == all nodes in one zone
func (q *Quorum) GridColumn() bool {
	for z, n := range q.zones {
		if n == config.npz[z] {
			return true
		}
	}
	return false
}

// FGridQ1 is flexible grid quorum for phase 1
func (q *Quorum) FGridQ1(Fz int) bool {
	zone := 0
	for z, n := range q.zones {
		if n > config.npz[z]/2 {
			zone++
		}
	}
	return zone >= config.z-Fz
}

// FGridQ2 is flexible grid quorum for phase 2
func (q *Quorum) FGridQ2(Fz int) bool {
	zone := 0
	for z, n := range q.zones {
		if n > config.npz[z]/2 {
			zone++
		}
	}
	return zone >= Fz+1
}

/*
// Q1 returns true if config.Quorum type is satisfied
func (q *Quorum) Q1() bool {
	switch config.Quorum {
	case "majority":
		return q.Majority()
	case "grid":
		return q.GridRow()
	case "fgrid":
		return q.FGridQ1()
	case "group":
		return q.ZoneMajority()
	case "count":
		return q.size >= config.n-config.F
	default:
		log.Error("Unknown quorum type")
		return false
	}
}

// Q2 returns true if config.Quorum type is satisfied
func (q *Quorum) Q2() bool {
	switch config.Quorum {
	case "majority":
		return q.Majority()
	case "grid":
		return q.GridColumn()
	case "fgrid":
		return q.FGridQ2()
	case "group":
		return q.ZoneMajority()
	case "count":
		return q.size > config.F
	default:
		log.Error("Unknown quorum type")
		return false
	}
}
*/


