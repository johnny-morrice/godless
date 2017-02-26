package godless

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"github.com/pkg/errors"
)

type IpfsNamespace struct {
	loaded bool
	FilePeer *IpfsPeer
	FilePath IpfsPath
	Namespace *Namespace
	Children []*IpfsNamespace
}

type IpfsNamespaceRecord struct {
	Namespace *Namespace
	Children []IpfsPath
}

func (ns *IpfsNamespace) LoadAll() (*Namespace, error) {
	out := &Namespace{}

	err := ns.LoadTraverse(func (child *IpfsNamespace) {
		out = out.JoinNamespace(child.Namespace)
	})

	if err != nil {
		return nil, errors.Wrap(err, "Error in IpfsNamespace LoadAll")
	}

	return out, nil
}

func (ns *IpfsNamespace) LoadTraverse(f func (*IpfsNamespace)) error {
	stack := make([]*IpfsNamespace, 1)
	stack[0] = ns

	for i := 0 ; i < len(stack); i++ {
		current := stack[i]
		err := current.Load()

		if err != nil {
			return errors.Wrap(err, "Error in IpfsNamespace LoadTraverse")
		}

		f(current)
		stack = append(stack, current.Children...)
	}

	return nil
}

// Load chunks over IPFS
func (ns *IpfsNamespace) Load() error {
	if ns.FilePath == "" {
		panic("Tried to load IpfsNamespace with empty FilePath")
	}

	if ns.loaded {
		log.Printf("WARN IpfsNamespace already loaded from: '%v'", ns.FilePath)
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
		log.Printf("WARN error draining reader: %v", drainerr)
	}

	if len(remainder) != 0 {
		log.Printf("WARN remaining bits after gob: %v", remainder)
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

// Write chunks to IPFS
func (ns *IpfsNamespace) Persist() error {
	if ns.loaded {
		log.Printf("WARN persisting loaded IpfsNamespace at: %v", ns.FilePath)
	}

	if ns.FilePath != "" {
		log.Printf("WARN IpfsNamespace already persisted at: %v", ns.FilePath)
		return nil
	}

	part := &IpfsNamespaceRecord{}
	part.Namespace = ns.Namespace

	part.Children = make([]IpfsPath, len(ns.Children))
	for i, child := range ns.Children {
		part.Children[i] = child.FilePath
	}

	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(part)

	if err != nil {
		return errors.Wrap(err, "Error encoding IpfsNamespace Gob")
	}

	addr, sherr := ns.FilePeer.Shell.Add(&buff)

	if sherr != nil {
		return errors.Wrap(err, "Error adding IpfsNamespace to Ipfs")
	}

	ns.FilePath = IpfsPath(addr)

	return nil
}
