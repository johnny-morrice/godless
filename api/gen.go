package api

import (
	"math/rand"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func GenResponse(rand *rand.Rand, size int) Response {
	gen := Response{}

	gen.Msg = testutil.RandLetters(rand, size)

	if rand.Float32() < 0.5 {
		gen.Type = API_QUERY
	} else {
		gen.Type = API_REFLECT
	}

	if rand.Float32() < 0.5 {
		if gen.Type == API_QUERY {
			genQueryResponse(rand, size, &gen)
		} else {
			genReflectResponse(rand, size, &gen)
		}
	} else {
		errText := testutil.RandLettersRange(rand, 1, size)
		gen.Err = errors.New(errText)
	}

	return gen
}

func genQueryResponse(rand *rand.Rand, size int, gen *Response) {
	ns := crdt.GenNamespace(rand, size)
	gen.Namespace = ns
	gen.Path = genResponsePath(rand, size)
}

func genReflectResponse(rand *rand.Rand, size int, gen *Response) {
	branch := rand.Float32()
	if branch < 0.333 {
		gen.Path = genResponsePath(rand, size)
	} else if branch < 0.666 {
		gen.Namespace = crdt.GenNamespace(rand, size)
	} else {
		gen.Index = crdt.GenIndex(rand, size)
	}
}

func genResponsePath(rand *rand.Rand, size int) crdt.IPFSPath {
	return crdt.IPFSPath(testutil.RandLettersRange(rand, 1, size))
}
