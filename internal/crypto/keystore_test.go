package crypto

import (
	"sync"
	"testing"
	"time"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestKeyStorePrivateKeyConcurrency(t *testing.T) {
	keyStore := &KeyStore{}

	keyCount := __CONCURRENCY_LEVEL / 2
	keys := genTestPrivateKeys(keyCount)

	wg := &sync.WaitGroup{}
	wg.Add(__CONCURRENCY_LEVEL)

	for _, priv := range keys {
		privateKey := priv
		publicKey := priv.GetPublicKey()
		hash, err := publicKey.Hash()

		testutil.AssertNil(t, err)

		go func() {
			defer wg.Done()
			err := keyStore.PutPrivateKey(privateKey)
			testutil.AssertNil(t, err)
		}()

		go func() {
			defer wg.Done()
			for {
				found, err := keyStore.GetPrivateKey(hash)

				if err != nil {
					continue
				}

				assertPrivEquals(t, privateKey, found)
				break
			}
		}()
	}

	testutil.WaitGroupTimeout(t, wg, time.Second*10)

}

func TestKeyStoreGetSetPublicKey(t *testing.T) {
	keyStore := &KeyStore{}

	_, pub, err := GenerateKey()

	testutil.AssertNil(t, err)

	err = keyStore.PutPublicKey(pub)

	testutil.AssertNil(t, err)

	goodHash, goodHashErr := pub.Hash()

	testutil.AssertNil(t, goodHashErr)

	found, foundErr := keyStore.GetPublicKey(goodHash)

	testutil.AssertNil(t, foundErr)

	if !pub.Equals(found) {
		t.Error("Unexpected public key")
	}

	_, badPub, badPubErr := GenerateKey()

	testutil.AssertNil(t, badPubErr)

	badHash, badHashErr := badPub.Hash()

	testutil.AssertNil(t, badHashErr)

	notFound, notFoundErr := keyStore.GetPublicKey(badHash)

	testutil.AssertNonNil(t, notFoundErr)

	if !notFound.Equals(PublicKey{}) {
		t.Error("Expected empty PublicKey")
	}
}

func TestKeyStoreGetSetPrivateKey(t *testing.T) {
	keyStore := &KeyStore{}

	priv, _, err := GenerateKey()

	testutil.AssertNil(t, err)

	putErr := keyStore.PutPrivateKey(priv)
	testutil.AssertNil(t, putErr)
	putErr = keyStore.PutPrivateKey(priv)
	testutil.AssertNonNil(t, putErr)

	goodHash, goodHashErr := priv.GetPublicKey().Hash()

	testutil.AssertNil(t, goodHashErr)

	found, noGetErr := keyStore.GetPrivateKey(goodHash)

	testutil.AssertNil(t, noGetErr)

	if priv != found {
		t.Error("Unexpected private key")
	}

	_, pub, err := GenerateKey()

	testutil.AssertNil(t, err)

	badHash, badHashErr := pub.Hash()

	testutil.AssertNil(t, badHashErr)

	notFound, getErr := keyStore.GetPrivateKey(badHash)

	testutil.AssertNonNil(t, getErr)

	if !notFound.Equals(PrivateKey{}) {
		t.Error("Expected empty Private Key")
	}
}

func genTestPublicKeys(count int) []PublicKey {
	pubKeys := make([]PublicKey, count)

	for i := 0; i < count; i++ {
		_, pub, err := GenerateKey()

		setupPanic(err)

		pubKeys[i] = pub
	}

	return pubKeys
}

func genTestPrivateKeys(count int) []PrivateKey {
	privKeys := make([]PrivateKey, count)

	for i := 0; i < count; i++ {
		priv, _, err := GenerateKey()

		setupPanic(err)

		privKeys[i] = priv
	}

	return privKeys
}

func assertPrivEquals(t *testing.T, expected PrivateKey, actual PrivateKey) {
	if !expected.Equals(actual) {
		t.Error("Unexpected private key")
	}
}

const __CONCURRENCY_LEVEL = 50
