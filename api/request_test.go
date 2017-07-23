package api

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/function"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func (request Request) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := GenRequest(rand, size)
	return reflect.ValueOf(gen)
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
	validator := RequestValidator{
		Functions: function.StandardFunctions(),
	}
	return request.Validate(validator) == nil
}

func TestRequestValidateFailure(t *testing.T) {
	badRequests := []Request{
		Request{},
		Request{Type: API_QUERY},
		Request{Type: API_REFLECT},
		Request{Type: API_REPLICATE},
	}

	for _, request := range badRequests {
		testutil.Assert(t, "Invalid request", requestIsValid(request))
	}
}
