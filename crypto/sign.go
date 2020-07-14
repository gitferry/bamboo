package crypto

// SigningAlgorithm is an identifier for a signing algorithm and curve.
type SigningAlgorithm int

// String returns the string representation of this signing algorithm.
func (f SigningAlgorithm) String() string {
	return [...]string{"UNKNOWN", "BLS_BLS12381", "ECDSA_P256", "ECDSA_SECp256k1"}[f]
}

const (
	// Supported signing algorithms
	UnknownSigningAlgorithm SigningAlgorithm = iota
	BLS_BLS12381
	ECDSA_P256
	ECDSA_SECp256k1
)

// PrivateKey is an unspecified signature scheme private key
type PrivateKey interface {
	// Algorithm returns the signing algorithm related to the private key.
	Algorithm() SigningAlgorithm
	// KeySize return the key size in bytes.
	KeySize() int
	// Sign generates a signature using the provided hasher.
	Sign([]byte, Hasher) (Signature, error)
	// PublicKey returns the public key.
	PublicKey() PublicKey
	// Encode returns a bytes representation of the private key
	Encode() ([]byte, error)
}

// PublicKey is an unspecified signature scheme public key.
type PublicKey interface {
	// Algorithm returns the signing algorithm related to the public key.
	Algorithm() SigningAlgorithm
	// KeySize return the key size in bytes.
	KeySize() int
	// Verify verifies a signature of an input message using the provided hasher.
	Verify(Signature, []byte, Hasher) (bool, error)
	// Encode returns a bytes representation of the public key.
	Encode() ([]byte, error)
}
