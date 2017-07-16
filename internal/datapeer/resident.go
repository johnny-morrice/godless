package datapeer

import (
	"bytes"
	"crypto"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/internal/util"
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
	ResidentMemoryStorageOptions
	hashes map[string][]byte
}

func MakeResidentMemoryStorage(options ResidentMemoryStorageOptions) api.ContentAddressableStorage {
	return &residentMemoryStorage{
		ResidentMemoryStorageOptions: options,
		hashes: map[string][]byte{},
	}
}

func (storage *residentMemoryStorage) Cat(hash string) (io.ReadCloser, error) {
	data, ok := storage.hashes[hash]

	if !ok {
		return nil, fmt.Errorf("Data not found for '%s'", hash)
	}

	buff := &bytes.Buffer{}
	_, err := buff.Write(data)

	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(buff), nil
}

func (storage *residentMemoryStorage) Add(r io.Reader) (string, error) {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return "", err
	}

	hash := storage.Hash.New()
	addressBytes := hash.Sum(data)
	address := util.EncodeBase58(addressBytes)

	_, present := storage.hashes[address]
	if !present {
		storage.hashes[address] = data
	}

	return address, nil
}

type residentMemoryPubSubBus struct {
	bus []residentSubscription
}

func MakeResidentMemoryPubSubBus() api.PubSubber {
	return &residentMemoryPubSubBus{}
}

func (pubsubber *residentMemoryPubSubBus) PubSubPublish(topic, data string) error {
	for _, sub := range pubsubber.bus {
		subscription := sub
		if subscription.topic == topic {
			go subscription.publish(topic, data)
		}
	}

	return nil
}

func (pubsubber *residentMemoryPubSubBus) PubSubSubscribe(topic string) (api.PubSubSubscription, error) {
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
