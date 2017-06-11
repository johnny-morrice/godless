package crdt

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"

	"github.com/johnny-morrice/godless/proto"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/pkg/errors"

	pb "github.com/gogo/protobuf/proto"
)

type Index struct {
	Index map[TableName][]RemoteStoreAddress
}

func EmptyIndex() Index {
	return MakeIndex(map[TableName]RemoteStoreAddress{})
}

func MakeIndex(indices map[TableName]RemoteStoreAddress) Index {
	out := Index{
		Index: map[TableName][]RemoteStoreAddress{},
	}

	for table, addr := range indices {
		out.Index[table] = []RemoteStoreAddress{addr}
	}

	return out
}

// Just encode as Gob for now.
func EncodeIndex(index Index, w io.Writer) error {
	const failMsg = "EncodeIndex failed"

	message := MakeIndexMessage(index)
	bs, err := pb.Marshal(message)

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

func DecodeIndex(r io.Reader) (Index, error) {
	const failMsg = "DecodeIndex failed"

	message := &proto.IndexMessage{}
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return __EMPTY_INDEX, errors.Wrap(err, failMsg)
	}

	err = pb.Unmarshal(bs, message)

	if err != nil {
		return __EMPTY_INDEX, errors.Wrap(err, failMsg)
	}

	return ReadIndexMessage(message), nil
}

func ReadIndexMessage(message *proto.IndexMessage) Index {
	stream := ReadIndexStreamMessage(message)
	return ReadIndexStream(stream)
}

func MakeIndexMessage(index Index) *proto.IndexMessage {
	stream := MakeIndexStream(index)
	return MakeIndexStreamMessage(stream)
}

func (index Index) IsEmpty() bool {
	return len(index.Index) == 0
}

func (index Index) JoinIndex(other Index) Index {
	cpy := index.Copy()

	for table, addrs := range other.Index {
		cpy.addTable(table, addrs...)
	}

	return cpy
}

func (index Index) joinStreamEntry(entry IndexStreamEntry) Index {
	cpy := index.Copy()
	addrs := make([]RemoteStoreAddress, len(entry.Links))

	for i, l := range entry.Links {
		addrs[i] = RemoteStoreAddress(l)
	}

	cpy.addTable(entry.TableName, addrs...)

	return cpy
}

func (index Index) Equals(other Index) bool {
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

func (index Index) AllTables() []TableName {
	tables := make([]TableName, len(index.Index))

	i := 0
	for name := range index.Index {
		tables[i] = name
		i++
	}

	return tables
}

func (index Index) GetTableAddrs(tableName TableName) ([]RemoteStoreAddress, error) {
	indices, ok := index.Index[tableName]

	if !ok {
		return nil, fmt.Errorf("No table in index: '%v'", tableName)
	}

	return indices, nil
}

func (index Index) JoinNamespace(addr RemoteStoreAddress, namespace Namespace) Index {
	tables := namespace.GetTableNames()

	joined := index.Copy()
	for _, t := range tables {
		joined.addTable(t, addr)
	}

	return joined
}

func (index Index) addTable(table TableName, addr ...RemoteStoreAddress) {
	if addrs, ok := index.Index[table]; ok {
		normal := normalStoreAddress(append(addrs, addr...))
		index.Index[table] = normal
	} else {
		index.Index[table] = addr
	}
}

func (index Index) Copy() Index {
	cpy := EmptyIndex()

	for table, addrs := range index.Index {
		addrCopy := make([]RemoteStoreAddress, len(addrs))
		for i, a := range addrs {
			addrCopy[i] = a
		}
		cpy.Index[table] = addrCopy
	}

	return cpy
}

func GenIndex(rand *rand.Rand, size int) Index {
	index := EmptyIndex()
	const ADDR_SCALE = 1
	const KEY_SCALE = 0.5
	const PATH_SCALE = 0.5

	for i := 0; i < size; i++ {
		keyCount := testutil.GenCountRange(rand, 1, size, KEY_SCALE)
		indexKey := TableName(testutil.RandPoint(rand, keyCount))
		addrCount := testutil.GenCountRange(rand, 1, size, ADDR_SCALE)
		addrs := make([]RemoteStoreAddress, addrCount)
		for j := 0; j < addrCount; j++ {
			pathCount := testutil.GenCountRange(rand, 1, size, PATH_SCALE)
			a := testutil.RandPoint(rand, pathCount)
			addrs[j] = IPFSPath(a)
		}

		index.Index[indexKey] = addrs
	}

	return index
}

var __EMPTY_INDEX Index
