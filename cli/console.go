package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/peterh/liner"

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

	query, err := query.Compile(queryText)

	if err != nil {
		console.printf("Compiliation error: %v", err.Error())
		return nil
	}

	sendTime := time.Now()
	request := api.MakeQueryRequest(query)
	resp, err := console.Client.Send(request)
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

func (console *Console) printResponseTables(resp api.Response, q *query.Query) {
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
	console.printf("\nFound %d Namespace Entries.\n", table.countrows())
}

func (console *Console) printPath(path crdt.IPFSPath) {
	if !crdt.IsNilPath(path) {
		fmt.Println(path)
	}
}

func (console *Console) printf(format string, args ...interface{}) {
	fmt.Fprintf(console.outputBuffer, format, args...)
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

	table.addColumn(columns...)

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

			table.addRow(row...)
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
const __ROW_PADDING_COUNT = 1
