package util

import (
	"math/big"
	"strings"
)

// Borrowed from jbenet/go-base58.

func EncodeBase58(b []byte) string {
	return encodeAlphabet(b, __BTCAlphabet)
}

func DecodeBase58(b string) []byte {
	return decodeAlphabet(b, __BTCAlphabet)
}

func IsBase58(b string) bool {
	parts := strings.Split(b, "")
	for _, p := range parts {
		partOk := strings.Contains(__BTCAlphabet, p)
		if !partOk {
			return false
		}
	}

	return true
}

// decodeAlphabet decodes a modified base58 string to a byte slice, using alphabet.
func decodeAlphabet(b, alphabet string) []byte {
	answer := big.NewInt(0)
	j := big.NewInt(1)

	for i := len(b) - 1; i >= 0; i-- {
		tmp := strings.IndexAny(alphabet, string(b[i]))
		if tmp == -1 {
			return []byte("")
		}
		idx := big.NewInt(int64(tmp))
		tmp1 := big.NewInt(0)
		tmp1.Mul(j, idx)

		answer.Add(answer, tmp1)
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
func encodeAlphabet(b []byte, alphabet string) string {
	x := new(big.Int)
	x.SetBytes(b)

	answer := make([]byte, 0, len(b)*136/100)
	for x.Cmp(bigZero) > 0 {
		mod := new(big.Int)
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
const __BTCAlphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var bigRadix = big.NewInt(58)
var bigZero = big.NewInt(0)
