package query

import (
	"testing"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
)

type placeholderTest struct {
	source   string
	values   []interface{}
	expected *Query
}

func TestPlaceholders(t *testing.T) {
	const tableName = crdt.TableName("The table")
	const rowName = crdt.RowName("The row")
	const entryName = crdt.EntryName("The entry")
	const literal = "The literal"
	const carTable = crdt.TableName("cars")
	const driverEntry = crdt.EntryName("driver")
	const specialFeature = crdt.EntryName("special_feature")
	const fourWheeler = crdt.PointText("4wd")
	const driverName = crdt.PointText("Mr Fast")
	const theLimit uint32 = 5

	placeholderTable := []placeholderTest{
		placeholderTest{
			source: "join ?? rows (@key=test)",
			values: []interface{}{tableName},
			expected: &Query{
				TableKey: tableName,
				OpCode:   JOIN,
				Join: QueryJoin{
					Rows: []QueryRowJoin{
						QueryRowJoin{
							RowKey: crdt.RowName("test"),
						},
					},
				},
			},
		},
		placeholderTest{
			source: "select ??",
			values: []interface{}{tableName},
			expected: &Query{
				TableKey: tableName,
				OpCode:   SELECT,
			},
		},
		placeholderTest{
			source: "join cars rows (@key=??, driver=?, ??=\"4wd\")",
			values: []interface{}{rowName, driverName, specialFeature},
			expected: &Query{
				TableKey: carTable,
				OpCode:   JOIN,
				Join: QueryJoin{
					Rows: []QueryRowJoin{
						QueryRowJoin{
							RowKey: rowName,
							Entries: map[crdt.EntryName]crdt.PointText{
								driverEntry:    driverName,
								specialFeature: fourWheeler,
							},
						},
					},
				},
			},
		},
		placeholderTest{
			source: "select cars where and(str_eq(??, ?), str_glob(??, ?)) limit ?",
			values: []interface{}{specialFeature, fourWheeler, driverEntry, driverName, theLimit},
			expected: &Query{
				TableKey: carTable,
				OpCode:   SELECT,
				Select: QuerySelect{
					Limit: theLimit,
					Where: QueryWhere{
						OpCode: AND,
						Clauses: []QueryWhere{
							QueryWhere{
								OpCode: PREDICATE,
								Predicate: QueryPredicate{
									FunctionName: "str_eq",
									Values:       []PredicateValue{PredicateKey(specialFeature), PredicateLiteral(fourWheeler)},
								},
							},
							QueryWhere{
								OpCode: PREDICATE,
								Predicate: QueryPredicate{
									FunctionName: "str_glob",
									Values:       []PredicateValue{PredicateKey(driverEntry), PredicateLiteral(driverName)},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range placeholderTable {
		actual, err := Compile(test.source, test.values...)

		testutil.AssertNil(t, err)
		testutil.Assert(t, "Unexpected query", test.expected.Equals(actual))
	}
}

func TestInvalidPlaceholder(t *testing.T) {
	const tableName = crdt.TableName("The table")
	const driverName = crdt.PointText("Mr Fast")
	const fourWheeler = crdt.PointText("4wd")
	const theLimit string = "5"

	placeholderTable := []placeholderTest{
		placeholderTest{
			source: "join ?? rows (@key=test,driver=?)",
			values: []interface{}{tableName},
		},
		placeholderTest{
			source: "join ?? rows (@key=test,driver=?)",
			values: []interface{}{tableName, driverName, fourWheeler},
		},
		placeholderTest{
			source: "select cars limit ?",
			values: []interface{}{theLimit},
		},
	}

	for _, test := range placeholderTable {
		actual, err := Compile(test.source, test.values)

		testutil.AssertNonNil(t, err)
		testutil.AssertNil(t, actual)
	}
}
