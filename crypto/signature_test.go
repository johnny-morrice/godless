package crypto

import (
	"testing"

	"github.com/johnny-morrice/godless/internal/testutil"
)

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

func setupPanic(err error) {
	if err != nil {
		panic(err)
	}
}
