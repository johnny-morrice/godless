// Godless is a peer-to-peer database running over IPFS.
//
// Godless uses a Consistent Replicated Data Type called a Namespace to share schemaless data with peers.
//
// This package is a facade to Godless internals.
//
// Godless is in alpha, and should be considered experimental software.
package godless

import (
	"sync"
	"time"

	gohttp "net/http"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/http"
	"github.com/johnny-morrice/godless/internal/ipfs"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
)

// Godless options.
type Options struct {
	// IpfsServiceUrl is required.
	IpfsServiceUrl string
	// WebServiceAddr is optional.  If not set, the webservice will be disabled.
	WebServiceAddr string
	// IndexHash is optional.  Set to load an existing index from IPFS.
	IndexHash string
	// FailEarly will cause the godless process to crash if it cannot contact IPFS on startup.
	FailEarly bool
	// ReplicateInterval is optional.  The duration between peer-to-peer replications.
	ReplicateInterval time.Duration
	// Topics is optional.  Two godless servers which share a topic will replicate indices. An empty topics slice will disable replication.
	Topics []string
	// IpfsClient is optional.  Specify a HTTP client for IPFS.
	IpfsClient *gohttp.Client
	// IpfsPingTimeout is optional.  Specify a lower timeout for "Am I Connected?" checks.
	IpfsPingTimeout time.Duration
}

// Godless is a peer-to-peer database.  It shares structured data between peers, using IPFS as a backing store.
// The core datastructure is a CRDT namespace which resembles a relational scheme in that it has tables, rows, and entries.
type Godless struct {
	Options
	errch    chan error
	errwg    sync.WaitGroup
	stopch   chan interface{}
	stoppers []chan<- interface{}
	store    api.RemoteStore
	remote   api.RemoteNamespace
	api      api.APIService
}

// New creates a godless instance, connecting to any services, and providing any services, specified in the options.
func New(options Options) (*Godless, error) {
	godless := &Godless{Options: options}
	setupFuncs := []func() error{
		godless.connectIpfs,
		godless.setupNamespace,
		godless.launchAPI,
		godless.serveWeb,
		godless.replicate,
	}

	err := breakOnError(setupFuncs)

	if err != nil {
		return nil, err
	}

	return godless, nil
}

// Errors provides a stream of errors from godless.  Godless will attempt to stay alive under difficult conditions.
// The returned channel is closed when godless shuts down.
func (godless *Godless) Errors() <-chan error {
	return godless.errch
}

// Shutdown stops all godless processes.  It does not wait for those goroutines to stop.
func (godless *Godless) Shutdown() {
	godless.stopch <- godless.stopch
}

func (godless *Godless) connectIpfs() error {
	client := godless.IpfsClient
	pingTimeout := godless.IpfsPingTimeout

	peer := &ipfs.IPFSPeer{
		Url:         godless.IpfsServiceUrl,
		Client:      client,
		PingTimeout: pingTimeout,
	}

	// if godless.FailEarly {
	// 	err := peer.Connect()
	//
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	godless.store = peer

	return nil
}

func (godless *Godless) setupNamespace() error {
	remote, err := godless.makeRemote()

	if err != nil {
		return err
	}

	godless.remote = remote
	return nil
}

// MakeRemoteNamespace creates a data store representing p2p data.
func (godless *Godless) makeRemote() (api.RemoteNamespace, error) {
	if godless.IndexHash == "" {
		if godless.FailEarly {
			namespace := crdt.EmptyNamespace()
			return service.PersistNewRemoteNamespace(godless.store, namespace)
		} else {
			return service.MakeRemoteNamespace(godless.store), nil
		}

	} else {
		path := crdt.IPFSPath(godless.IndexHash)
		return service.LoadRemoteNamespace(godless.store, path)
	}
}

func (godless *Godless) launchAPI() error {
	api, errch := service.LaunchKeyValueStore(godless.remote)

	godless.addErrors(errch)
	godless.api = api

	return nil
}

// Serve serves the Godless webservice.
func (godless *Godless) serveWeb() error {
	addr := godless.WebServiceAddr

	if addr == "" {
		return nil
	}

	webService := &service.WebService{API: godless.api}
	stopch, err := http.Serve(addr, webService.Handler())

	if err != nil {
		return err
	}

	godless.addStopper(stopch)
	return nil
}

// Replicate shares data via the IPFS pubsub mechanism.
func (godless *Godless) replicate() error {
	topics := godless.Topics
	interval := godless.ReplicateInterval

	if len(topics) == 0 {
		return nil
	}

	ipfsTopics := make([]crdt.IPFSPath, len(topics))

	for i, t := range topics {
		ipfsTopics[i] = crdt.IPFSPath(t)
	}

	stopch, errch := service.Replicate(godless.api, godless.store, interval, ipfsTopics)
	godless.addStopper(stopch)
	godless.addErrors(errch)
	return nil
}

func (godless *Godless) addStopper(stopch chan<- interface{}) {
	if godless.stopch == nil {
		godless.stopch = make(chan interface{})
		go func() {
			godless.handleShutdown()
		}()
	}

	godless.stoppers = append(godless.stoppers, stopch)
}

func (godless *Godless) handleShutdown() {
	<-godless.stopch
	log.Info("Shutting down")
	for _, stopper := range godless.stoppers {
		go close(stopper)
	}

}

func (godless *Godless) addErrors(errch <-chan error) {
	godless.errwg.Add(1)

	if godless.errch == nil {
		godless.errch = make(chan error)
		go func() {
			godless.errwg.Wait()
			close(godless.errch)
		}()
	}

	go func() {
		for err := range errch {
			godless.errch <- err
		}
		godless.errwg.Done()
	}()
}

// Client is a Godless HTTP client.
type Client interface {
	SendQuery(*query.Query) (api.APIResponse, error)
	SendReflection(api.APIReflectionType) (api.APIResponse, error)
}

// MakeClient creates a Godless HTTP Client.
func MakeClient(serviceAddr string) Client {
	return service.MakeClient(serviceAddr)
}

func MakeClientWithHttp(serviceAddr string, webClient gohttp.Client) Client {
	return nil
}

func breakOnError(pipeline []func() error) error {
	for _, f := range pipeline {
		err := f()

		if err != nil {
			return err
		}
	}

	return nil
}
