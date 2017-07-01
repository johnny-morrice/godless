package api

import (
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
)

type NamespaceTree interface {
	JoinTable(crdt.TableName, crdt.Table) error
	LoadTraverse(searcher NamespaceSearcher) error
}

type RemoteNamespaceTree interface {
	RemoteNamespace
	NamespaceTree
}

type SearchResult struct {
	Namespace            crdt.Namespace
	NamespaceLoadFailure bool
	IndexLoadFailure     bool
}

type NamespaceTreeReader interface {
	ReadSearchResult(result SearchResult) TraversalUpdate
}

type TraversalUpdate struct {
	More  bool
	Error error
}

type IndexSearch interface {
	Search(index crdt.Index) []crdt.Link
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

func (searcher SignedTableSearcher) ReadSearchResult(result SearchResult) TraversalUpdate {
	return searcher.Reader.ReadSearchResult(result)
}

func (searcher SignedTableSearcher) Search(index crdt.Index) []crdt.Link {
	verified := []crdt.Link{}

	needSignature := len(searcher.Keys) > 0

	for _, t := range searcher.Tables {
		index.ForTable(t, func(link crdt.Link) {
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

type NamespaceTreeLambda func(result SearchResult) TraversalUpdate

func (lambda NamespaceTreeLambda) ReadSearchResult(result SearchResult) TraversalUpdate {
	return lambda(result)
}
