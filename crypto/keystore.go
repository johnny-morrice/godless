package crypto

import (
	"fmt"
	"sync"

	crypto "github.com/libp2p/go-libp2p-crypto"
	mh "github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
)

// KeyStore is an in-memory association between private and public keys.
type KeyStore struct {
	sync.Mutex
	privKeys     []PrivateKey
	pubKeys      []PublicKey
	pubKeyHashes []PublicKeyHash
}

func (keys *KeyStore) PutPrivateKey(priv PrivateKey) error {
	const failMsg = "PutPrivateKey failed"

	keys.Lock()
	defer keys.Unlock()

	keys.init()

	for _, other := range keys.privKeys {
		// Regardless of the crypto, prevent this key store being tricked.
		if priv.SamePublicKey(other) {
			return errors.New("private key has duplicate public key")
		}
	}

	isPublicKeyPresent := false

	pub := priv.GetPublicKey()
	for _, otherPub := range keys.pubKeys {
		if pub.Equals(otherPub) {
			isPublicKeyPresent = true
			break
		}
	}

	if !isPublicKeyPresent {
		pubKeyErr := keys.insertPublicKey(priv.GetPublicKey())

		if pubKeyErr != nil {
			return errors.Wrap(pubKeyErr, failMsg)
		}
	}

	keys.privKeys = append(keys.privKeys, priv)
	return nil
}

func (keys *KeyStore) GetPrivateKey(hash PublicKeyHash) (PrivateKey, error) {
	const failMsg = "KeyStore.GetPrivateKey faild"

	keys.Lock()
	defer keys.Unlock()

	keys.init()

	pub, err := keys.lookupPublicKey(hash)

	if err != nil {
		return PrivateKey{}, errors.Wrap(err, failMsg)
	}

	for _, priv := range keys.privKeys {
		p2pPrivPub := priv.p2pKey.GetPublic()
		privPub := PublicKey{p2pKey: p2pPrivPub}

		if privPub.Equals(pub) {
			return priv, nil
		}
	}

	return PrivateKey{}, fmt.Errorf("No private key found for: %s", string(hash))
}

func (keys *KeyStore) GetAllPrivateKeys() []PrivateKey {
	keys.Lock()
	defer keys.Unlock()

	keys.init()

	cpy := make([]PrivateKey, len(keys.privKeys))

	for i, priv := range keys.privKeys {
		cpy[i] = priv
	}

	return cpy
}

func (keys *KeyStore) PutPublicKey(pub PublicKey) error {
	const failMsg = "KeyStore.PutPublicKey failed"

	keys.Lock()
	defer keys.Unlock()

	keys.init()

	err := keys.insertPublicKey(pub)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func (keys *KeyStore) insertPublicKey(pub PublicKey) error {
	hash, err := pub.Hash()

	if err != nil {
		return err
	}

	for _, other := range keys.pubKeyHashes {
		if hash.Equals(other) {
			err := errors.New("duplicate hash")
			return err
		}
	}

	for _, other := range keys.pubKeys {
		if pub.Equals(other) {
			err := errors.New("duplicate key")
			return err
		}
	}

	keys.pubKeys = append(keys.pubKeys, pub)
	keys.pubKeyHashes = append(keys.pubKeyHashes, hash)

	return nil
}

func (keys *KeyStore) lookupPublicKey(hash PublicKeyHash) (PublicKey, error) {
	for i, otherHash := range keys.pubKeyHashes {
		if hash.Equals(otherHash) {
			pub := keys.pubKeys[i]
			return pub, nil
		}
	}

	return PublicKey{}, fmt.Errorf("No Public Key found for %s", string(hash))
}

func (keys *KeyStore) GetPublicKey(hash PublicKeyHash) (PublicKey, error) {
	keys.Lock()
	defer keys.Unlock()

	keys.init()

	return keys.lookupPublicKey(hash)
}

func (keys *KeyStore) GetAllPublicKeys() []PublicKey {
	keys.Lock()
	defer keys.Unlock()

	keys.init()

	pubKeys := make([]PublicKey, len(keys.pubKeys))

	for i, pub := range keys.pubKeys {
		pubKeys[i] = pub
	}

	return pubKeys
}

func (keys *KeyStore) init() {
	if keys.privKeys == nil {
		keys.privKeys = []PrivateKey{}
	}

	if keys.pubKeys == nil {
		keys.pubKeys = []PublicKey{}
	}

	if keys.pubKeyHashes == nil {
		keys.pubKeyHashes = []PublicKeyHash{}
	}
}

// keyHash hashes a key.
func keyHash(k crypto.Key) ([]byte, error) {
	kb, err := k.Bytes()
	if err != nil {
		return nil, err
	}

	h, _ := mh.Sum(kb, mh.SHA2_256, -1)
	return []byte(h), nil
}
