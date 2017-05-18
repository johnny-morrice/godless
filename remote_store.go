package godless

import (
	"encoding/gob"
	"fmt"
	"io"
	"sort"
)

//go:generate mockgen -destination mock/mock_remote_store.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless RemoteStore

type RemoteStore interface {
	Connect() error
	AddNamespace(RemoteNamespaceRecord) (RemoteStoreAddress, error)
	AddIndex(RemoteNamespaceIndex) (RemoteStoreAddress, error)
	CatNamespace(RemoteStoreAddress) (RemoteNamespaceRecord, error)
	CatIndex(RemoteStoreAddress) (RemoteNamespaceIndex, error)
	Disconnect() error
}

type RemoteStoreAddress interface {
	Path() string
}

type byPath []RemoteStoreAddress

func (addrs byPath) Len() int {
	return len(addrs)
}

func (addrs byPath) Swap(i, j int) {
	addrs[i], addrs[j] = addrs[j], addrs[i]
}

func (addrs byPath) Less(i, j int) bool {
	return addrs[i].Path() < addrs[j].Path()
}

func normalStoreAddress(addrs []RemoteStoreAddress) []RemoteStoreAddress {
	uniq := uniqStoreAddress(addrs)
	sort.Sort(byPath(uniq))
	return uniq
}

func uniqStoreAddress(addrs []RemoteStoreAddress) []RemoteStoreAddress {
	dedupe := map[string]RemoteStoreAddress{}

	for _, a := range addrs {
		path := a.Path()
		if _, present := dedupe[path]; !present {
			dedupe[path] = a
		}
	}

	uniq := make([]RemoteStoreAddress, len(dedupe))

	i := 0
	for _, a := range dedupe {
		uniq[i] = a
		i++
	}

	return uniq
}

type RemoteNamespaceRecord struct {
	Namespace Namespace
}

// TODO tiny type for table name.
type RemoteNamespaceIndex struct {
	Index map[TableName][]RemoteStoreAddress
}

func EmptyRemoteNamespaceIndex() RemoteNamespaceIndex {
	return MakeRemoteNamespaceIndex(map[TableName]RemoteStoreAddress{})
}

func MakeRemoteNamespaceIndex(indices map[TableName]RemoteStoreAddress) RemoteNamespaceIndex {
	out := RemoteNamespaceIndex{
		Index: map[TableName][]RemoteStoreAddress{},
	}

	for table, addr := range indices {
		out.Index[table] = []RemoteStoreAddress{addr}
	}

	return out
}

// Just encode as Gob for now.
func EncodeIndex(index RemoteNamespaceIndex, w io.Writer) error {
	stream := MakeIndexStream(index)
	enc := gob.NewEncoder(w)
	return enc.Encode(stream)
}

func DecodeIndex(r io.Reader) (RemoteNamespaceIndex, error) {
	var stream []IndexStreamEntry
	dec := gob.NewDecoder(r)
	err := dec.Decode(&stream)
	return ReadIndexStream(stream), err
}

func (rni RemoteNamespaceIndex) joinStreamEntry(entry IndexStreamEntry) RemoteNamespaceIndex {
	cpy := rni.Copy()
	addrs := make([]RemoteStoreAddress, len(entry.Links))

	for i, l := range entry.Links {
		addrs[i] = RemoteStoreAddress(l)
	}

	cpy.addTable(entry.TableName, addrs...)

	return cpy
}

func (rni RemoteNamespaceIndex) APIIndex() APIRemoteIndex {
	apiIndex := APIRemoteIndex{
		Index: map[string][]string{},
	}

	for table, addrs := range rni.Index {
		apiAddrs := make([]string, len(addrs))
		for i, a := range addrs {
			apiAddrs[i] = a.Path()
		}
		apiIndex.Index[string(table)] = apiAddrs
	}

	return apiIndex
}

func (rni RemoteNamespaceIndex) AllTables() []TableName {
	tables := make([]TableName, len(rni.Index))

	i := 0
	for name := range rni.Index {
		tables[i] = name
		i++
	}

	return tables
}

func (rni RemoteNamespaceIndex) GetTableAddrs(tableName TableName) ([]RemoteStoreAddress, error) {
	indices, ok := rni.Index[tableName]

	if !ok {
		return nil, fmt.Errorf("No table in index: '%v'", tableName)
	}

	return indices, nil
}

func (rni RemoteNamespaceIndex) JoinNamespace(addr RemoteStoreAddress, namespace Namespace) RemoteNamespaceIndex {
	tables := namespace.GetTableNames()

	joined := rni.Copy()
	for _, t := range tables {
		joined.addTable(t, addr)
	}

	return joined
}

func (rni RemoteNamespaceIndex) addTable(table TableName, addr ...RemoteStoreAddress) {
	if addrs, ok := rni.Index[table]; ok {
		normal := normalStoreAddress(append(addrs, addr...))
		rni.Index[table] = normal
	} else {
		rni.Index[table] = addr
	}
}

func (rni RemoteNamespaceIndex) Copy() RemoteNamespaceIndex {
	cpy := EmptyRemoteNamespaceIndex()

	for table, addrs := range rni.Index {
		addrCopy := make([]RemoteStoreAddress, len(addrs))
		for i, a := range addrs {
			addrCopy[i] = a
		}
		cpy.Index[table] = addrCopy
	}

	return cpy
}

var EMPTY_RECORD RemoteNamespaceRecord
var EMPTY_INDEX RemoteNamespaceIndex
