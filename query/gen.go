package query

import (
	"math/rand"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func GenQuery(rand *rand.Rand, size int) *Query {
	const TABLE_NAME_MAX = 20

	gen := &Query{}
	gen.TableKey = crdt.TableName(testutil.RandStr(rand, __ALPHABET, 1, TABLE_NAME_MAX))

	if rand.Float32() > 0.5 {
		gen.OpCode = SELECT
		gen.Select = genQuerySelect(rand, size)
	} else {
		gen.OpCode = JOIN
		gen.Join = genQueryJoin(rand, size)
	}

	return gen
}

func genQuerySelect(rand *rand.Rand, size int) QuerySelect {
	gen := QuerySelect{}
	limit := rand.Intn(__GEN_QUERY_LIMIT)
	gen.Limit = uint32(limit)
	gen.Where = genQueryWhere(rand, size, 1)

	return gen
}

func genQueryWhere(rand *rand.Rand, size int, depth int) QueryWhere {
	gen := QueryWhere{}

	if size > 10 {
		size = 10
	}

	if rand.Float32()/float32(depth) > 0.4 {
		if rand.Float32() > 0.5 {
			gen.OpCode = AND
		} else {
			gen.OpCode = OR
		}

		clauseCount := testutil.GenCountRange(rand, 1, size)
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
	const MAX_POINT = 10

	gen := QueryPredicate{}
	if rand.Float32() > 0.5 {
		gen.IncludeRowKey = true
	}

	gen.FunctionName = "str_eq"

	keyCount := testutil.GenCountRange(rand, 1, size)
	litCount := testutil.GenCountRange(rand, 1, size)
	gen.Keys = make([]crdt.EntryName, keyCount)
	gen.Literals = make([]string, litCount)

	for i := 0; i < keyCount; i++ {
		entry := testutil.RandKey(rand, MAX_POINT)
		gen.Keys[i] = crdt.EntryName(entry)
	}

	for i := 0; i < litCount; i++ {
		lit := testutil.RandPoint(rand, MAX_POINT)
		gen.Literals[i] = lit
	}

	return gen
}

func genQueryJoin(rand *rand.Rand, size int) QueryJoin {
	const MAX_STR_LEN = 10

	if size > 10 {
		size = 10
	}

	rowCount := testutil.GenCountRange(rand, 1, size)

	gen := QueryJoin{Rows: make([]QueryRowJoin, rowCount)}

	for i := 0; i < rowCount; i++ {
		gen.Rows[i] = QueryRowJoin{Entries: map[crdt.EntryName]crdt.PointText{}}
		row := &gen.Rows[i]
		row.RowKey = crdt.RowName(testutil.RandKey(rand, MAX_STR_LEN))

		entryCount := testutil.GenCount(rand, size)
		for i := 0; i < entryCount; i++ {
			entry := testutil.RandKey(rand, MAX_STR_LEN)
			point := testutil.RandPoint(rand, MAX_STR_LEN)
			row.Entries[crdt.EntryName(entry)] = crdt.PointText(point)
		}
	}

	return gen
}

const __ALPHABET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const __DIGITS = "0123456789"

const __KEY_SYMS = __ALPHABET + __DIGITS
const __GEN_QUERY_LIMIT = 1000
