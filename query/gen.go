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
	gen.Limit = rand.Uint32()
	gen.Where = genQueryWhere(rand, size, 1)

	return gen
}

func genQueryWhere(rand *rand.Rand, size int, depth int) QueryWhere {
	const CLAUSE_SCALE = 0.8

	gen := QueryWhere{}
	if rand.Float32()/float32(depth) > 0.4 {
		if rand.Float32() > 0.5 {
			gen.OpCode = AND
		} else {
			gen.OpCode = OR
		}

		clauseCount := testutil.GenCount(rand, size, CLAUSE_SCALE)
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

	keyCount := testutil.GenCount(rand, size, SCALE)
	litCount := testutil.GenCount(rand, size, SCALE)
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
	const ROW_SCALE = 1.0
	const ENTRY_SCALE = 0.2
	const MAX_STR_LEN = 10
	rowCount := testutil.GenCountRange(rand, 1, size, ROW_SCALE)

	gen := QueryJoin{Rows: make([]QueryRowJoin, rowCount)}

	for i := 0; i < rowCount; i++ {
		gen.Rows[i] = QueryRowJoin{Entries: map[crdt.EntryName]crdt.Point{}}
		row := &gen.Rows[i]
		row.RowKey = crdt.RowName(testutil.RandKey(rand, MAX_STR_LEN))

		entryCount := testutil.GenCount(rand, size, ENTRY_SCALE)
		for i := 0; i < entryCount; i++ {
			entry := testutil.RandKey(rand, MAX_STR_LEN)
			point := testutil.RandPoint(rand, MAX_STR_LEN)
			row.Entries[crdt.EntryName(entry)] = crdt.Point(point)
		}
	}

	return gen
}

const __ALPHABET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const __DIGITS = "0123456789"

const __KEY_SYMS = __ALPHABET + __DIGITS
