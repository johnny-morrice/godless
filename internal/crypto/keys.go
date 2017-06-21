package crypto

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/johnny-morrice/godless/log"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/errors"
)

type PublicKeyText []byte

type PrivateKeyText []byte

type PublicKeyHash []byte

func (hash PublicKeyHash) Equals(other PublicKeyHash) bool {
	cmp := bytes.Compare(hash, other)
	return cmp == 0
}

func ParsePublicKey(text PublicKeyText) (PublicKey, error) {
	p2pKey, err := crypto.UnmarshalPublicKey(text)
	if err != nil {
		return PublicKey{}, nil
	}

	return PublicKey{p2pKey: p2pKey}, nil
}

func ParsePrivateKey(text PrivateKeyText) (PrivateKey, error) {
	p2pKey, err := crypto.UnmarshalEd25519PrivateKey(text)
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
	key, err := priv.p2pKey.Bytes()

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

func (priv PrivateKey) SamePublicKey(other PrivateKey) bool {
	myPub := priv.GetPublicKey()
	otherPub := other.GetPublicKey()
	return myPub.Equals(otherPub)
}

func (priv PrivateKey) GetPublicKey() PublicKey {
	return PublicKey{p2pKey: priv.p2pKey.GetPublic()}
}

func (priv PrivateKey) Equals(other PrivateKey) bool {
	const failMsg = "Could not serialize PrivateKey for Equals"

	if priv.p2pKey == nil && other.p2pKey == nil {
		return true
	}

	myBytes, myErr := PrintPrivateKey(priv)

	if myErr != nil {
		log.Error(failMsg)
		return false
	}

	otherBytes, otherErr := PrintPrivateKey(priv)

	if otherErr != nil {
		log.Error(failMsg)
		return false
	}

	return bytes.Compare(myBytes, otherBytes) == 0
}

type PublicKey struct {
	p2pKey crypto.PubKey
}

func (pub PublicKey) Equals(other PublicKey) bool {
	const errFmt = "Could not serialize PublicKey for Equals: %v"

	if pub.p2pKey == nil && other.p2pKey == nil {
		return true
	}

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

func (pub PublicKey) Hash() (PublicKeyHash, error) {
	const failMsg = "PublicKey.Hash failed"

	bs, err := pub.p2pKey.Hash()

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	return PublicKeyHash(bs), nil
}

var NIL_PUBLIC_KEY_TEXT = PublicKeyText([]byte(""))
var NIL_PRIVATE_KEY_TEXT = PrivateKeyText([]byte(""))

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

	err := keys.insertPublicKey(priv.GetPublicKey())

	if err != nil {
		return errors.Wrap(err, failMsg)
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

	return PrivateKey{}, fmt.Errorf("No private key found for: %v", pub)
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

	return PublicKey{}, fmt.Errorf("No Public Key found for %v", hash)
}

func (keys *KeyStore) GetPublicKey(hash PublicKeyHash) (PublicKey, error) {
	keys.Lock()
	defer keys.Unlock()

	keys.init()

	return keys.lookupPublicKey(hash)
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
