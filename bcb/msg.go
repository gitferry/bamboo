package bcb

import "github.com/gitferry/zeitgeber"

type TmoMsg struct {
	View   zeitgeber.View
	NodeID zeitgeber.ID
	HighQC zeitgeber.QC
}

type TCMsg struct {
	View   zeitgeber.View
	NodeID zeitgeber.ID
}
