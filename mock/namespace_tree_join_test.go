package mock_godless

import (
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
	"github.com/pkg/errors"
)

func TestRunQueryJoinSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	query := &lib.Query{
		OpCode:   lib.JOIN,
		TableKey: MAIN_TABLE_KEY,
		Join: lib.QueryJoin{
			Rows: []lib.QueryRowJoin{
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[lib.EntryName]lib.Point{
						"Entry A": "Point A",
						"Entry B": "Point B",
					},
				},
				lib.QueryRowJoin{
					RowKey: "Row B",
					Entries: map[lib.EntryName]lib.Point{
						"Entry C": "Point C",
					},
				},
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[lib.EntryName]lib.Point{
						"Entry A": "Point D",
						"Entry D": "Point E",
					},
				},
			},
		},
	}

	table := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row A": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]lib.Point{"Point A", "Point D"}),
			"Entry B": lib.MakeEntry([]lib.Point{"Point B"}),
			"Entry D": lib.MakeEntry([]lib.Point{"Point E"}),
		}),
		"Row B": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry C": lib.MakeEntry([]lib.Point{"Point C"}),
		}),
	})
	mock.EXPECT().JoinTable(MAIN_TABLE_KEY, mtchtable(table)).Return(nil)

	joiner := lib.MakeNamespaceTreeJoin(mock)
	query.Visit(joiner)
	resp := joiner.RunQuery()

	if !lib.RESPONSE_QUERY.Equals(resp) {
		t.Error("Expected", lib.RESPONSE_QUERY, "but was", resp)
	}
}

func TestRunQueryJoinFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	failQuery := &lib.Query{
		OpCode:   lib.JOIN,
		TableKey: MAIN_TABLE_KEY,
		Join: lib.QueryJoin{
			Rows: []lib.QueryRowJoin{
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[lib.EntryName]lib.Point{
						"Entry A": "Point A",
						"Entry B": "Point B",
					},
				},
			},
		},
	}

	table := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row A": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]lib.Point{"Point A"}),
			"Entry B": lib.MakeEntry([]lib.Point{"Point B"}),
		}),
	})

	mock.EXPECT().JoinTable(MAIN_TABLE_KEY, mtchtable(table)).Return(errors.New("Expected error"))

	joiner := lib.MakeNamespaceTreeJoin(mock)
	failQuery.Visit(joiner)
	resp := joiner.RunQuery()

	if resp.Msg != "error" {
		t.Error("Expected Msg error but received", resp.Msg)
	}

	if resp.Err == nil {
		t.Error("Expected response Err")
	}

	if resp.Type != lib.API_QUERY {
		t.Error("Unexpected response Type")
	}
}

func TestRunQueryJoinInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	invalidQueries := []*lib.Query{
		// Basically wrong.
		&lib.Query{},
		&lib.Query{OpCode: lib.SELECT},
	}

	for _, q := range invalidQueries {
		joiner := lib.MakeNamespaceTreeJoin(mock)
		q.Visit(joiner)
		resp := joiner.RunQuery()

		if resp.Msg != "error" {
			t.Error("Expected Msg error but received", resp.Msg)
		}

		if resp.Err == nil {
			t.Error("Expected response Err")
		}
	}
}

func mtchtable(t lib.Table) gomock.Matcher {
	return tablematcher{t}
}

type tablematcher struct {
	t lib.Table
}

func (tm tablematcher) String() string {
	return "is matching Table"
}

func (tm tablematcher) Matches(v interface{}) bool {
	other, ok := v.(lib.Table)

	if !ok {
		return false
	}

	return tm.t.Equals(other)
}
