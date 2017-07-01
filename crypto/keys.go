package crypto

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strings"
	"sync"

	"github.com/johnny-morrice/godless/log"
	crypto "github.com/libp2p/go-libp2p-crypto"
	mh "github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
)

type PublicKeyText []byte

type PrivateKeyText []byte

type PublicKeyHash []byte

func PublicKeysAsText(publicKeys []PublicKey) string {
	pubTexts := make([]string, 0, len(publicKeys))

	for _, pub := range publicKeys {
		text, err := SerializePublicKey(pub)

		if err != nil {
			log.Error("Failed to serialize PublicKey: %s", err.Error())
		}

		pubTexts = append(pubTexts, string(text))
	}

	return strings.Join(pubTexts, __KEY_SEPERATOR)
}

// TODO maybe return the errors.
// These are high level functions extracted from client code.
func PrivateKeysAsText(privateKeys []PrivateKey) string {
	privTexts := make([]string, 0, len(privateKeys))

	for _, priv := range privateKeys {
		text, err := SerializePrivateKey(priv)
		if err != nil {
			log.Error("Failed to serialize PrivateKey: %s", err.Error())
		}

		privTexts = append(privTexts, string(text))
	}

	return strings.Join(privTexts, __KEY_SEPERATOR)
}

func PrivateKeysFromText(privateKeyText string) []PrivateKey {
	parts := strings.Split(privateKeyText, __KEY_SEPERATOR)

	keys := make([]PrivateKey, 0, len(parts))

	for _, text := range parts {
		priv, err := ParsePrivateKey(PrivateKeyText(text))

		if err != nil {
			log.Error("Failed to parse PrivateKey: %s", err.Error())
		}

		keys = append(keys, priv)
	}

	return keys
}

func PublicKeysFromText(publicKeyText string) []PublicKey {
	parts := strings.Split(publicKeyText, __KEY_SEPERATOR)

	keys := make([]PublicKey, 0, len(parts))

	for _, text := range parts {
		pub, err := ParsePublicKey(PublicKeyText(text))

		if err != nil {
			log.Error("Failed to parse PublicKey: %s", err.Error())
		}

		keys = append(keys, pub)
	}

	return keys
}

func (hash PublicKeyHash) Equals(other PublicKeyHash) bool {
	cmp := bytes.Compare(hash, other)
	return cmp == 0
}

func ParsePublicKey(text PublicKeyText) (PublicKey, error) {
	unBased := decodeBase58(string(text))
	p2pKey, err := crypto.UnmarshalPublicKey(unBased)
	if err != nil {
		return PublicKey{}, nil
	}

	return PublicKey{p2pKey: p2pKey}, nil
}

func ParsePrivateKey(text PrivateKeyText) (PrivateKey, error) {
	unBased := decodeBase58(string(text))
	p2pKey, err := crypto.UnmarshalPrivateKey(unBased)
	if err != nil {
		return PrivateKey{}, nil
	}

	return PrivateKey{p2pKey: p2pKey}, nil
}

func SerializePublicKey(pub PublicKey) (PublicKeyText, error) {
	key, err := crypto.MarshalPublicKey(pub.p2pKey)

	if err != nil {
		return NIL_PUBLIC_KEY_TEXT, err
	}

	return PublicKeyText(encodeBase58(key)), nil
}

func SerializePrivateKey(priv PrivateKey) (PrivateKeyText, error) {
	key, err := crypto.MarshalPrivateKey(priv.p2pKey)

	if err != nil {
		return NIL_PRIVATE_KEY_TEXT, err
	}

	return PrivateKeyText(encodeBase58(key)), nil
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

	myBytes, myErr := SerializePrivateKey(priv)

	if myErr != nil {
		log.Error(failMsg)
		return false
	}

	otherBytes, otherErr := SerializePrivateKey(priv)

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
	const errFmt = "Could not serialize PublicKey for Equals: %s"

	if pub.p2pKey == nil && other.p2pKey == nil {
		return true
	}

	myBytes, myErr := SerializePublicKey(pub)

	if myErr != nil {
		log.Error(errFmt, myErr.Error())
		return false
	}

	theirBytes, theirErr := SerializePublicKey(other)

	if theirErr != nil {
		log.Error(errFmt, theirErr.Error())
		return false
	}

	cmp := bytes.Compare(myBytes, theirBytes)
	return cmp == 0
}

func (pub PublicKey) Hash() (PublicKeyHash, error) {
	const failMsg = "PublicKey.Hash failed"

	bs, err := keyHash(pub.p2pKey)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	return PublicKeyHash(encodeBase58(bs)), nil
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

const __KEY_SEPERATOR = ":"
