package cli

import (
	"io"

	"github.com/pkg/errors"
)

type monospaceTable struct {
	columns      []string
	columnWidths []int
	rows         [][]string
	frozen       bool
	totalWidth   int
}

func (table *monospaceTable) countrows() int {
	return len(table.rows)
}

func (table *monospaceTable) fprint(w io.Writer) {
	table.printline(w)
	table.fprintrow(w, table.columns)
	table.printline(w)

	for _, r := range table.rows {
		table.fprintrow(w, r)
	}
	table.printline(w)
}

func (table *monospaceTable) gettotalwidth() int {
	if table.totalWidth == 0 {
		for _, w := range table.columnWidths {
			table.totalWidth += w
			table.totalWidth += __VALUE_PADDING_COUNT
		}

		table.totalWidth += __ROW_PADDING_COUNT
	}

	return table.totalWidth
}

func (table *monospaceTable) printline(w io.Writer) {
	total := table.gettotalwidth()
	table.repeat(w, "-", total)

	table.newline(w)
}

func (table *monospaceTable) repeat(w io.Writer, text string, count int) {
	for i := 0; i < count; i++ {
		table.write(w, text)
	}
}

func (table *monospaceTable) fprintrow(w io.Writer, row []string) {
	for i, value := range row {
		table.write(w, "| ")
		table.fprintvalue(w, i, value)
		table.write(w, " ")
	}

	table.write(w, " |")
	table.newline(w)
}

func (table *monospaceTable) newline(w io.Writer) {
	table.write(w, "\n")
}

func (table *monospaceTable) write(w io.Writer, text string) {
	w.Write([]byte(text))
}

func (table *monospaceTable) fprintvalue(w io.Writer, columnIndex int, value string) {
	table.write(w, value)
	padding := table.padsize(columnIndex, value)
	table.repeat(w, " ", padding)
}

func (table *monospaceTable) padsize(columnIndex int, text string) int {
	columnWidth := table.columnWidths[columnIndex]
	return columnWidth - len(text)
}

func (table *monospaceTable) addColumn(column string) error {
	if table.frozen {
		return errors.New("don't addColumn after addRow")
	}

	table.columns = append(table.columns, column)
	table.columnWidths = append(table.columnWidths, 0)
	index := len(table.columns) - 1
	table.recalcWidth(index, column)
	return nil
}

func (table *monospaceTable) recalcWidth(columnIndex int, addition string) {
	current := table.columnWidths[columnIndex]
	newWidth := len(addition)
	if current < newWidth {
		table.columnWidths[columnIndex] = newWidth
	}
}

func (table *monospaceTable) addRow(values []string) error {
	if len(values) != len(table.columns) {
		return errors.New("row length does not match column length")
	}

	table.frozen = true

	table.rows = append(table.rows, values)

	for i, value := range values {
		table.recalcWidth(i, value)
	}

	return nil
}
