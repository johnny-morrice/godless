package datapeer

import (
	"io"
	gohttp "net/http"
	"time"

	ipfs "github.com/ipfs/go-ipfs-api"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/log"
)

type IpfsWebServiceOptions struct {
	Url         string
	Http        *gohttp.Client
	PingTimeout time.Duration
	Shell       *ipfs.Shell
}

type ipfsWebService struct {
	IpfsWebServiceOptions
	pinger *ipfs.Shell
}

func MakeIpfsWebService(options IpfsWebServiceOptions) api.DataPeer {
	return &ipfsWebService{IpfsWebServiceOptions: options}
}

func (client *ipfsWebService) Connect() error {
	if client.PingTimeout == 0 {
		client.PingTimeout = __DEFAULT_PING_TIMEOUT
	}

	if client.Http == nil {
		log.Info("Using default HTTP client")
		client.Http = defaultBackendClient()
	}

	log.Info("Connecting to IPFS API...")
	pingClient := defaultPingClient()
	client.Shell = ipfs.NewShellWithClient(client.Url, client.Http)
	client.pinger = ipfs.NewShellWithClient(client.Url, pingClient)

	return nil
}

func (client *ipfsWebService) IsUp() bool {
	return client.pinger.IsUp()
}

func (client *ipfsWebService) Disconnect() error {
	return nil
}

func (client ipfsWebService) Cat(path string) (io.ReadCloser, error) {
	return client.Shell.Cat(path)
}

func (client ipfsWebService) Add(r io.Reader) (string, error) {
	return client.Shell.Add(r)
}

func (client ipfsWebService) PubSubPublish(topic, data string) error {
	return client.Shell.PubSubPublish(topic, data)
}

type subscription struct {
	sub *ipfs.PubSubSubscription
}

func (sub subscription) Next() (api.PubSubRecord, error) {
	rec, err := sub.sub.Next()

	if err != nil {
		return nil, err
	}

	return record{rec: rec}, nil
}

type record struct {
	rec ipfs.PubSubRecord
}

func (rec record) From() string {
	return string(rec.rec.From())
}

func (rec record) Data() []byte {
	return rec.rec.Data()
}

func (rec record) SeqNo() int64 {
	return rec.rec.SeqNo()
}

func (rec record) TopicIDs() []string {
	return rec.rec.TopicIDs()
}

func (client ipfsWebService) PubSubSubscribe(topic string) (api.PubSubSubscription, error) {
	sub, err := client.Shell.PubSubSubscribe(topic)

	if err != nil {
		return nil, err
	}

	return subscription{sub: sub}, nil
}

var __backendClient *gohttp.Client
var __pingClient *gohttp.Client

func defaultBackendClient() *gohttp.Client {
	if __backendClient == nil {
		__backendClient = http.MakeBackendHttpClient(__BACKEND_TIMEOUT)
	}

	return __backendClient
}

func defaultPingClient() *gohttp.Client {
	if __pingClient == nil {
		__pingClient = http.MakeBackendHttpClient(__PING_TIMEOUT)
	}

	return __pingClient
}

const __PING_TIMEOUT = 10 * time.Second
const __BACKEND_TIMEOUT = 10 * time.Minute

const __DEFAULT_PING_TIMEOUT = time.Second * 5
