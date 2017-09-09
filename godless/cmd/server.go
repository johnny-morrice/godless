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
	"path"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/log"
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

func serve(options lib.Options) {
	godless, err := lib.New(options)
	defer shutdown(godless)

	if err != nil {
		die(err)
	}

	shutdownOnTrap(godless)

	for runError := range godless.Errors() {
		log.Error("%s", runError.Error())
	}
}

func makeBoltCache(params *Parameters) (api.Cache, error) {
	factory := getBoltFactoryInstance(params)
	return factory.MakeCache()
}

func makeBoltMemoryImage(params *Parameters) (api.MemoryImage, error) {
	factory := getBoltFactoryInstance(params)
	return factory.MakeMemoryImage()
}

var boltFactory *cache.BoltFactory

func getBoltFactoryInstance(params *Parameters) *cache.BoltFactory {
	if boltFactory == nil {
		databaseFilePath := *params.String(__SERVER_DATABASE_FLAG)
		log.Info("Using database file: '%s'", databaseFilePath)
		options := cache.BoltOptions{
			FilePath: databaseFilePath,
			Mode:     0600,
		}
		factory, err := cache.MakeBoltFactory(options)

		if err != nil {
			die(err)
		}

		boltFactory = &factory
	}

	return boltFactory
}

func shutdownOnTrap(godless *lib.Godless) {
	installTrapHandler(func(signal os.Signal) {
		log.Warn("Caught signal: %s", signal.String())
		go func() {
			shutdown(godless)
		}()
	})
}

func installTrapHandler(handler func(signal os.Signal)) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, os.Kill)
	sig := <-sigch
	signal.Reset(os.Interrupt, os.Kill)
	handler(sig)
}

func makePriorityQueue(params *Parameters) api.RequestPriorityQueue {
	apiQueueLength := *params.Int(__SERVER_QUEUE_FLAG)
	return cache.MakeResidentBufferQueue(apiQueueLength)
}

func shutdown(godless *lib.Godless) {
	godless.Shutdown()
	os.Exit(0)
}

func homePath(relativePath string) string {
	home := os.Getenv("HOME")
	return path.Join(home, relativePath)
}

func defaultConcurrency() int {
	return runtime.NumCPU()
}

var storeServerParams *Parameters = &Parameters{}

func init() {
	storeCmd.AddCommand(serveCmd)

	defaultBoltDb := homePath(__DEFAULT_BOLT_DB_FILENAME)

	addServerParams(serveCmd, storeServerParams, defaultBoltDb)
}

func addServerParams(cmd *cobra.Command, params *Parameters, dbPath string) {
	cmd.PersistentFlags().StringVar(params.String(__SERVER_ADDR_FLAG), __SERVER_ADDR_FLAG, __DEFAULT_LISTEN_ADDR, "Listen address for server")
	cmd.PersistentFlags().DurationVar(params.Duration(__SERVER_SYNC_FLAG), __SERVER_SYNC_FLAG, __DEFAULT_REPLICATION_INTERVAL, "Interval between peer replications")
	cmd.PersistentFlags().DurationVar(params.Duration(__SERVER_PULSE_FLAG), __SERVER_PULSE_FLAG, __DEFAULT_PULSE, "Interval between writes to IPFS")
	cmd.PersistentFlags().BoolVar(params.Bool(__SERVER_EARLY_FLAG), __SERVER_EARLY_FLAG, __DEFAULT_EARLY_CONNECTION, "Early check on IPFS API access")
	cmd.PersistentFlags().IntVar(params.Int(__SERVER_CONCURRENT_FLAG), __SERVER_CONCURRENT_FLAG, defaultConcurrency(), "Number of simulataneous queries run by the API. limit < 0 for no restrictions.")
	cmd.PersistentFlags().BoolVar(params.Bool(__SERVER_PUBLIC_FLAG), __SERVER_PUBLIC_FLAG, __DEFAULT_SERVER_PUBLIC_STATUS, "Don't limit pubsub updates to the public key list")
	cmd.PersistentFlags().DurationVar(params.Duration(__SERVER_TIMEOUT_FLAG), __SERVER_TIMEOUT_FLAG, __DEFAULT_SERVER_TIMEOUT, "Timeout for serverside HTTP queries")
	cmd.PersistentFlags().IntVar(params.Int(__SERVER_QUEUE_FLAG), __SERVER_QUEUE_FLAG, __DEFAULT_QUEUE_LENGTH, "API Priority queue length")
	cmd.PersistentFlags().IntVar(params.Int(__SERVER_BUFFER_FLAG), __SERVER_BUFFER_FLAG, __DEFAULT_BUFFER_LENGTH, "Buffer length")
	cmd.PersistentFlags().StringVar(params.String(__SERVER_DATABASE_FLAG), __SERVER_DATABASE_FLAG, dbPath, "Embedded database file path")
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
