package crypto

import (
	"hash"

	"golang.org/x/crypto/sha3"
)

const (
	HashLenSha3_224 = 32
	HashLenSha3_256 = 32
	HashLenSha3_384 = 32
	HashLenSha3_512 = 32
)

// sha3_224Algo, embeds commonHasher
type sha3_224Algo struct {
	*commonHasher
	hash.Hash
}

// sha3_256Algo, embeds commonHasher
type sha3_256Algo struct {
	*commonHasher
	hash.Hash
}

// sha3_384Algo, embeds commonHasher
type sha3_384Algo struct {
	*commonHasher
	hash.Hash
}

// sha3_512Algo, embeds commonHasher
type sha3_512Algo struct {
	*commonHasher
	hash.Hash
}

// NewSHA3_224 returns a new instance of SHA3-224 hasher
func NewSHA3_224() Hasher {
	return &sha3_224Algo{
		commonHasher: &commonHasher{
			outputSize: HashLenSha3_224},
		Hash: sha3.New224()}
}

// NewSHA3_256 returns a new instance of SHA3-256 hasher
func NewSHA3_256() Hasher {
	return &sha3_256Algo{
		commonHasher: &commonHasher{
			outputSize: HashLenSha3_256},
		Hash: sha3.New256()}
}

// NewSHA3_384 returns a new instance of SHA3-384 hasher
func NewSHA3_384() Hasher {
	return &sha3_384Algo{
		commonHasher: &commonHasher{
			outputSize: HashLenSha3_384},
		Hash: sha3.New384()}
}

// NewSHA3_512 returns a new instance of SHA3-512 hasher
func NewSHA3_512() Hasher {
	return &sha3_512Algo{
		commonHasher: &commonHasher{
			outputSize: HashLenSha3_512},
		Hash: sha3.New512()}
}

// ComputeHash calculates and returns the SHA3-224 output of input byte array
func (s *sha3_224Algo) ComputeHash(data []byte) Hash {
	s.Reset()
	_, _ = s.Write(data)
	digest := make(Hash, 0, HashLenSha3_224)
	return s.Sum(digest)
}

// ComputeHash calculates and returns the SHA3-256 output of input byte array
func (s *sha3_256Algo) ComputeHash(data []byte) Hash {
	s.Reset()
	_, _ = s.Write(data)
	digest := make(Hash, 0, HashLenSha3_256)
	return s.Sum(digest)
}

// ComputeHash calculates and returns the SHA3-384 output of input byte array
func (s *sha3_384Algo) ComputeHash(data []byte) Hash {
	s.Reset()
	_, _ = s.Write(data)
	digest := make(Hash, 0, HashLenSha3_384)
	return s.Sum(digest)
}

// ComputeHash calculates and returns the SHA3-512 output of input byte array
func (s *sha3_512Algo) ComputeHash(data []byte) Hash {
	s.Reset()
	_, _ = s.Write(data)
	digest := make(Hash, 0, HashLenSha3_512)
	return s.Sum(digest)
}

// SumHash returns the SHA3-224 output and resets the hash state
func (s *sha3_224Algo) SumHash() Hash {
	digest := make(Hash, HashLenSha3_224)
	s.Sum(digest[:0])
	s.Reset()
	return digest
}

// SumHash returns the SHA3-256 output and resets the hash state
func (s *sha3_256Algo) SumHash() Hash {
	digest := make(Hash, HashLenSha3_256)
	s.Sum(digest[:0])
	s.Reset()
	return digest
}

// SumHash returns the SHA3-384 output and resets the hash state
func (s *sha3_384Algo) SumHash() Hash {
	digest := make(Hash, HashLenSha3_384)
	s.Sum(digest[:0])
	s.Reset()
	return digest
}

// SumHash returns the SHA3-512 output and resets the hash state
func (s *sha3_512Algo) SumHash() Hash {
	digest := make(Hash, HashLenSha3_512)
	s.Sum(digest[:0])
	s.Reset()
	return digest
}
