package godless

import (
	"time"

	gohttp "net/http"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/http"
	"github.com/johnny-morrice/godless/internal/ipfs"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/query"
)

type WebService interface {
	Handler() gohttp.Handler
}
type Client interface {
	SendQuery(*query.Query) (api.APIResponse, error)
	SendReflection(api.APIReflectionType) (api.APIResponse, error)
}

// MakeIPFSPeer provides an interface to IPFS.  The url parameter should take the address of an IPFS webservice API.
func MakeIPFSPeer(url string) api.RemoteStore {
	peer := &ipfs.IPFSPeer{
		Url:    url,
		Client: http.DefaultBackendClient(),
	}

	return peer
}

func MakeRemoteNamespace(store api.RemoteStore, hash crdt.IPFSPath, earlyConnect bool) (api.RemoteNamespace, error) {
	if hash == "" {
		namespace := crdt.EmptyNamespace()
		if earlyConnect {
			return service.PersistNewRemoteNamespace(store, namespace)
		} else {
			return service.MakeRemoteNamespace(store, namespace), nil
		}

	} else {
		return service.LoadRemoteNamespace(store, hash)
	}
}

func Run(remoteNamespace api.RemoteNamespace) (api.APIService, <-chan error) {
	return service.LaunchKeyValueStore(remoteNamespace)
}

func MakeWebService(api api.APIService) WebService {
	return &service.WebService{API: api}
}

func Serve(addr string, webService WebService) (chan<- interface{}, error) {
	return http.Serve(addr, webService.Handler())
}

func Replicate(api api.APIService, store api.RemoteStore, interval time.Duration, topics []crdt.RemoteStoreAddress) (chan<- interface{}, <-chan error) {
	return service.Replicate(api, store, interval, topics)
}

func MakeClient(serviceAddr string) Client {
	return service.MakeClient(serviceAddr)
}
