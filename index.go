package godless

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

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
	const failMsg = "EncodeIndex failed"

	message := MakeIndexMessage(index)
	bs, err := proto.Marshal(message)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	var written int
	written, err = w.Write(bs)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%v after %v bytes", failMsg, written))
	}

	return nil
}

func DecodeIndex(r io.Reader) (RemoteNamespaceIndex, error) {
	const failMsg = "DecodeIndex failed"

	message := &IndexMessage{}
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return EMPTY_INDEX, errors.Wrap(err, failMsg)
	}

	err = proto.Unmarshal(bs, message)

	if err != nil {
		return EMPTY_INDEX, errors.Wrap(err, failMsg)
	}

	return ReadIndexMessage(message), nil
}

func ReadIndexMessage(message *IndexMessage) RemoteNamespaceIndex {
	stream := ReadIndexStreamMessage(message)
	return ReadIndexStream(stream)
}

func MakeIndexMessage(index RemoteNamespaceIndex) *IndexMessage {
	stream := MakeIndexStream(index)
	return MakeIndexStreamMessage(stream)
}

func (index RemoteNamespaceIndex) joinStreamEntry(entry IndexStreamEntry) RemoteNamespaceIndex {
	cpy := index.Copy()
	addrs := make([]RemoteStoreAddress, len(entry.Links))

	for i, l := range entry.Links {
		addrs[i] = RemoteStoreAddress(l)
	}

	cpy.addTable(entry.TableName, addrs...)

	return cpy
}

func (index RemoteNamespaceIndex) Equals(other RemoteNamespaceIndex) bool {
	stream := MakeIndexStream(index)
	otherStream := MakeIndexStream(other)

	for i, entry := range stream {
		otherEntry := otherStream[i]
		if !entry.Equals(otherEntry) {
			return false
		}
	}

	return true
}

func (index RemoteNamespaceIndex) AllTables() []TableName {
	tables := make([]TableName, len(index.Index))

	i := 0
	for name := range index.Index {
		tables[i] = name
		i++
	}

	return tables
}

func (index RemoteNamespaceIndex) GetTableAddrs(tableName TableName) ([]RemoteStoreAddress, error) {
	indices, ok := index.Index[tableName]

	if !ok {
		return nil, fmt.Errorf("No table in index: '%v'", tableName)
	}

	return indices, nil
}

func (index RemoteNamespaceIndex) JoinNamespace(addr RemoteStoreAddress, namespace Namespace) RemoteNamespaceIndex {
	tables := namespace.GetTableNames()

	joined := index.Copy()
	for _, t := range tables {
		joined.addTable(t, addr)
	}

	return joined
}

func (index RemoteNamespaceIndex) addTable(table TableName, addr ...RemoteStoreAddress) {
	if addrs, ok := index.Index[table]; ok {
		normal := normalStoreAddress(append(addrs, addr...))
		index.Index[table] = normal
	} else {
		index.Index[table] = addr
	}
}

func (index RemoteNamespaceIndex) Copy() RemoteNamespaceIndex {
	cpy := EmptyRemoteNamespaceIndex()

	for table, addrs := range index.Index {
		addrCopy := make([]RemoteStoreAddress, len(addrs))
		for i, a := range addrs {
			addrCopy[i] = a
		}
		cpy.Index[table] = addrCopy
	}

	return cpy
}

var EMPTY_INDEX RemoteNamespaceIndex
