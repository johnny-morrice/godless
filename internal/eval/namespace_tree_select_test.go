package eval

import (
	"reflect"
	"testing"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/query"
)

// TODO implement test.
// func TestRowCriteria_selectMatching(t *testing.T) {
// 	t.Fail()
// }

// The various where predicates are tested elsewhere.  This test focusses
// on whether the correct rows will be discovered for any predicate.
func TestRowCriteria_findRows(t *testing.T) {
	pointA := crdt.UnsignedPoint("hello")
	pointB := crdt.UnsignedPoint("world")

	rowA := crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
		"foo": crdt.MakeEntry([]crdt.Point{pointA}),
	})
	rowB := crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
		"bar": crdt.MakeEntry([]crdt.Point{pointB}),
	})

	streamPointA := makeStreamPoint(pointA.Text, crypto.Signature{})
	streamPointB := makeStreamPoint(pointB.Text, crypto.Signature{})

	namespace := crdt.MakeNamespace(map[crdt.TableName]crdt.Table{
		TABLE_KEY: crdt.MakeTable(map[crdt.RowName]crdt.Row{
			"a": rowA,
			"b": rowB,
			"c": rowB,
		}),
	})

	streamEntryA := crdt.NamespaceStreamEntry{
		Table: TABLE_KEY,
		Row:   "a",
		Entry: "foo",
		Point: streamPointA,
	}

	streamEntryB := crdt.NamespaceStreamEntry{
		Table: TABLE_KEY,
		Row:   "b",
		Entry: "bar",
		Point: streamPointB,
	}

	streamEntryC := crdt.NamespaceStreamEntry{
		Table: TABLE_KEY,
		Row:   "c",
		Entry: "bar",
		Point: streamPointB,
	}

	expected := [][]crdt.NamespaceStreamEntry{
		[]crdt.NamespaceStreamEntry{
			streamEntryA,
		},
		[]crdt.NamespaceStreamEntry{
			streamEntryB, streamEntryC,
		},
	}

	where := []query.QueryWhere{
		query.QueryWhere{
			OpCode: query.PREDICATE,
			Predicate: query.QueryPredicate{
				OpCode:        query.STR_EQ,
				Literals:      []string{"a"},
				IncludeRowKey: true,
			},
		},
		query.QueryWhere{
			OpCode: query.PREDICATE,
			Predicate: query.QueryPredicate{
				OpCode:   query.STR_EQ,
				Literals: []string{"world"},
				Keys:     []crdt.EntryName{"bar"},
			},
		},
	}

	for i, e := range expected {
		w := where[i]
		rc := &rowCriteria{
			tableKey:  TABLE_KEY,
			limit:     10,
			rootWhere: &w,
		}

		actual, invalid := rc.findRows(namespace)

		testutil.AssertLenEquals(t, 0, invalid)

		crdt.SortNamespaceStream(actual)

		if !reflect.DeepEqual(e, actual) {
			t.Error(i, "Expected", e, "but was", actual)
		}
	}
}

func TestRowCriteria_isReady(t *testing.T) {
	bad := []*rowCriteria{
		&rowCriteria{},
		&rowCriteria{rootWhere: &query.QueryWhere{}},
		&rowCriteria{tableKey: TABLE_KEY},
	}

	for i, b := range bad {
		if b.isReady() {
			t.Error("Unexpected rowCriteria isReady() at", i)
		}
	}

	okay := &rowCriteria{}
	okay.limit = 10
	okay.rootWhere = &query.QueryWhere{}
	okay.tableKey = TABLE_KEY

	if !okay.isReady() {
		t.Error("Expected rowCriteria isReady()")
	}
}

func makeStreamPoint(text crdt.PointText, sig crypto.Signature) crdt.StreamPoint {
	streamPoint, err := crdt.MakeStreamPoint(text, sig)
	setupPanic(err)
	return streamPoint
}

func setupPanic(err error) {
	if err != nil {
		panic(err)
	}
}

const TABLE_KEY = "The Table"
