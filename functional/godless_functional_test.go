package functional_godless

import (
	"testing"

	"github.com/johnny-morrice/godless"
)

func TestGodlessRequestFunctional(t *testing.T) {
	options := godless.Options{
		RemoteStore: nil,
	}
	godless.New(options)
}

func TestGodlessReplicateFunctional(t *testing.T) {
	// This test checks online replication service.
	t.FailNow()
}
