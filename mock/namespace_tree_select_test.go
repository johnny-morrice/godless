package mock_godless

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
)

func TestRunQuerySelectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	whereA := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			Literals: []string{"Hi"},
			Keys: []string{"Entry A"},
		},
	}

	whereB := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			Literals: []string{"Hi"},
			Keys: []string{"Entry B"},
		},
	}

	whereC := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_NEQ,
			Literals: []string{"Hello World"},
			Keys: []string{"Entry B"},
		},
	}

	whereD := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			Literals: []string{"Apple"},
			Keys: []string{"Entry C"},
		},
	}

	whereE := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			Literals: []string{"Orange"},
			Keys: []string{"Entry D"},
		},
	}

	queries := []*lib.Query{
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 1,
				Where: whereA,
			},
		},
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereB,
			},
		},
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereC,
			},
		},
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: lib.QueryWhere{
					OpCode: lib.AND,
					Clauses: []lib.QueryWhere{whereD, whereE},
				},
			},
		},
	}

	responseA := lib.RESPONSE_OK
	responseA.Rows = rowsA()

	responseB := lib.RESPONSE_OK
	responseB.Rows = append(rowsB(), rowsC()...)

	responseC := lib.RESPONSE_OK
	responseC.Rows = rowsC()

	responseD := lib.RESPONSE_OK
	responseD.Rows = rowsA()
	var _ interface{} = responseD

	expect := []lib.APIResponse{
		responseA,
		responseB,
		responseC,
		responseD,
	}

	if len(queries) != len(expect) {
		panic("mismatched input and expect")
	}

	mock.EXPECT().LoadTraverse(gomock.Any()).Return(nil).Do(feedNamespace).Times(len(queries))

	for i, q := range queries {
		selector := lib.MakeNamespaceTreeSelect(mock)
		q.Visit(selector)
		resp := selector.RunQuery()

		if !reflect.DeepEqual(resp, expect[i]) {
			if resp.Rows == nil {
				t.Error("resp.Rows was nil")
			}
			if resp.Err != nil {
				t.Error("resp.Err was", resp.Err)
			}

			t.Error("Case", i, "Expected", expect[i], "but receieved", resp)
		}
	}
}

func TestRunQuerySelectFailure(t *testing.T) {
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

func rowsB() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				"Entry B": []string{"Hi", "Hello World"},
			},
		},
	}
}

func rowsC() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				"Entry B": []string{"Hi", "Hello Dude"},
			},
		},
	}
}

func rowsD() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				"Entry C": []string{"Apple"},
				"Entry D": []string{"Orange"},
			},
		},
	}
}

func rowsZ() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				"Entry A": []string{"No", "Match"},
				"Entry C": []string{"No", "Match", "Here"},
				"Entry D": []string{"Nada!"},
			},
		},
	}
}

func tableA() lib.Table {
	return mktable("A", rowsA())
}

func tableB() lib.Table {
	return mktable("B", rowsB())
}

func tableC() lib.Table {
	return mktable("C", rowsC())
}

func tableD() lib.Table {
	return mktable("D", rowsD())
}

func tableZ() lib.Table {
	return mktable("Z", rowsZ())
}

func feedNamespace(ntr lib.NamespaceTreeReader) {
	ntr.ReadNamespace(mkselectns())
}

func mkselectns() *lib.Namespace {
	namespace := lib.MakeNamespace()
	tables := []lib.Table{
		tableA(),
		tableB(),
		tableC(),
		tableD(),
		tableZ(),
	}

	for _, t := range tables {
		var err error
		namespace, err = namespace.JoinTable(mainTableKey, t)

		if err != nil {
			panic(err)
		}
	}

	return namespace
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
