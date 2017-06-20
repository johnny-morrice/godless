package crdt

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"

	"github.com/ethereum/go-ethereum/log"
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"

	pb "github.com/gogo/protobuf/proto"
)

type Index struct {
	Index map[TableName][]Link
}

func EmptyIndex() Index {
	return MakeIndex(map[TableName]Link{})
}

func MakeIndex(indices map[TableName]Link) Index {
	out := Index{
		Index: map[TableName][]Link{},
	}

	for table, addr := range indices {
		out.Index[table] = []Link{addr}
	}

	return out
}

// Just encode as Gob for now.
func EncodeIndex(index Index, w io.Writer) ([]InvalidIndexEntry, error) {
	const failMsg = "EncodeIndex failed"

	message, invalid := MakeIndexMessage(index)

	invalidCount := len(invalid)
	if invalidCount > 0 {
		log.Error("EncodeIndex: %v invalid entries", invalidCount)
	}

	bs, err := pb.Marshal(message)

	if err != nil {
		return invalid, errors.Wrap(err, failMsg)
	}

	var written int
	written, err = w.Write(bs)

	if err != nil {
		return invalid, errors.Wrap(err, fmt.Sprintf("%v after %v bytes", failMsg, written))
	}

	return invalid, nil
}

func DecodeIndex(r io.Reader) (Index, []InvalidIndexEntry, error) {
	const failMsg = "DecodeIndex failed"

	message := &proto.IndexMessage{}
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return __EMPTY_INDEX, nil, errors.Wrap(err, failMsg)
	}

	err = pb.Unmarshal(bs, message)

	if err != nil {
		return __EMPTY_INDEX, nil, errors.Wrap(err, failMsg)
	}

	index, invalid := ReadIndexMessage(message)

	return index, invalid, nil
}

func ReadIndexMessage(message *proto.IndexMessage) (Index, []InvalidIndexEntry) {
	stream := ReadIndexStreamMessage(message)
	return ReadIndexStream(stream)
}

func MakeIndexMessage(index Index) (*proto.IndexMessage, []InvalidIndexEntry) {
	stream, invalid := MakeIndexStream(index)
	return MakeIndexStreamMessage(stream), invalid
}

func (index Index) IsEmpty() bool {
	return len(index.Index) == 0
}

func (index Index) ForTable(tableName TableName, f func(link Link)) error {
	const failMsg = "index.ForTable failed"

	links, err := index.GetTableAddrs(tableName)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	for _, l := range links {
		f(l)
	}

	return nil
}

func (index Index) JoinIndex(other Index) Index {
	cpy := index.Copy()

	for table, addrs := range other.Index {
		cpy.addTable(table, addrs...)
	}

	return cpy
}

func (index Index) joinStreamEntry(entry IndexStreamEntry) (Index, error) {
	const failMsg = "joinStreamEntry failed"

	cpy := index.Copy()

	sig, err := crypto.ParseSignature(entry.Signature)

	if err != nil {
		return __EMPTY_INDEX, errors.Wrap(err, failMsg)
	}

	cpy.addTable(entry.TableName, PreSignedLink(entry.Link, sig))

	return cpy, nil
}

// Equals does not take into account any invalid signatures.
func (index Index) Equals(other Index) bool {
	stream, _ := MakeIndexStream(index)
	otherStream, _ := MakeIndexStream(other)

	if len(stream) != len(otherStream) {
		return false
	}

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

func (index Index) GetTableAddrs(tableName TableName) ([]Link, error) {
	indices, ok := index.Index[tableName]

	if !ok {
		return nil, fmt.Errorf("No table in index: '%v'", tableName)
	}

	return indices, nil
}

func (index Index) JoinNamespace(addr Link, namespace Namespace) Index {
	tables := namespace.GetTableNames()

	joined := index.Copy()
	for _, t := range tables {
		joined.addTable(t, addr)
	}

	return joined
}

func (index Index) JoinTable(table TableName, addr ...Link) Index {
	cpy := index.Copy()

	cpy.addTable(table, addr...)

	return cpy
}

func (index Index) addTable(table TableName, addr ...Link) {
	if addrs, ok := index.Index[table]; ok {
		normal := MergeLinks(append(addrs, addr...))
		index.Index[table] = normal
	} else {
		index.Index[table] = addr
	}
}

func (index Index) Copy() Index {
	cpy := EmptyIndex()

	for table, addrs := range index.Index {
		addrCopy := make([]Link, len(addrs))
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
		addrs := make([]Link, addrCount)
		for j := 0; j < addrCount; j++ {
			pathCount := testutil.GenCountRange(rand, 1, size, PATH_SCALE)
			a := testutil.RandPoint(rand, pathCount)
			addrs[j] = UnsignedLink(IPFSPath(a))
		}

		index.Index[indexKey] = addrs
	}

	return index
}

var __EMPTY_INDEX Index
