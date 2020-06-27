package zeitgeber


type Synchronizer interface {
	AdvanceView(view int)
}
