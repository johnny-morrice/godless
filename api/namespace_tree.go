package api

import (
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/crypto"
)

type NamespaceTree interface {
	JoinTable(crdt.TableName, crdt.Table) error
	LoadTraverse(searcher NamespaceSearcher) error
}

type RemoteNamespaceTree interface {
	RemoteNamespace
	NamespaceTree
}

type NamespaceTreeReader interface {
	ReadNamespace(ns crdt.Namespace) TraversalUpdate
}

type TraversalUpdate struct {
	More  bool
	Error error
}

type IndexSearch interface {
	Search(index crdt.Index) []crdt.SignedLink
}

type NamespaceSearcher interface {
	IndexSearch
	NamespaceTreeReader
}

type SignedTableSearcher struct {
	Reader NamespaceTreeReader
	Tables []crdt.TableName
	Keys   []crypto.PublicKey
}

func (searcher SignedTableSearcher) ReadNamespace(ns crdt.Namespace) TraversalUpdate {
	return searcher.Reader.ReadNamespace(ns)
}

func (searcher SignedTableSearcher) Search(index crdt.Index) []crdt.SignedLink {
	verified := []crdt.SignedLink{}

	needSignature := len(searcher.Keys) > 0

	for _, t := range searcher.Tables {
		index.ForTable(t, func(link crdt.SignedLink) {
			if !needSignature {
				verified = append(verified, link)
				return
			}

			if link.IsVerifiedByAny(searcher.Keys) {
				verified = append(verified, link)
			}
		})
	}

	return verified
}

// NamespaceTreeReader functions return true when they have finished reading
// the tree.
type NamespaceTreeLambda func(ns crdt.Namespace) TraversalUpdate

func (ntl NamespaceTreeLambda) ReadNamespace(ns crdt.Namespace) TraversalUpdate {
	return ntl(ns)
}
