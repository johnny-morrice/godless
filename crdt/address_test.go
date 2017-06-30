package crdt

import (
	"testing"

	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestSignedLink(t *testing.T) {
	const keyCount = 5
	const text = "hello"
	keys := generateTestKeys(keyCount)
	duplicateKeys := append(keys, keys...)
	badKeys := append(keys, crypto.PrivateKey{})

	signed, err := SignedLink(text, keys)

	testutil.AssertNil(t, err)
	testutil.AssertLenEquals(t, len(keys), signed.Signatures())
	testutil.AssertEquals(t, "Unexpected text", signed.Path(), text)

	signed, err = SignedLink(text, duplicateKeys)

	testutil.AssertNil(t, err)
	testutil.AssertLenEquals(t, len(keys), signed.Signatures())
	testutil.AssertEquals(t, "Unexpected text", signed.Path(), text)

	signed, err = SignedLink(text, badKeys)

	testutil.AssertNonNil(t, err)
	testutil.AssertLenEquals(t, 0, signed.Signatures())
	testutil.Assert(t, "Expected empty Path for failed SignedLink", signed.Path() == "")
}

func TestLinkIsVerifiedBy(t *testing.T) {
	const keyCount = 5
	const text = "hello"
	myPriv, myPub, err := crypto.GenerateKey()

	if err != nil {
		panic(err)
	}

	notMyKeys := generateTestKeys(keyCount)

	point, err := SignedLink(text, []crypto.PrivateKey{myPriv})

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

func TestLinkIsVerifiedByAny(t *testing.T) {
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

	point, err := SignedLink(text, []crypto.PrivateKey{myPriv})

	if err != nil {
		panic(err)
	}

	isVerified := point.IsVerifiedByAny(allPubKeys)

	testutil.Assert(t, "Expected verification", isVerified)

	isVerified = point.IsVerifiedByAny(notMyPubKeys)

	testutil.Assert(t, "Unexpected verification", !isVerified)
}
