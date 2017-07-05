package crypto

import (
	"bytes"
	"crypto/rand"
	"strings"

	"github.com/johnny-morrice/godless/log"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/errors"
)

type PublicKeyText []byte

type PrivateKeyText []byte

type PublicKeyHash []byte

func PublicKeysAsText(publicKeys []PublicKey) (string, error) {
	pubTexts := make([]string, 0, len(publicKeys))

	for _, pub := range publicKeys {
		text, err := SerializePublicKey(pub)

		if err != nil {
			return "", err
		}

		pubTexts = append(pubTexts, string(text))
	}

	text := strings.Join(pubTexts, __KEY_SEPERATOR)
	return text, nil
}

func PrivateKeysAsText(privateKeys []PrivateKey) (string, error) {
	privTexts := make([]string, 0, len(privateKeys))

	for _, priv := range privateKeys {
		text, err := SerializePrivateKey(priv)
		if err != nil {
			return "", err
		}

		privTexts = append(privTexts, string(text))
	}

	text := strings.Join(privTexts, __KEY_SEPERATOR)
	return text, nil
}

func PrivateKeysFromText(text string) ([]PrivateKey, error) {
	parts := strings.Split(text, __KEY_SEPERATOR)

	keys := make([]PrivateKey, 0, len(parts))

	for _, text := range parts {
		priv, err := ParsePrivateKey(PrivateKeyText(text))

		if err != nil {
			return nil, err
		}

		keys = append(keys, priv)
	}

	return keys, nil
}

func PublicKeysFromText(text string) ([]PublicKey, error) {
	parts := strings.Split(text, __KEY_SEPERATOR)

	keys := make([]PublicKey, 0, len(parts))

	for _, text := range parts {
		pub, err := ParsePublicKey(PublicKeyText(text))

		if err != nil {
			return nil, err
		}

		keys = append(keys, pub)
	}

	return keys, nil
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

const __KEY_SEPERATOR = ":"
