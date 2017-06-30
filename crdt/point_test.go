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
	testutil.Assert(t, "Expected empty text for failed SignedPoint", signed.Text() == "")
}

func TestPointIsVerifiedBy(t *testing.T) {
	const keyCount = 5
	const text = "hello"
	myPriv, myPub, err := crypto.GenerateKey()

	if err != nil {
		panic(err)
	}

	notMyKeys := generateTestKeys(keyCount)

	point, err := SignedPoint(text, []crypto.PrivateKey{myPriv})

	if err != nil {
		panic(err)
	}

	isVerified := point.IsVerifiedBy(myPub)

	testutil.Assert(t, "Expected verification", isVerified)

	for _, notMyPriv := range notMyKeys {
		isVerified = point.IsVerifiedBy(notMyPriv.GetPublicKey())
		testutil.Assert(t, "Unexpected verification", !isVerified)
	}
}

func TestPointIsVerifiedByAny(t *testing.T) {
	const keyCount = 5
	const text = "hello"
	myPriv, myPub, err := crypto.GenerateKey()

	if err != nil {
		panic(err)
	}

	notMyKeys := generateTestKeys(keyCount)
	notMyPubKeys := make([]crypto.PublicKey, len(notMyKeys))

	for i, notMyKey := range notMyKeys {
		notMyPubKeys[i] = notMyKey.GetPublicKey()
	}

	allPubKeys := append(notMyPubKeys, myPub)

	point, err := SignedPoint(text, []crypto.PrivateKey{myPriv})

	if err != nil {
		panic(err)
	}

	isVerified := point.IsVerifiedByAny(allPubKeys)

	testutil.Assert(t, "Expected verification", isVerified)

	isVerified = point.IsVerifiedByAny(notMyPubKeys)

	testutil.Assert(t, "Unexpected verification", !isVerified)
}
