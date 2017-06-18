package api

import (
	"crypto"

	"github.com/johnny-morrice/godless/crdt"
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

// TODO implement
func (searcher SignedTableSearcher) Search(index crdt.Index) []crdt.SignedLink {
	panic("not implemented")
}

// NamespaceTreeReader functions return true when they have finished reading
// the tree.
type NamespaceTreeLambda func(ns crdt.Namespace) TraversalUpdate

func (ntl NamespaceTreeLambda) ReadNamespace(ns crdt.Namespace) TraversalUpdate {
	return ntl(ns)
}
