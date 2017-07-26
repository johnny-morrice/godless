package function

import (
	"testing"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestStandardFunctions(t *testing.T) {
	functions := StandardFunctions()

	_, err := functions.GetFunction(StrEq{}.FuncName())
	testutil.AssertNil(t, err)

	_, err = functions.GetFunction("Foo")
	testutil.AssertNonNil(t, err)
}

func TestFunctionNamespace(t *testing.T) {
	functions := MakeFunctionNamespace()

	err := functions.PutFunction(StrEq{})
	testutil.AssertNil(t, err)

	_, err = functions.GetFunction(StrEq{}.FuncName())
	testutil.AssertNil(t, err)

	_, err = functions.GetFunction("Foo")
	testutil.AssertNonNil(t, err)
}

func TestStrEq(t *testing.T) {
	testMatchFunction(t, StrEq{}, "foo", "foo")
}

func TestStrGlob(t *testing.T) {
	glob := &StrGlob{}
	testMatchFunction(t, glob, "fo*", "foo")
	testMatchFunction(t, glob, "fo*", "foo")
	testMatchFunction(t, glob, "ba*", "bar")
}

func TestStrRegexp(t *testing.T) {
	regexp := &StrRegexp{}
	testMatchFunction(t, regexp, "foox.*", "fooxxxxx")
	testMatchFunction(t, regexp, "foox.*", "fooxxxxx")
	testMatchFunction(t, regexp, "barx.*", "barxxxxx")
}

func testMatchFunction(t *testing.T, function MatchFunction, pattern, text string) {
	const size = 50

	noMatchEntries := make([]crdt.Entry, size)

	for i := 0; i < size; i++ {
		entry := crdt.GenEntry(testutil.Rand(), size)
		noMatchEntries[i] = entry
	}

	matchEntries := []crdt.Entry{
		crdt.MakeEntry([]crdt.Point{
			crdt.UnsignedPoint(crdt.PointText(text)),
		}),
	}

	literals := []string{pattern}

	isMatch := function.Match(literals, noMatchEntries)
	testutil.Assert(t, "Unexpected match", !isMatch)

	isMatch = function.Match(literals, matchEntries)
	testutil.Assert(t, "Expected match on entries", isMatch)
}
