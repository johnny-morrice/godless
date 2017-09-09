// Copyright © 2017 Johnny Morrice <john@functorama.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"os"

	"github.com/johnny-morrice/godless/cli"
	"github.com/johnny-morrice/godless/log"
	"github.com/spf13/cobra"
)

// client_consoleCmd represents the client_console command
var clientConsoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Godless terminal console",
	Long:  `A REPL console for evaluating godless queries.`,
	Run: func(cmd *cobra.Command, args []string) {
		options := cli.TerminalOptions{
			Client: makeClient(),
		}

		historyFile := openConsoleHistory(&options)

		if historyFile != nil {
			defer historyFile.Close()
			options.History = historyFile
		}

		console := cli.MakeConsole(options)
		err := console.ReadEvalPrintLoop()

		if err != nil {
			die(err)
		}
	},
}

var clientConsoleParams *Parameters = &Parameters{}

func openOrCreate(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
}

func openConsoleHistory(options *cli.TerminalOptions) *os.File {
	consoleHistoryFilePath := *clientConsoleParams.String(__CONSOLE_HISTORY_FLAG)
	if consoleHistoryFilePath != "" {
		historyFile, err := openOrCreate(consoleHistoryFilePath)

		if err != nil {
			log.Error("Error opening history file: %s", err.Error())
			return nil
		}

		return historyFile
	}

	log.Warn("Not using history file")

	return nil
}

func init() {
	queryCmd.AddCommand(clientConsoleCmd)

	defaultHistoryPath := homePath(__DEFAULT_CONSOLE_HISTORY)
	clientConsoleCmd.Flags().StringVar(clientConsoleParams.String(__CONSOLE_HISTORY_FLAG), __CONSOLE_HISTORY_FLAG, defaultHistoryPath, "Console history file")
}

const __DEFAULT_CONSOLE_HISTORY = ".godless_history"
const __CONSOLE_HISTORY_FLAG = "history"
