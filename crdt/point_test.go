package crdt

import (
	"testing"

	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestSignedPoint(t *testing.T) {
	const keyCount = 5
	const text = "hello"
	keys := generateTestKeys(keyCount)
	duplicateKeys := append(keys, keys...)
	badKeys := append(keys, crypto.PrivateKey{})

	signed, err := SignedPoint(text, keys)

	testutil.AssertNil(t, err)
	testutil.AssertLenEquals(t, len(keys), signed.Signatures)
	testutil.AssertEquals(t, "Unexpected text", signed.Text, text)

	signed, err = SignedPoint(text, duplicateKeys)

	testutil.AssertNil(t, err)
	testutil.AssertLenEquals(t, len(keys), signed.Signatures)
	testutil.AssertEquals(t, "Unexpected text", signed.Text, text)

	signed, err = SignedPoint(text, badKeys)

	testutil.AssertNonNil(t, err)
	testutil.AssertLenEquals(t, 0, signed.Signatures)
	testutil.Assert(t, "Expected empty text for failed SignedPoint", signed.Text == "")
}

func TestPointIsVerifiedBy(t *testing.T) {
	t.FailNow()
}

func TestPointIsVerifiedByAny(t *testing.T) {
	t.FailNow()
}
