package godless

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"github.com/pkg/errors"
)

// TODO unload stored nodes - garbage collection.
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

func (ns *IpfsNamespace) GetMap(namespaceKey string) (Map, error) {
	obj, err := ns.lazyGetObject(namespaceKey)

	if err != nil {
		return Map{}, errors.Wrap(err, "GetMap failed")
	}

	out, ok := obj.Obj.(Map)

	if !ok {
		return Map{}, fmt.Errorf("Not a Set: %v", namespaceKey)
	}

	return out, nil
}

func (ns *IpfsNamespace) GetMapValues(namespaceKey string, mapKey string) ([]string, error) {
	m, err := ns.GetMap(namespaceKey)

	if err != nil {
		return nil, errors.Wrap(err, "GetMapValues failed")
	}

	vs, present := m.Members[mapKey]

	if !present {
		return nil, fmt.Errorf("No values in map for: '%v'", mapKey)
	}

	return vs, nil
}

func (ns *IpfsNamespace) AddMapValues(namespaceKey string, mapKey string, values []string) error {
	members := map[string][]string{
		mapKey: values,
	}

	newObj := Object{
		Obj: Map{Members: members},
	}

	err := ns.Namespace.JoinObject(namespaceKey, newObj)

	if err != nil {
		return errors.Wrap(err, "AddMapValues failed")
	}

	return nil
}

func (ns *IpfsNamespace) GetSet(namespaceKey string) (Set, error) {
	obj, err := ns.lazyGetObject(namespaceKey)

	if err != nil {
		return Set{}, errors.Wrap(err, "GetSet failed")
	}

	out, ok := obj.Obj.(Set)

	if !ok {
		return Set{}, fmt.Errorf("Not a Set: %v", namespaceKey)
	}

	return out, nil
}

func (ns *IpfsNamespace) AddSetValues(namespaceKey string, values []string) error {
	newObj := Object{
		Obj: Set{Members: values},
	}

	err := ns.Namespace.JoinObject(namespaceKey, newObj)

	if err != nil {
		return errors.Wrap(err, "AddSetValues failed")
	}

	return nil
}

func (ns *IpfsNamespace) lazyGetObject(namespaceKey string) (Object, error) {
	// TODO implement
	return Object{}, nil
}

func (ns *IpfsNamespace) LoadAll() (*Namespace, error) {
	out := &Namespace{}

	err := ns.LoadTraverse(func (child *IpfsNamespace) (bool, error) {
		joined, joinerr := out.JoinNamespace(child.Namespace)

		if joinerr != nil {
			return true, joinerr
		}

		out = joined
		return false, nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "Error in IpfsNamespace LoadAll")
	}

	return out, nil
}

// Return whether to abort early, and any error.
type IpfsNamespaceVisitor func (*IpfsNamespace) (bool, error)

func (ns *IpfsNamespace) LoadTraverse(f IpfsNamespaceVisitor) error {
	stack := make([]*IpfsNamespace, 1)
	stack[0] = ns

	for i := 0 ; i < len(stack); i++ {
		current := stack[i]
		err := current.Load()

		if err != nil {
			return errors.Wrap(err, "Error in IpfsNamespace LoadTraverse")
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
