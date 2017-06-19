package api

import (
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/crypto"
)

type HeadCache interface {
	SetHead(head crdt.IPFSPath) error
	GetHead() (crdt.IPFSPath, error)
}

type RequestPriorityQueue interface {
	Len() int
	Enqueue(request APIRequest, data interface{}) error
	Drain() <-chan interface{}
	Close() error
}

type IndexCache interface {
	GetIndex(indexAddr crdt.IPFSPath) (crdt.Index, error)
	SetIndex(indexAddr crdt.IPFSPath, index crdt.Index) error
}

type KeyStore interface {
	PutPrivateKey(priv crypto.PrivateKey) error
	GetPrivateKey(pub crypto.PublicKey) (crypto.PrivateKey, error)
	GetAllPrivateKeys() []crypto.PrivateKey
}
