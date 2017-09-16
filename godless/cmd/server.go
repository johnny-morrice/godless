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

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/http"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a Godless server",
	Long:  `A godless server listens to queries over HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		readKeysFromViper()
		serve(serveOptions())
	},
}

func serveOptions() lib.Options {
	params := storeServerParams.Merge(storeParams)
	addr := *params.String(__SERVER_ADDR_FLAG)
	earlyConnect := *params.Bool(__SERVER_EARLY_FLAG)
	interval := *params.Duration(__SERVER_SYNC_FLAG)
	apiQueryLimit := *params.Int(__SERVER_CONCURRENT_FLAG)
	publicServer := *params.Bool(__SERVER_PUBLIC_FLAG)
	pulse := *params.Duration(__SERVER_PULSE_FLAG)
	serverTimeout := *params.Duration(__SERVER_TIMEOUT_FLAG)
	ipfsService := *params.String(__STORE_IPFS_FLAG)
	hash := *params.String(__STORE_HASH_FLAG)
	topics := *params.StringSlice(__STORE_TOPICS_FLAG)

	client := http.MakeBackendHttpClient(serverTimeout)

	queue := makePriorityQueue(storeServerParams)
	memimg, err := makeBoltMemoryImage(storeServerParams)

	if err != nil {
		die(err)
	}

	cache, err := makeBoltCache(storeServerParams)

	if err != nil {
		die(err)
	}

	return lib.Options{
		IpfsServiceUrl:    ipfsService,
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
	}
}

var storeServerParams *Parameters = &Parameters{}

func init() {
	storeCmd.AddCommand(serveCmd)

	defaultBoltDb := homePath(__DEFAULT_BOLT_DB_FILENAME)

	addServerParams(serveCmd, storeServerParams, defaultBoltDb)
}

const __SERVER_ADDR_FLAG = "address"
const __SERVER_SYNC_FLAG = "synctime"
const __SERVER_PULSE_FLAG = "pulse"
const __SERVER_EARLY_FLAG = "early"
const __SERVER_CONCURRENT_FLAG = "concurrent"
const __SERVER_PUBLIC_FLAG = "public"
const __SERVER_TIMEOUT_FLAG = "timeout"
const __SERVER_QUEUE_FLAG = "qlength"
const __SERVER_BUFFER_FLAG = "buffer"
const __SERVER_DATABASE_FLAG = "dbpath"

const __DEFAULT_BOLT_DB_FILENAME = ".godless.bolt"
const __DEFAULT_EARLY_CONNECTION = false
const __DEFAULT_SERVER_PUBLIC_STATUS = false
const __DEFAULT_CACHE_TYPE = __BOLT_CACHE_TYPE
const __DEFAULT_LISTEN_ADDR = "localhost:8085"
const __DEFAULT_SERVER_TIMEOUT = time.Minute * 10
const __DEFAULT_QUEUE_LENGTH = 4096
const __DEFAULT_PULSE = time.Second * 10
const __DEFAULT_REPLICATION_INTERVAL = time.Minute
const __DEFAULT_BUFFER_LENGTH = -1
