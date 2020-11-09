package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
)

type ecdsa_p256_PrivateKey struct {
	SignAlg    string
	PrivateKey *ecdsa.PrivateKey
}

type ecdsa_p256_PublicKey struct {
	SignAlg   string
	PublicKey ecdsa.PublicKey
}

func (priv *ecdsa_p256_PrivateKey) PublicKey() PublicKey {
	pub := &ecdsa_p256_PublicKey{SignAlg: ECDSA_P256, PublicKey: priv.PrivateKey.PublicKey}
	return pub
}

func (priv *ecdsa_p256_PrivateKey) Algorithm() string {
	return priv.SignAlg
}

// This function is commented for now.
// func (priv *ecdsa_p256_PrivateKey) KeySize() int {
//	return len([]byte(*priv))
// }

// ecdsa.Sign returns two *big.Int variables. In order to save it in the Signature type,
// I first turn them into strings, and then I turn the strings to byte arrays.
// I have implemented a Signature to ecdsa signature parser (toECDSA in signature.go) in oder to
// cast the byte array Signature into the original signature of the ECDSA signing method.
func (priv *ecdsa_p256_PrivateKey) Sign(msg []byte, hasher Hasher) (Signature, error) {
	var r, s *big.Int
	var err error
	if hasher != nil {
		r, s, err = ecdsa.Sign(rand.Reader, priv.PrivateKey, hasher.ComputeHash(msg))
		if err != nil {
			return nil, err
		}
	} else {
		r, s, err = ecdsa.Sign(rand.Reader, priv.PrivateKey, msg)
		if err != nil {
			return nil, err
		}
	}
	sig := make([][]byte, 2)
	sig[0] = []byte(r.String())
	sig[1] = []byte(s.String())
	return sig, err
}

func (pub *ecdsa_p256_PublicKey) Algorithm() string {
	return pub.SignAlg
}

func (pub *ecdsa_p256_PublicKey) Verify(sig Signature, hash Hash) (bool, error) {
	ecdsaSig := sig.ToECDSA()
	isVerified := ecdsa.Verify(&pub.PublicKey, hash, ecdsaSig.r, ecdsaSig.s)
	return isVerified, nil
}
