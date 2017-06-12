package api

import "github.com/johnny-morrice/godless/crdt"

type RemoteStore interface {
	Connect() error
	AddNamespace(crdt.Namespace) (crdt.IPFSPath, error)
	AddIndex(crdt.Index) (crdt.IPFSPath, error)
	CatNamespace(crdt.IPFSPath) (crdt.Namespace, error)
	CatIndex(crdt.IPFSPath) (crdt.Index, error)
	SubscribeAddrStream(topic crdt.IPFSPath) (<-chan crdt.IPFSPath, <-chan error)
	PublishAddr(addr crdt.IPFSPath, topics []crdt.IPFSPath) error
	Disconnect() error
}
