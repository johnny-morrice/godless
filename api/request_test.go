package api

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/query"
)

func (request Request) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := Request{}

	chooseType := rand.Float32()

	if chooseType < 0.3 {
		generateQueryRequest(rand, size, &gen)
	} else if chooseType < 0.6 {
		generateReflectRequest(rand, size, &gen)
	} else {
		generateReplicateRequest(rand, size, &gen)
	}

	return reflect.ValueOf(gen)
}

func generateQueryRequest(rand *rand.Rand, size int, gen *Request) {
	gen.Type = API_QUERY
	gen.Query = query.GenQuery(rand, size)
}

func generateReflectRequest(rand *rand.Rand, size int, gen *Request) {
	gen.Type = API_REFLECT

	chooseType := rand.Float32()

	if chooseType < 0.3 {
		gen.Reflection = REFLECT_HEAD_PATH
	} else if chooseType < 0.6 {
		gen.Reflection = REFLECT_INDEX
	} else {
		gen.Reflection = REFLECT_DUMP_NAMESPACE
	}
}

func generateReplicateRequest(rand *rand.Rand, size int, gen *Request) {
	gen.Type = API_REPLICATE

	gen.Replicate = make([]crdt.Link, size)

	for i := 0; i < size; i++ {
		gen.Replicate[i] = crdt.GenLink(rand, size)
	}
}

func TestEncodeRequest(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(requestEncodeOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func requestEncodeOk(expected Request) bool {
	buff := &bytes.Buffer{}
	err := EncodeRequest(expected, buff)

	if err != nil {
		return false
	}

	actual, err := DecodeRequest(buff)

	if err != nil {
		return false
	}

	return expected.Equals(actual)
}

func TestRequestValidateSuccess(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(requestIsValid, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func requestIsValid(request Request) bool {
	return request.Validate() == nil
}

func TestRequestValidateFailure(t *testing.T) {
	badRequests := []Request{
		Request{},
		Request{Type: API_QUERY},
		Request{Type: API_REFLECT},
		Request{Type: API_REPLICATE},
	}

	for _, request := range badRequests {
		err := request.Validate()
		testutil.AssertNonNil(t, err)
	}
}
