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
	}

	for _, test := range placeholderTable {
		actual, err := Compile(test.source, test.values...)

		testutil.AssertNil(t, err)
		testutil.Assert(t, "Unexpected query", test.expected.Equals(actual))
	}
}

// func TestJoinRowKeyPlaceholder(t *testing.T) {
// 	t.FailNow()
// }
//
// func TestJoinValuePlaceholder(t *testing.T) {
// 	t.FailNow()
// }
//
// func TestJoinKeyPlaceholder(t *testing.T) {
// 	t.FailNow()
// }
//
// func TestLimitPlaceholder(t *testing.T) {
// 	t.FailNow()
// }
//
// func TestPredicateKeyPlaceholder(t *testing.T) {
// 	t.FailNow()
// }
//
// func TestPredicateLiteralPlaceholder(t *testing.T) {
// 	t.FailNow()
// }

func TestInvalidPlaceholder(t *testing.T) {
	// Count not match
	// Type not match
	t.FailNow()
}

func TestLongTextPlaceholders(t *testing.T) {
	t.FailNow()
}
