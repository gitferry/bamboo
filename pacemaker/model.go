package pacemaker

import "github.com/gitferry/zeitgeber"

type TMO struct {
	View   zeitgeber.View
	NodeID zeitgeber.NodeID
	HighTC *TC
}

type TC struct {
	zeitgeber.View
}

func NewTC(view zeitgeber.View) *TC {
	return &TC{View: view}
}
