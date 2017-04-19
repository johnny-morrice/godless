package godless

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	ipfs "github.com/ipfs/go-ipfs-api"
)


type IPFSPath string

func (path IPFSPath) Path() string {
	return string(path)
}

type IPFSPeer struct {
	Url string
	Client *http.Client
	Shell *ipfs.Shell
}

type IPFSRecord struct {
	Namespace *Namespace
	Children []IPFSPath
}

func MakeIPFSPeer(url string) RemoteStore {
	peer := &IPFSPeer{
		Url: url,
		Client: defaultHttpClient(),
	}

	return peer
}

func (peer *IPFSPeer) Connect() error {
	peer.Shell = ipfs.NewShellWithClient(peer.Url, peer.Client)

	if !peer.Shell.IsUp() {
		return fmt.Errorf("IPFSPeer is not up at '%v'", peer.Url)
	}

	return nil
}

func (peer *IPFSPeer) Disconnect() error {
	// Nothing to do.
	return nil
}

func (peer *IPFSPeer) Add(record RemoteNamespaceRecord) (RemoteStoreIndex, error) {
	chunk := IPFSRecord{
		Namespace: record.Namespace,
		Children: make([]IPFSPath, len(record.Children)),
	}

	for i, index := range record.Children {
		path, ok := index.(IPFSPath)

		if !ok {
			panic("Child was not IPFSPath")
		}

		chunk.Children[i] = path
	}

	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(chunk)

	if err != nil {
		return nil, errors.Wrap(err, "Error encoding IPFSRecord Gob")
	}

	path, sherr := peer.Shell.Add(&buff)

	if sherr != nil {
		return nil, errors.Wrap(err, "Error adding to IPFS")
	}

	return RemoteStoreIndex(IPFSPath(path)), nil
}

func (peer *IPFSPeer) Cat(index RemoteStoreIndex) (RemoteNamespaceRecord, error) {
	path, ok := index.(IPFSPath)

	if !ok {
		panic("index was not IPFSPath")
	}

	dat, err := peer.Shell.Cat(string(path))

	if err != nil {
		return EMPTY_RECORD, errors.Wrap(err, "Could not Cat index from IPFS")
	}

	defer dat.Close()

	dec := gob.NewDecoder(dat)
	chunk := &IPFSRecord{}
	err = dec.Decode(chunk)

	// According to IPFS binding docs we must drain the reader.
	remainder, drainerr := ioutil.ReadAll(dat)

	if (drainerr != nil) {
		logwarn("error draining reader: %v", drainerr)
	}

	if len(remainder) != 0 {
		logwarn("remaining bits after gob: %v", remainder)
	}

	record := RemoteNamespaceRecord{
		Namespace: chunk.Namespace,
		Children: make([]RemoteStoreIndex, len(chunk.Children)),
	}

	for i, path := range chunk.Children {
		record.Children[i] = RemoteStoreIndex(path)
	}

	return record, nil
}
