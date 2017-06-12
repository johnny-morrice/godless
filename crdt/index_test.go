package crdt

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func (index Index) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := GenIndex(rand, size)
	return reflect.ValueOf(gen)
}

func TestEncodeIndex(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(indexEncodeOk, config)

	if err != nil {
		t.Error("Unexpected error:", testutil.Trim(err))
	}
}

func TestEncodeIndexStable(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	const size = 50

	index := GenIndex(__RAND, size)

	buff := &bytes.Buffer{}

	encoder := func(w io.Writer) {
		err := EncodeIndex(index, w)

		if err != nil {
			panic(err)
		}
	}

	encoder(buff)

	expected := buff.Bytes()

	testutil.AssertEncodingStable(t, expected, encoder)
}

func TestEmptyIndex(t *testing.T) {
	index := EmptyIndex()

	testutil.AssertEquals(t, "Expected empty index", 0, len(index.Index))
}

func TestMakeIndex(t *testing.T) {
	const table = "Hi"
	const value = "world"
	index := MakeIndex(map[TableName]IPFSPath{
		table: value,
	})

	testutil.AssertEquals(t, "Expected index length 1", 1, len(index.Index))

	expected := []IPFSPath{value}
	actual, err := index.GetTableAddrs(table)
	if err != nil {
		panic(err)
	}

	testutil.AssertEquals(t, "Unexpected index addr", expected, actual)
}

func TestIndexAllTables(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(indexAllTablesOk, config)

	if err != nil {
		t.Error("Unexpected error:", testutil.Trim(err))
	}
}

func indexAllTablesOk(index Index) bool {
	tables := index.AllTables()

	if len(index.Index) != len(tables) {
		return false
	}

	for _, t := range tables {
		_, err := index.GetTableAddrs(t)

		if err != nil {
			return false
		}
	}

	return true
}

func TestIndexCopy(t *testing.T) {
	t.FailNow()
}

func TestIndexEquals(t *testing.T) {
	t.FailNow()
}

func TestIndexGetTableAddrs(t *testing.T) {
	t.FailNow()
}

func TestIndexIsEmpty(t *testing.T) {
	t.FailNow()
}

func TestIndexJoinIndex(t *testing.T) {
	t.FailNow()
}

func TestIndexJoinNamespace(t *testing.T) {
	t.FailNow()
}

func indexEncodeOk(expected Index) bool {
	actual := indexSerializationPass(expected)
	return expected.Equals(actual)
}

func indexSerializationPass(expected Index) Index {
	buff := &bytes.Buffer{}
	err := EncodeIndex(expected, buff)

	if err != nil {
		panic(err)
	}

	var actual Index
	actual, err = DecodeIndex(buff)

	if err != nil {
		panic(err)
	}

	return actual
}
