package crypto

import (
	"bytes"
	"errors"
	"sort"
)

type SignatureText string

func IsNilSignature(sigText SignatureText) bool {
	return sigText == ""
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

func OrderSignatures(sigs []Signature) []Signature {
	sortSignatures(sigs)
	return uniqSigSorted(sigs)
}

func uniqSigSorted(signatures []Signature) []Signature {
	if len(signatures) < 2 {
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
