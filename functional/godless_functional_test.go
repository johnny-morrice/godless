package functional_godless

import (
	"crypto"
	"testing"

	"github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/datapeer"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestGodlessRequestFunctionalWithoutCache(t *testing.T) {
	peerOptions := datapeer.ResidentMemoryStorageOptions{
		Hash: crypto.MD5,
	}
	dataPeer := datapeer.MakeResidentMemoryDataPeer(peerOptions)
	keyStore := godless.MakeKeyStore()
	memoryImage := cache.MakeResidentMemoryImage()

	options := godless.Options{
		DataPeer:    dataPeer,
		KeyStore:    keyStore,
		MemoryImage: memoryImage,
	}
	godless, err := godless.New(options)
	testutil.AssertNil(t, err)
	client := localClient{godless: godless}
	RunRequestResultTests(t, client)
}

func TestGodlessReplicateFunctional(t *testing.T) {
	// This test checks online replication service.
	t.FailNow()
}

// FIXME should be provided by main lib - needs API change.
type localClient struct {
	godless *godless.Godless
}

func (client localClient) Send(request api.Request) (api.Response, error) {
	respch, err := client.godless.Call(request)

	if err != nil {
		return api.RESPONSE_FAIL, err
	}

	resp := <-respch

	if resp.Err != nil {
		return resp, resp.Err
	}

	return resp, nil
}
