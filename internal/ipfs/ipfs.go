package ipfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"
	"time"

	ipfs "github.com/ipfs/go-ipfs-api"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/http"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type IPFSRecord struct {
	Namespace crdt.Namespace
}

func makeIpfsRecord(namespace crdt.Namespace) *IPFSRecord {
	return &IPFSRecord{
		Namespace: namespace,
	}
}

func (record *IPFSRecord) encode(w io.Writer) error {
	invalid, err := crdt.EncodeNamespace(record.Namespace, w)

	record.logInvalid(invalid)

	return err
}

func (record *IPFSRecord) decode(r io.Reader) error {
	ns, invalid, err := crdt.DecodeNamespace(r)

	record.logInvalid(invalid)

	if err != nil {
		return err
	}

	record.Namespace = ns
	return nil
}

func (record *IPFSRecord) logInvalid(invalid []crdt.InvalidNamespaceEntry) {
	invalidCount := len(invalid)

	if invalidCount > 0 {
		log.Error("IPFSRecord: %v invalid entries", invalidCount)
	}
}

type encoder interface {
	encode(io.Writer) error
}

type decoder interface {
	decode(io.Reader) error
}

type IPFSIndex struct {
	Index crdt.Index
}

func makeIpfsIndex(index crdt.Index) *IPFSIndex {
	return &IPFSIndex{
		Index: index,
	}
}

func (index *IPFSIndex) encode(w io.Writer) error {
	invalid, err := crdt.EncodeIndex(index.Index, w)

	index.logInvalid(invalid)

	return err
}

func (index *IPFSIndex) decode(r io.Reader) error {
	dx, invalid, err := crdt.DecodeIndex(r)

	index.logInvalid(invalid)

	if err != nil {
		return err
	}

	// TODO should cache the invalid details.
	if len(invalid) > 0 {
		log.Warn("IPFSIndex Decoded invalid index entries")
	}

	index.Index = dx
	return nil
}

func (index *IPFSIndex) logInvalid(invalid []crdt.InvalidIndexEntry) {
	invalidCount := len(invalid)

	if invalidCount > 0 {
		log.Error("IPFSRecord: %v invalid entries", invalidCount)
	}
}

// TODO Don't use Shell directly - invent an interface.  This would enable mocking.
type IPFSPeer struct {
	Url         string
	Client      *gohttp.Client
	Shell       *ipfs.Shell
	PingTimeout time.Duration
	pinger      *ipfs.Shell
}

func (peer *IPFSPeer) Connect() error {
	if peer.PingTimeout == 0 {
		peer.PingTimeout = __DEFAULT_PING_TIMEOUT
	}

	if peer.Client == nil {
		log.Info("Using default HTTP client")
		peer.Client = http.DefaultBackendClient()
	}

	log.Info("Connecting to IPFS API...")
	pingClient := http.DefaultBackendClient()
	pingClient.Timeout = peer.PingTimeout
	peer.Shell = ipfs.NewShellWithClient(peer.Url, peer.Client)
	peer.pinger = ipfs.NewShellWithClient(peer.Url, pingClient)
	err := peer.validateConnection()

	if err == nil {
		log.Info("IPFS API Connection OK")
	}

	return err
}

func (peer *IPFSPeer) Disconnect() error {
	// Nothing to do.
	return nil
}

func (peer *IPFSPeer) validateShell() error {
	if peer.Shell == nil {
		return peer.Connect()
	}

	return peer.validateConnection()
}

func (peer *IPFSPeer) validateConnection() error {
	if !peer.pinger.IsUp() {
		return fmt.Errorf("IPFSPeer is not up at '%v'", peer.Url)
	}

	return nil
}

func (peer *IPFSPeer) PublishAddr(addr crdt.Link, topics []api.PubSubTopic) error {
	const failMsg = "IPFSPeer.PublishAddr failed"

	if verr := peer.validateShell(); verr != nil {
		return verr
	}

	publishValue, printErr := crdt.PrintLink(addr)

	if printErr != nil {
		return errors.Wrap(printErr, failMsg)
	}

	for _, t := range topics {
		topicText := string(t)
		log.Info("Publishing to topic: %v", t)
		pubsubErr := peer.Shell.PubSubPublish(topicText, string(publishValue))

		if pubsubErr != nil {
			log.Warn("Pubsub failed (topic %v): %v", t, pubsubErr.Error())
			continue
		}

		log.Info("Published to topic: %v", t)
	}

	return nil
}

