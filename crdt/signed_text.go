package crdt

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/log"
)

type signedText struct {
	text       []byte
	signatures []crypto.Signature
}

func makeSignedText(text []byte, keys []crypto.PrivateKey) (signedText, error) {
	const failMsg = "makeSignedText failed"

	signed := signedText{
		text:       text,
		signatures: make([]crypto.Signature, len(keys)),
	}

	for i, priv := range keys {
		sig, err := crypto.Sign(priv, signed.text)

		if err != nil {
			return signedText{}, errors.Wrap(err, failMsg)
		}

		signed.signatures[i] = sig
	}

	signed.signatures = crypto.OrderSignatures(signed.signatures)

	return signed, nil
}

func (signed signedText) textLess(other signedText) bool {
	cmp := bytes.Compare(signed.text, other.text)
	return cmp == -1
}

func (signed signedText) SameText(other signedText) bool {
	cmp := bytes.Compare(signed.text, other.text)
	return cmp == 0
}

func (signed signedText) Equals(other signedText) bool {
	ok := signed.SameText(other)
	ok = ok && len(signed.signatures) == len(other.signatures)

	if !ok {
		return false
	}

	for i, mySig := range signed.signatures {
		theirSig := signed.signatures[i]

		if !mySig.Equals(theirSig) {
			return false
		}
	}

	return true
}

func (signed signedText) IsVerifiedByAny(keys []crypto.PublicKey) bool {
	for _, pub := range keys {
		if signed.IsVerifiedBy(pub) {
			return true
		}
	}

	return false
}

func (signed signedText) IsVerifiedBy(publicKey crypto.PublicKey) bool {
	for _, sig := range signed.signatures {
		ok, err := crypto.Verify(publicKey, signed.text, sig)

		if err != nil {
			log.Warn("Bad key while verifying signedText signature")
			continue
		}

		if ok {
			return true
		}
	}

	return false
}
