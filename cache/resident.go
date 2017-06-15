package cache

import (
	"errors"
	"sync"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
)

type residentHeadCache struct {
	sync.RWMutex
	current  crdt.IPFSPath
	previous crdt.IPFSPath
}

func (cache *residentHeadCache) SetHead(head crdt.IPFSPath) error {
	cache.Lock()
	cache.previous = cache.current
	cache.current = head
	return nil
}

func (cache *residentHeadCache) Commit() error {
	cache.previous = ""
	cache.Unlock()
	return nil
}

func (cache *residentHeadCache) Rollback() error {
	if crdt.IsNilPath(cache.previous) {
		return errors.New("Cannot rollback: no previous value")
	}

	cache.current = cache.previous
	return cache.Commit()
}

func (cache *residentHeadCache) GetHead() (crdt.IPFSPath, error) {
	cache.RLock()
	head := cache.current
	cache.RUnlock()
	return head, nil
}

func MakeResidentHeadCache() api.HeadCache {
	return &residentHeadCache{}
}
