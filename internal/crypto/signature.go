package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"hash"
	"math/big"
)

type PublicKeyText string

type PrivateKeyText string

func ParsePublicKey(text PublicKeyText) (PublicKey, error) {
	return PublicKey{}, nil
}

func ParsePrivateKey(text PrivateKeyText) (PrivateKey, error) {
	return PrivateKey{}, nil
}

type Signature struct {
	r *big.Int
	s *big.Int
}

type PrivateKey struct {
	curveKey *ecdsa.PrivateKey
}

type PublicKey struct {
	curveKey *ecdsa.PublicKey
}

func Sign(priv PrivateKey, message []byte) (Signature, error) {
	hash := sum(message)
	r, s, err := ecdsa.Sign(rand.Reader, priv.curveKey, hash)

	if err != nil {
		return Signature{}, err
	}

	sig := Signature{r: r, s: s}
	return sig, nil
}

func Verify(pub PublicKey, message []byte, sig Signature) bool {
	hash := sum(message)
	return ecdsa.Verify(pub.curveKey, hash, sig.r, sig.s)
}

func GenerateKey() (PrivateKey, PublicKey, error) {
	curvePriv, err := ecdsa.GenerateKey(curve(), rand.Reader)

	if err != nil {
		return PrivateKey{}, PublicKey{}, err
	}

	priv := PrivateKey{curveKey: curvePriv}
	pub := PublicKey{curveKey: &curvePriv.PublicKey}
	return priv, pub, nil
}

func curve() elliptic.Curve {
	return elliptic.P384()
}

var hasher hash.Hash

func init() {
	hasher = crypto.SHA512.New()
}

func sum(message []byte) []byte {
	return hasher.Sum(message)
}
