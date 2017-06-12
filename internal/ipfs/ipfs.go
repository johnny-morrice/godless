package ipfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"

	ipfs "github.com/ipfs/go-ipfs-api"
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
	return crdt.EncodeNamespace(record.Namespace, w)
}

func (record *IPFSRecord) decode(r io.Reader) error {
	ns, err := crdt.DecodeNamespace(r)

	if err != nil {
		return err
	}

	record.Namespace = ns
	return nil
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
	return crdt.EncodeIndex(index.Index, w)
}

func (index *IPFSIndex) decode(r io.Reader) error {
	dx, err := crdt.DecodeIndex(r)

	if err != nil {
		return err
	}

	index.Index = dx
	return nil
}

// TODO Don't use Shell directly - invent an interface.  This would enable mocking.
type IPFSPeer struct {
	Url    string
	Client *gohttp.Client
	Shell  *ipfs.Shell
	pinger *ipfs.Shell
}

func (peer *IPFSPeer) Connect() error {
	log.Info("Connecting to IPFS API...")
	peer.Shell = ipfs.NewShellWithClient(peer.Url, peer.Client)
	peer.pinger = ipfs.NewShellWithClient(peer.Url, http.BackendPingClient())
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
	if peer.pinger.IsUp() {
		log.Debug("IPFS API Connection check OK")
	} else {
		return fmt.Errorf("IPFSPeer is not up at '%v'", peer.Url)
	}

	return nil
}

func (peer *IPFSPeer) PublishAddr(addr crdt.IPFSPath, topics []crdt.IPFSPath) error {
	const failMsg = "IPFSPeer.PublishAddr failed"

	if verr := peer.validateShell(); verr != nil {
		return verr
	}

	publishValue := string(addr)

	for _, t := range topics {
		topicText := string(t)
		err := peer.Shell.PubSubPublish(topicText, publishValue)

		if err != nil {
			return errors.Wrap(err, failMsg)
		}
	}

	return nil
}

func (peer *IPFSPeer) SubscribeAddrStream(topic crdt.IPFSPath) (<-chan crdt.IPFSPath, <-chan error) {
	stream := make(chan crdt.IPFSPath)
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
		subscription, launchErr := peer.Shell.PubSubSubscribe(topicText)

		if launchErr != nil {
			errch <- launchErr
			return
		}

		for {
			record, recordErr := subscription.Next()

			if recordErr != nil {
				errch <- recordErr
				continue
			}

			pubsubPeer := record.From()
			bs := record.Data()
			addr := crdt.IPFSPath(string(bs))

			stream <- addr
			log.Info("Subscription update: '%v' from '%v'", addr, pubsubPeer)
		}

	}()

	return stream, errch
}

func (peer *IPFSPeer) AddIndex(index crdt.Index) (crdt.IPFSPath, error) {
	const failMsg = "IPFSPeer.AddIndex failed"

	if verr := peer.validateShell(); verr != nil {
		return crdt.NIL_PATH, verr
	}

	chunk := makeIpfsIndex(index)

	path, addErr := peer.add(chunk)

	if addErr != nil {
		return crdt.NIL_PATH, errors.Wrap(addErr, failMsg)
	}

	return path, nil
}

func (peer *IPFSPeer) CatIndex(addr crdt.IPFSPath) (crdt.Index, error) {
	if verr := peer.validateShell(); verr != nil {
		return crdt.EmptyIndex(), verr
	}

	chunk := &IPFSIndex{}
	caterr := peer.cat(addr, chunk)

	if caterr != nil {
		return crdt.EmptyIndex(), errors.Wrap(caterr, "IPFSPeer.CatNamespace failed")
	}

	return chunk.Index, nil
}

func (peer *IPFSPeer) AddNamespace(namespace crdt.Namespace) (crdt.IPFSPath, error) {
	if verr := peer.validateShell(); verr != nil {
		return crdt.NIL_PATH, verr
	}

	chunk := makeIpfsRecord(namespace)

	path, err := peer.add(chunk)

	if err != nil {
		return crdt.NIL_PATH, errors.Wrap(err, "IPFSPeer.AddNamespace failed")
	}

	return path, nil
}

func (peer *IPFSPeer) CatNamespace(addr crdt.IPFSPath) (crdt.Namespace, error) {
	if verr := peer.validateShell(); verr != nil {
		return crdt.EmptyNamespace(), verr
	}

	chunk := &IPFSRecord{}
	caterr := peer.cat(addr, chunk)

	if caterr != nil {
		return crdt.EmptyNamespace(), errors.Wrap(caterr, "IPFSPeer.CatNamespace failed")
	}

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
