package crdt

import (
	"sort"

	"github.com/johnny-morrice/godless/internal/crypto"
)

type IPFSPath string

const NIL_PATH IPFSPath = ""

func IsNilPath(path IPFSPath) bool {
	return path == NIL_PATH
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
		link.Signatures = crypto.UniqSignatures(link.Signatures)
	}

	return uniq
}
