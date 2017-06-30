// Godless is a peer-to-peer database running over IPFS.
//
// Godless uses a Consistent Replicated Data Type called a Namespace to share schemaless data with peers.
//
// This package is a facade to Godless internals.
//
// Godless is in alpha, and should be considered experimental software.
package godless

import (
	"fmt"
	"sync"
	"time"

	gohttp "net/http"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/internal/http"
	"github.com/johnny-morrice/godless/internal/ipfs"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

// Godless options.
type Options struct {
	// IpfsServiceUrl is required, unless specifying your own DataPeer or RemoteStore.
	IpfsServiceUrl string
	// DataPeer is optional.
	DataPeer api.DataPeer
	// RemoteStore is optional.
	RemoteStore api.RemoteStore
	// KeyStore is required. A private Key store.
	KeyStore api.KeyStore
	// MemoryImage is required.
	MemoryImage api.MemoryImage
	// WebServiceAddr is optional.  If not set, the webservice will be disabled.
	WebServiceAddr string
	// IndexHash is optional.  Set to load an existing index from IPFS.
	IndexHash string
	// FailEarly will cause the godless process to crash if it cannot contact IPFS on startup.
	FailEarly bool
	// ReplicateInterval is optional.  The duration between peer-to-peer replications.
	ReplicateInterval time.Duration
	Pulse             time.Duration
	// Topics is optional.  Two godless servers which share a topic will replicate indices. An empty topics slice will disable replication.
	Topics []string
	// IpfsClient is optional.  Specify a HTTP client for IPFS.
	IpfsClient *gohttp.Client
	// IpfsPingTimeout is optional.  Specify a lower timeout for "Am I Connected?" checks.
	IpfsPingTimeout time.Duration
	// Cache is optional. Build a 12-factor app by supplying your own remote cache.
	// HeadCache, IndexCache and NamespaceCache can be used to specify different caches for different data types.
	Cache api.Cache
	// HeadCache is optional.  Build a 12-factor app by supplying your own remote cache.
	HeadCache api.HeadCache
	// IndexCache is optional.  Build a 12-factor app by supplying your own remote cache.
	IndexCache api.IndexCache
	// NamespaceCache is optional. Build a 12-factor app by supplying your own remote cache.
	NamespaceCache api.NamespaceCache
	// PriorityQueue is optional. Build a 12-factor app by supplying your own remote cache.
	PriorityQueue api.RequestPriorityQueue
	// APIQueryLimit is optional.  Tune performance by setting the number of simultaneous queries.
	APIQueryLimit int
	// PublicServer is optional.  If false, the index will only be updated from peers who are in your public key list.
	PublicServer bool
}

// Godless is a peer-to-peer database.  It shares structured data between peers, using IPFS as a backing store.
// The core datastructure is a CRDT namespace which resembles a relational scheme in that it has tables, rows, and entries.
type Godless struct {
	Options
	errch    chan error
	errwg    sync.WaitGroup
	stopch   chan struct{}
	stoppers []chan<- struct{}
	remote   api.RemoteNamespace
	api      api.APIService
}

// New creates a godless instance, connecting to any services, and providing any services, specified in the options.
func New(options Options) (*Godless, error) {
	godless := &Godless{Options: options}

	missing := godless.findMissingParameters()

	if missing != nil {
		return nil, missing
	}

	setupFuncs := []func() error{
		godless.connectDataPeer,
		godless.connectRemoteStore,
		godless.connectCache,
		godless.setupNamespace,
		godless.launchAPI,
		godless.serveWeb,
		godless.replicate,
	}

	err := breakOnError(setupFuncs)

	if err != nil {
		return nil, err
	}

	godless.report()

	return godless, nil
}

func (godless *Godless) report() {
	if godless.PublicServer {
		log.Info("Running public Godless API")
	} else {
		log.Info("Running private Godless API")
	}

	privCount := len(godless.KeyStore.GetAllPrivateKeys())
	pubCount := len(godless.KeyStore.GetAllPublicKeys())

	log.Info("Godless API using %v private and %v public keys", privCount, pubCount)
}

func (godless *Godless) findMissingParameters() error {
	var missing error

	if godless.KeyStore == nil {
		msg := godless.missingParameterText("KeyStore")
		missing = addErrorMessage(missing, msg)
	}

	if godless.MemoryImage == nil {
		msg := godless.missingParameterText("MemoryImage")
		missing = addErrorMessage(missing, msg)
	}

	return missing
}

func addErrorMessage(err error, msg string) error {
	if err == nil {
		return errors.New(msg)
	} else {
		return errors.Wrap(err, msg)
	}
}

func (godless *Godless) missingParameterText(param string) string {
	return fmt.Sprintf("Missing required parameter '%v'", param)
}

// Errors provides a stream of errors from godless.  Godless will attempt to handle any errors it can.  Any errors received here indicate that bad things have happened.
func (godless *Godless) Errors() <-chan error {
	return godless.errch
}

// Shutdown stops all godless processes.  It does not wait for those goroutines to stop.
func (godless *Godless) Shutdown() {
	godless.stopch <- struct{}{}
}

func (godless *Godless) connectDataPeer() error {
	if godless.RemoteStore != nil {
		return nil
	}

	if godless.DataPeer == nil {
		if godless.IpfsServiceUrl == "" {
			msg := godless.missingParameterText("IpfsServiceUrl")
			return errors.New(msg)
		}

		peer := &ipfs.WebServiceClient{
			Url:         godless.IpfsServiceUrl,
			PingTimeout: godless.IpfsPingTimeout,
			Http:        godless.IpfsClient,
		}

		godless.DataPeer = peer
	}

	return nil
}

func (godless *Godless) connectRemoteStore() error {
	if godless.RemoteStore == nil {
		ipfs := &ipfs.IpfsRemoteStore{
			Shell: godless.DataPeer,
		}

		if godless.FailEarly {
			err := ipfs.Connect()

			if err != nil {
				return err
			}
		}

		godless.RemoteStore = ipfs
	}

	return nil
}

func (godless *Godless) makeCacheStopper(cache api.Cache) chan<- struct{} {
	stopper := make(chan struct{}, 1)
	go func() {
		<-stopper
		err := cache.CloseCache()
		if err == nil {
			log.Info("Closed cache")
		} else {
			log.Error("Error closing cache: %v", err.Error())
		}
	}()

	return stopper
}

func (godless *Godless) connectCache() error {
	if godless.Cache != nil {
		godless.HeadCache = godless.Cache
		godless.IndexCache = godless.Cache
		godless.NamespaceCache = godless.Cache
		godless.addStopper(godless.makeCacheStopper(godless.Cache))
		return nil
	}

	if godless.HeadCache == nil {
		godless.HeadCache = cache.MakeResidentHeadCache()
	}

	if godless.IndexCache == nil {
		godless.IndexCache = cache.MakeResidentIndexCache(__UNKNOWN_BUFFER_SIZE)
	}

	if godless.NamespaceCache == nil {
		godless.NamespaceCache = cache.MakeResidentNamespaceCache(__UNKNOWN_BUFFER_SIZE)
	}

	return nil
}

func (godless *Godless) setupNamespace() error {
	if godless.IndexHash != "" {
		head := crdt.IPFSPath(godless.IndexHash)

		err := godless.HeadCache.SetHead(head)

		if err != nil {
			return err
		}
	}

	namespaceOptions := service.RemoteNamespaceOptions{
		Pulse:          godless.Pulse,
		Store:          godless.RemoteStore,
		HeadCache:      godless.HeadCache,
		IndexCache:     godless.IndexCache,
		NamespaceCache: godless.NamespaceCache,
		KeyStore:       godless.KeyStore,
		IsPublicIndex:  godless.PublicServer,
		MemoryImage:    godless.MemoryImage,
	}

	godless.remote = service.MakeRemoteNamespace(namespaceOptions)
	return nil
}

func (godless *Godless) launchAPI() error {
	limit := godless.APIQueryLimit

	if limit == 0 {
		limit = 1
	}

	queue := godless.PriorityQueue

	if queue == nil {
		queue = cache.MakeResidentBufferQueue(__UNKNOWN_BUFFER_SIZE)
	}

	api, errch := service.LaunchKeyValueStore(godless.remote, queue, limit)

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

	pubsubTopics := make([]api.PubSubTopic, len(topics))

	for i, t := range topics {
		pubsubTopics[i] = api.PubSubTopic(t)
	}

	options := service.ReplicateOptions{
		API:         godless.api,
		RemoteStore: godless.RemoteStore,
		Interval:    interval,
		Topics:      pubsubTopics,
		KeyStore:    godless.KeyStore,
	}
	stopch, errch := service.Replicate(options)
	godless.addStopper(stopch)
	godless.addErrors(errch)
	return nil
}

func (godless *Godless) addStopper(stopch chan<- struct{}) {
	if godless.stopch == nil {
		godless.stopch = make(chan struct{})
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

func MakeClientWithHttp(serviceAddr string, webClient *gohttp.Client) Client {
	return service.MakeClientWithHttp(serviceAddr, webClient)
}

func MakeKeyStore() api.KeyStore {
	return &crypto.KeyStore{}
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

// We don't know the right buffer size here, so let the cache package handle it.
const __UNKNOWN_BUFFER_SIZE = -1
