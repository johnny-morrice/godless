package functional_godless

import (
	"crypto"
	"testing"
	"time"

	"github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
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

	godless, err := godlessWithoutCache()
	defer godless.Shutdown()

	testutil.AssertNil(t, err)
	joinThenQuery(godless, LOCAL_DATA_SIZE, t)
}

func TestGodlessRequestFunctionalWithCache(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	godless, err := godlessWithCache()
	defer godless.Shutdown()
	testutil.AssertNil(t, err)

	joinThenQuery(godless, LOCAL_DATA_SIZE, t)
}

func TestGodlessRequestFunctionalWithHttp(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	godless, err := godlessWithHttp()
	defer godless.Shutdown()
	testutil.AssertNil(t, err)

	clientOptions := http.ClientOptions{
		ServerAddr: SERVER_ADDR,
	}
	client, err := http.MakeClient(clientOptions)
	testutil.AssertNil(t, err)

	joinThenQuery(client, HTTP_DATA_SIZE, t)
}

func TestGodlessReplicateFunctional(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	const topic = "The Topic"
	peerOptions := datapeer.ResidentMemoryStorageOptions{
		Hash: crypto.MD5,
	}
	dataPeer := datapeer.MakeResidentMemoryDataPeer(peerOptions)

	peers := make([]*godless.Godless, REPLICATING_PEER_COUNT)
	for i := 0; i < REPLICATING_PEER_COUNT; i++ {
		peer, err := godlessWithPeer(dataPeer, topic)
		if err != nil {
			t.Error(err)
			t.FailNow()
			return
		}

		peers[i] = peer
	}

	// FIXME wait for peers to subscribe.
	timeout := time.NewTimer(TIMEOUT)
	<-timeout.C

	for _, peer := range peers {
		request := api.MakeReflectRequest(api.REFLECT_DUMP_NAMESPACE)
		resp, _ := peer.Send(request)
		testutil.Assert(t, "Expected empty namespace", resp.Namespace.IsEmpty())
	}

	query := forceCompile("join office rows (@key=headquarters, manager=\"Bob\")")
	request := api.MakeQueryRequest(query)

	_, err := peers[0].Send(request)
	testutil.AssertNil(t, err)

	for i, peer := range peers {
		assertFindNamespaceWithin(t, peer, i, TIMEOUT)
	}
}

func assertFindNamespaceWithin(t *testing.T, godless *godless.Godless, peerNumber int, duration time.Duration) {
	timeout := time.NewTimer(duration)
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			t.Errorf("Timeout expired without finding namespace on peer %d", peerNumber)
			return
		default:
			request := api.MakeReflectRequest(api.REFLECT_DUMP_NAMESPACE)
			resp, _ := godless.Send(request)

			if !resp.Namespace.IsEmpty() {
				timeout.Stop()
				return
			}
		}
	}

}

func panicOnBadInit(err error) {
	if err != nil {
		panic(err)
	}
}

func godlessWithPeer(dataPeer api.DataPeer, topics ...string) (*godless.Godless, error) {
	keyStore := godless.MakeKeyStore()
	memoryImage := cache.MakeResidentMemoryImage()

	options := godless.Options{
		DataPeer:          dataPeer,
		KeyStore:          keyStore,
		MemoryImage:       memoryImage,
		ApiConcurrency:    8,
		Pulse:             PULSE,
		Topics:            topics,
		ReplicateInterval: REPLICATE_INTERVAL,
		PublicServer:      true,
	}

	return godless.New(options)

}

func godlessWithoutCache() (*godless.Godless, error) {
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
	return godless.New(options)
}

func godlessWithCache() (*godless.Godless, error) {
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
	return godless.New(options)
}

func godlessWithHttp() (*godless.Godless, error) {
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
	return godless.New(options)
}

const BUFFER_SIZE = 10000
const PULSE = time.Millisecond * 100
const BIND_ADDR = "localhost:33333"
const SERVER_ADDR = "http://localhost:33333"
const LOCAL_DATA_SIZE = 500
const HTTP_DATA_SIZE = 100
const REPLICATING_PEER_COUNT = 10
const REPLICATE_INTERVAL = time.Millisecond * 100
const TIMEOUT = time.Millisecond * 500
