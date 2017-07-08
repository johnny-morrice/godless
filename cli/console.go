package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/peterh/liner"
	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
)

type TerminalOptions struct {
	Client  api.Client
	Verbose bool
	History io.ReadWriter
	Output  io.Writer
}

type Console struct {
	TerminalOptions
	line         *liner.State
	outputBuffer *bufio.Writer
}

func MakeConsole(options TerminalOptions) *Console {
	console := &Console{TerminalOptions: options}

	if console.Output == nil {
		console.Output = os.Stdout
	}

	console.outputBuffer = bufio.NewWriter(console.Output)

	return console
}

func (console *Console) ReadEvalPrintLoop() error {
	err := console.setupLiner()
	defer console.shutdownLiner()

	if err != nil {
		return err
	}

	for {
		err := console.readEvalPrint()
		if err != nil {
			return err
		}
	}

	return nil
}

func (console *Console) setupLiner() error {
	console.line = liner.NewLiner()
	console.line.SetCtrlCAborts(true)

	if console.History != nil {
		_, err := console.line.ReadHistory(console.History)
		return err
	}

	return nil
}

func (console *Console) shutdownLiner() {
	defer console.outputBuffer.Flush()

	_, err := console.line.WriteHistory(console.History)

	if err != nil {
		console.printf("Failed to write history\n")
	}

	console.line.Close()
}

func (console Console) readEvalPrint() error {
	defer console.outputBuffer.Flush()
	queryText, err := console.line.Prompt("> ")

	if err != nil {
		return err
	}

	console.line.AppendHistory(queryText)

	query, err := query.CompileQuery(queryText)

	if err != nil {
		console.printf("Compiliation error: %v", err.Error())
		return nil
	}

	sendTime := time.Now()
	resp, err := console.Client.SendQuery(query)
	receiveTime := time.Now()
	waitTime := receiveTime.Sub(sendTime)

	if err != nil {
		console.printf("Error: %v\n", err.Error())

		if console.Verbose {
			if resp.Msg != "" {
				console.printf("Response message: %v\n", resp.Msg)
			}

			if resp.Err != nil && resp.Err != err {
				console.printf("Response error: %v\n", resp.Err.Error())
			}
		}
	}

	console.printResponseTables(resp, query)

	console.printf("Waited %v for response from server.\n", waitTime)

	return nil
}

func (console *Console) printResponseTables(resp api.APIResponse, q *query.Query) {
	if q.OpCode == query.SELECT {
		console.printNamespaceTables(resp.Namespace)
	}

	console.printPath(resp.Path)
}

func (console *Console) printIndexTables(index crdt.Index) {
	if index.IsEmpty() {
		return
	}

	table := makeIndexTable(index)
	table.fprint(console.outputBuffer)
}

func (console *Console) printNamespaceTables(namespace crdt.Namespace) {
	if namespace.IsEmpty() {
		console.printf("No results returned.\n")
		return
	}

	table := makeNamespaceTable(namespace)
	table.fprint(console.outputBuffer)
	console.printf("Found %d Namespace Entries.\n", table.countrows())
}

func (console *Console) printPath(path crdt.IPFSPath) {
	if !crdt.IsNilPath(path) {
		fmt.Println(path)
	}
}

func (console *Console) printf(format string, args ...interface{}) {
	fmt.Fprintf(console.outputBuffer, format, args...)
}

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

func makeIndexTable(index crdt.Index) *monospaceTable {
	panic("not implemented")
}

// TODO figure out how to make signatures look nice
func makeNamespaceTable(namespace crdt.Namespace) *monospaceTable {
	table := &monospaceTable{}
	columns := []string{
		"Table",
		"Row",
		"Entry",
		"Point",
		// "Signatures",
	}

	for _, c := range columns {
		table.addColumn(c)
	}

	namespace.ForeachEntry(func(t crdt.TableName, r crdt.RowName, e crdt.EntryName, entry crdt.Entry) {
		for _, point := range entry.GetValues() {
			// sigText := makeSigText(point.Signatures())
			row := []string{
				string(t),
				string(r),
				string(e),
				string(point.Text()),
				// sigText,
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
