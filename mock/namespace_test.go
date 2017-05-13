package mock_godless

import (
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
)

func TestTableForeachrow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emptyRow := lib.EmptyRow()
	fullRow := lib.MakeRow(map[lib.EntryName]lib.Entry{
		"baz": lib.EmptyEntry(),
	})

	mock := NewMockRowConsumer(ctrl)
	mock.EXPECT().Accept(lib.RowName("foo"), emptyRow)
	mock.EXPECT().Accept(lib.RowName("bar"), fullRow)

	table := lib.MakeTable(map[lib.RowName]lib.Row{
		"foo": emptyRow,
		"bar": fullRow,
	})

	table.Foreachrow(mock)
}
