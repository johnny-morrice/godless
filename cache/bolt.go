package cache

import (
	"os"

	"github.com/boltdb/bolt"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/pkg/errors"
)

type BoltOptions struct {
	DBOptions *bolt.Options
	FilePath  string
	Mode      os.FileMode
	Db        *bolt.DB
}

type BoltFactory struct {
	BoltOptions
}

func (factory BoltFactory) MakeCache() (api.Cache, error) {
	const failMsg = "BoltFactory.MakeCache failed"

	cache := boltCache{db: factory.Db}

	err := cache.initBuckets()

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	return cache, nil
}

func (factory BoltFactory) MakeMemoryImage() (api.MemoryImage, error) {
	const failMsg = "BoltFactory.MakeMemoryImage failed"

	memImg := boltMemoryImage{db: factory.Db}

	err := memImg.initBuckets()

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	return memImg, nil
}

func MakeBoltCacheFactory(options BoltOptions) (BoltFactory, error) {
	const failMsg = "MakeBoltCacheFactory"

	db, err := connectBolt(options)

	if err != nil {
		return BoltFactory{}, errors.Wrap(err, failMsg)
	}

	factory := BoltFactory{
		BoltOptions: options,
	}
	factory.Db = db

	return factory, nil
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

type boltMemoryImage struct {
	db *bolt.DB
}

func (memimg boltMemoryImage) initBuckets() error {
	panic("not implemented")
}

func (memimg boltMemoryImage) GetIndex() (crdt.Index, error) {
	panic("not implemented")
}

func (memimg boltMemoryImage) JoinIndex(index crdt.Index) error {
	panic("not implemented")
}

func connectBolt(options BoltOptions) (*bolt.DB, error) {
	return bolt.Open(options.FilePath, options.Mode, options.DBOptions)
}
