package api

import "github.com/johnny-morrice/godless/crdt"

type RemoteStore interface {
	Connect() error
	AddNamespace(crdt.Namespace) (crdt.RemoteStoreAddress, error)
	AddIndex(crdt.Index) (crdt.RemoteStoreAddress, error)
	CatNamespace(crdt.RemoteStoreAddress) (crdt.Namespace, error)
	CatIndex(crdt.RemoteStoreAddress) (crdt.Index, error)
	SubscribeAddrStream(topic crdt.RemoteStoreAddress) (<-chan crdt.RemoteStoreAddress, <-chan error)
	PublishAddr(addr crdt.RemoteStoreAddress, topics []crdt.RemoteStoreAddress) error
	Disconnect() error
}
