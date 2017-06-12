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

func TestMakeIndex(t *testing.T) {
	t.FailNow()
}

func TestIndexAllTables(t *testing.T) {
	t.FailNow()
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
