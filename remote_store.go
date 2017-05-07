package godless

import "fmt"

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

type RemoteNamespaceRecord struct {
	Namespace Namespace
}

type RemoteNamespaceIndex struct {
	Index map[string][]RemoteStoreAddress
}

func MakeRemoteNamespaceIndex(indices map[string]RemoteStoreAddress) RemoteNamespaceIndex {
	out := RemoteNamespaceIndex{
		Index: map[string][]RemoteStoreAddress{},
	}

	for table, addr := range indices {
		out.Index[table] = []RemoteStoreAddress{addr}
	}

	return out
}

func (rni RemoteNamespaceIndex) GetTableIndices(tableName string) ([]RemoteStoreAddress, error) {
	indices, ok := rni.Index[tableName]

	if !ok {
		return nil, fmt.Errorf("No table in index: '%v'", tableName)
	}

	return indices, nil
}

func (rni RemoteNamespaceIndex) JoinNamespace(addr RemoteStoreAddress, namespace Namespace) RemoteNamespaceIndex {
	tables := namespace.GetTableNames()

	out := rni.Copy()
	for _, t := range tables {
		out = out.addTable(t, addr)
	}

	return out
}

func (rni RemoteNamespaceIndex) addTable(table string, addr RemoteStoreAddress) RemoteNamespaceIndex {
	if addrs, ok := rni.Index[table]; ok {
		rni.Index[table] = append(addrs, addr)
	} else {
		rni.Index[table] = []RemoteStoreAddress{addr}
	}

	return rni
}

func (rni RemoteNamespaceIndex) Copy() RemoteNamespaceIndex {
	out := RemoteNamespaceIndex{Index: map[string][]RemoteStoreAddress{}}

	for table, addrs := range rni.Index {
		addrCopy := make([]RemoteStoreAddress, len(addrs))
		for i, a := range addrs {
			addrCopy[i] = a
		}
		out.Index[table] = addrCopy
	}

	return out
}

var EMPTY_RECORD RemoteNamespaceRecord
var EMPTY_INDEX RemoteNamespaceIndex
