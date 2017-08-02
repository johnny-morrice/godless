package query

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

func init() {
	log.SetLevel(log.LOG_DEBUG)
}

func TestParseQuery(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: PARSE_REPEAT_COUNT,
	}

	err := quick.Check(queryParseOk, config)

	if err != nil {
		t.Error("Unexpected error:", testutil.Trim(err))
	}
}

func TestQueryEncode(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(queryEncodeOk, config)

	if err != nil {
		t.Error("Unexpected error:", testutil.Trim(err))
	}
}

// Generate a *valid* query.
func (query *Query) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := GenQuery(rand, size)

	return reflect.ValueOf(gen)
}

func prettyQuery(query *Query) string {
	text, err := query.PrettyText()
	if err != nil {
		panic(err)
	}
	return text
}

func queryParseOk(expected *Query) bool {
	source := prettyQuery(expected)
	log.Debug("Pretty Printed input: \"%s\"", source)

	actual, err := Compile(source)

	if err != nil {
		panic(errors.Wrap(err, "Parse error"))
	}

	same := expected.Equals(actual)

	if !same {
		actualSource := prettyQuery(actual)
		log.Debug("Pretty Printed output: \"%s\"", actualSource)
		testutil.LogDiff(source, prettyQuery(actual))
	}

	return same
}

func queryEncodeOk(expected *Query) bool {
	actual := querySerializationPass(expected)
	same := expected.Equals(actual)

	if !same {
		expectedSource := prettyQuery(expected)
		actualSource := prettyQuery(actual)
		testutil.LogDiff(expectedSource, actualSource)
	}

	return same
}

func querySerializationPass(expected *Query) *Query {
	buff := &bytes.Buffer{}
	err := EncodeQuery(expected, buff)

	if err != nil {
		panic(err)
	}

	var actual *Query
	actual, err = DecodeQuery(buff)

	if err != nil {
		panic(err)
	}

	return actual
}

const PARSE_REPEAT_COUNT = 50
