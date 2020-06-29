package zeitgeber

type Synchronizer interface {
	NewView(view View)
	TimeoutFor(view View)
	GetCurView() View
}
