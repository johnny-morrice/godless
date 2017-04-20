package mock_godless

import (
	"fmt"
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

	whereF := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			Literals: []string{"Train"},
			Keys: []string{"Entry E"},
		},
	}

	whereG := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			Literals: []string{"Bus"},
			Keys: []string{"Entry E"},
		},
	}

	whereH := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			Literals: []string{"Boat"},
			Keys: []string{"Entry E"},
		},
	}

	whereI := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode: lib.STR_EQ,
			IncludeRowKey: true,
			Literals: []string{"Row F0"},
		},
	}

	queries := []*lib.Query{
		// One result
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereA,
			},
		},
		// Multiple results
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereB,
			},
		},
		// STR_NEQ
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereC,
			},
		},
		// AND
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
		// OR
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: lib.QueryWhere{
					OpCode: lib.OR,
					Clauses: []lib.QueryWhere{whereF, whereG},
				},
			},
		},
		// No results
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereH,
			},
		},
		// Row key
		&lib.Query{
			OpCode: lib.SELECT,
			TableKey: mainTableKey,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereI,
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
	responseD.Rows = rowsD()

	responseE := lib.RESPONSE_OK
	responseE.Rows = rowsE()

	responseF := lib.RESPONSE_OK

	responseG := lib.RESPONSE_OK
	responseG.Rows = rowsF()

	expect := []lib.APIResponse{
		responseA,
		responseB,
		responseC,
		responseD,
		responseE,
		responseF,
		responseG,
	}

	if len(queries) != len(expect) {
		panic("mismatched input and expect")
	}

	mock.EXPECT().LoadTraverse(gomock.Any()).Return(nil).Do(feedNamespace).Times(len(queries))

	for i, q := range queries {
		selector := lib.MakeNamespaceTreeSelect(mock)
		q.Visit(selector)
		resp := selector.RunQuery()

		if !apiResponseEq(resp, expect[i]) {
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

func rowsE() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				"Entry E": []string{"Bus"},
			},
		},
		lib.Row{
			Entries: map[string][]string {
				"Entry E": []string{"Train"},
			},
		},
	}
}

func rowsF() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				"Entry F": []string{"This row", "rocks"},
			},
		},
	}
}


// Non matching rows.
func rowsZ() []lib.Row {
	return []lib.Row{
		lib.Row{
			Entries: map[string][]string {
				"Entry A": []string{"No", "Match"},
			},
		},
		lib.Row{
			Entries: map[string][]string {
				"Entry C": []string{"No", "Match", "Here"},
				"Entry D": []string{"Nada!"},
			},
		},
		lib.Row{
			Entries: map[string][]string {
				"Entry E": []string{"Horse"},
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

func tableE() lib.Table {
	return mktable("E", rowsE())
}

func tableF() lib.Table {
	return mktable("F", rowsF())
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
		tableE(),
		tableF(),
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
		rowKey := fmt.Sprintf("Row %v%v", name, i)
		table.Rows[rowKey] = r
	}

	return table
}

func apiResponseEq(a, b lib.APIResponse) bool {
	if a.Msg != b.Msg {
		return false
	}

	if a.Err != b.Err {
		return false
	}

	if len(a.Rows) != len(b.Rows) {
		return false
	}

	if !rowsEq(a.Rows, b.Rows) {
		return false
	}

	return true
}

func rowsEq(a, b []lib.Row) bool {
	for _, ar := range a {
		found := false
		for _, br := range b {
			if rowEq(ar, br) {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func rowEq(a, b lib.Row) bool {
	for ak, av := range a.Entries {
		bv, present := b.Entries[ak]

		if !present {
			return false
		}

		if !entryEq(av, bv) {
			return false
		}
	}

	return true
}

func entryEq(a, b []string) bool {
	for _, av := range a {
		found := false
		for _, bv := range b {
			if av == bv {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}

const mainTableKey = "The Table"
