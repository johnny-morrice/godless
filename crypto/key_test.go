package crypto

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/internal/testutil"
)

type pubKeyList []PublicKey
type privKeyList []PrivateKey

func (keys privKeyList) Generate(rand *rand.Rand, size int) reflect.Value {
	privKeys := generateKeys(size)
	gen := privKeyList(privKeys)
	return reflect.ValueOf(gen)
}

func (keys pubKeyList) Generate(rand *rand.Rand, size int) reflect.Value {
	privKeys := generateKeys(size)

	pubs := make([]PublicKey, len(privKeys))

	for i, priv := range privKeys {
		pubs[i] = priv.GetPublicKey()
	}

	gen := pubKeyList(pubs)
	return reflect.ValueOf(gen)
}

func (pub PublicKey) Generate(rand *rand.Rand, size int) reflect.Value {
	_, pub, err := GenerateKey()

	if err != nil {
		setupPanic(err)
	}

	return reflect.ValueOf(pub)
}

func (priv PrivateKey) Generate(rand *rand.Rand, size int) reflect.Value {
	priv, _, err := GenerateKey()

	if err != nil {
		setupPanic(err)
	}

	return reflect.ValueOf(priv)
}

func TestParsePublicKeys(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(parsePublicKeyOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func parsePublicKeyOk(expected PublicKey) bool {
	text, err := SerializePublicKey(expected)

	if err != nil {
		return false
	}

	actual, err := ParsePublicKey(text)

	if err != nil {
		return false
	}

	return expected.Equals(actual)
}

func TestParsePrivateKeys(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(parsePrivateKeyOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func parsePrivateKeyOk(expected PrivateKey) bool {
	text, err := SerializePrivateKey(expected)

	if err != nil {
		return false
	}

	actual, err := ParsePrivateKey(text)

	if err != nil {
		return false
	}

	return expected.Equals(actual)
}

func TestPrivateKeyListText(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(privateKeyListTextOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func privateKeyListTextOk(expected privKeyList) bool {
	text, err := PrivateKeysAsText(expected)

	if err != nil {
		return false
	}

	actual, err := PrivateKeysFromText(text)

	if err != nil {
		return false
	}

	for i, ex := range expected {
		ac := actual[i]
		if !ex.Equals(ac) {
			return false
		}
	}

	return true
}

func TestPublicKeyListText(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(publicKeyListTextOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func publicKeyListTextOk(expected pubKeyList) bool {
	text, err := PublicKeysAsText(expected)

	if err != nil {
		return false
	}

	actual, err := PublicKeysFromText(text)

	if err != nil {
		return false
	}

	for i, ex := range expected {
		ac := actual[i]
		if !ex.Equals(ac) {
			return false
		}
	}

	return true
}

func generateKeys(size int) []PrivateKey {
	keys := make([]PrivateKey, size)
	for i := 0; i < size; i++ {
		priv, _, err := GenerateKey()
		setupPanic(err)
		keys[i] = priv
	}

	return keys
}
