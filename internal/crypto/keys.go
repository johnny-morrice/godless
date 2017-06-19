package crypto

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

type PublicKeyText []byte

type PrivateKeyText []byte

func ParsePublicKey(text PublicKeyText) (PublicKey, error) {
	p2pKey, err := crypto.UnmarshalPublicKey(text)
	if err != nil {
		return PublicKey{}, nil
	}

	return PublicKey{p2pKey: p2pKey}, nil
}

func ParsePrivateKey(text PrivateKeyText) (PrivateKey, error) {
	p2pKey, err := crypto.UnmarshalPrivateKey(text)
	if err != nil {
		return PrivateKey{}, nil
	}

	return PrivateKey{p2pKey: p2pKey}, nil
}

func PrintPublicKey(pub PublicKey) (PublicKeyText, error) {
	key, err := crypto.MarshalPublicKey(pub.p2pKey)

	if err != nil {
		return NIL_PUBLIC_KEY_TEXT, err
	}

	return PublicKeyText(key), nil
}

func PrintPrivateKey(priv PrivateKey) (PrivateKeyText, error) {
	key, err := crypto.MarshalPrivateKey(priv.p2pKey)

	if err != nil {
		return NIL_PRIVATE_KEY_TEXT, err
	}

	return PrivateKeyText(key), nil
}

func GenerateKey() (PrivateKey, PublicKey, error) {
	p2pPriv, p2pPub, err := crypto.GenerateEd25519Key(rand.Reader)

	if err != nil {
		return PrivateKey{}, PublicKey{}, err
	}

	priv := PrivateKey{p2pKey: p2pPriv}
	pub := PublicKey{p2pKey: p2pPub}
	return priv, pub, err
}

type PrivateKey struct {
	p2pKey crypto.PrivKey
}

type PublicKey struct {
	p2pKey crypto.PubKey
}

func (pub PublicKey) Equals(other PublicKey) bool {
	const errFmt = "Could not serialize PublicKey for Equals: %v"

	myBytes, myErr := PrintPublicKey(pub)

	if myErr != nil {
		log.Error(errFmt, myErr)
		return false
	}

	theirBytes, theirErr := PrintPublicKey(other)

	if theirErr != nil {
		log.Error(errFmt, theirErr)
		return false
	}

	cmp := bytes.Compare(myBytes, theirBytes)
	return cmp == 0
}

var NIL_PUBLIC_KEY_TEXT = PublicKeyText([]byte(""))
var NIL_PRIVATE_KEY_TEXT = PrivateKeyText([]byte(""))

// KeyStore is an in-memory association between private and public keys.
type KeyStore struct {
	privKeys []PrivateKey
}

func (keys KeyStore) PutPrivateKey(priv PrivateKey) {
	keys.privKeys = append(keys.privKeys, priv)
}

func (keys KeyStore) GetPrivateKey(pub PublicKey) (PrivateKey, error) {
	for _, priv := range keys.privKeys {
		p2pPrivPub := priv.p2pKey.GetPublic()
		privPub := PublicKey{p2pKey: p2pPrivPub}

		if privPub.Equals(pub) {
			return priv, nil
		}
	}

	return PrivateKey{}, fmt.Errorf("No private key found for: %v", pub)
}
