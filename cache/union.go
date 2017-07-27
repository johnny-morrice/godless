package cache

import (
	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
)

type Union struct {
	HeadCache      api.HeadCache
	IndexCache     api.IndexCache
	NamespaceCache api.NamespaceCache
}

func (cache Union) GetHead() (crdt.IPFSPath, error) {
	if cache.HeadCache == nil {
		return crdt.NIL_PATH, noSuchCache()
	}

	return cache.HeadCache.GetHead()
}

func (cache Union) SetHead(head crdt.IPFSPath) error {
	if cache.HeadCache == nil {
		return noSuchCache()
	}

	return cache.HeadCache.SetHead(head)
}

func (cache Union) GetIndex(indexAddr crdt.IPFSPath) (crdt.Index, error) {
	if cache.IndexCache == nil {
		return crdt.EmptyIndex(), noSuchCache()
	}

	return cache.IndexCache.GetIndex(indexAddr)
}

func (cache Union) SetIndex(indexAddr crdt.IPFSPath, index crdt.Index) error {
	if cache.IndexCache == nil {
		return noSuchCache()
	}

	return cache.IndexCache.SetIndex(indexAddr, index)
}

func (cache Union) GetNamespace(namespaceAddr crdt.IPFSPath) (crdt.Namespace, error) {
	if cache.NamespaceCache == nil {
		return crdt.EmptyNamespace(), noSuchCache()
	}

	return cache.NamespaceCache.GetNamespace(namespaceAddr)
}

func (cache Union) SetNamespace(namespaceAddr crdt.IPFSPath, namespace crdt.Namespace) error {
	if cache.NamespaceCache == nil {
		return noSuchCache()
	}

	return cache.NamespaceCache.SetNamespace(namespaceAddr, namespace)
}

func (cache Union) CloseCache() error {
	return nil
}

func noSuchCache() error {
	return errors.New("No cache available")
}
