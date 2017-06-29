package api

import (
	"github.com/johnny-morrice/godless/crypto"
)

type KeyStore interface {
	PutPrivateKey(priv crypto.PrivateKey) error
	GetPrivateKey(hash crypto.PublicKeyHash) (crypto.PrivateKey, error)
	GetAllPrivateKeys() []crypto.PrivateKey
	GetAllPublicKeys() []crypto.PublicKey
	PutPublicKey(pub crypto.PublicKey) error
	GetPublicKey(hash crypto.PublicKeyHash) (crypto.PublicKey, error)
}
