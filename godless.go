// Godless is a peer-to-peer database running over IPFS.
//
// Godless uses a Consistent Replicated Data Type called a Namespace to share schemaless data with peers.
//
// This package is a facade to Godless internals.
//
// Godless is in alpha, and should be considered experimental software.
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

// WebService is the Godless webservice.
type WebService interface {
	Handler() gohttp.Handler
}

// Client is a Godless HTTP client.
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

// MakeRemoteNamespace creates a data store representing p2p data.
func MakeRemoteNamespace(store api.RemoteStore, hash crdt.IPFSPath, earlyConnect bool) (api.RemoteNamespace, error) {
	if hash == "" {
		if earlyConnect {
			namespace := crdt.EmptyNamespace()
			return service.PersistNewRemoteNamespace(store, namespace)
		} else {
			return service.MakeRemoteNamespace(store), nil
		}

	} else {
		return service.LoadRemoteNamespace(store, hash)
	}
}

// Run launches the Godless API service.  The API is safe for concurrent access.
func Run(remoteNamespace api.RemoteNamespace) (api.APIService, <-chan error) {
	return service.LaunchKeyValueStore(remoteNamespace)
}

// MakeWebService creates the Godless webservice.
func MakeWebService(api api.APIService) WebService {
	return &service.WebService{API: api}
}

// Serve serves the Godless webservice.  Send an item down the channel to stop the HTTP server.
func Serve(addr string, webService WebService) (chan<- interface{}, error) {
	return http.Serve(addr, webService.Handler())
}

// Replicate shares data via the IPFS pubsub mechanism.  Send an item down the channel to stop replication.
func Replicate(api api.APIService, store api.RemoteStore, interval time.Duration, topics []crdt.IPFSPath) (chan<- interface{}, <-chan error) {
	return service.Replicate(api, store, interval, topics)
}

// MakeClient creates a Godless HTTP Client.
func MakeClient(serviceAddr string) Client {
	return service.MakeClient(serviceAddr)
}
