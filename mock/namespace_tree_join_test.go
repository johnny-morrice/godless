package mock_godless


import (
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
)

func TestRunQueryJoinSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockNamespaceTree(ctrl)

	query := &lib.Query{
		OpCode: lib.JOIN,
		TableKey: mainTableKey,
		Join: lib.QueryJoin{
			Rows: []lib.QueryRowJoin{
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[string]string{
						"Entry A": "Value A",
						"Entry B": "Value B",
					},
				},
				lib.QueryRowJoin{
					RowKey: "Row B",
					Entries: map[string]string{
						"Entry C": "Value C",
					},
				},
				lib.QueryRowJoin{
					RowKey: "Row A",
					Entries: map[string]string{
						"Entry A": "Value D",
						"Entry D": "Value E",
					},
				},
			},
		},
	}


	table := lib.Table{
		Rows: map[string]lib.Row{
			"Row A": lib.Row{
				Entries: map[string][]string {
					"Entry A": []string{"Value A", "Value D"},
					"Entry B": []string{"Value B"},
					"Entry D": []string{"Value E"},
				},
			},
			"Row B": lib.Row{
				Entries: map[string][]string {
					"Entry C": []string{"Value C"},
				},
			},
		},
	}
	// TODO table equality matcher.
	mock.EXPECT().JoinTable(mainTableKey, table).Return(nil)

	joiner := lib.MakeNamespaceTreeJoin(mock)
	query.Visit(joiner)
	resp := joiner.RunQuery()

	if !apiResponseEq(lib.RESPONSE_OK, resp) {
		t.Error("Expected", lib.RESPONSE_OK, "but was", resp)
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
