package godless

import (
	"bytes"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/pkg/errors"
)

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

func TestQueryEncode(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(queryEncodeOk, config)

	if err != nil {
		t.Error("Unexpected error:", trim(err))
	}
}

// Generate a *valid* query.
func (query *Query) Generate(rand *rand.Rand, size int) reflect.Value {
	const TABLE_NAME_MAX = 20

	gen := &Query{}
	gen.TableKey = TableName(randStr(rand, __ALPHABET, 1, TABLE_NAME_MAX))

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
	gen := QuerySelect{}
	gen.Limit = rand.Uint32()
	gen.Where = genQueryWhere(rand, size, 1)

	return gen
}

func genQueryWhere(rand *rand.Rand, size int, depth int) QueryWhere {
	const CLAUSE_SCALE = 0.8

	gen := QueryWhere{}
	if rand.Float32()/float32(depth) > 0.8 {
		if rand.Float32() > 0.5 {
			gen.OpCode = AND
		} else {
			gen.OpCode = OR
		}

		clauseCount := genCount(rand, size, CLAUSE_SCALE)
		gen.Clauses = make([]QueryWhere, clauseCount)

		nextDepth := depth + 1
		for i := 0; i < clauseCount; i++ {
			gen.Clauses[i] = genQueryWhere(rand, size, nextDepth)
		}
	} else {
		gen.OpCode = PREDICATE
		gen.Predicate = genQueryPredicate(rand, size)
	}

	return gen
}

func genQueryPredicate(rand *rand.Rand, size int) QueryPredicate {
	const SCALE = 0.5
	const MAX_POINT = 10

	gen := QueryPredicate{}
	if rand.Float32() > 0.5 {
		gen.IncludeRowKey = true
	}

	if rand.Float32() > 0.5 {
		gen.OpCode = STR_EQ
	} else {
		gen.OpCode = STR_NEQ
	}

	keyCount := genCount(rand, size, SCALE)
	litCount := genCount(rand, size, SCALE)
	gen.Keys = make([]EntryName, keyCount)
	gen.Literals = make([]string, litCount)

	for i := 0; i < keyCount; i++ {
		entry := randKey(rand, MAX_POINT)
		gen.Keys[i] = EntryName(entry)
	}

	for i := 0; i < litCount; i++ {
		lit := randPoint(rand, MAX_POINT)
		gen.Literals[i] = lit
	}

	return gen
}

func genQueryJoin(rand *rand.Rand, size int) QueryJoin {
	const ROW_SCALE = 1.0
	const ENTRY_SCALE = 0.2
	const MAX_STR_LEN = 10
	rowCount := genCountRange(rand, 1, size, ROW_SCALE)

	gen := QueryJoin{Rows: make([]QueryRowJoin, rowCount)}

	for i := 0; i < rowCount; i++ {
		gen.Rows[i] = QueryRowJoin{Entries: map[EntryName]Point{}}
		row := &gen.Rows[i]
		row.RowKey = RowName(randKey(rand, MAX_STR_LEN))

		entryCount := genCount(rand, size, ENTRY_SCALE)
		for i := 0; i < entryCount; i++ {
			entry := randKey(rand, MAX_STR_LEN)
			point := randPoint(rand, MAX_STR_LEN)
			row.Entries[EntryName(entry)] = Point(point)
		}
	}

	return gen
}

func randKey(rand *rand.Rand, max int) string {
	return randLetters(rand, max)
}

func randPoint(rand *rand.Rand, max int) string {
	const MIN_POINT_LENGTH = 0
	const pointSyms = __ALPHABET + __DIGITS + SYMBOLS
	const injectScale = 0.1
	point := randStr(rand, pointSyms, MIN_POINT_LENGTH, max)

	if len(point) == 0 {
		return point
	}

	if rand.Float32() > 0.5 {
		position := rand.Intn(len(point))
		inject := randEscape(rand)
		point = insert(point, inject, position)
	}

	return point
}

func insert(old, ins string, pos int) string {
	before := old[:pos]
	after := old[pos:]
	return before + ins + after
}

func randEscape(rand *rand.Rand) string {
	const chars = "\\tnav"
	const MIN_CHARS = 1
	const CHARS_LIM = 2
	return "\\" + randStr(rand, chars, MIN_CHARS, CHARS_LIM)
}

func queryParseOk(expected *Query) bool {
	source := prettyQueryText(expected)
	logdbg("Pretty Printed input: \"%v\"", source)

	actual, err := CompileQuery(source)

	if err != nil {
		panic(errors.Wrap(err, "Parse error"))
	}

	same := expected.Equals(actual)

	if !same {
		actualSource := prettyQueryText(actual)
		logdbg("Pretty Printed output: \"%v\"", actualSource)
		logDiff(source, prettyQueryText(actual))
	}

	return same
}

func logDiff(old, new string) {
	oldParts := strings.Split(old, "")
	newParts := strings.Split(new, "")

	minSize := imin(len(oldParts), len(newParts))

	for i := 0; i < minSize; i++ {
		oldChar := oldParts[i]
		newChar := newParts[i]

		if oldChar != newChar {
			fragmentStart := i - 10
			if fragmentStart < 0 {
				fragmentStart = 0
			}

			fragmentEnd := i + 100

			oldEnd := fragmentEnd
			if len(old) < fragmentEnd {
				oldEnd = len(old) - 1
			}

			newEnd := fragmentEnd
			if len(new) < fragmentEnd {
				newEnd = len(new) - 1
			}

			oldFragment := old[fragmentStart:oldEnd]
			newFragment := new[fragmentStart:newEnd]

			logerr("First difference at %v", i)
			logerr("Old was: '%v'", oldFragment)
			logerr("New was: '%v'", newFragment)
			return
		}
	}

}

func queryEncodeOk(expected *Query) bool {
	actual := querySerializationPass(expected)
	same := expected.Equals(actual)

	if !same {
		expectedSource := prettyQueryText(expected)
		actualSource := prettyQueryText(actual)
		logDiff(expectedSource, actualSource)
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
