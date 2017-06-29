package cache

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"

	pb "github.com/gogo/protobuf/proto"
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
	return initBucket(memimg.db, BOLT_MEMORY_IMAGE_BUCKET_NAME)
}

func (memimg boltMemoryImage) GetIndex() (crdt.Index, error) {
	const failMsg = "boltMemoryIndex.GetIndex failed"
	indexMessage := &proto.IndexMessage{}

	err := memimg.view(func(bucket *bolt.Bucket) error {
		return getMessage(bucket, BOLT_MEMORY_IMAGE_INDEX_KEY, indexMessage)
	})

	if err != nil {
		return crdt.EmptyIndex(), errors.Wrap(err, failMsg)
	}

	// TODO handle the invalid entries.
	index, _ := crdt.ReadIndexMessage(indexMessage)

	return index, nil
}

func (memimg boltMemoryImage) JoinIndex(index crdt.Index) error {
	const failMsg = "boltMemoryIndex.JoinIndex failed"

	currentIndex, err := memimg.GetIndex()

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	joinedIndex := currentIndex.JoinIndex(index)

	// TODO handle the invalid entries.
	joinedMessage, _ := crdt.MakeIndexMessage(joinedIndex)

	err = memimg.update(func(bucket *bolt.Bucket) error {
		return putMessage(bucket, BOLT_MEMORY_IMAGE_INDEX_KEY, joinedMessage)
	})

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func (memimg boltMemoryImage) view(viewer func(bucket *bolt.Bucket) error) error {
	return memimg.db.View(func(transaction *bolt.Tx) error {
		bucket := transaction.Bucket(BOLT_MEMORY_IMAGE_BUCKET_NAME)
		return viewer(bucket)
	})
}

func (memimg boltMemoryImage) update(updater func(bucket *bolt.Bucket) error) error {
	return memimg.db.Update(func(transaction *bolt.Tx) error {
		bucket := transaction.Bucket(BOLT_MEMORY_IMAGE_BUCKET_NAME)
		return updater(bucket)
	})
}

func connectBolt(options BoltOptions) (*bolt.DB, error) {
	return bolt.Open(options.FilePath, options.Mode, options.DBOptions)
}

func initBucket(db *bolt.DB, bucketName []byte) error {
	return db.Update(func(transaction *bolt.Tx) error {
		_, err := transaction.CreateBucketIfNotExists(bucketName)
		return err
	})
}

func putMessage(bucket *bolt.Bucket, key []byte, value pb.Message) error {
	keyText := string(key)
	valueBytes, err := pb.Marshal(value)

	if err != nil {
		msg := fmt.Sprintf("Failed to Marshal protobuf message for Bolt key: %v", keyText)
		return errors.Wrap(err, msg)
	}

	err = bucket.Put(key, valueBytes)

	if err != nil {
		msg := fmt.Sprintf("Failed to Put value at Bolt key: %v", keyText)
		return errors.Wrap(err, msg)
	}

	return nil
}

func getMessage(bucket *bolt.Bucket, key []byte, value pb.Message) error {
	keyText := string(key)
	valueBytes := bucket.Get(key)

	if valueBytes == nil {
		return fmt.Errorf("Failed to Get value at Bolt key: %v", keyText)
	}

	err := pb.Unmarshal(valueBytes, value)

	if err != nil {
		msg := fmt.Sprintf("Failed to Unmarshal protobuf message for Bolt key: %v", keyText)
		return errors.Wrap(err, msg)
	}

	return nil
}

var BOLT_MEMORY_IMAGE_INDEX_KEY = []byte("current_index")
var BOLT_MEMORY_IMAGE_BUCKET_NAME = []byte("memory_image")
var BOLT_CACHE_BUCKET_NAME = []byte("cache")
