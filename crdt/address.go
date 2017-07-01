package crdt

import (
	"bytes"
	"sort"

	pb "github.com/gogo/protobuf/proto"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"
)

type IPFSPath string

const NIL_PATH IPFSPath = ""

func IsNilPath(path IPFSPath) bool {
	return path == NIL_PATH
}

type Link struct {
	signedText
}

func (link Link) Path() IPFSPath {
	return IPFSPath(link.text)
}

func (link Link) Signatures() []crypto.Signature {
	return link.signatures
}

func SignedLink(path IPFSPath, keys []crypto.PrivateKey) (Link, error) {
	const failMsg = "SignedLink failed"

	signed, err := makeSignedText([]byte(path), keys)

	if err != nil {
		return Link{}, errors.Wrap(err, failMsg)
	}

	return Link{signedText: signed}, nil
}

func UnsignedLink(path IPFSPath) Link {
	return Link{signedText: signedText{
		text: []byte(path),
	}}
}

func PresignedLink(path IPFSPath, sigs []crypto.Signature) Link {
	return Link{signedText: signedText{
		text:       []byte(path),
		signatures: sigs,
	}}
}

func (link Link) Equals(other Link) bool {
	return link.signedText.Equals(other.signedText)
}

func (link Link) SameLink(other Link) bool {
	return link.signedText.SameText(other.signedText)
}

type LinkText string

func PrintLink(link Link) (LinkText, error) {
	const failMsg = "PrintLink failed"

	message, makeErr := MakeLinkMessage(link)

	if makeErr != nil {
		return "", errors.Wrap(makeErr, failMsg)
	}

	buff := bytes.Buffer{}

	pbErr := pb.MarshalText(&buff, message)

	if pbErr != nil {
		return "", errors.Wrap(pbErr, failMsg)
	}

	text := LinkText(buff.String())

	return text, nil
}

func ParseLink(text LinkText) (Link, error) {
	const failMsg = "ParseLink failed"

	message := proto.LinkMessage{}
	pbErr := pb.UnmarshalText(string(text), &message)

	if pbErr != nil {
		return Link{}, errors.Wrap(pbErr, failMsg)
	}

	link, readErr := ReadLinkMessage(&message)

	if readErr != nil {
		return Link{}, errors.Wrap(readErr, failMsg)
	}

	return link, nil
}

func MakeLinkMessage(link Link) (*proto.LinkMessage, error) {
	const failMsg = "MakeLinkMessage failed"

	messageSigs := make([]string, len(link.Signatures()))

	for i, sig := range link.Signatures() {
		text, err := crypto.PrintSignature(sig)

		if err != nil {
			return nil, errors.Wrap(err, failMsg)
		}

		messageSigs[i] = string(text)
	}

	message := &proto.LinkMessage{
		Link:       string(link.Path()),
		Signatures: messageSigs,
	}

	return message, nil
}

func ReadLinkMessage(message *proto.LinkMessage) (Link, error) {
	const failMsg = "ReadLinkMessage failed"

	sigs := make([]crypto.Signature, len(message.Signatures))

	for i, messageSig := range message.Signatures {
		sigText := crypto.SignatureText(messageSig)
		sig, err := crypto.ParseSignature(sigText)

		if err != nil {
			return Link{}, errors.Wrap(err, failMsg)
		}

		sigs[i] = sig
	}

	link := PresignedLink(IPFSPath(message.Link), sigs)
	return link, nil
}

func MergeLinks(links []Link) []Link {
	sort.Sort(byLinkPath(links))
	return uniqLinkSorted(links)
}

type byLinkPath []Link

func (addrs byLinkPath) Len() int {
	return len(addrs)
}

func (addrs byLinkPath) Swap(i, j int) {
	addrs[i], addrs[j] = addrs[j], addrs[i]
}

func (addrs byLinkPath) Less(i, j int) bool {
	a := addrs[i]
	b := addrs[j]

	return a.Path() < b.Path()
}

func uniqLinkSorted(links []Link) []Link {
	if len(links) < 2 {
		return links
	}

	uniqIndex := 0
	for i := 1; i < len(links); i++ {
		p := links[i]
		last := &links[uniqIndex]
		if p.Path() == last.Path() {
			last.signedText.signatures = append(last.signedText.signatures, p.Signatures()...)
		} else {
			uniqIndex++
			links[uniqIndex] = p
		}
	}

	links = links[:uniqIndex+1]

	for i := 0; i < len(links); i++ {
		link := &links[i]
		link.signedText.signatures = crypto.OrderSignatures(link.Signatures())
	}

	return links
}
