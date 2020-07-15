package blockchain

import (
	"github.com/gitferry/zeitgeber"
	"github.com/gitferry/zeitgeber/crypto"
)

type Vote struct {
}

type AggSig []crypto.Signature

type QC struct {
	View    zeitgeber.View
	BlockID crypto.Identifier
	AggSig
}
