package api

import (
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
)

type RemoteNamespace interface {
	JoinTable(crdt.TableName, crdt.Table) (crdt.IPFSPath, error)
	LoadTraverse(searcher NamespaceSearcher) error
}

type RemoteNamespaceCore interface {
	Core
	RemoteNamespace
}

type SearchResult struct {
	Namespace            crdt.Namespace
	NamespaceLoadFailure bool
	IndexLoadFailure     bool
}

type SearchResultTraverser interface {
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
	SearchResultTraverser
}

type SignedTableSearcher struct {
	Reader SearchResultTraverser
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

type SearchResultLambda func(result SearchResult) TraversalUpdate

func (lambda SearchResultLambda) ReadSearchResult(result SearchResult) TraversalUpdate {
	return lambda(result)
}
