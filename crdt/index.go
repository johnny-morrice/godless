package crdt

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/log"
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

func EncodeIndex(index Index, w io.Writer) ([]InvalidIndexEntry, error) {
	const failMsg = "EncodeIndex failed"

	message, invalid := MakeIndexMessage(index)

	invalidCount := len(invalid)
	if invalidCount > 0 {
		log.Error("EncodeIndex: %d invalid entries", invalidCount)
	}

	bs, err := pb.Marshal(message)

	if err != nil {
		return invalid, errors.Wrap(err, failMsg)
	}

	var written int
	written, err = w.Write(bs)

	if err != nil {
		return invalid, errors.Wrap(err, fmt.Sprintf("%s after %d bytes", failMsg, written))
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

func (index Index) addStreamEntry(entry IndexStreamEntry) error {
	const failMsg = "joinStreamEntry failed"

	if crypto.IsNilSignature(entry.Signature) {
		index.addTable(entry.TableName, UnsignedLink(entry.Link))
		return nil
	}

	sig, err := crypto.ParseSignature(entry.Signature)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	link := PresignedLink(entry.Link, []crypto.Signature{sig})
	index.addTable(entry.TableName, link)

	return nil
}

// Equals does not take into account any invalid signatures.
func (index Index) Equals(other Index) bool {
	if len(index.Index) != len(other.Index) {
		return false
	}

	indexTables := index.AllTables()

	for _, tableName := range indexTables {
		myLinks := index.Index[tableName]
		theirLinks, present := other.Index[tableName]

		if !present {
			return false
		}

		if len(myLinks) != len(theirLinks) {
			return false
		}

		for i, mine := range myLinks {
			theirs := theirLinks[i]

			if !mine.Equals(theirs) {
				return false
			}
		}
	}

	return true
}

func (index Index) AllTables() []TableName {
	tables := make([]TableName, 0, len(index.Index))

	for name := range index.Index {
		tables = append(tables, name)
	}

	return tables
}

func (index Index) GetTableAddrs(tableName TableName) ([]Link, error) {
	indices, ok := index.Index[tableName]

	if !ok {
		return nil, fmt.Errorf("No table in index: '%s'", tableName)
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
		index.Index[table] = MergeLinks(addr)
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

var __EMPTY_INDEX Index
