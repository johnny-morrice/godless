package mock_godless

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
	"github.com/pkg/errors"
)

func TestRunQuerySelectSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	whereA := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_EQ,
			Literals: []string{"Hi"},
			Keys:     []lib.EntryName{"Entry A"},
		},
	}

	whereB := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_EQ,
			Literals: []string{"Hi"},
			Keys:     []lib.EntryName{"Entry B"},
		},
	}

	whereC := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_NEQ,
			Literals: []string{"Hello World"},
			Keys:     []lib.EntryName{"Entry B"},
		},
	}

	whereD := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_EQ,
			Literals: []string{"Apple"},
			Keys:     []lib.EntryName{"Entry C"},
		},
	}

	whereE := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_EQ,
			Literals: []string{"Orange"},
			Keys:     []lib.EntryName{"Entry D"},
		},
	}

	whereF := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_EQ,
			Literals: []string{"Train"},
			Keys:     []lib.EntryName{"Entry E"},
		},
	}

	whereG := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_EQ,
			Literals: []string{"Bus"},
			Keys:     []lib.EntryName{"Entry E"},
		},
	}

	whereH := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:   lib.STR_EQ,
			Literals: []string{"Boat"},
			Keys:     []lib.EntryName{"Entry E"},
		},
	}

	whereI := lib.QueryWhere{
		OpCode: lib.PREDICATE,
		Predicate: lib.QueryPredicate{
			OpCode:        lib.STR_EQ,
			IncludeRowKey: true,
			Literals:      []string{"Row F0"},
		},
	}

	queries := []*lib.Query{
		// One result
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereA,
			},
		},
		// Multiple results
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereB,
			},
		},
		// STR_NEQ
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereC,
			},
		},
		// AND
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: lib.QueryWhere{
					OpCode:  lib.AND,
					Clauses: []lib.QueryWhere{whereD, whereE},
				},
			},
		},
		// OR
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: lib.QueryWhere{
					OpCode:  lib.OR,
					Clauses: []lib.QueryWhere{whereF, whereG},
				},
			},
		},
		// No results
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereH,
			},
		},
		// Row key
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: MAIN_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 2,
				Where: whereI,
			},
		},
		// No where clause
		&lib.Query{
			OpCode:   lib.SELECT,
			TableKey: ALT_TABLE_KEY,
			Select: lib.QuerySelect{
				Limit: 3,
			},
		},
	}

	responseA := lib.RESPONSE_OK
	responseA.QueryResponse.Rows = rowsA()

	responseB := lib.RESPONSE_OK
	responseB.QueryResponse.Rows = append(rowsB(), rowsC()...)

	responseC := lib.RESPONSE_OK
	responseC.QueryResponse.Rows = rowsC()

	responseD := lib.RESPONSE_OK
	responseD.QueryResponse.Rows = rowsD()

	responseE := lib.RESPONSE_OK
	responseE.QueryResponse.Rows = rowsE()

	responseF := lib.RESPONSE_OK

	responseG := lib.RESPONSE_OK
	responseG.QueryResponse.Rows = rowsF()

	responseH := lib.RESPONSE_OK
	responseH.QueryResponse.Rows = rowsG()

	expect := []lib.APIResponse{
		responseA,
		responseB,
		responseC,
		responseD,
		responseE,
		responseF,
		responseG,
		responseH,
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
			if resp.QueryResponse.Rows == nil {
				t.Error("resp.QueryResponse.Rows was nil")
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

	mock.EXPECT().LoadTraverse(gomock.Any()).Return(errors.New("Expected Error"))

	failQuery := &lib.Query{
		OpCode:   lib.SELECT,
		TableKey: MAIN_TABLE_KEY,
		Select: lib.QuerySelect{
			Limit: 2,
			Where: lib.QueryWhere{
				OpCode: lib.PREDICATE,
				Predicate: lib.QueryPredicate{
					OpCode:   lib.STR_EQ,
					Literals: []string{"Hi"},
					Keys:     []lib.EntryName{"Entry A"},
				},
			},
		},
	}

	selector := lib.MakeNamespaceTreeSelect(mock)
	failQuery.Visit(selector)
	resp := selector.RunQuery()

	if resp.Msg != "error" {
		t.Error("Expected Msg error but received", resp.Msg)
	}

	if resp.Err == nil {
		t.Error("Expected response Err")
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
						OpCode:   lib.STR_EQ,
						Literals: []string{"Hi"},
						Keys:     []lib.EntryName{"Entry A"},
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
						OpCode:   lib.STR_EQ,
						Literals: []string{"Hi"},
						Keys:     []lib.EntryName{"Entry A"},
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
						Keys:     []lib.EntryName{"Entry A"},
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
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			// TODO use user concepts to match only the Hi.
			"Entry A": lib.MakeEntry([]lib.Point{"Hi", "Hello"}),
		}),
	}
}

