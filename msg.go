package zeitgeber

type ProposalMsg struct {
	NodeID   ID
	View     View
	TimeCert *TC
	HighQC   *QC
	Command  Command
}

type TmoMsg struct {
	View   View
	NodeID ID
	HighQC QC
}

type TCMsg struct {
	View   View
	NodeID ID
}
