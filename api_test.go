package godless

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

func (resp APIResponse) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := APIResponse{}

	if rand.Float32() < 0.5 {

	}

	return reflect.ValueOf(gen)
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
