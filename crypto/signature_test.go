package crypto

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/internal/testutil"
)

// Worth stating here that testutil.Rand() provides a non-cryptographically
// safe random number generator.  This is fine for generating test data but
// should not be used for crypto.

var alicePub PublicKey
var malloryPub PublicKey

var alicePriv PrivateKey
var malloryPriv PrivateKey

func init() {
	var err error
	alicePriv, alicePub, err = GenerateKey()

	setupPanic(err)

	malloryPriv, malloryPub, err = GenerateKey()

	setupPanic(err)
}

var plainText []byte = []byte("Hello world")

func TestAlice(t *testing.T) {
	sig, signErr := Sign(alicePriv, plainText)
	testutil.AssertNil(t, signErr)
	ok, verifyErr := Verify(alicePub, plainText, sig)
	testutil.AssertNil(t, verifyErr)
	testutil.Assert(t, "Unexpected signature failure", ok)
}

func TestMallory(t *testing.T) {
	sig, signErr := Sign(alicePriv, plainText)
	testutil.AssertNil(t, signErr)
	ok, verifyErr := Verify(malloryPub, plainText, sig)
	testutil.AssertNil(t, verifyErr)
	testutil.Assert(t, "Expected signature failure", !ok)
}

func TestSignNil(t *testing.T) {
	message := []byte("hello")
	sig, err := Sign(PrivateKey{}, message)
	testutil.AssertNonNil(t, err)
	testutil.Assert(t, "Expected zero value", Signature{}.Equals(sig))
}

func TestVerifyNil(t *testing.T) {
	message := []byte("hello")
	priv, pub, err := GenerateKey()
	realSig, err := Sign(priv, message)

	testutil.AssertNil(t, err)
	testutil.Assert(t, "Unexpected empty signature", len(realSig.sig) > 0)

	ok, err := Verify(PublicKey{}, message, realSig)

	testutil.Assert(t, "Unexpected verification", !ok)
	testutil.AssertNonNil(t, err)

	ok, err = Verify(pub, message, Signature{})

	testutil.Assert(t, "Unexpected verification", !ok)
	testutil.AssertNonNil(t, err)
}

func TestIsNilSignature(t *testing.T) {
	const nilSigText = ""
	text := []byte("hello")

	priv, _, err := GenerateKey()
	setupPanic(err)
	sig, err := Sign(priv, text)
	setupPanic(err)
	nonNilSigText, err := PrintSignature(sig)
	setupPanic(err)

	testutil.Assert(t, "Expected IsNil Signature", IsNilSignature(nilSigText))
	testutil.Assert(t, "Unexpected IsNil Signature", !IsNilSignature(nonNilSigText))
}

func TestParsePrintSignature(t *testing.T) {
	const count = 10
	const maxSignage = 3
	const invalidProbability = 1 / 8

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(testParseSignatureOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func TestOrderSignatures(t *testing.T) {
	const size = 50
	sigs := genManySigs(testutil.Rand(), size)
	unorderedSignatures := append(sigs, sigs...)

	ordered := OrderSignatures(unorderedSignatures)

	testutil.AssertLenEquals(t, len(sigs), ordered)

	for i := 1; i < len(ordered); i++ {
		last := ordered[i-1]
		current := ordered[i]
		cmp := last.Cmp(current)
		isLessThan := cmp == -1
		testutil.Assert(t, "Unexpected signature order", isLessThan)
	}

MY_SIGS:
	for _, mySig := range sigs {
		for _, otherSig := range ordered {
			if mySig.Equals(otherSig) {
				continue MY_SIGS
			}
		}

		t.Error("Signature not found in ordered")
		t.FailNow()
	}
}

func TestSignatureEquals(t *testing.T) {
	const size = 50
	sigs := genManySigs(testutil.Rand(), size)

	for i := 0; i < size; i++ {
		mySig := sigs[i]
		for j := 0; j < size; j++ {
			otherSig := sigs[j]
			shouldBeEquals := i == j
			isEquals := mySig.Equals(otherSig)
			msg := "Expected equality"
			if !shouldBeEquals {
				msg = "Unexpected equality"
			}
			testutil.Assert(t, msg, shouldBeEquals == isEquals)
		}
	}
}

func TestSignatureCmp(t *testing.T) {
	const size = 50
	sigs := genManySigs(testutil.Rand(), size)

	for i := 0; i < size; i++ {
		mySig := sigs[i]
		for j := 0; j < size; j++ {
			otherSig := sigs[j]
			shouldBeEquals := i == j

			cmp := mySig.Cmp(otherSig)
			isEquals := cmp == 0
			isValidCmp := cmp == -1 || cmp == 0 || cmp == 1

			msg := "Expected equality"
			if !shouldBeEquals {
				msg = "Unexpected equality"
			}

			testutil.Assert(t, msg, shouldBeEquals == isEquals)
			testutil.Assert(t, "Unexpected Cmp return", isValidCmp)
		}
	}
}

func testParseSignatureOk(inputSig Signature) bool {
	sigText, err := PrintSignature(inputSig)

	if err != nil {
		return false
	}

	outputSig, err := ParseSignature(sigText)

	if err != nil {
		return false
	}

	return inputSig.Equals(outputSig)
}

func setupPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func genManySigs(rand *rand.Rand, size int) []Signature {
	sigs := make([]Signature, size)

	for i := 0; i < size; i++ {
		sigs[i] = genSignature(rand)
	}

	return sigs
}

func genSignature(rand *rand.Rand) Signature {
	text := []byte(__GEN_SIG_TEXT)
	priv, _, err := GenerateKey()
	setupPanic(err)
	sig, err := Sign(priv, text)
	setupPanic(err)
	return sig
}

func (sig Signature) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := genSignature(rand)
	return reflect.ValueOf(gen)
}

const __GEN_SIG_TEXT = "Hello World"
