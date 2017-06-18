package mock_godless

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/eval"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

func TestRunQueryJoinSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	query := &query.Query{
		OpCode:   query.JOIN,
		TableKey: MAIN_TABLE_KEY,
		Join: query.QueryJoin{
			Rows: []query.QueryRowJoin{
				query.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[crdt.EntryName]crdt.PointText{
						"Entry A": "Point A",
						"Entry B": "Point B",
					},
				},
				query.QueryRowJoin{
					RowKey: "Row B",
					Entries: map[crdt.EntryName]crdt.PointText{
						"Entry C": "Point C",
					},
				},
				query.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[crdt.EntryName]crdt.PointText{
						"Entry A": "Point D",
						"Entry D": "Point E",
					},
				},
			},
		},
	}

	table := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point A"), crdt.UnsignedPoint("Point D")}),
			"Entry B": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point B")}),
			"Entry D": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point E")}),
		}),
		"Row B": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry C": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point C")}),
		}),
	})
	mock.EXPECT().JoinTable(MAIN_TABLE_KEY, mtchtable(table)).Return(nil)

	joiner := eval.MakeNamespaceTreeJoin(mock)
	query.Visit(joiner)
	resp := joiner.RunQuery()

	if !api.RESPONSE_QUERY.Equals(resp) {
		t.Error("Expected", api.RESPONSE_QUERY, "but was", resp)
	}
}

func TestRunQueryJoinFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	failQuery := &query.Query{
		OpCode:   query.JOIN,
		TableKey: MAIN_TABLE_KEY,
		Join: query.QueryJoin{
			Rows: []query.QueryRowJoin{
				query.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[crdt.EntryName]crdt.PointText{
						"Entry A": "Point A",
						"Entry B": "Point B",
					},
				},
			},
		},
	}

	table := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point A")}),
			"Entry B": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point B")}),
		}),
	})

	mock.EXPECT().JoinTable(MAIN_TABLE_KEY, mtchtable(table)).Return(errors.New("Expected error"))

	joiner := eval.MakeNamespaceTreeJoin(mock)
	failQuery.Visit(joiner)
	resp := joiner.RunQuery()

	if resp.Msg != "error" {
		t.Error("Expected Msg error but received", resp.Msg)
	}

	if resp.Err == nil {
		t.Error("Expected response Err")
	}

	if resp.Type != api.API_QUERY {
		t.Error("Unexpected response Type")
	}
}

func TestRunQueryJoinInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	invalidQueries := []*query.Query{
		// Basically wrong.
		&query.Query{},
		&query.Query{OpCode: query.SELECT},
	}

	for _, q := range invalidQueries {
		joiner := eval.MakeNamespaceTreeJoin(mock)
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

func mtchtable(t crdt.Table) gomock.Matcher {
	return tablematcher{t}
}

type tablematcher struct {
	t crdt.Table
}

func (tm tablematcher) String() string {
	return "is matching Table"
}

func (tm tablematcher) Matches(v interface{}) bool {
	other, ok := v.(crdt.Table)

	if !ok {
		return false
	}

	return tm.t.Equals(other)
}
