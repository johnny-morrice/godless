package crdt

import (
	"bytes"
	"math/big"
	"sort"

	pb "github.com/gogo/protobuf/proto"
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"
)

type IPFSPath string

const NIL_PATH IPFSPath = ""

func IsNilPath(path IPFSPath) bool {
	return path == NIL_PATH
}

// func ParseHash(hash string) (IPFSPath, error) {
// 	ok := crypto.IsBase58(hash)
//
// 	if !ok {
// 		return NIL_PATH, errors.New("Hash was not base58 encoded")
// 	}
//
// 	return IPFSPath(hash), nil
// }

type Link struct {
	Path       IPFSPath
	Signatures []crypto.Signature
}

func SignedLink(path IPFSPath, keys []crypto.PrivateKey) (Link, error) {
	const failMsg = "SignedLink failed"

	signed := Link{
		Path:       path,
		Signatures: make([]crypto.Signature, len(keys)),
	}

	for i, priv := range keys {
		sig, err := crypto.Sign(priv, []byte(signed.Path))

		if err != nil {
			return Link{}, errors.Wrap(err, failMsg)
		}

		signed.Signatures[i] = sig
	}

	return signed, nil
}

func UnsignedLink(path IPFSPath) Link {
	return Link{Path: path}
}

func PreSignedLink(path IPFSPath, sig crypto.Signature) Link {
	return Link{Path: path, Signatures: []crypto.Signature{sig}}
}

func (link Link) IsVerifiedByAny(keys []crypto.PublicKey) bool {
	for _, pub := range keys {
		if link.IsVerifiedBy(pub) {
			return true
		}
	}

	return false
}

func (link Link) IsVerifiedBy(pub crypto.PublicKey) bool {
	for _, sig := range link.Signatures {
		ok, err := crypto.Verify(pub, []byte(link.Path), sig)

		if err != nil {
			log.Warn("Bad key while verifying SignedLink signature")
			continue
		}

		if ok {
			return true
		}
	}

	log.Warn("Signature verification failed for SignedLink")

	return false
}

func (link Link) Equals(other Link) bool {
	ok := link.SameLink(other)
	ok = ok && len(link.Signatures) == len(other.Signatures)

	if !ok {
		return false
	}

	for i, mySig := range link.Signatures {
		theirSig := other.Signatures[i]

		if !mySig.Equals(theirSig) {
			return false
		}
	}

	return true
}

func (link Link) SameLink(other Link) bool {
	return link.Path == other.Path
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

	messageSigs := make([]string, len(link.Signatures))

	for i, sig := range link.Signatures {
		text, err := crypto.PrintSignature(sig)

		if err != nil {
			return nil, errors.Wrap(err, failMsg)
		}

		messageSigs[i] = string(text)
	}

	message := &proto.LinkMessage{
		Link:       string(link.Path),
		Signatures: messageSigs,
	}

	log.Debug("Created link message: %v", message)

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

	link := Link{
		Path:       IPFSPath(message.Link),
		Signatures: sigs,
	}

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

	if a.Path < b.Path {
		return true
	} else if a.Path > b.Path {
		return false
	}

	return false
}

func uniqLinkSorted(links []Link) []Link {
	if len(links) < 2 {
		return links
	}

	uniqIndex := 0
	for i := 1; i < len(links); i++ {
		p := links[i]
		last := &links[uniqIndex]
		if p.Path == last.Path {
			last.Signatures = append(last.Signatures, p.Signatures...)
		} else {
			uniqIndex++
			links[uniqIndex] = p
		}
	}

	links = links[:uniqIndex+1]

	for i := 0; i < len(links); i++ {
		link := &links[i]
		link.Signatures = crypto.OrderSignatures(link.Signatures)
	}

	return links
}

var bigRadix = big.NewInt(58)
