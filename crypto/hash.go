package crypto

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
)

const (
	SHA3_224 = "sha3_224"
	SHA3_256 = "sha3_256"
	SHA3_384 = "sha3_384"
	SHA3_512 = "sha3_512"
)

//var hashTypes = []string{SHA3_224, SHA3_256, SHA3_384, SHA3_512}

// Hash is the hash algorithms output types
type Hash []byte

// Equal checks if a hash is equal to a given hash
func (h Hash) Equal(input Hash) bool {
	return bytes.Equal(h, input)
}

// Hex returns the hex string representation of the hash.
func (h Hash) Hex() string {
	return hex.EncodeToString(h)
}

// Hasher interface
type Hasher interface {
	// Size returns the hash output length
	Size() int
	// ComputeHash returns the hash output regardless of the hash state
	ComputeHash([]byte) Hash
	// Write([]bytes) (using the io.Writer interface) adds more bytes to the
	// current hash state
	io.Writer
	// SumHash returns the hash output and resets the hash state
	SumHash() Hash
	// Reset resets the hash state
	Reset()
}

// commonHasher holds the common data for all hashers
type commonHasher struct {
	outputSize int
}

func BytesToHash(b []byte) Hash {
	h := make([]byte, len(b))
	copy(h, b)
	return h
}

// HashesToBytes converts a slice of hashes to a slice of byte slices.
func HashesToBytes(hashes []Hash) [][]byte {
	b := make([][]byte, len(hashes))
	for i, h := range hashes {
		b[i] = h
	}
	return b
}

// NewHasher chooses and initializes a hashing algorithm
// Deprecated and will removed later: use dedicated hash generation functions instead.
// UPDATE By Ali: This function is used to build a new hasher for a replica
func NewHasher(hashType string) (Hasher, error) {
	if hashType == SHA3_224 {
		return NewSHA3_224(), nil
	} else if hashType == SHA3_256 {
		return NewSHA3_256(), nil
	} else if hashType == SHA3_384 {
		return NewSHA3_384(), nil
	} else if hashType == SHA3_512 {
		return NewSHA3_512(), nil
	} else {
		return nil, errors.New("Invalid hasher type!")
	}
}
