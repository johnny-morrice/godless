package godless

import (
	"bytes"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/pkg/errors"
)

// Generate a *valid* query.
func (query *Query) Generate(rand *rand.Rand, size int) reflect.Value {
	const TABLE_NAME_MAX = 20

	gen := &Query{}
	gen.TableKey = TableName(randStr(rand, TABLE_NAME_MAX))

	if rand.Float32() > 0.5 {
		gen.OpCode = SELECT
		gen.Select = genQuerySelect(rand, size)
	} else {
		gen.OpCode = JOIN
		gen.Join = genQueryJoin(rand, size)
	}

	return reflect.ValueOf(gen)
}

func genQuerySelect(rand *rand.Rand, size int) QuerySelect {
	return QuerySelect{}
}

func genQueryJoin(rand *rand.Rand, size int) QueryJoin {
	return QueryJoin{}
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
		t.Error("Unexpected error:", trim(err))
	}
}

func queryParseOk(expected *Query) bool {
	source := prettyQueryString(expected)
	logdbg("Pretty Printed: \"%v\"", source)

	actual, err := CompileQuery(source)

	if err != nil {
		panic(errors.Wrap(err, "Parse error"))
	}

	return expected.Equals(actual)
}

func prettyQueryString(query *Query) string {
	buff := &bytes.Buffer{}
	err := query.PrettyPrint(buff)

	if err != nil {
		panic(err)
	}

	return buff.String()
}

func queryEncodeOk(expected *Query) bool {
	actual := querySerializationPass(expected)
	return expected.Equals(actual)
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
