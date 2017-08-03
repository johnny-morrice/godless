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
		stop, err := console.readEvalPrint()

		if err != nil {
			return err
		}

		if stop {
			return nil
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

	if console.History != nil {
		_, err := console.line.WriteHistory(console.History)

		if err != nil {
			console.printf("Failed to write history: %v\n", err)
		}
	}

	console.line.Close()
}

func (console Console) readEvalPrint() (bool, error) {
	defer console.outputBuffer.Flush()
	command, err := console.line.Prompt("> ")

	if err != nil {
		return true, err
	}

	if command == ":exit" {
		return true, nil
	}

	console.line.AppendHistory(command)

	query, err := query.Compile(command)

	if err != nil {
		console.printf("Compiliation error: %v", err.Error())
		return false, nil
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

	return false, nil
}

func (console *Console) printResponseTables(resp api.Response, q *query.Query) {
	if q.OpCode == query.SELECT {
		FprintNamespaceTable(console.outputBuffer, resp)
	}
}

func (console *Console) printIndexTables(index crdt.Index) {
	if index.IsEmpty() {
		return
	}

	table := makeIndexTable(index)
	table.fprint(console.outputBuffer)
}
func (console *Console) printf(format string, args ...interface{}) {
	fmt.Fprintf(console.outputBuffer, format, args...)
}

func FprintNamespaceTable(w io.Writer, resp api.Response) {
	if resp.Namespace.IsEmpty() {
		fmt.Fprintln(w, "No results returned.")
		return
	}

	table := makeNamespaceTable(resp.Namespace)
	table.fprint(w)
	fmt.Fprintf(w, "\nFound %d Namespace Entries.\n", table.countrows())

	if !crdt.IsNilPath(resp.Path) {
		fmt.Println(resp.Path)
	}
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
