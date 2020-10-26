package crypto

import "math/big"

type Signature [][]byte
type AggSig []Signature

type ECDSASignature struct {
	r, s *big.Int
}

// TODO: I have to generate an error here
func (sig *Signature) ToECDSA() ECDSASignature {
	var ecdsaSig = new(ECDSASignature)
	signature := *sig
	var r, s big.Int
	r.SetString(string(signature[0]), 10)
	s.SetString(string(signature[1]), 10)
	ecdsaSig.r = &r
	ecdsaSig.s = &s
	return *ecdsaSig
}
