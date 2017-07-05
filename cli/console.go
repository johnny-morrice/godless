package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterh/liner"
	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
)

type TerminalOptions struct {
	Client api.Client
}

func RunTerminalConsole(options TerminalOptions) error {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	for {
		queryText, err := line.Prompt(">")

		if err != nil {
			return err
		}

		query, err := query.CompileQuery(queryText)

		if err != nil {
			fmt.Printf("Compiliation error: %v", err.Error())
			continue
		}

		resp, err := options.Client.SendQuery(query)

		if err != nil {
			fmt.Printf("Error: %v", err.Error())

			if resp.Msg != "" {
				fmt.Printf("Response message: %v", resp.Msg)
			}

			if resp.Err != nil && resp.Err != err {
				fmt.Printf("Response error: %v", resp.Err.Error())
			}
		}

		printResponseTables(resp)
	}

	return nil
}

func printResponseTables(resp api.APIResponse) {
	printNamespaceTables(resp.Namespace)
	printIndexTables(resp.Index)
	printPath(resp.Path)
}

func printIndexTables(index crdt.Index) {
	if index.IsEmpty() {
		return
	}

	table := makeIndexTable(index)
	table.fprint(os.Stdout)
}

func printNamespaceTables(namespace crdt.Namespace) {
	if namespace.IsEmpty() {
		return
	}

	table := makeNamespaceTable(namespace)
	table.fprint(os.Stdout)
}

func printPath(path crdt.IPFSPath) {
	if !crdt.IsNilPath(path) {
		fmt.Println(path)
	}
}

type monospaceTable struct {
	columns      []string
	columnWidths []int
	rows         [][]string
	frozen       bool
	totalWidth   int
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

func makeIndexTable(index crdt.Index) *monospaceTable {
	panic("not implemented")
}

func makeNamespaceTable(namespace crdt.Namespace) *monospaceTable {
	table := &monospaceTable{}
	columns := []string{
		"Table",
		"Row",
		"Entry",
		"Point",
		"Signatures",
	}

	for _, c := range columns {
		table.addColumn(c)
	}

	namespace.ForeachEntry(func(t crdt.TableName, r crdt.RowName, e crdt.EntryName, entry crdt.Entry) {
		for _, point := range entry.GetValues() {
			sigText := makeSigText(point.Signatures())
			row := []string{
				string(t),
				string(r),
				string(e),
				string(point.Text()),
				sigText,
			}

			table.addRow(row)
		}
	})

	return table
}

func makeSigText(signatures []crypto.Signature) string {
	sigText := make([]string, 0, len(signatures))

	for _, sig := range signatures {
		text, err := crypto.PrintSignature(sig)

		if err != nil {
			log.Error("Error serializing crypto.Signature: %v", err)
			continue
		}

		sigText = append(sigText, string(text))
	}

	return strings.Join(sigText, " ")
}

const __VALUE_PADDING_COUNT = 3
const __ROW_PADDING_COUNT = 2
