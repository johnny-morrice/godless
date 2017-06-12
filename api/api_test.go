package api

import (
	"bytes"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func (resp APIResponse) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := APIResponse{}

	gen.Msg = testutil.RandLetters(rand, size)

	if rand.Float32() < 0.5 {
		gen.Type = API_QUERY
	} else {
		gen.Type = API_REFLECT
	}

	if rand.Float32() < 0.5 {
		if gen.Type == API_QUERY {
			gen.QueryResponse = genQueryResponse(rand, size)
		} else {
			gen.ReflectResponse = genReflectResponse(rand, size)
		}
	} else {
		errText := testutil.RandPoint(rand, size)
		gen.Err = errors.New(errText)
	}

	return reflect.ValueOf(gen)
}

func genQueryResponse(rand *rand.Rand, size int) APIQueryResponse {
	gen := APIQueryResponse{}
	ns := crdt.GenNamespace(rand, size)
	stream := crdt.MakeNamespaceStream(ns)
	gen.Entries = stream
	return gen
}

func genReflectResponse(rand *rand.Rand, size int) APIReflectResponse {
	gen := APIReflectResponse{}

	branch := rand.Float32()
	if branch < 0.333 {
		gen.Type = REFLECT_HEAD_PATH
		gen.Path = crdt.IPFSPath(testutil.RandLetters(rand, size))
	} else if branch < 0.666 {
		gen.Type = REFLECT_DUMP_NAMESPACE
		gen.Namespace = crdt.GenNamespace(rand, size)
	} else {
		gen.Type = REFLECT_INDEX
		gen.Index = crdt.GenIndex(rand, size)
	}

	return gen
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

	if err != nil {
		t.Error("Unexpected error:", testutil.Trim(err))
	}
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

	if err != nil {
		t.Error("Unexpected error:", testutil.Trim(err))
	}
}

func apiResponseEncodeOk(resp APIResponse) bool {
	return apiResponseTestWithSerializer(resp, apiResponseSerializationPass)
}

func apiResponseEncodeTextOk(resp APIResponse) bool {
	return apiResponseTestWithSerializer(resp, apiResponseSerializationTextPass)
}

func apiResponseTestWithSerializer(expected APIResponse, pass func(APIResponse) APIResponse) bool {
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

func apiResponseSerializationPass(resp APIResponse) APIResponse {
	buff := &bytes.Buffer{}
	err := EncodeAPIResponse(resp, buff)

	if err != nil {
		panic(err)
	}

	var decoded APIResponse
	decoded, err = DecodeAPIResponse(buff)

	if err != nil {
		panic(err)
	}

	return decoded
}

func apiResponseSerializationTextPass(resp APIResponse) APIResponse {
	buff := &bytes.Buffer{}
	err := EncodeAPIResponseText(resp, buff)

	if err != nil {
		panic(err)
	}

	var decoded APIResponse
	decoded, err = DecodeAPIResponseText(buff)

	if err != nil {
		panic(err)
	}

	return decoded
}
