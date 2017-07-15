package api

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func (resp Response) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := GenResponse(rand, size)

	return reflect.ValueOf(gen)
}

func TestEncodeAPIResponse(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(apiResponseEncodeOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func TestEncodeAPIResponseText(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(apiResponseEncodeTextOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func apiResponseEncodeOk(resp Response) bool {
	return apiResponseTestWithSerializer(resp, apiResponseSerializationPass)
}

func apiResponseEncodeTextOk(resp Response) bool {
	return apiResponseTestWithSerializer(resp, apiResponseSerializationTextPass)
}

func apiResponseTestWithSerializer(expected Response, pass func(Response) Response) bool {
	actual := pass(expected)
	same := expected.Equals(actual)

	if !same {
		expectText, exErr := expected.AsText()

		if exErr != nil {
			panic(exErr)
		}

		actualText, acErr := actual.AsText()

		if acErr != nil {
			panic(acErr)
		}

		testutil.LogDiff(expectText, actualText)
	}

	return same
}

func apiResponseSerializationPass(resp Response) Response {
	buff := &bytes.Buffer{}
	err := EncodeResponse(resp, buff)

	if err != nil {
		panic(err)
	}

	var decoded Response
	decoded, err = DecodeResponse(buff)

	if err != nil {
		panic(err)
	}

	return decoded
}

func apiResponseSerializationTextPass(resp Response) Response {
	buff := &bytes.Buffer{}
	err := EncodeResponseText(resp, buff)

	if err != nil {
		panic(err)
	}

	var decoded Response
	decoded, err = DecodeResponseText(buff)

	if err != nil {
		panic(err)
	}

	return decoded
}

func makeNamespaceStream(ns crdt.Namespace) []crdt.NamespaceStreamEntry {
	stream, invalid := crdt.MakeNamespaceStream(ns)

	invalidCount := len(invalid)
	if invalidCount > 0 {
		panic(fmt.Sprintf("%d invalid entries", invalidCount))
	}

	return stream
}
