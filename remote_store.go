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
		joined = joined.addTable(t, addr)
	}

	return joined
}

func (rni RemoteNamespaceIndex) addTable(table TableName, addr RemoteStoreAddress) RemoteNamespaceIndex {
	if addrs, ok := rni.Index[table]; ok {
		rni.Index[table] = append(addrs, addr)
	} else {
		rni.Index[table] = []RemoteStoreAddress{addr}
	}

	return rni
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
