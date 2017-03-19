package godless

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"github.com/pkg/errors"
)

type IpfsNamespace struct {
	loaded bool
	dirty bool

	updateCache *Namespace

	FilePeer *IpfsPeer
	FilePath IpfsPath
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
	ns.updateCache = MakeNamespace()
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
	ns.updateCache = namespace
	ns.Namespace = MakeNamespace()
	ns.dirty = true
	ns.Children = []*IpfsNamespace{}

	return ns.Persist()
}

// Return whether to abort early, and any error.
type IpfsNamespaceVisitor func (*IpfsNamespace) (bool, error)

func (ns *IpfsNamespace) loadTraverse(f IpfsNamespaceVisitor) error {
	stack := make([]*IpfsNamespace, 1)
	stack[0] = ns

	for i := 0 ; i < len(stack); i++ {
		current := stack[i]
		err := current.load()

		if err != nil {
			return errors.Wrap(err, "Error in IpfsNamespace loadTraverse")
		}

		abort, visiterr := f(current)

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
	if !ns.dirty {
		logwarn("persisting unchanged IpfsNamespace at: %v", ns.FilePath)
		return nil, nil
	}

	part := &IpfsNamespaceRecord{}
	part.Namespace = ns.updateCache

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
	out.Namespace = ns.updateCache
	out.updateCache = MakeNamespace()
	out.Children = []*IpfsNamespace{ns}

	return out, nil
}
