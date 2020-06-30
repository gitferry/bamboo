package zeitgeber

type ProposalMsg struct {
	NodeID   ID
	View     View
	TimeCert *TC
	HighQC   *QC
	Command  Command
}