func (peer *IPFSPeer) SubscribeAddrStream(topic api.PubSubTopic) (<-chan crdt.Link, <-chan error) {
	stream := make(chan crdt.Link)
	errch := make(chan error)

	tidy := func() {
		close(stream)
		close(errch)
	}

	if verr := peer.validateShell(); verr != nil {
		go func() {
			errch <- verr
			defer tidy()
		}()

		return stream, errch
	}

	go func() {
		defer tidy()

		topicText := string(topic)

		var subscription *ipfs.PubSubSubscription

	RESTART:
		for {
			var launchErr error
			log.Info("(Re)starting subscription on %v", topic)
			subscription, launchErr = peer.Shell.PubSubSubscribe(topicText)

			if launchErr != nil {
				log.Error("Subcription launch failed, retrying: %v", launchErr.Error())
				continue
			}

			for {
				log.Info("Fetching next subscription message on %v...", topic)
				record, recordErr := subscription.Next()

				if recordErr != nil {
					log.Error("Subscription read failed (topic %v), continuing: %v", topic, recordErr.Error())
					continue RESTART
				}

				pubsubPeer := record.From()
				bs := record.Data()
				addr, err := crdt.ParseLink(crdt.LinkText(bs))

				if err != nil {
					log.Warn("Bad link from peer (topic %v): %v", topic, pubsubPeer)
					continue
				}

				stream <- addr
				log.Info("Subscription update: '%v' from '%v'", addr, pubsubPeer)
			}
		}

	}()

	return stream, errch
}

func (peer *IPFSPeer) AddIndex(index crdt.Index) (crdt.IPFSPath, error) {
	const failMsg = "IPFSPeer.AddIndex failed"

	log.Info("Adding index to IPFS...")

	if verr := peer.validateShell(); verr != nil {
		return crdt.NIL_PATH, verr
	}

	chunk := makeIpfsIndex(index)

	path, addErr := peer.add(chunk)

	if addErr != nil {
		return crdt.NIL_PATH, errors.Wrap(addErr, failMsg)
	}

	log.Info("Added index")

	return path, nil
}

func (peer *IPFSPeer) CatIndex(addr crdt.IPFSPath) (crdt.Index, error) {
	log.Info("Catting index from IPFS at: %v ...", addr)

	if verr := peer.validateShell(); verr != nil {
		return crdt.EmptyIndex(), verr
	}

	chunk := &IPFSIndex{}
	caterr := peer.cat(addr, chunk)

	if caterr != nil {
		return crdt.EmptyIndex(), errors.Wrap(caterr, "IPFSPeer.CatNamespace failed")
	}

	log.Info("Catted index")

	return chunk.Index, nil
}

func (peer *IPFSPeer) AddNamespace(namespace crdt.Namespace) (crdt.IPFSPath, error) {
	log.Info("Adding Namespace to IPFS...")

	if verr := peer.validateShell(); verr != nil {
		return crdt.NIL_PATH, verr
	}

	chunk := makeIpfsRecord(namespace)

	path, err := peer.add(chunk)

	if err != nil {
		return crdt.NIL_PATH, errors.Wrap(err, "IPFSPeer.AddNamespace failed")
	}

	log.Info("Added namespace")

	return path, nil
}

func (peer *IPFSPeer) CatNamespace(addr crdt.IPFSPath) (crdt.Namespace, error) {
	log.Info("Catting namespace from IPFS at: %v ...", addr)

	if verr := peer.validateShell(); verr != nil {
		return crdt.EmptyNamespace(), verr
	}

	chunk := &IPFSRecord{}
	caterr := peer.cat(addr, chunk)

	if caterr != nil {
		return crdt.EmptyNamespace(), errors.Wrap(caterr, "IPFSPeer.CatNamespace failed")
	}

	log.Info("Catted namespace")

	return chunk.Namespace, nil
}

func (peer *IPFSPeer) add(chunk encoder) (crdt.IPFSPath, error) {
	const failMsg = "IPFSPeer.add failed"
	buff := &bytes.Buffer{}
	err := chunk.encode(buff)

	if err != nil {
		return crdt.NIL_PATH, errors.Wrap(err, failMsg)
	}

	path, sherr := peer.Shell.Add(buff)

	if sherr != nil {
		return crdt.NIL_PATH, errors.Wrap(err, failMsg)
	}

	return crdt.IPFSPath(path), nil
}

func (peer *IPFSPeer) cat(path crdt.IPFSPath, out decoder) error {
	const failMsg = "IPFSPeer.cat failed"
	reader, err := peer.Shell.Cat(string(path))

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	defer reader.Close()

	err = out.decode(reader)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	// According to IPFS binding docs we must drain the reader.
	remainder, drainerr := ioutil.ReadAll(reader)

	if drainerr != nil {
		log.Warn("error draining reader: %v", drainerr)
	}

	if len(remainder) != 0 {
		log.Warn("remaining bits after gob: %v", remainder)
	}

	return nil
}

const __DEFAULT_PING_TIMEOUT = time.Second * 5
