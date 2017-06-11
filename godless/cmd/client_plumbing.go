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

	"github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/query"
)

// clientPlumbingCmd represents the query command
var clientPlumbingCmd = &cobra.Command{
	Use:   "plumbing",
	Short: "Query a godless server",
	Long:  `Send a query to a godless server over HTTP.  User must specify either --reflect or --query.`,
	// TODO tidy method.
	Run: func(cmd *cobra.Command, args []string) {
		var q *query.Query
		var err error
		client := godless.MakeClient(serverAddr)

		validateClientPlumbingArgs(cmd)

		if source != "" || analyse {
			q, err = query.CompileQuery(source)

			if err != nil {
				die(err)
			}
		}

		if analyse {
			err = analyseQuery(q)

			if err != nil {
				die(err)
			}
		}

		if dryrun {
			return
		}

		var response api.APIResponse
		response, err = sendQuery(client, q)

		if err != nil {
			die(err)
		}

		outputResponse(response)
	},
}

var source string
var analyse bool
var dryrun bool
var queryBinary bool
var reflect string

func outputResponse(response api.APIResponse) {
	var err error
	if queryBinary {
		err = api.EncodeAPIResponse(response, os.Stdout)

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

func analyseQuery(query *query.Query) error {
	format := "Query analysis for:\n\n%s\n\n%v\n\n"
	fmt.Printf(format, source, query.Analyse())
	fmt.Println("Syntax tree:\n\n")
	query.Parser.PrintSyntaxTree()
	return query.PrettyPrint(os.Stdout)
}

func sendQuery(client godless.Client, query *query.Query) (api.APIResponse, error) {
	if query != nil {
		return client.SendQuery(query)
	} else if reflect != "" {
		reflectType, err := parseReflect()

		if err != nil {
			return api.RESPONSE_FAIL, err
		}

		return client.SendReflection(reflectType)
	} else {
		panic("Validation should prevent this contingency")
	}
}

func parseReflect() (api.APIReflectionType, error) {
	switch reflect {
	case "index":
		return api.REFLECT_INDEX, nil
	case "head":
		return api.REFLECT_HEAD_PATH, nil
	case "namespace":
		return api.REFLECT_DUMP_NAMESPACE, nil
	default:
		return api.REFLECT_NOOP, fmt.Errorf("Unknown reflect type: %v", reflect)
	}
}

func init() {
	queryCmd.AddCommand(clientPlumbingCmd)

	clientPlumbingCmd.Flags().StringVar(&reflect, "reflect", "", "Reflect on server state. (index|head|namespace)")
	clientPlumbingCmd.Flags().BoolVar(&queryBinary, "binary", false, "Output protocol buffer binary")
	clientPlumbingCmd.Flags().BoolVar(&dryrun, "dryrun", false, "Don't send query to server")
	clientPlumbingCmd.Flags().StringVar(&source, "query", "", "Godless NoSQL query text")
	clientPlumbingCmd.Flags().BoolVar(&analyse, "analyse", false, "Analyse query")
}
