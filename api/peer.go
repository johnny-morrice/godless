package api

import (
	"io"
)

type ContentAddressableStorage interface {
	Cat(address string) (io.Reader, error)
	Add(r io.Reader) (string, error)
}

type PubSubPublisher interface {
	PubSubPublish(topic, data string) error
}

type PubSubSubscriber interface {
	PubSubSubscribe(topic string) (PubSubSubscription, error)
}

type PubSubSubscription interface {
	Next() (PubSubRecord, error)
}

type PubSubRecord interface {
	// From returns the peer ID of the node that published this record
	From() string

	// Data returns the data field
	Data() []byte

	// SeqNo is the sequence number of this record
	SeqNo() int64

	//TopicIDs is the list of topics this record belongs to
	TopicIDs() []string
}
