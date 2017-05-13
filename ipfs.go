package godless

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net/http"

	ipfs "github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
)

type IPFSPath string

func (path IPFSPath) Path() string {
	return string(path)
}

func castIPFSPath(addr RemoteStoreAddress) IPFSPath {
	path, ok := addr.(IPFSPath)

	if !ok {
		panic("addr was not IPFSPath")
	}

	return path
}

type IPFSPeer struct {
	Url    string
	Client *http.Client
	Shell  *ipfs.Shell
}

type IPFSRecord struct {
	Namespace Namespace
}

type IPFSIndex struct {
	Index map[TableName][]IPFSPath
}

func (index IPFSIndex) remoteIndex() RemoteNamespaceIndex {
	out := EmptyRemoteNamespaceIndex()

	for indexKey, paths := range index.Index {
		addrs := make([]RemoteStoreAddress, len(paths))
		for i, p := range paths {
			addrs[i] = p
		}
		out.Index[indexKey] = addrs
	}

	return out
}

func makeIpfsIndex(index RemoteNamespaceIndex) IPFSIndex {
	out := IPFSIndex{
		Index: map[TableName][]IPFSPath{},
	}

	for indexKey, addrs := range index.Index {
		paths := make([]IPFSPath, len(addrs))
		for i, a := range addrs {
			paths[i] = castIPFSPath(a)
		}
		out.Index[indexKey] = paths
	}

	return out
}

func MakeIPFSPeer(url string) RemoteStore {
	peer := &IPFSPeer{
		Url:    url,
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

func (peer *IPFSPeer) AddIndex(index RemoteNamespaceIndex) (RemoteStoreAddress, error) {
	chunk := makeIpfsIndex(index)

	path, err := peer.add(&chunk)

	if err != nil {
		return nil, errors.Wrap(err, "IPFSPeer.AddIndex failed")
	}

	return path, nil
}

func (peer *IPFSPeer) CatIndex(addr RemoteStoreAddress) (RemoteNamespaceIndex, error) {
	path := castIPFSPath(addr)

	chunk := IPFSIndex{}
	caterr := peer.cat(path, &chunk)

	if caterr != nil {
		return EMPTY_INDEX, errors.Wrap(caterr, "IPFSPeer.CatNamespace failed")
	}

	index := chunk.remoteIndex()
	return index, nil
}

func (peer *IPFSPeer) AddNamespace(record RemoteNamespaceRecord) (RemoteStoreAddress, error) {
	chunk := IPFSRecord{
		Namespace: record.Namespace,
	}

	path, err := peer.add(&chunk)

	if err != nil {
		return nil, errors.Wrap(err, "IPFSPeer.AddNamespace failed")
	}

	return path, nil
}

func (peer *IPFSPeer) CatNamespace(addr RemoteStoreAddress) (RemoteNamespaceRecord, error) {
	path := castIPFSPath(addr)

	chunk := IPFSRecord{}
	caterr := peer.cat(path, &chunk)

	if caterr != nil {
		return EMPTY_RECORD, errors.Wrap(caterr, "IPFSPeer.CatNamespace failed")
	}

	record := RemoteNamespaceRecord{Namespace: chunk.Namespace}
	return record, nil
}

func (peer *IPFSPeer) add(chunk interface{}) (IPFSPath, error) {
	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(chunk)

	if err != nil {
		return "", errors.Wrap(err, "Error encoding Gob")
	}

	path, sherr := peer.Shell.Add(&buff)

	if sherr != nil {
		return "", errors.Wrap(err, "IPFSPeer.add failed")
	}

	return IPFSPath(path), nil
}

func (peer *IPFSPeer) cat(path IPFSPath, out interface{}) error {
	reader, err := peer.Shell.Cat(string(path))

	if err != nil {
		return errors.Wrap(err, "IPFSPeer.cat failed")
	}

	defer reader.Close()

	dec := gob.NewDecoder(reader)
	err = dec.Decode(out)

	if err != nil {
		return errors.Wrap(err, "failed to decode gob")
	}

	// According to IPFS binding docs we must drain the reader.
	remainder, drainerr := ioutil.ReadAll(reader)

	if drainerr != nil {
		logwarn("error draining reader: %v", drainerr)
	}

	if len(remainder) != 0 {
		logwarn("remaining bits after gob: %v", remainder)
	}

	return nil
}
