package pacemaker

import (
	"github.com/gitferry/bamboo/blockchain"
	"github.com/gitferry/bamboo/crypto"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/types"
)

type TMO struct {
	View   types.View
	NodeID identity.NodeID
	HighQC *blockchain.QC
}

type TC struct {
	types.View
	crypto.AggSig
	crypto.Signature
}

func NewTC(view types.View, requesters map[identity.NodeID]*TMO) *TC {
	// TODO: add crypto
	return &TC{View: view}
}
