package zeitgeber

type Synchronizer interface {
	NewView(view View)
	TimeoutFor(view View)
	HandleTC(tc TCMsg)
	HandleTmo(tmo TmoMsg)
	GetCurView() View
	GetHighCert() *TC
	EnteringViewEvent() chan View
}
