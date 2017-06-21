package crdt

import (
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type PointText string

type Point struct {
	Text       PointText
	Signatures []crypto.Signature
}

func (p Point) HasText(text string) bool {
	return p.Text == PointText(text)
}

func (p Point) Equals(other Point) bool {
	ok := p.Text == other.Text
	ok = ok && len(p.Signatures) == len(other.Signatures)

	if !ok {
		return false
	}

	for i, mySig := range p.Signatures {
		theirSig := p.Signatures[i]

		if !mySig.Equals(theirSig) {
			return false
		}
	}

	return true
}

func (p Point) IsVerifiedByAny(keys []crypto.PublicKey) bool {
	for _, pub := range keys {
		if p.IsVerifiedBy(pub) {
			return true
		}
	}

	return false
}

func (p Point) IsVerifiedBy(publicKey crypto.PublicKey) bool {
	for _, sig := range p.Signatures {
		ok, err := crypto.Verify(publicKey, []byte(p.Text), sig)

		if err != nil {
			log.Warn("Bad key while verifying Point signature")
			continue
		}

		if ok {
			return true
		}
	}

	log.Warn("Signature verification failed for Point")

	return false
}

func UnsignedPoint(text PointText) Point {
	return Point{Text: text}
}

func SignedPoint(text PointText, keys []crypto.PrivateKey) (Point, error) {
	const failMsg = "SignedPoint failed"

	signed := Point{
		Text:       text,
		Signatures: make([]crypto.Signature, len(keys)),
	}

	for i, priv := range keys {
		sig, err := crypto.Sign(priv, []byte(signed.Text))

		if err != nil {
			return Point{}, errors.Wrap(err, failMsg)
		}

		signed.Signatures[i] = sig
	}

	return signed, nil
}

type byPointValue []Point

func (p byPointValue) Len() int {
	return len(p)
}

func (p byPointValue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p byPointValue) Less(i, j int) bool {
	return p[i].Text < p[j].Text
}

func uniqPointSorted(set []Point) []Point {
	if len(set) < 2 {
		return set
	}

	uniq := make([]Point, 1, len(set))

	uniq[0] = set[0]
	for _, p := range set[1:] {
		last := &uniq[len(uniq)-1]
		if p.Text == last.Text {
			last.Signatures = append(last.Signatures, p.Signatures...)
		} else {
			uniq = append(uniq, p)
		}
	}

	for i := 0; i < len(uniq); i++ {
		point := &uniq[i]
		point.Signatures = crypto.OrderSignatures(point.Signatures)
	}

	return uniq
}
