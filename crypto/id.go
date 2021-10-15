package crypto

import (
	"github.com/gitferry/bamboo/types/encoding"
)

type Identifier [32]byte

// MakeID creates an Hash from the hash of encoded data.
func MakeID(body interface{}) Identifier {
	data := encoding.DefaultEncoder.MustEncode(body)
	hasher := NewSHA3_256()
	hash := hasher.ComputeHash(data)
	return HashToID(hash)
}

func HashToID(hash []byte) Identifier {
	var id Identifier
	copy(id[:], hash)
	return id
}

func IDToByte(id Identifier) []byte {
	return id[:]
}
