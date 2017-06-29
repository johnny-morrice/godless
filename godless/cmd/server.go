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
		serve(cmd)
	},
}

func serve(cmd *cobra.Command) {
	client := http.DefaultBackendClient()
	client.Timeout = serverTimeout

	queue := makePriorityQueue(cmd)
	memimg, err := makeMemoryImage()
	cache, err := makeCache(cmd)

	if err != nil {
		die(err)
	}

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
		Cache:             cache,
		MemoryImage:       memimg,
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
var memoryBufferLength int
var publicServer bool
var serverTimeout time.Duration
var cacheType string
var databaseFilePath string
var boltFactory *cache.BoltFactory

func makeCache(cmd *cobra.Command) (api.Cache, error) {
	switch cacheType {
	case __MEMORY_CACHE_TYPE:
		return makeMemoryCache()
	case __BOLT_CACHE_TYPE:
		return makeBoltCache()
	default:
		err := fmt.Errorf("Unknown cache: '%v'", cacheType)
		cmd.Help()
		die(err)
	}

	return nil, fmt.Errorf("Bug in makeCache")
}

func makeBoltCache() (api.Cache, error) {
	factory := getBoltFactoryInstance()
	return factory.MakeCache()
}

func makeMemoryCache() (api.Cache, error) {
	memCache := cache.MakeResidentMemoryCache(memoryBufferLength, memoryBufferLength)
	return memCache, nil
}

func makeMemoryImage() (api.MemoryImage, error) {
	factory := getBoltFactoryInstance()
	return factory.MakeMemoryImage()
}

func getBoltFactoryInstance() *cache.BoltFactory {
	if boltFactory == nil {
		options := cache.BoltOptions{
			FilePath: databaseFilePath,
			Mode:     0600,
		}
		factory, err := cache.MakeBoltCacheFactory(options)

		if err != nil {
			die(err)
		}

		boltFactory = &factory
	}

	return boltFactory
}

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

func makePriorityQueue(cmd *cobra.Command) api.RequestPriorityQueue {
	return cache.MakeResidentBufferQueue(apiQueueLength)
}

func shutdown(godless *lib.Godless) {
	godless.Shutdown()
	os.Exit(0)
}

func init() {
	storeCmd.AddCommand(serveCmd)

	defaultLimit := runtime.NumCPU()
	serveCmd.PersistentFlags().StringVar(&addr, "address", __DEFAULT_LISTEN_ADDR, "Listen address for server")
	serveCmd.PersistentFlags().DurationVar(&interval, "synctime", __DEFAULT_REPLICATION_INTERVAL, "Interval between peer replications")
	serveCmd.PersistentFlags().DurationVar(&pulse, "pulse", __DEFAULT_PULSE, "Interval between writes to IPFS")
	serveCmd.PersistentFlags().BoolVar(&earlyConnect, "early", __DEFAULT_EARLY_CONNECTION, "Early check on IPFS API access")
	serveCmd.PersistentFlags().IntVar(&apiQueryLimit, "limit", defaultLimit, "Number of simulataneous queries run by the API. limit < 0 for no restrictions.")
	serveCmd.PersistentFlags().BoolVar(&publicServer, "public", __DEFAULT_SERVER_PUBLIC_STATUS, "Don't limit pubsub updates to the public key list")
	serveCmd.PersistentFlags().DurationVar(&serverTimeout, "timeout", __DEFAULT_SERVER_TIMEOUT, "Timeout for serverside HTTP queries")
	serveCmd.PersistentFlags().IntVar(&apiQueueLength, "qlength", __DEFAULT_QUEUE_LENGTH, "API Priority queue length")
	serveCmd.PersistentFlags().StringVar(&cacheType, "cache", __DEFAULT_CACHE_TYPE, "Cache type (disk|memory)")
	serveCmd.PersistentFlags().IntVar(&memoryBufferLength, "buffer", __DEFAULT_MEMORY_BUFFER_LENGTH, "Buffer length if using memory cache")
	serveCmd.PersistentFlags().StringVar(&databaseFilePath, "dbpath", __DEFAULT_BOLT_DB_PATH, "Embedded database file path")
}

const __MEMORY_CACHE_TYPE = "memory"
const __BOLT_CACHE_TYPE = "disk"

const __DEFAULT_BOLT_DB_PATH = "godless.bolt"
const __DEFAULT_EARLY_CONNECTION = false
const __DEFAULT_SERVER_PUBLIC_STATUS = false
const __DEFAULT_CACHE_TYPE = __MEMORY_CACHE_TYPE
const __DEFAULT_LISTEN_ADDR = "localhost:8085"
const __DEFAULT_SERVER_TIMEOUT = time.Minute * 10
const __DEFAULT_QUEUE_LENGTH = 4096
const __DEFAULT_PULSE = time.Second
const __DEFAULT_REPLICATION_INTERVAL = time.Minute
const __DEFAULT_MEMORY_BUFFER_LENGTH = -1
