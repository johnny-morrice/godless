package godless

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func (index Index) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := genIndex(rand, size)
	return reflect.ValueOf(gen)
}

func genIndex(rand *rand.Rand, size int) Index {
	index := EmptyIndex()
	const ADDR_SCALE = 1
	const KEY_SCALE = 0.5
	const PATH_SCALE = 0.5

	for i := 0; i < size; i++ {
		keyCount := genCountRange(rand, 1, size, KEY_SCALE)
		indexKey := TableName(randPoint(rand, keyCount))
		addrCount := genCountRange(rand, 1, size, ADDR_SCALE)
		addrs := make([]RemoteStoreAddress, addrCount)
		for j := 0; j < addrCount; j++ {
			pathCount := genCountRange(rand, 1, size, PATH_SCALE)
			a := randPoint(rand, pathCount)
			addrs[j] = IPFSPath(a)
		}

		index.Index[indexKey] = addrs
	}

	return index
}

func TestEncodeIndex(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(indexEncodeOk, config)

	if err != nil {
		t.Error("Unexpected error:", trim(err))
	}
}

func TestEncodeIndexStable(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	const size = 50

	index := genIndex(__RAND, size)

	buff := &bytes.Buffer{}

	encoder := func(w io.Writer) {
		err := EncodeIndex(index, w)

		if err != nil {
			panic(err)
		}
	}

	encoder(buff)

	expected := buff.Bytes()

	assertEncodingStable(t, expected, encoder)
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
