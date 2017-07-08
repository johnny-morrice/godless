package api

import (
	"github.com/johnny-morrice/godless/crdt"
)

type CacheCloser interface {
	CloseCache() error
}

type Cache interface {
	HeadCache
	IndexCache
	NamespaceCache
	CacheCloser
}

type HeadCache interface {
	SetHead(head crdt.IPFSPath) error
	GetHead() (crdt.IPFSPath, error)
}

type RequestPriorityQueue interface {
	Len() int
	Enqueue(request Request, data interface{}) error
	Drain() <-chan interface{}
	Close() error
}

type IndexCache interface {
	GetIndex(indexAddr crdt.IPFSPath) (crdt.Index, error)
	SetIndex(indexAddr crdt.IPFSPath, index crdt.Index) error
}

type NamespaceCache interface {
	GetNamespace(namespaceAddr crdt.IPFSPath) (crdt.Namespace, error)
	SetNamespace(namespaceAddr crdt.IPFSPath, namespace crdt.Namespace) error
}
