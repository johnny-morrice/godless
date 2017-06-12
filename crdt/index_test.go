package crdt

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
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

	testutil.AssertVerboseErrorIsNil(t, err)
}

func TestEncodeIndexStable(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	const size = 50

	index := GenIndex(testutil.Rand(), size)

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

	testutil.AssertVerboseErrorIsNil(t, err)
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
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(indexCopyOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func indexCopyOk(expected Index) bool {
	actual := expected.Copy()
	return expected.Equals(actual)
}

func TestIndexEquals(t *testing.T) {
	if !testing.Short() {
		testIndexEqualsQuick(t)
	}

	indexA := MakeIndex(map[TableName]IPFSPath{
		"hi": "world",
	})
	indexB := MakeIndex(map[TableName]IPFSPath{
		"hello": "world",
	})

	testutil.Assert(t, "Expected indexA to be equal to itself", indexA.Equals(indexA))
	testutil.Assert(t, "Expected index to be not equal to indexB", !indexA.Equals(indexB))
}

func testIndexEqualsQuick(t *testing.T) {
	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(indexEqualsOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func indexEqualsOk(index Index) bool {
	return index.Equals(index)
}

func TestIndexGetTableAddrs(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(indexGetTableAddrsOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func indexGetTableAddrsOk(index Index) bool {
	return isIndexSubset(index, index)
}

func isIndexSubset(subset, superset Index) bool {
	for table, expected := range subset.Index {
		actual, err := superset.GetTableAddrs(table)

		if err != nil {
			log.Debug("Subset key '%v' not found in superset")
			return false
		}

	LOOP:
		for _, exAddr := range expected {
			for _, acAddr := range actual {
				if exAddr == acAddr {
					continue LOOP
				}
			}
			log.Debug("Subset key %v not found in %v", exAddr, actual)

			return false
		}
	}

	return true
}

func assertIndexSubset(t *testing.T, subset, superset Index) {
	testutil.Assert(t, "Expected index subset", isIndexSubset(subset, superset))
}

func TestIndexIsEmpty(t *testing.T) {
	empty := EmptyIndex()

	testutil.Assert(t, "Expected empty index", empty.IsEmpty())

	full := MakeIndex(map[TableName]IPFSPath{
		"Hi": "world",
	})

	testutil.Assert(t, "Expected non empty index", !full.IsEmpty())
}

func TestIndexJoinIndex(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	const size = 50

	for i := 0; i < testutil.ENCODE_REPEAT_COUNT; i++ {
		indexA := GenIndex(testutil.Rand(), size)
		indexB := GenIndex(testutil.Rand(), size)

		joinedA := indexA.JoinIndex(indexB)
		joinedB := indexB.JoinIndex(indexA)

		// Should have AssertEquals check for Equals method with reflection.
		testutil.Assert(t, "Expected equal indices", joinedA.Equals(joinedB))

		log.Debug("Checking subsets")
		assertIndexSubset(t, indexA, joinedA)
		assertIndexSubset(t, indexB, joinedA)
	}
}

func TestIndexJoinNamespace(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	const size = 50
	const expected = "hello"

	for i := 0; i < testutil.ENCODE_REPEAT_COUNT; i++ {
		index := GenIndex(testutil.Rand(), size)
		namespace := GenNamespace(testutil.Rand(), size)

		joined := index.JoinNamespace(expected, namespace)

	LOOP:
		for _, table := range namespace.GetTableNames() {
			actual, err := joined.GetTableAddrs(table)
			testutil.AssertNil(t, err)

			for _, ac := range actual {
				if ac == expected {
					continue LOOP
				}
			}

			t.Errorf("%v not found in %v", expected, actual)
		}

	}
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
