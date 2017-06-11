package mock_godless

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/crdt"
)

func TestTableForeachrow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emptyRow := crdt.EmptyRow()
	fullRow := crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
		"baz": crdt.EmptyEntry(),
	})

	mock := NewMockRowConsumer(ctrl)
	mock.EXPECT().Accept(crdt.RowName("foo"), emptyRow)
	mock.EXPECT().Accept(crdt.RowName("bar"), fullRow)

	table := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"foo": emptyRow,
		"bar": fullRow,
	})

	table.Foreachrow(mock)
}
