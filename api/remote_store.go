package api

import "github.com/johnny-morrice/godless/crdt"

type PubSubTopic string

type RemoteStore interface {
	Connect() error
	AddNamespace(crdt.Namespace) (crdt.IPFSPath, error)
	AddIndex(crdt.Index) (crdt.IPFSPath, error)
	CatNamespace(crdt.IPFSPath) (crdt.Namespace, error)
	CatIndex(crdt.IPFSPath) (crdt.Index, error)
	SubscribeAddrStream(topic PubSubTopic) (<-chan crdt.Link, <-chan error)
	PublishAddr(addr crdt.Link, topics []PubSubTopic) error
	Disconnect() error
}
