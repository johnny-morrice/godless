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
	gohttp "net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/query"

	"github.com/pkg/errors"
)

var clientPlumbingCmd = &cobra.Command{
	Use:   "plumbing",
	Short: "Query a godless server",
	Long:  `Send a query to a godless server over HTTP.  User must specify either --reflect or --query.`,
	// TODO tidy method.
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		client := makeClient()

		validateClientPlumbingArgs(cmd)

		readKeysFromViper()

		query := parseQuery()

		dryrun := *clientPlumbingParams.Bool(__PLUMBING_DRYRUN_FLAG)
		if dryrun {
			return
		}

		var response api.Response
		response, err = sendRequest(client, query)

		if err != nil {
			die(err)
		}

		outputResponse(response)
	},
}

var clientPlumbingParams *Parameters = &Parameters{}

func parseQuery() *query.Query {
	var q *query.Query
	var err error
	source := *clientPlumbingParams.String(__PLUMBING_QUERY_FLAG)
	if source != "" {
		q, err = query.Compile(source)

		if err != nil {
			die(err)
		}
	}

	analyse := *clientPlumbingParams.Bool(__PLUMBING_ANALYSE_FLAG)
	if analyse && q == nil {
		die(errors.New("Cannot analyse without query"))
	}

	if analyse {
		err = analyseQuery(q)

		if err != nil {
			die(err)
		}
	}

	return q
}

func makeClient() api.Client {
	timeout := *queryParams.Duration(__QUERY_TIMEOUT_FLAG)
	serverAddr := *queryParams.String(__QUERY_SERVER_FLAG)

	webClient := &gohttp.Client{
		Timeout: timeout,
	}

	options := http.ClientOptions{
		Http:       webClient,
		ServerAddr: serverAddr,
	}

	client, err := http.MakeClient(options)

	if err != nil {
		die(err)
	}

	return client
}

func outputResponse(response api.Response) {
	var err error
	queryBinary := *clientPlumbingParams.Bool(__PLUMBING_BINARY_FLAG)
	if queryBinary {
		err = api.EncodeResponse(response, os.Stdout)

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
	source := *clientPlumbingParams.String(__PLUMBING_QUERY_FLAG)
	reflect := *clientPlumbingParams.String(__PLUMBING_REFLECT_FLAG)
	replicate := *clientPlumbingParams.String(__PLUMBING_REPLICATE_FLAG)
	if source == "" && reflect == "" && replicate == "" {
		err := cmd.Help()

		if err != nil {
			die(err)
		}
	}
}

func analyseQuery(query *query.Query) error {
	format := "Query analysis for:\n\n%s\n\n%v\n\n"
	source := *clientPlumbingParams.String(__PLUMBING_QUERY_FLAG)
	fmt.Printf(format, source, query.Analyse())
	fmt.Print("Syntax tree:\n\n\n")
	query.Parser.PrintSyntaxTree()
	return query.PrettyPrint(os.Stdout)
}

func sendRequest(client api.Client, query *query.Query) (api.Response, error) {
	reflect := *clientPlumbingParams.String(__PLUMBING_REFLECT_FLAG)
	replicate := *clientPlumbingParams.String(__PLUMBING_REPLICATE_FLAG)
	if query != nil {
		request := api.MakeQueryRequest(query)
		return client.Send(request)
	} else if reflect != "" {
		reflectType, err := parseReflect()

		if err != nil {
			die(err)
		}

		request := api.MakeReflectRequest(reflectType)
		return client.Send(request)
	} else if replicate != "" {
		replicatePath := crdt.IPFSPath(replicate)
		keys := keyStore.GetAllPrivateKeys()
		link, err := crdt.SignedLink(replicatePath, keys)

		if err != nil {
			die(err)
		}

		request := api.MakeReplicateRequest([]crdt.Link{link})
		return client.Send(request)
	} else {
		panic("BUG validation should prevent this contingency")
	}
}

func parseReflect() (api.ReflectionType, error) {
	reflect := *clientPlumbingParams.String(__PLUMBING_REFLECT_FLAG)
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

	clientPlumbingCmd.Flags().StringVar(clientPlumbingParams.String(__PLUMBING_REPLICATE_FLAG), __PLUMBING_REPLICATE_FLAG, "", "Replicate index from hash")
	clientPlumbingCmd.Flags().StringVar(clientPlumbingParams.String(__PLUMBING_REFLECT_FLAG), __PLUMBING_REFLECT_FLAG, "", "Reflect on server state. (index|head|namespace)")
	clientPlumbingCmd.Flags().BoolVar(clientPlumbingParams.Bool(__PLUMBING_BINARY_FLAG), __PLUMBING_BINARY_FLAG, false, "Output protocol buffer binary")
	clientPlumbingCmd.Flags().BoolVar(clientPlumbingParams.Bool(__PLUMBING_DRYRUN_FLAG), __PLUMBING_DRYRUN_FLAG, false, "Don't send query to server")
	clientPlumbingCmd.Flags().StringVar(clientPlumbingParams.String(__PLUMBING_QUERY_FLAG), __PLUMBING_QUERY_FLAG, "", "Godless NoSQL query text")
	clientPlumbingCmd.Flags().BoolVar(clientPlumbingParams.Bool(__PLUMBING_ANALYSE_FLAG), __PLUMBING_ANALYSE_FLAG, false, "Analyse query")
}

const __PLUMBING_REPLICATE_FLAG = "replicate"
const __PLUMBING_REFLECT_FLAG = "reflect"
const __PLUMBING_BINARY_FLAG = "binary"
const __PLUMBING_DRYRUN_FLAG = "dryrun"
const __PLUMBING_QUERY_FLAG = "query"
const __PLUMBING_ANALYSE_FLAG = "analyse"
