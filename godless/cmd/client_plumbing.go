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
	"os"

	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
)

// clientPlumbingCmd represents the query command
var clientPlumbingCmd = &cobra.Command{
	Use:   "plumbing",
	Short: "Query a godless server",
	Long:  `Send a query to a godless server over HTTP.  User must specify either --reflect or --query.`,
	// TODO tidy method.
	Run: func(cmd *cobra.Command, args []string) {
		var query *lib.Query
		client := lib.MakeClient(serverAddr)

		validateClientPlumbingArgs(cmd)

		if source != "" || analyse {
			q, err := lib.CompileQuery(source)

			if err != nil {
				die(err)
			}

			query = q
		}

		if analyse {
			err := analyseQuery(query)

			if err != nil {
				die(err)
			}
		}

		if dryrun {
			return
		}

		response, err := sendQuery(client, query)

		if err != nil {
			die(err)
		}

		outputResponse(response)
	},
}

var source string
var analyse bool
var dryrun bool
var binary bool
var reflect string

func outputResponse(response lib.APIResponse) {
	var err error
	if binary {
		err = lib.EncodeAPIResponse(response, os.Stdout)

		if err != nil {
			die(err)
		}
	} else {
		var text string
		text, err = response.AsText()

		if err != nil {
			die(err)
		}

		fmt.Println(text)
	}
}

func validateClientPlumbingArgs(cmd *cobra.Command) {
	if source == "" && reflect == "" {
		err := cmd.Help()

		if err != nil {
			die(err)
		}
	}
}

func analyseQuery(query *lib.Query) error {
	format := "Query analysis for:\n\n%s\n\n%v\n\n"
	fmt.Printf(format, source, query.Analyse())
	fmt.Println("Syntax tree:\n\n")
	query.Parser.PrintSyntaxTree()
	return query.PrettyPrint(os.Stdout)
}

func sendQuery(client *lib.Client, query *lib.Query) (lib.APIResponse, error) {
	if query != nil {
		return client.SendQuery(query)
	} else if reflect != "" {
		reflectType, err := parseReflect()

		if err != nil {
			return lib.RESPONSE_FAIL, err
		}

		return client.SendReflection(reflectType)
	} else {
		panic("Validation should prevent this contingency")
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
	clientCmd.AddCommand(clientPlumbingCmd)

	clientPlumbingCmd.Flags().StringVar(&reflect, "reflect", "", "Reflect on server state. (index|head|namespace)")
	clientPlumbingCmd.Flags().BoolVar(&binary, "binary", false, "Output protocol buffer binary")
	clientPlumbingCmd.Flags().BoolVar(&dryrun, "dryrun", false, "Don't send query to server")
	clientPlumbingCmd.Flags().StringVar(&source, "query", "", "Godless NoSQL query text")
	clientPlumbingCmd.Flags().BoolVar(&analyse, "analyse", false, "Analyse query")
}
