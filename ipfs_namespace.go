package godless

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"github.com/pkg/errors"
)

type IpfsNamespace struct {
	loaded bool
	dirty bool

	queryCache *Namespace
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
	ns.queryCache = NewNamespace()
	ns.updateCache = NewNamespace()
	ns.Namespace = NewNamespace()
	ns.Children = []*IpfsNamespace{}

	err := ns.cacheChildNamespaces()

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error loading namespace at '%v'", filePath))
	}

	return ns, nil
}

func PersistNewIpfsNamespace(filePeer *IpfsPeer, namespace *Namespace) (*IpfsNamespace, error) {
	ns := &IpfsNamespace{}
	ns.FilePeer = filePeer
	ns.queryCache = NewNamespace()
	ns.updateCache = namespace
	ns.Namespace = NewNamespace()
	ns.Children = []*IpfsNamespace{}

	return ns.Persist()
}

func (ns *IpfsNamespace) GetMap(namespaceKey string) (Map, error) {
	obj, err := ns.loadObject(namespaceKey)

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

func (ns *IpfsNamespace) JoinMap(namespaceKey string, update map[string][]string) error {
	newObj := Object{
		Obj: Map{Members: update},
	}

	err := ns.updateCache.JoinObject(namespaceKey, newObj)

	if err != nil {
		return errors.Wrap(err, "AddMapValues failed")
	}

	ns.dirty = true

	return nil
}

func (ns *IpfsNamespace) GetSet(namespaceKey string) (Set, error) {
	obj, err := ns.loadObject(namespaceKey)

	if err != nil {
		return Set{}, errors.Wrap(err, "GetSet failed")
	}

	out, ok := obj.Obj.(Set)

	if !ok {
		return Set{}, fmt.Errorf("Not a Set: %v", namespaceKey)
	}

	return out, nil
}

func (ns *IpfsNamespace) JoinSet(namespaceKey string, values []string) error {
	newObj := Object{
		Obj: Set{Members: values},
	}

	err := ns.updateCache.JoinObject(namespaceKey, newObj)

	if err != nil {
		return errors.Wrap(err, "AddSetValues failed")
	}

	ns.dirty = true

	return nil
}

func (ns *IpfsNamespace) loadObject(namespaceKey string) (Object, error) {
	old, oldPresent := ns.queryCache.Objects[namespaceKey]
	new, newPresent := ns.updateCache.Objects[namespaceKey]
	out := Object{}

	if !newPresent && !oldPresent {
		return Object{}, fmt.Errorf("No such object: %v", namespaceKey)
	}

	if oldPresent {
		out = old
	}

	if newPresent {
		joined, err := new.JoinObject(old)

		if err != nil {
			return Object{}, errors.Wrap(err, fmt.Sprintf("loadObject failed for '%v'", namespaceKey))
		}

		out = joined
	}

	return out, nil
}

func (ns *IpfsNamespace) cacheChildNamespaces() error {
	err := ns.loadTraverse(func (child *IpfsNamespace) (bool, error) {
		joined, joinerr := ns.queryCache.JoinNamespace(child.Namespace)

		if joinerr != nil {
			return true, joinerr
		}

		ns.queryCache = joined
		return false, nil
	})

	if err != nil {
		return errors.Wrap(err, "Error in IpfsNamespace LoadAll")
	}

	return nil
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

// Write pending changes to IPFS and return the new parent namespace.
// TODO allow child namespace merges
func (ns *IpfsNamespace) Persist() (*IpfsNamespace, error) {
	if !ns.dirty {
		log.Printf("WARN persiting unchanged IpfsNamespace at: %v", ns.FilePath)
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

	out := &IpfsNamespace{}
	out.loaded = true
	out.FilePeer = ns.FilePeer
	out.FilePath = IpfsPath(addr)
	out.Namespace = ns.updateCache
	out.updateCache = NewNamespace()

	// Join query cache
	cache, jerr := ns.queryCache.JoinNamespace(ns.updateCache)

	if jerr != nil {
		return nil, errors.Wrap(err, "Error joining new parent query cache")
	}

	out.queryCache = cache
	out.Children = []*IpfsNamespace{ns}

	return out, nil
}
