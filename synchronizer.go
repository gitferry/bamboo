package zeitgeber

type Synchronizer interface {
	NewView(view View)
	TimeoutFor(view View)
	HandleTC(tc *TC)
	GetCurView() View
	ResetTimer() chan bool
}
