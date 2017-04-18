package godless

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"github.com/pkg/errors"
)

type ipfsNamespace struct {
	loaded bool
	Update *Namespace

	FilePeer *IpfsPeer
	FilePath IpfsPath
	// Breaking 12 factor rule in caching namespace...
	Namespace *Namespace
	Children []*ipfsNamespace
}

type ipfsNamespaceRecord struct {
	Namespace *Namespace
	Children []IpfsPath
}

func LoadIPFSNamespace(filePeer *IpfsPeer, filePath IpfsPath) (KvNamespace, error) {
	ns := &ipfsNamespace{}
	ns.FilePeer = filePeer
	ns.FilePath = filePath
	ns.Update = MakeNamespace()
	ns.Namespace = MakeNamespace()
	ns.Children = []*ipfsNamespace{}

	err := ns.load()

	if err != nil {
		return nil, errors.Wrap(err, "Error loading new namespace")
	}

	return ns, nil
}

func PersistNewIPFSNamespace(filePeer *IpfsPeer, namespace *Namespace) (KvNamespace, error) {
	ns := &ipfsNamespace{}
	ns.FilePeer = filePeer
	ns.Update = namespace
	ns.Namespace = MakeNamespace()
	ns.Children = []*ipfsNamespace{}

	return ns.Persist()
}

func (ns *ipfsNamespace) RunKvQuery(kvq KvQuery) {
	query := kvq.Query
	var runner QueryRun

	switch query.OpCode {
	case JOIN:
		visitor := MakeQueryJoinVisitor(ns)
		query.Visit(visitor)
		runner = visitor
	case SELECT:
		visitor := MakeQuerySelectVisitor(ns)
		query.Visit(visitor)
		runner = visitor
	default:
		query.opcodePanic()
	}

	runner.RunQuery(kvq)
}

func (ns *ipfsNamespace) IsChanged() bool {
	return !ns.Namespace.IsEmpty()
}

func (ns *ipfsNamespace) JoinTable(tableKey string, table Table) error {
	joined, joinerr := ns.Update.JoinTable(tableKey, table)

	if joinerr != nil {
		return errors.Wrap(joinerr, "ipfsNamespace.JoinTable failed")
	}

	ns.Update = joined

	return nil
}

func (ns *ipfsNamespace) NamespaceLeaf() *Namespace {
	return ns.Namespace
}

func (ns *ipfsNamespace) LoadTraverse(f RemoteNamespaceReader) error {
	stack := make([]*ipfsNamespace, 1)
	stack[0] = ns

	for i := 0 ; i < len(stack); i++ {
		current := stack[i]
		err := current.load()

		if err != nil {
			return errors.Wrap(err, "Error in ipfsNamespace loadTraverse")
		}

		leaf := current.NamespaceLeaf()
		abort, visiterr := f.ReadNamespace(leaf)

		if visiterr != nil {
			return errors.Wrap(err, "Error in ipfsNamespace traversal")
		}

		if abort {
			return nil
		}

		stack = append(stack, current.Children...)
	}

	return nil
}

// Load chunks over IPFS
// TODO opportunity to query IPFS in parallel?
func (ns *ipfsNamespace) load() error {
	if ns.FilePath == "" {
		logdie("tried to load ipfsNamespace with empty FilePath")
	}

	if ns.loaded {
		logwarn("ipfsNamespace already loaded from: '%v'", ns.FilePath)
		return nil
	}

	reader, err := ns.FilePeer.Shell.Cat(string(ns.FilePath))

	if err != nil {
		return errors.Wrap(err, "Error in ipfsNamespace Cat")
	}

	defer reader.Close()

	dec := gob.NewDecoder(reader)
	part := &ipfsNamespaceRecord{}
	err = dec.Decode(part)

	// According to IPFS binding docs we must drain the reader.
	remainder, drainerr := ioutil.ReadAll(reader)

	if (drainerr != nil) {
		logwarn("error draining reader: %v", drainerr)
	}

	if len(remainder) != 0 {
		logwarn("remaining bits after gob: %v", remainder)
	}

	ns.Namespace = part.Namespace
	ns.Children = make([]*ipfsNamespace, len(part.Children))

	for i, file := range part.Children {
		child := &ipfsNamespace{}
		child.FilePath = file
		child.FilePeer = ns.FilePeer
		ns.Children[i] = child
	}

	if err != nil {
		return errors.Wrap(err, "Error decoding ipfsNamespace Gob")
	}

	ns.loaded = true
	return nil
}

// Write pending changes to IPFS and return the new parent namespace.
func (ns *ipfsNamespace) Persist() (KvNamespace, error) {
	part := &ipfsNamespaceRecord{}
	part.Namespace = ns.Update

	// If this is the first namespace in the chain, don't save children.
	// TODO become parent of multiple children.
	if ns.FilePath != "" {
		part.Children = make([]IpfsPath, 1)
		part.Children[0] = ns.FilePath
	}

	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(part)

	if err != nil {
		return nil, errors.Wrap(err, "Error encoding ipfsNamespace Gob")
	}

	addr, sherr := ns.FilePeer.Shell.Add(&buff)

	if sherr != nil {
		return nil, errors.Wrap(err, "Error adding ipfsNamespace to Ipfs")
	}

	logdbg("Persisted Namespace at: %v", addr)

	out := &ipfsNamespace{}
	out.loaded = true
	out.FilePeer = ns.FilePeer
	out.FilePath = IpfsPath(addr)
	out.Namespace = ns.Update
	out.Update = MakeNamespace()
	out.Children = []*ipfsNamespace{ns}

	return out, nil
}
