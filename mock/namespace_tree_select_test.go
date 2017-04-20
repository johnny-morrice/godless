package mock_godless

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
)

func TestRunQuerySelectA(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)
	mock.EXPECT().LoadTraverse(gomock.Any()).Return(nil).Do(feedNamespace)

	query := &lib.Query{
		OpCode: lib.SELECT,
		TableKey: mainTableKey,
		Select: lib.QuerySelect{
			Limit: 1,
			Where: lib.QueryWhere{
				OpCode: lib.PREDICATE,
				Predicate: lib.QueryPredicate{
					OpCode: lib.STR_EQ,
					Literals: []string{"Hi"},
					Keys: []string{"Entry A"},
				},
			},
		},
	}

	selector := lib.MakeNamespaceTreeSelect(mock)
	query.Visit(selector)
	resp := selector.RunQuery()

	if !reflect.DeepEqual(resp.Rows, rowsA()) {
		if resp.Rows == nil {
			t.Error("resp.Rows was nil")
		}
		t.Error("Expected", rowsA(), "but receieved", resp.Rows)
	}

	if resp.Msg != "ok" {
		t.Error("Expected Msg ok but received", resp.Msg)
	}

	if resp.Err != nil {
		t.Error(resp.Err)
	}
}

func TestRunQuerySelectInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	invalidQueries := []*lib.Query{
		// Basically wrong.
		&lib.Query{},
		&lib.Query{OpCode: lib.JOIN},
		// No limit
		&lib.Query{
			Select: lib.QuerySelect{
				Where: lib.QueryWhere{
					OpCode: lib.PREDICATE,
					Predicate: lib.QueryPredicate{
						OpCode: lib.STR_EQ,
						Literals: []string{"Hi"},
						Keys: []string{"Entry A"},
					},
				},
			},
		},
		// No where OpCode
		&lib.Query{
			Select: lib.QuerySelect{
				Limit: 1,
				Where: lib.QueryWhere{
					Predicate: lib.QueryPredicate{
						OpCode: lib.STR_EQ,
						Literals: []string{"Hi"},
						Keys: []string{"Entry A"},
					},
				},
			},
		},
		// No predicate OpCode
		&lib.Query{
			Select: lib.QuerySelect{
				Limit: 1,
				Where: lib.QueryWhere{
					OpCode: lib.PREDICATE,
					Predicate: lib.QueryPredicate{
						Literals: []string{"Hi"},
						Keys: []string{"Entry A"},
					},
				},
			},
		},
	}

	for _, q := range invalidQueries {
		selector := lib.MakeNamespaceTreeSelect(mock)
		q.Visit(selector)
		resp := selector.RunQuery()

		if resp.Msg != "error" {
			t.Error("Expected Msg error but received", resp.Msg)
		}

		if resp.Err == nil {
			t.Error("Expected response Err")
		}
	}

}

func TestRunQuerySelectFail(t *testing.T) {

}

func feedNamespace(ntr lib.NamespaceTreeReader) {
	ntr.ReadNamespace(mkselectns())
}

func mkselectns() *lib.Namespace {
	namespace := lib.MakeNamespace()
	joinA, err := namespace.JoinTable(mainTableKey, tableA())

	if err != nil {
		panic(err)
	}

	joinZ, err := joinA.JoinTable(mainTableKey, nomatchtable())

	if err != nil {
		panic(err)
	}

	return joinZ
}

func rowsA() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				// TODO use user concepts to match only the Hi.
				"Entry A": []string{"Hi", "Hello"},
			},
		},
	}
}

func rowsZ() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				// TODO use user concepts to match only the Hi.
				"Entry Z": []string{"No", "Match"},
				"Entry ZZ": []string{"No", "Match", "Here"},
			},
		},
	}
}

func tableA() lib.Table {
	return mktable("A", rowsA())
}

func nomatchtable() lib.Table {
	return mktable("Z", rowsZ())
}

func mktable(name string, rows []lib.Row) lib.Table {
	table := lib.Table{
		Rows: map[string]lib.Row{},
	}

	for i, r := range rows {
		table.Rows[fmt.Sprintf("Row %v%v", name, i)] = r
	}

	return table
}

const mainTableKey = "The Table"
