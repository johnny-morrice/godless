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
	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/http"
)

// benchStoreServerCmd represents the benchStoreServer command
var benchStoreServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Benchmark `store server` command",

	Run: func(cmd *cobra.Command, args []string) {
		readKeysFromViper()
		serve(benchServeOptions(cmd))
	},
}

func benchServeOptions(cmd *cobra.Command) lib.Options {
	params := benchStoreParams.Merge(benchStoreServerParams)
	serverTimeout := *params.Duration(__SERVER_TIMEOUT_FLAG)
	addr := *params.String(__SERVER_ADDR_FLAG)
	earlyConnect := *params.Bool(__SERVER_EARLY_FLAG)
	interval := *params.Duration(__SERVER_SYNC_FLAG)
	apiQueryLimit := *params.Int(__SERVER_CONCURRENT_FLAG)
	publicServer := *params.Bool(__SERVER_PUBLIC_FLAG)
	pulse := *params.Duration(__SERVER_PULSE_FLAG)
	client := http.MakeBackendHttpClient(serverTimeout)

	hash := *params.String(__STORE_HASH_FLAG)
	topics := *params.StringSlice(__STORE_TOPICS_FLAG)

	queue := makePriorityQueue(benchStoreServerParams)
	webService := makeBenchWebService(benchStoreServerParams)

	memimg, err := makeBoltMemoryImage(benchStoreServerParams)

	if err != nil {
		die(err)
	}

	cache, err := makeBoltCache(benchStoreServerParams)

	if err != nil {
		die(err)
	}

	peer, err := makeBenchDataPeer(cmd, benchStoreServerParams)

	if err != nil {
		die(err)
	}

	return lib.Options{
		DataPeer:          peer,
		WebServiceAddr:    addr,
		IndexHash:         hash,
		FailEarly:         earlyConnect,
		ReplicateInterval: interval,
		Topics:            topics,
		ApiConcurrency:    apiQueryLimit,
		KeyStore:          keyStore,
		PublicServer:      publicServer,
		IpfsClient:        client,
		Pulse:             pulse,
		PriorityQueue:     queue,
		Cache:             cache,
		MemoryImage:       memimg,
		WebService:        webService,
	}
}

func makeBenchWebService(params *Parameters) api.WebService {
	panic("not implemented")
}

func makeBenchDataPeer(cmd *cobra.Command, params *Parameters) (api.DataPeer, error) {
	panic("not implemented")
}

var benchStoreServerParams *Parameters = &Parameters{}

func init() {
	benchStoreCmd.AddCommand(benchStoreServerCmd)

	dbPath := makeTempDbFile(__BENCH_DB_DIR_PREFIX)
	addServerParams(benchStoreServerCmd, benchStoreServerParams, dbPath)
}

const __BENCH_DB_DIR_PREFIX = "godless_bench_store_server"
