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
	"time"

	"github.com/spf13/cobra"
)

// queryCmd represents the client command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Godless frontend",
	Long: `godless query provides a frontend to the godless p2p node.  For a terminal console, do:

	godless query console`,
}

func init() {
	RootCmd.AddCommand(queryCmd)

	queryCmd.PersistentFlags().StringVar(&serverAddr, "server", __DEFAULT_QUERY_SERVER, "Server address")
	queryCmd.PersistentFlags().DurationVar(&queryTimeout, "timeout", __DEFAULT_QUERY_TIMEOUT, "Query timeout")
}

var serverAddr string
var queryTimeout time.Duration

const __DEFAULT_QUERY_TIMEOUT = time.Minute
const __DEFAULT_QUERY_SERVER = "http://localhost:8085"
