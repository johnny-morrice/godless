package api

import (
	"crypto"

	"github.com/johnny-morrice/godless/crdt"
)

type HeadCache interface {
	SetHead(head crdt.IPFSPath) error
	GetHead() (crdt.IPFSPath, error)
	Rollback() error
	Commit() error
}

type RequestPriorityQueue interface {
	Enqueue(request APIRequest)
	PopFront() APIRequest
}

type PublicKeyId string
type PrivateKeyId string

type KeyCache interface {
	StorePrivateKey(priv crypto.PrivateKey) PrivateKeyId
	GetPrivateKey(privId PrivateKeyId) crypto.PrivateKey
	StorePublicKey(pub crypto.PublicKey) PublicKeyId
	GetPublicKey(pubId PublicKeyId) crypto.PublicKey
}
