package crypto

import (
	"sync"
	"testing"
	"time"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestKeyStoreConcurrency(t *testing.T) {
	keyStore := &KeyStore{}

	keyCount := __CONCURRENCY_LEVEL / 2
	keys := genTestPrivateKeys(keyCount)

	wg := &sync.WaitGroup{}
	wg.Add(__CONCURRENCY_LEVEL)

	for _, priv := range keys {
		privateKey := priv
		publicKey := priv.GetPublicKey()

		go func() {
			defer wg.Done()
			err := keyStore.PutPrivateKey(privateKey)
			testutil.AssertNil(t, err)
		}()

		go func() {
			defer wg.Done()
			for {
				found, err := keyStore.GetPrivateKey(publicKey)

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

func TestKeyStoreDuplicate(t *testing.T) {
	keyStore := &KeyStore{}

	priv, _, err := GenerateKey()

	testutil.AssertNil(t, err)

	putErr := keyStore.PutPrivateKey(priv)
	testutil.AssertNil(t, putErr)
	putErr = keyStore.PutPrivateKey(priv)
	testutil.AssertNonNil(t, putErr)
}

func TestKeyStoreNotPresent(t *testing.T) {
	keyStore := &KeyStore{}

	_, pub, err := GenerateKey()

	testutil.AssertNil(t, err)

	priv, getErr := keyStore.GetPrivateKey(pub)

	testutil.AssertNonNil(t, getErr)

	if !priv.Equals(PrivateKey{}) {
		t.Error("Expected empty Private Key")
	}
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
