package godless

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"github.com/pkg/errors"
)

type IpfsNamespace struct {
	loaded bool
	Update *Namespace

	FilePeer *IpfsPeer
	FilePath IpfsPath
	// Breaking 12 factor rule in caching namespace...
	Namespace *Namespace
	Children []*IpfsNamespace
}

type IpfsNamespaceRecord struct {
	Namespace *Namespace
	Children []IpfsPath
}

func LoadIpfsNamespace(filePeer *IpfsPeer, filePath IpfsPath) (*IpfsNamespace, error) {
	ns := &IpfsNamespace{}
	ns.FilePeer = filePeer
	ns.FilePath = filePath
	ns.Update = MakeNamespace()
	ns.Namespace = MakeNamespace()
	ns.Children = []*IpfsNamespace{}

	err := ns.load()

	if err != nil {
		return nil, errors.Wrap(err, "Error loading new namespace")
	}

	return ns, nil
}

func PersistNewIpfsNamespace(filePeer *IpfsPeer, namespace *Namespace) (*IpfsNamespace, error) {
	ns := &IpfsNamespace{}
	ns.FilePeer = filePeer
	ns.Update = namespace
	ns.Namespace = MakeNamespace()
	ns.Children = []*IpfsNamespace{}

	return ns.Persist()
}

func (ns *IpfsNamespace) JoinTable(tableKey string, table Table) error {
	joined, joinerr := ns.Update.JoinTable(tableKey, table)

	if joinerr != nil {
		return errors.Wrap(joinerr, "IpfsNamespace.JoinTable failed")
	}

	ns.Update = joined

	return nil
}

func (ns *IpfsNamespace) NamespaceLeaf() *Namespace {
	return ns.Namespace
}

func (ns *IpfsNamespace) LoadTraverse(f RemoteNamespaceReader) error {
	stack := make([]*IpfsNamespace, 1)
	stack[0] = ns

	for i := 0 ; i < len(stack); i++ {
		current := stack[i]
		err := current.load()

		if err != nil {
			return errors.Wrap(err, "Error in IpfsNamespace loadTraverse")
		}

		leaf := current.NamespaceLeaf()
		abort, visiterr := f.ReadNamespace(leaf)

		if visiterr != nil {
			return errors.Wrap(err, "Error in IpfsNamespace traversal")
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
func (ns *IpfsNamespace) load() error {
	if ns.FilePath == "" {
		logdie("tried to load IpfsNamespace with empty FilePath")
	}

	if ns.loaded {
		logwarn("IpfsNamespace already loaded from: '%v'", ns.FilePath)
		return nil
	}

	reader, err := ns.FilePeer.Shell.Cat(string(ns.FilePath))

	if err != nil {
		return errors.Wrap(err, "Error in IpfsNamespace Cat")
	}

	defer reader.Close()

	dec := gob.NewDecoder(reader)
	part := &IpfsNamespaceRecord{}
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
	ns.Children = make([]*IpfsNamespace, len(part.Children))

	for i, file := range part.Children {
		child := &IpfsNamespace{}
		child.FilePath = file
		child.FilePeer = ns.FilePeer
		ns.Children[i] = child
	}

	if err != nil {
		return errors.Wrap(err, "Error decoding IpfsNamespace Gob")
	}

	ns.loaded = true
	return nil
}

// Write pending changes to IPFS and return the new parent namespace.
// TODO allow child namespace merges
func (ns *IpfsNamespace) Persist() (*IpfsNamespace, error) {
	part := &IpfsNamespaceRecord{}
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
		return nil, errors.Wrap(err, "Error encoding IpfsNamespace Gob")
	}

	addr, sherr := ns.FilePeer.Shell.Add(&buff)

	if sherr != nil {
		return nil, errors.Wrap(err, "Error adding IpfsNamespace to Ipfs")
	}

	logdbg("Persisted Namespace at: %v", addr)

	out := &IpfsNamespace{}
	out.loaded = true
	out.FilePeer = ns.FilePeer
	out.FilePath = IpfsPath(addr)
	out.Namespace = ns.Update
	out.Update = MakeNamespace()
	out.Children = []*IpfsNamespace{ns}

	return out, nil
}
