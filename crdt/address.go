package crdt

import (
	"errors"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/log"
	"github.com/johnny-morrice/godless/internal/crypto"
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

type SignedLink struct {
	Link       IPFSPath
	Signatures []crypto.Signature
}

func UnsignedLink(path IPFSPath) SignedLink {
	return SignedLink{Link: path}
}

func PreSignedLink(path IPFSPath, sig crypto.Signature) SignedLink {
	return SignedLink{Link: path, Signatures: []crypto.Signature{sig}}
}

func (link SignedLink) IsVerifiedByAny(keys []crypto.PublicKey) bool {
	for _, pub := range keys {
		if link.IsVerifiedBy(pub) {
			return true
		}
	}

	return false
}

func (link SignedLink) IsVerifiedBy(pub crypto.PublicKey) bool {
	for _, sig := range link.Signatures {
		ok, err := crypto.Verify(pub, []byte(link.Link), sig)

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

func (link SignedLink) Equals(other SignedLink) bool {
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

func (link SignedLink) SameLink(other SignedLink) bool {
	return link.Link == other.Link
}

func MergeLinks(links []SignedLink) []SignedLink {
	sort.Sort(byLinkPath(links))
	return uniqLinkSorted(links)
}

type byLinkPath []SignedLink

func (addrs byLinkPath) Len() int {
	return len(addrs)
}

func (addrs byLinkPath) Swap(i, j int) {
	addrs[i], addrs[j] = addrs[j], addrs[i]
}

func (addrs byLinkPath) Less(i, j int) bool {
	a := addrs[i]
	b := addrs[j]

	if a.Link < b.Link {
		return true
	} else if a.Link > b.Link {
		return false
	}

	return false
}

func uniqLinkSorted(links []SignedLink) []SignedLink {
	if len(links) == 0 {
		return links
	}

	uniq := make([]SignedLink, 1, len(links))

	uniq[0] = links[0]
	for _, p := range links[1:] {
		last := &uniq[len(uniq)-1]
		if p.Link == last.Link {
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
