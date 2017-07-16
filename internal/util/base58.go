package util

import (
	"bytes"
	"math/big"
	"strings"
)

// Borrowed from jbenet/go-base58.

func EncodeBase58(b []byte) string {
	return encodeAlphabet(b, __BTC_ALPHABET)
}

func DecodeBase58(b string) []byte {
	return decodeAlphabet(b, __BTC_ALPHABET)
}

func IsBase58(b string) bool {
	parts := strings.Split(b, "")
	for _, p := range parts {
		partOk := strings.Contains(__BTC_ALPHABET_TEXT, p)
		if !partOk {
			return false
		}
	}

	return true
}

// decodeAlphabet decodes a modified base58 string to a byte slice, using alphabet.
func decodeAlphabet(b string, alphabet []byte) []byte {
	answer := big.NewInt(0)
	j := big.NewInt(1)
	idx := big.NewInt(0)

	for i := len(b) - 1; i >= 0; i-- {
		tmp := bytes.IndexByte(alphabet, b[i])
		if tmp == -1 {
			return []byte("")
		}
		idx.SetInt64(int64(tmp))
		idx.Mul(idx, j)

		answer.Add(answer, idx)
		j.Mul(j, bigRadix)
	}

	tmpval := answer.Bytes()

	var numZeros int
	for numZeros = 0; numZeros < len(b); numZeros++ {
		if b[numZeros] != alphabet[0] {
			break
		}
	}
	flen := numZeros + len(tmpval)
	val := make([]byte, flen, flen)
	copy(val[numZeros:], tmpval)

	return val
}

// encodeAlphabet encodes a byte slice to a modified base58 string, using alphabet
func encodeAlphabet(b []byte, alphabet []byte) string {
	x := new(big.Int)
	x.SetBytes(b)

	mod := new(big.Int)
	answer := make([]byte, 0, len(b)*136/100)
	for x.Cmp(bigZero) > 0 {
		x.DivMod(x, bigRadix, mod)
		answer = append(answer, alphabet[mod.Int64()])
	}

	// leading zero bytes
	for _, i := range b {
		if i != 0 {
			break
		}
		answer = append(answer, alphabet[0])
	}

	// reverse
	alen := len(answer)
	for i := 0; i < alen/2; i++ {
		answer[i], answer[alen-1-i] = answer[alen-1-i], answer[i]
	}

	return string(answer)
}

// alphabet is the modified base58 alphabet used by Bitcoin.
const __BTC_ALPHABET_TEXT = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var __BTC_ALPHABET = []byte(__BTC_ALPHABET_TEXT)

var bigRadix = big.NewInt(58)
var bigZero = big.NewInt(0)
