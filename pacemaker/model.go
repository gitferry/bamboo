package pacemaker

import (
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/types"
)

type TMO struct {
	View   types.View
	NodeID identity.NodeID
	HighTC *TC
}

type TC struct {
	types.View
}

func NewTC(view types.View) *TC {
	return &TC{View: view}
}
