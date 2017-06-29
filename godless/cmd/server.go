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
	"os/signal"
	"runtime"
	"time"

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/internal/http"
	"github.com/johnny-morrice/godless/log"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a Godless server",
	Long:  `A godless server listens to queries over HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		readKeysFromViper()
		serve()
	},
}

func serve() {
	client := http.DefaultBackendClient()
	client.Timeout = serverTimeout

	queue := makePriorityQueue()

	options := lib.Options{
		IpfsServiceUrl:    ipfsService,
		WebServiceAddr:    addr,
		IndexHash:         hash,
		FailEarly:         earlyConnect,
		ReplicateInterval: interval,
		Topics:            topics,
		APIQueryLimit:     apiQueryLimit,
		KeyStore:          keyStore,
		PublicServer:      publicServer,
		IpfsClient:        client,
		Pulse:             pulse,
		PriorityQueue:     queue,
	}

	godless, err := lib.New(options)

	if err != nil {
		die(err)
	}

	shutdownOnTrap(godless)

	for runError := range godless.Errors() {
		log.Error("%v", runError)
	}

	defer shutdown(godless)
}

var addr string
var interval time.Duration
var pulse time.Duration
var earlyConnect bool
var apiQueryLimit int
var apiQueueLength int
var publicServer bool
var serverTimeout time.Duration

func shutdownOnTrap(godless *lib.Godless) {
	onTrap(func(signal os.Signal) {
		log.Warn("Caught signal: %v", signal.String())
		shutdown(godless)
	})
}

func onTrap(handler func(signal os.Signal)) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, os.Kill)
	sig := <-sigch
	handler(sig)
}

func makePriorityQueue() api.RequestPriorityQueue {
	return cache.MakeResidentBufferQueue(apiQueueLength)
}

func shutdown(godless *lib.Godless) {
	godless.Shutdown()
	os.Exit(0)
}

func init() {
	storeCmd.AddCommand(serveCmd)

	defaultLimit := runtime.NumCPU()
	serveCmd.PersistentFlags().StringVar(&addr, "address", "localhost:8085", "Listen address for server")
	serveCmd.PersistentFlags().DurationVar(&interval, "synctime", __DEFAULT_REPLICATION_INTERVAL, "Interval between peer replications")
	serveCmd.PersistentFlags().DurationVar(&pulse, "pulse", __DEFAULT_PULSE, "Interval between writes to IPFS")
	serveCmd.PersistentFlags().BoolVar(&earlyConnect, "early", false, "Early check on IPFS API access")
	serveCmd.PersistentFlags().IntVar(&apiQueryLimit, "limit", defaultLimit, "Number of simulataneous queries run by the API. limit < 0 for no restrictions.")
	serveCmd.PersistentFlags().BoolVar(&publicServer, "public", false, "Don't limit pubsub updates to the public key list")
	serveCmd.PersistentFlags().DurationVar(&serverTimeout, "timeout", __DEFAULT_SERVER_TIMEOUT, "Timeout for serverside HTTP queries")
	serveCmd.PersistentFlags().IntVar(&apiQueueLength, "qlength", __DEFAULT_QUEUE_LENGTH, "API Priority queue length")
}

const __DEFAULT_SERVER_TIMEOUT = time.Minute * 10
const __DEFAULT_QUEUE_LENGTH = 4096
const __DEFAULT_PULSE = time.Second
const __DEFAULT_REPLICATION_INTERVAL = time.Minute
