package godless

import (
	"bytes"
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func (resp APIResponse) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := APIResponse{}

	text := randLetters(rand, size)
	if rand.Float32() < 0.5 {
		gen.Msg = text
	} else {
		gen.Err = errors.New(text)
	}

	if rand.Float32() < 0.5 {
		gen.Type = API_QUERY
		if gen.Err != nil {
			gen.QueryResponse = genQueryResponse(rand, size)
		}
	} else {
		gen.Type = API_REFLECT
		if gen.Err != nil {
			gen.ReflectResponse = genReflectResponse(rand, size)
		}
	}

	return reflect.ValueOf(gen)
}

func genQueryResponse(rand *rand.Rand, size int) APIQueryResponse {
	gen := APIQueryResponse{}
	ns := genNamespace(rand, size)
	stream := MakeNamespaceStream(ns)
	gen.Rows = stream
	return gen
}

func genReflectResponse(rand *rand.Rand, size int) APIReflectResponse {
	gen := APIReflectResponse{}

	branch := rand.Float32()
	if branch < 0.333 {
		gen.Type = REFLECT_HEAD_PATH
		gen.Path = randLetters(rand, size)
	} else if branch < 0.666 {
		gen.Type = REFLECT_DUMP_NAMESPACE
		gen.Namespace = genNamespace(rand, size)
	} else {
		gen.Type = REFLECT_INDEX
		gen.Index = genIndex(rand, size)
	}

	return gen
}

func TestEncodeAPIResponse(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(apiResponseEncodeOk, config)

	if err != nil {
		t.Error("Unexpected error:", trim(err))
	}
}

func TestEncodeAPIResponseText(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(apiResponseEncodeTextOk, config)

	if err != nil {
		t.Error("Unexpected error:", trim(err))
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
	return expected.Equals(actual)
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
