package util

import (
	"fmt"
	"strconv"
	"testing"
	"testing/quick"
)

func TestEncodeBase58(t *testing.T) {
	err := quick.Check(encodesOk, nil)
	if err != nil {
		t.Error(err)
	}
}

func encodesOk(expected []byte) bool {
	encoded := EncodeBase58(expected)
	actual := DecodeBase58(encoded)

	if len(expected) > 0 && len(encoded) == 0 {
		return false
	}

	if len(expected) != len(actual) {
		return false
	}

	for i, ex := range expected {
		ac := actual[i]

		if ex != ac {
			return false
		}
	}

	return true
}

func BenchmarkBase58(b *testing.B) {
	for i := 0; i < b.N; i++ {
		text := strconv.Itoa(i)
		data := []byte(text)
		encoded := EncodeBase58(data)
		decoded := DecodeBase58(encoded)

		if len(data) != len(decoded) {
			msg := fmt.Sprintf("Expected: '%v' (len %d) but received: '%v' (len %d)", string(data), len(data), string(decoded), len(encoded))
			panic(msg)
		}
	}
}
