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
		TableKey: mainTableKey,
		Join: lib.QueryJoin{
			Rows: []lib.QueryRowJoin{
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[string]string{
						"Entry A": "Point A",
						"Entry B": "Point B",
					},
				},
				lib.QueryRowJoin{
					RowKey: "Row B",
					Entries: map[string]string{
						"Entry C": "Point C",
					},
				},
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[string]string{
						"Entry A": "Point D",
						"Entry D": "Point E",
					},
				},
			},
		},
	}

	table := lib.MakeTable(map[RowName]lib.Row{
		"Row A": lib.MakeRow(map[EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]Point{"Point A", "Point D"}),
			"Entry B": lib.MakeEntry([]Point{"Point B"}),
			"Entry D": lib.MakeEntry([]Point{"Point E"}),
		}),
		"Row B": lib.MakeRow(map[EntryName]lib.Entry{
			"Entry C": lib.MakeEntry([]Point{"Point C"}),
		}),
	})
	mock.EXPECT().JoinTable(mainTableKey, mtchtable(table)).Return(nil)

	joiner := lib.MakeNamespaceTreeJoin(mock)
	query.Visit(joiner)
	resp := joiner.RunQuery()

	if !apiResponseEq(lib.RESPONSE_OK, resp) {
		t.Error("Expected", lib.RESPONSE_OK, "but was", resp)
	}
}

func TestRunQueryJoinFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	failQuery := &lib.Query{
		OpCode:   lib.JOIN,
		TableKey: mainTableKey,
		Join: lib.QueryJoin{
			Rows: []lib.QueryRowJoin{
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[string]string{
						"Entry A": "Point A",
						"Entry B": "Point B",
					},
				},
			},
		},
	}

	table := lib.MakeTable(map[RowName]lib.Row{
		"Row A": lib.MakeRow(map[EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]Point{"Point A"}),
			"Entry B": lib.MakeEntry([]Point{"Point B"}),
		}),
	})

	mock.EXPECT().JoinTable(mainTableKey, mtchtable(table)).Return(errors.New("Expected error"))

	joiner := lib.MakeNamespaceTreeJoin(mock)
	failQuery.Visit(joiner)
	resp := joiner.RunQuery()

	if resp.Msg != "error" {
		t.Error("Expected Msg error but received", resp.Msg)
	}

	if resp.Err == nil {
		t.Error("Expected response Err")
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
