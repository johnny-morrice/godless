package cache

import (
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
)

type cacheUnion struct {
	headCache      api.HeadCache
	indexCache     api.IndexCache
	namespaceCache api.NamespaceCache
}

func (cache cacheUnion) GetHead() (crdt.IPFSPath, error) {
	return cache.headCache.GetHead()
}

func (cache cacheUnion) SetHead(head crdt.IPFSPath) error {
	return cache.headCache.SetHead(head)
}

func (cache cacheUnion) GetIndex(indexAddr crdt.IPFSPath) (crdt.Index, error) {
	return cache.indexCache.GetIndex(indexAddr)
}

func (cache cacheUnion) SetIndex(indexAddr crdt.IPFSPath, index crdt.Index) error {
	return cache.indexCache.SetIndex(indexAddr, index)
}

func (cache cacheUnion) GetNamespace(namespaceAddr crdt.IPFSPath) (crdt.Namespace, error) {
	return cache.namespaceCache.GetNamespace(namespaceAddr)
}

func (cache cacheUnion) SetNamespace(namespaceAddr crdt.IPFSPath, namespace crdt.Namespace) error {
	return cache.namespaceCache.SetNamespace(namespaceAddr, namespace)
}
