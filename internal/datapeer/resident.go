package datapeer

import (
	"crypto"

	"github.com/johnny-morrice/godless/api"
)

type ResidentMemoryStorageOptions struct {
	Hash crypto.Hash
}

type residentMemoryStorage struct {
}

func MakeResidentMemoryStorage(options ResidentMemoryStorageOptions) api.ContentAddressableStorage {
	panic("not implemented")
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
