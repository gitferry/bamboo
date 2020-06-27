package bdb

import "github.com/gitferry/zeitgeber"

type WishMsg struct {
	View zeitgeber.View
	NodeID zeitgeber.ID
	HighQC zeitgeber.QC
}
