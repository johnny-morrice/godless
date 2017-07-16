package functional_godless

import (
	"crypto"
	"testing"
	"time"

	"github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/datapeer"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestGodlessRequestFunctionalWithoutCache(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	peerOptions := datapeer.ResidentMemoryStorageOptions{
		Hash: crypto.MD5,
	}
	dataPeer := datapeer.MakeResidentMemoryDataPeer(peerOptions)
	keyStore := godless.MakeKeyStore()
	memoryImage := cache.MakeResidentMemoryImage()

	options := godless.Options{
		DataPeer:       dataPeer,
		KeyStore:       keyStore,
		MemoryImage:    memoryImage,
		ApiConcurrency: 8,
		Pulse:          PULSE,
	}
	godless, err := godless.New(options)
	defer godless.Shutdown()

	testutil.AssertNil(t, err)
	RunRequestResultTests(t, godless, LOCAL_DATA_SIZE)
}

func TestGodlessRequestFunctionalWithCache(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	peerOptions := datapeer.ResidentMemoryStorageOptions{
		Hash: crypto.MD5,
	}
	dataPeer := datapeer.MakeResidentMemoryDataPeer(peerOptions)
	keyStore := godless.MakeKeyStore()
	memoryImage := cache.MakeResidentMemoryImage()
	cache := cache.MakeResidentMemoryCache(BUFFER_SIZE, BUFFER_SIZE)

	options := godless.Options{
		DataPeer:       dataPeer,
		KeyStore:       keyStore,
		MemoryImage:    memoryImage,
		ApiConcurrency: 8,
		Cache:          cache,
		Pulse:          PULSE,
	}
	godless, err := godless.New(options)
	defer godless.Shutdown()
	testutil.AssertNil(t, err)

	RunRequestResultTests(t, godless, LOCAL_DATA_SIZE)
}

func TestGodlessRequestFunctionalWithHttp(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	peerOptions := datapeer.ResidentMemoryStorageOptions{
		Hash: crypto.MD5,
	}
	dataPeer := datapeer.MakeResidentMemoryDataPeer(peerOptions)
	keyStore := godless.MakeKeyStore()
	memoryImage := cache.MakeResidentMemoryImage()

	options := godless.Options{
		DataPeer:       dataPeer,
		KeyStore:       keyStore,
		MemoryImage:    memoryImage,
		ApiConcurrency: 8,
		Pulse:          PULSE,
		WebServiceAddr: BIND_ADDR,
	}
	godless, err := godless.New(options)
	defer godless.Shutdown()
	testutil.AssertNil(t, err)

	clientOptions := http.ClientOptions{
		ServerAddr: SERVER_ADDR,
	}
	client, err := http.MakeClient(clientOptions)
	testutil.AssertNil(t, err)

	RunRequestResultTests(t, client, HTTP_DATA_SIZE)
}

func TestGodlessReplicateFunctional(t *testing.T) {
	// This test checks online replication service.
	t.FailNow()
}

const BUFFER_SIZE = 10000
const PULSE = time.Millisecond * 100
const BIND_ADDR = "localhost:33333"
const SERVER_ADDR = "http://localhost:33333"
const LOCAL_DATA_SIZE = 500
const HTTP_DATA_SIZE = 100
