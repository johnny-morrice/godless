package cache

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type BoltOptions struct {
	DBOptions *bolt.Options
	FilePath  string
	Mode      os.FileMode
	Db        *bolt.DB
}

type boltCache struct {
	db *bolt.DB
}

func (cache boltCache) initBuckets() error {
	panic("not implemented")
}

func (cache boltCache) Close() error {
	panic("not implemented")
}

func (cache boltCache) GetHead() (crdt.IPFSPath, error) {
	panic("not implemented")
}

func (cache boltCache) SetHead(head crdt.IPFSPath) error {
	panic("not implemented")
}

func (cache boltCache) GetIndex(indexAddr crdt.IPFSPath) (crdt.Index, error) {
	panic("not implemented")
}

func (cache boltCache) SetIndex(indexAddr crdt.IPFSPath, index crdt.Index) error {
	panic("not implemented")
}

func (cache boltCache) GetNamespace(namespaceAddr crdt.IPFSPath) (crdt.Namespace, error) {
	panic("not implemented")
}

func (cache boltCache) SetNamespace(namespaceAddr crdt.IPFSPath, namespace crdt.Namespace) error {
	panic("not implemented")
}

func (cache boltCache) CloseCache() error {
	panic("not implemented")
}

func MakeBoltCache(options BoltOptions) (api.Cache, error) {
	const failMsg = "MakeBoltCache failed"

	db, err := connectBolt(options)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	cache := boltCache{db: db}

	err = cache.initBuckets()

	if err != nil {
		closeErr := cache.Close()

		if closeErr != nil {
			log.Error("Error closing database: %v", err.Error())
		}

		return nil, errors.Wrap(err, failMsg)
	}

	return cache, nil
}

func connectBolt(options BoltOptions) (*bolt.DB, error) {
	return bolt.Open(options.FilePath, options.Mode, options.DBOptions)
}

func MakeBoltMemoryImage() api.MemoryImage {
	panic("not implemented")
}
