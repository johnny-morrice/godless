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
	const ROW_SCALE = 1.0
	const ENTRY_SCALE = 0.2
	const MAX_STR_LEN = 10
	rowCount := genCount(rand, size, ROW_SCALE)

	gen := QueryJoin{Rows: make([]QueryRowJoin, rowCount)}

	for i := 0; i < rowCount; i++ {
		gen.Rows[i] = QueryRowJoin{Entries: map[EntryName]Point{}}
		row := &gen.Rows[i]
		row.RowKey = RowName(randStr(rand, MAX_STR_LEN))

		entryCount := genCount(rand, size, ENTRY_SCALE)
		for i := 0; i < entryCount; i++ {
			entry := randStr(rand, MAX_STR_LEN)
			point := randStr(rand, MAX_STR_LEN)
			row.Entries[EntryName(entry)] = Point(point)
		}
	}

	return gen
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