func rowsB() []lib.Row {
	return []lib.Row{
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry B": lib.MakeEntry([]lib.Point{"Hi", "Hello World"}),
		}),
	}
}

func rowsC() []lib.Row {
	return []lib.Row{
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry B": lib.MakeEntry([]lib.Point{"Hi", "Hello Dude"}),
		}),
	}
}

func rowsD() []lib.Row {
	return []lib.Row{
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry C": lib.MakeEntry([]lib.Point{"Apple"}),
			"Entry D": lib.MakeEntry([]lib.Point{"Orange"}),
		}),
	}
}

func rowsE() []lib.Row {
	return []lib.Row{
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry E": lib.MakeEntry([]lib.Point{"Bus"}),
		}),
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry E": lib.MakeEntry([]lib.Point{"Train"}),
		}),
	}
}

func rowsF() []lib.Row {
	return []lib.Row{
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry F": lib.MakeEntry([]lib.Point{"This row", "rocks"}),
		}),
	}
}

func rowsG() []lib.Row {
	return []lib.Row{
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry Q": lib.MakeEntry([]lib.Point{"Hi", "Folks"}),
		}),
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry R": lib.MakeEntry([]lib.Point{"Wowzer"}),
		}),
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry S": lib.MakeEntry([]lib.Point{"Trumpet"}),
		}),
	}
}

// Non matching rows.
func rowsZ() []lib.Row {
	return []lib.Row{
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]lib.Point{"No", "Match"}),
		}),
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry C": lib.MakeEntry([]lib.Point{"No", "Match", "Here"}),
			"Entry D": lib.MakeEntry([]lib.Point{"Nada!"}),
		}),
		lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry E": lib.MakeEntry([]lib.Point{"Horse"}),
		}),
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

func tableG() lib.Table {
	return mktable("G", rowsG())
}

func tableZ() lib.Table {
	return mktable("Z", rowsZ())
}

func feedNamespace(ntr lib.NamespaceTreeReader) {
	ntr.ReadNamespace(mkselectns())
}

func mkselectns() lib.Namespace {
	namespace := lib.EmptyNamespace()
	mainTables := []lib.Table{
		tableA(),
		tableB(),
		tableC(),
		tableD(),
		tableE(),
		tableF(),
		tableZ(),
	}
	altTables := []lib.Table{
		tableG(),
	}

	tables := map[lib.TableName][]lib.Table{
		MAIN_TABLE_KEY: mainTables,
		ALT_TABLE_KEY:  altTables,
	}

	for tableKey, ts := range tables {
		for _, t := range ts {
			namespace = namespace.JoinTable(tableKey, t)
		}
	}

	return namespace
}

func mktable(name string, rows []lib.Row) lib.Table {
	table := lib.EmptyTable()

	for i, r := range rows {
		rowKey := lib.RowName(fmt.Sprintf("Row %v%v", name, i))
		table = table.JoinRow(rowKey, r)
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

	if len(a.QueryResponse.Rows) != len(b.QueryResponse.Rows) {
		return false
	}

	if !rowsEq(a.QueryResponse.Rows, b.QueryResponse.Rows) {
		return false
	}

	return true
}

func rowsEq(a, b []lib.Row) bool {
	for _, ar := range a {
		found := false
		for _, br := range b {
			if ar.Equals(br) {
				found = true
			}
		}
		if !found {
			return false
		}
	}

	return true
}

const MAIN_TABLE_KEY = lib.TableName("The Table")
const ALT_TABLE_KEY = lib.TableName("Another table")
