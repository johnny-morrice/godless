package crypto

import (
	"bytes"
	"crypto/rand"
	"errors"
	"sort"

	crypto "github.com/libp2p/go-libp2p-crypto"
)

type PublicKeyText []byte

type PrivateKeyText []byte

type SignatureText string

func ParsePublicKey(text PublicKeyText) (PublicKey, error) {
	p2pKey, err := crypto.UnmarshalPublicKey(text)
	if err != nil {
		return PublicKey{}, nil
	}

	return PublicKey{p2pKey: p2pKey}, nil
}

func ParsePrivateKey(text PrivateKeyText) (PrivateKey, error) {
	p2pKey, err := crypto.UnmarshalPrivateKey(text)
	if err != nil {
		return PrivateKey{}, nil
	}

	return PrivateKey{p2pKey: p2pKey}, nil
}

func ParseSignature(text SignatureText) (Signature, error) {
	sigStr := string(text)
	if !IsBase58(sigStr) {
		return Signature{}, errors.New("Signature was not base58 encoded")
	}

	bs := decodeBase58(sigStr)

	return Signature{sig: bs}, nil
}

func PrintSignature(sig Signature) (SignatureText, error) {
	text := encodeBase58(sig.sig)
	return SignatureText(text), nil
}

func PrintPublicKey(pub PublicKey) (PublicKeyText, error) {
	key, err := crypto.MarshalPublicKey(pub.p2pKey)

	if err != nil {
		return NIL_PUBLIC_KEY_TEXT, err
	}

	return PublicKeyText(key), nil
}

func PrintPrivateKey(priv PrivateKey) (PrivateKeyText, error) {
	key, err := crypto.MarshalPrivateKey(priv.p2pKey)

	if err != nil {
		return NIL_PRIVATE_KEY_TEXT, err
	}

	return PrivateKeyText(key), nil
}

func OrderSignatures(sigs []Signature) []Signature {
	sortSignatures(sigs)
	return uniqSigSorted(sigs)
}

func uniqSigSorted(signatures []Signature) []Signature {
	if len(signatures) == 0 {
		return signatures
	}

	uniq := make([]Signature, 1, len(signatures))

	uniq[0] = signatures[0]
	for _, sig := range signatures[1:] {
		last := uniq[len(uniq)-1]
		if !sig.Equals(last) {
			uniq = append(uniq, sig)
		}
	}

	return uniq
}

func sortSignatures(sigs []Signature) {
	sort.Sort(bySignatureText(sigs))
}

type bySignatureText []Signature

func (sigs bySignatureText) Len() int {
	return len(sigs)
}

func (sigs bySignatureText) Swap(i, j int) {
	sigs[i], sigs[j] = sigs[j], sigs[i]
}

func (sigs bySignatureText) Less(i, j int) bool {
	return sigs[i].TextLess(sigs[j])
}

type Signature struct {
	sig []byte
}

func (sig Signature) Equals(other Signature) bool {
	return sig.Cmp(other) == 0
}

func (sig Signature) TextLess(other Signature) bool {
	return sig.Cmp(other) < 0
}

func (sig Signature) Cmp(other Signature) int {
	return bytes.Compare(sig.sig, other.sig)
}

type PrivateKey struct {
	p2pKey crypto.PrivKey
}

type PublicKey struct {
	p2pKey crypto.PubKey
}

func Sign(priv PrivateKey, message []byte) (Signature, error) {
	bs, err := priv.p2pKey.Sign(message)

	if err != nil {
		return Signature{}, err
	}

	return Signature{sig: bs}, err
}

func Verify(pub PublicKey, message []byte, sig Signature) (bool, error) {
	return pub.p2pKey.Verify(message, sig.sig)
}

func GenerateKey() (PrivateKey, PublicKey, error) {
	p2pPriv, p2pPub, err := crypto.GenerateEd25519Key(rand.Reader)

	if err != nil {
		return PrivateKey{}, PublicKey{}, err
	}

	priv := PrivateKey{p2pKey: p2pPriv}
	pub := PublicKey{p2pKey: p2pPub}
	return priv, pub, err
}

var NIL_PUBLIC_KEY_TEXT = PublicKeyText([]byte(""))
var NIL_PRIVATE_KEY_TEXT = PrivateKeyText([]byte(""))
