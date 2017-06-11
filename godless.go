package godless

import (
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/internal/http"
	"github.com/johnny-morrice/godless/internal/ipfs"
)

func MakeIPFSPeer(url string) api.RemoteStore {
	peer := &ipfs.IPFSPeer{
		Url:    url,
		Client: http.DefaultBackendClient(),
	}

	return peer
}
