package api

import (
	"crypto"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
)

type HeadCache interface {
	SetHead(crdt.IPFSPath)
	GetHead(crdt.IPFSPath)
}

type RequestPriorityQueue interface {
	Enqueue(api.APIRequest)
	PopFront(api.APIRequest)
}

type PublicKeyId string
type PrivateKeyId string

type KeyCache interface {
	StorePrivateKey(crypto.PrivateKey) PrivateKeyId
	GetPrivateKey(PrivateKeyId) crypto.PrivateKey
	StorePublicKey(crypto.PublicKey) PublicKeyId
	GetPublicKey(PublicKeyId) crypto.PublicKey
}
