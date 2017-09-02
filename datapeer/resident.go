package datapeer

import (
	"bytes"
	"crypto"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/log"
)

func MakeResidentMemoryDataPeer(options ResidentMemoryStorageOptions) api.DataPeer {
	storage := MakeResidentMemoryStorage(options)
	pubsubber := MakeResidentMemoryPubSubBus()

	return Union{
		Storage:    storage,
		Publisher:  pubsubber,
		Subscriber: pubsubber,
	}
}

type ResidentMemoryStorageOptions struct {
	Hash crypto.Hash
}

type residentMemoryStorage struct {
	sync.RWMutex
	ResidentMemoryStorageOptions
	hashes map[string][]byte
}

func MakeResidentMemoryStorage(options ResidentMemoryStorageOptions) api.ContentAddressableStorage {
	if options.Hash == 0 {
		options.Hash = crypto.MD5
	}

	return &residentMemoryStorage{
		ResidentMemoryStorageOptions: options,
		hashes: map[string][]byte{},
	}
}

func (storage *residentMemoryStorage) Cat(hash string) (io.ReadCloser, error) {
	log.Info("Catting '%s' from residentMemoryStorage", hash)
	storage.RLock()
	defer storage.RUnlock()
	data, ok := storage.hashes[hash]

	if !ok {
		return nil, fmt.Errorf("Data not found for '%s'", hash)
	}

	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

func (storage *residentMemoryStorage) Add(r io.Reader) (string, error) {
	log.Info("Adding to residentMemoryStorage...")
	storage.Lock()
	defer storage.Unlock()
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return "", err
	}

	hash := storage.Hash.New()
	hash.Write(data)
	addressBytes := hash.Sum(nil)
	address := util.EncodeBase58(addressBytes)

	_, present := storage.hashes[address]
	if !present {
		storage.hashes[address] = data
	}

	log.Info("Added '%s' to residentMemoryStorage", address)

	return address, nil
}

type residentMemoryPubSubBus struct {
	sync.RWMutex
	bus []residentSubscription
}

func MakeResidentMemoryPubSubBus() api.PubSubber {
	return &residentMemoryPubSubBus{}
}

func (pubsubber *residentMemoryPubSubBus) PubSubPublish(topic, data string) error {
	log.Debug("Publishing '%s' to '%s'...", topic, data)

	pubsubber.RLock()
	defer pubsubber.RUnlock()

	for _, sub := range pubsubber.bus {
		subscription := sub
		if subscription.topic == topic {
			go subscription.publish(topic, data)
		}
	}

	return nil
}

func (pubsubber *residentMemoryPubSubBus) PubSubSubscribe(topic string) (api.PubSubSubscription, error) {
	log.Debug("Subscribing to '%s'...", topic)
	pubsubber.Lock()
	defer pubsubber.Unlock()

	subscription := residentSubscription{
		topic:  topic,
		nextch: make(chan api.PubSubRecord),
	}

	pubsubber.bus = append(pubsubber.bus, subscription)

	return subscription, nil
}

type residentSubscription struct {
	topic  string
	nextch chan api.PubSubRecord
}

func (subscription residentSubscription) publish(topic, data string) {
	record := residentPubSubRecord{
		data:   []byte(data),
		topics: []string{topic},
	}

	subscription.nextch <- record
}

func (subscription residentSubscription) Next() (api.PubSubRecord, error) {
	record := <-subscription.nextch
	return record, nil
}

type residentPubSubRecord struct {
	data   []byte
	topics []string
}

func (record residentPubSubRecord) From() string {
	return "Local Memory Peer"
}

func (record residentPubSubRecord) Data() []byte {
	return record.data
}

func (record residentPubSubRecord) SeqNo() int64 {
	return 0
}

// Not sure if TopicIds==Topics.
func (record residentPubSubRecord) TopicIDs() []string {
	return record.topics
}
