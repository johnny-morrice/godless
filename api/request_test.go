package api

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/davecgh/go-spew/spew"

	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
)

func init() {
	log.SetLevel(log.LOG_DEBUG)
}

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

	same := expected.Equals(actual)

	if !same {
		testutil.LogDiff(spew.Sprint(expected), spew.Sprint(actual))
	}

	return same
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
	return request.Validate(StandardRequestValidator()) == nil
}

func TestRequestValidateFailure(t *testing.T) {
	badRequests := []Request{
		Request{},
		Request{Type: API_QUERY},
		Request{Type: API_REFLECT},
		Request{Type: API_REPLICATE},
	}

	for _, request := range badRequests {
		testutil.Assert(t, "Invalid request", !requestIsValid(request))
	}
}
