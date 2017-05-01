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
	fullRow := lib.MakeRow(map[string]lib.Entry{
		"baz": lib.EmptyEntry(),
	})

	mock := NewMockRowConsumer(ctrl)
	mock.EXPECT().Accept("foo", emptyRow)
	mock.EXPECT().Accept("bar", fullRow)

	table := lib.MakeTable(map[string]lib.Row{
		"foo": emptyRow,
		"bar": fullRow,
	})

	table.Foreachrow(mock)
}
