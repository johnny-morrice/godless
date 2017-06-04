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
	"fmt"

	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query a godless server",
	Long:  `Send a query to a godless server over HTTP.`,
	// TODO tidy method.
	Run: func(cmd *cobra.Command, args []string) {
		client := lib.MakeClient(serverAddr)

		if source != "" || analyse {
			q, err := lib.CompileQuery(source)

			if err != nil {
				die(err)
			}

			query = q
		}

		if analyse {
			format := "Query analysis for:\n\n%s\n\n%v\n\n"
			fmt.Printf(format, source, query.Analyse())
			fmt.Println("Syntax tree:\n\n")
			query.Parser.PrintSyntaxTree()
		}

		if dryrun {
			return
		}

		response, err := runClient(client)

		if err != nil {
			die(err)
		}

		if response == nil {
			return
		}

		var text string
		text, err = response.AsText()

		if err != nil {
			die(err)
		}

		fmt.Println(text)
	},
}

var query *lib.Query
var source string
var analyse bool
var noparse bool
var serverAddr string
var dryrun bool
var reflect string

func runClient(client *lib.Client) (*lib.APIResponse, error) {
	if query != nil {
		if noparse {
			return client.SendRawQuery(source)
		} else {
			return client.SendQuery(query)
		}
	} else if reflect != "" {
		reflectType, err := parseReflect()

		if err != nil {
			return nil, err
		}

		return client.SendReflection(reflectType)
	} else {
		return nil, nil
	}
}

func parseReflect() (lib.APIReflectionType, error) {
	switch reflect {
	case "index":
		return lib.REFLECT_INDEX, nil
	case "head":
		return lib.REFLECT_HEAD_PATH, nil
	case "namespace":
		return lib.REFLECT_DUMP_NAMESPACE, nil
	default:
		return lib.REFLECT_NOOP, fmt.Errorf("Unknown reflect type: %v", reflect)
	}
}

func init() {
	RootCmd.AddCommand(queryCmd)

	queryCmd.Flags().BoolVar(&dryrun, "dryrun", false, "Don't send query to server")
	queryCmd.Flags().StringVar(&source, "query", "", "Godless NoSQL query text")
	queryCmd.Flags().BoolVar(&analyse, "analyse", false, "Analyse query")
	queryCmd.Flags().BoolVar(&noparse, "noparse", false, "Send raw query text to server.")
	queryCmd.Flags().StringVar(&serverAddr, "server", "localhost:8085", "Server address")
	queryCmd.Flags().StringVar(&reflect, "reflect", "", "Reflect on server state")
}
