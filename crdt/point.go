package crdt

import (
	"github.com/johnny-morrice/godless/crypto"
	"github.com/pkg/errors"
)

type PointText string

type Point struct {
	signedText
}

func (p Point) Text() PointText {
	return PointText(p.text)
}

func (p Point) Signatures() []crypto.Signature {
	return p.signatures
}

func (p Point) HasText(text string) bool {
	return p.Text() == PointText(text)
}

func (p Point) Equals(other Point) bool {
	return p.signedText.Equals(other.signedText)
}

func PresignedPoint(text PointText, sigs []crypto.Signature) Point {
	return Point{signedText: signedText{
		text:       []byte(text),
		signatures: sigs,
	}}
}

func UnsignedPoint(text PointText) Point {
	return Point{signedText: signedText{text: []byte(text)}}
}

func SignedPoint(text PointText, keys []crypto.PrivateKey) (Point, error) {
	const failMsg = "SignedPoint failed"

	signed, err := makeSignedText([]byte(text), keys)

	if err != nil {
		return Point{}, errors.Wrap(err, failMsg)
	}

	return Point{signedText: signed}, nil
}

type byPointValue []Point

func (p byPointValue) Len() int {
	return len(p)
}

func (p byPointValue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p byPointValue) Less(i, j int) bool {
	return p[i].Text() < p[j].Text()
}

func uniqPointSorted(set []Point) []Point {
	if len(set) < 2 {
		return set
	}

	uniqIndex := 0
	for i := 1; i < len(set); i++ {
		p := set[i]
		last := &set[uniqIndex]
		if p.Text() == last.Text() {
			last.signedText.signatures = append(last.signedText.signatures, p.Signatures()...)
		} else {
			uniqIndex++
			set[uniqIndex] = p
		}
	}

	set = set[:uniqIndex+1]

	for i := 0; i < len(set); i++ {
		point := &set[i]
		point.signedText.signatures = crypto.OrderSignatures(point.Signatures())
	}

	return set
}
