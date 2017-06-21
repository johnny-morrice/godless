package crdt

import (
	"math/big"
	"sort"

	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type IPFSPath string

const NIL_PATH IPFSPath = ""

func IsNilPath(path IPFSPath) bool {
	return path == NIL_PATH
}

func ParseHash(hash string) (IPFSPath, error) {
	ok := crypto.IsBase58(hash)

	if !ok {
		return NIL_PATH, errors.New("Hash was not base58 encoded")
	}

	return IPFSPath(hash), nil
}

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

	uniq := make([]Link, 1, len(links))

	uniq[0] = links[0]
	for _, p := range links[1:] {
		last := &uniq[len(uniq)-1]
		if p.Path == last.Path {
			last.Signatures = append(last.Signatures, p.Signatures...)
		} else {
			uniq = append(uniq, p)
		}
	}

	for i := 0; i < len(uniq); i++ {
		link := &uniq[i]
		link.Signatures = crypto.OrderSignatures(link.Signatures)
	}

	return uniq
}

var bigRadix = big.NewInt(58)
