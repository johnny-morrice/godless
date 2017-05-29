package godless

import (
	"reflect"
	"testing"
)

// TODO implement test.
// func TestRowCriteria_selectMatching(t *testing.T) {
// 	t.Fail()
// }

// The various where predicates are tested elsewhere.  This test focusses
// on whether the correct rows will be discovered for any predicate.
func TestRowCriteria_findRows(t *testing.T) {
	pointA := []Point{"hello"}
	pointB := []Point{"world"}

	rowA := MakeRow(map[EntryName]Entry{
		"foo": MakeEntry(pointA),
	})
	rowB := MakeRow(map[EntryName]Entry{
		"bar": MakeEntry(pointB),
	})

	namespace := MakeNamespace(map[TableName]Table{
		TABLE_KEY: MakeTable(map[RowName]Row{
			"a": rowA,
			"b": rowB,
			"c": rowB,
		}),
	})

	streamEntryA := NamespaceStreamEntry{
		Table:  TABLE_KEY,
		Row:    "a",
		Entry:  "foo",
		Points: pointA,
	}

	streamEntryB := NamespaceStreamEntry{
		Table:  TABLE_KEY,
		Row:    "b",
		Entry:  "bar",
		Points: pointB,
	}

	streamEntryC := NamespaceStreamEntry{
		Table:  TABLE_KEY,
		Row:    "c",
		Entry:  "bar",
		Points: pointB,
	}

	expected := [][]NamespaceStreamEntry{
		[]NamespaceStreamEntry{
			streamEntryA,
		},
		[]NamespaceStreamEntry{
			streamEntryB, streamEntryC,
		},
	}

	where := []QueryWhere{
		QueryWhere{
			OpCode: PREDICATE,
			Predicate: QueryPredicate{
				OpCode:        STR_EQ,
				Literals:      []string{"a"},
				IncludeRowKey: true,
			},
		},
		QueryWhere{
			OpCode: PREDICATE,
			Predicate: QueryPredicate{
				OpCode:   STR_EQ,
				Literals: []string{"world"},
				Keys:     []EntryName{"bar"},
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

		actual := rc.findRows(namespace)

		if !reflect.DeepEqual(e, actual) {
			t.Error(i, "Expected", e, "but was", actual)
		}
	}
}

func TestRowCriteria_isReady(t *testing.T) {
	bad := []*rowCriteria{
		&rowCriteria{},
		&rowCriteria{limit: 10},
		&rowCriteria{rootWhere: &QueryWhere{}},
		&rowCriteria{tableKey: TABLE_KEY},
		&rowCriteria{limit: 10, rootWhere: &QueryWhere{}},
		&rowCriteria{limit: 10, tableKey: TABLE_KEY},
		&rowCriteria{rootWhere: &QueryWhere{}, tableKey: TABLE_KEY},
	}

	for _, b := range bad {
		if b.isReady() {
			t.Error("Unexpected rowCriteria isReady()")
		}
	}

	okay := &rowCriteria{}
	okay.limit = 10
	okay.rootWhere = &QueryWhere{}
	okay.tableKey = TABLE_KEY

	if !okay.isReady() {
		t.Error("Expected rowCriteria isReady()")
	}
}

const TABLE_KEY = "The Table"
