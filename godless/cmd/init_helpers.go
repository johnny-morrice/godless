package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"runtime"

	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/datapeer"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/log"
)

// This file contains helpers for server initialisation.

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

func makeTempDbFile(directoryPrefix string) string {
	tempDir, err := ioutil.TempDir(__TEMP_ROOT, directoryPrefix+"_")

	if err != nil {
		die(err)
	}

	return path.Join(tempDir, __TEMP_DB_NAME)
}

func makeDataPeer(cmd *cobra.Command, params *Parameters) (api.DataPeer, error) {
	testMode := *params.Bool(__MOCK_SERVER_TESTMODE_FLAG)
	if testMode {
		return makeMemoryDataPeer()
	} else {
		return makeIpfsDataPeer(params)
	}
}

func makeIpfsDataPeer(params *Parameters) (api.DataPeer, error) {
	timeout := *params.Duration(__SERVER_TIMEOUT_FLAG)
	ipfsService := *params.String(__STORE_IPFS_FLAG)

	options := datapeer.IpfsWebServiceOptions{
		Url:  ipfsService,
		Http: http.MakeBackendHttpClient(timeout),
	}
	peer := datapeer.MakeIpfsWebService(options)
	return peer, nil
}

func makeMemoryDataPeer() (api.DataPeer, error) {
	// Use default options
	options := datapeer.ResidentMemoryStorageOptions{}
	peer := datapeer.MakeResidentMemoryDataPeer(options)
	return peer, nil
}

func makeCache(cmd *cobra.Command, params *Parameters) (api.Cache, error) {
	cacheType := *params.String(__MOCK_SERVER_CACHETYPE_FLAG)
	switch cacheType {
	case __MEMORY_CACHE_TYPE:
		return makeMemoryCache(params)
	case __BOLT_CACHE_TYPE:
		return makeBoltCache(params)
	}

	err := fmt.Errorf("Unknown cache: '%s'", cacheType)
	cmd.Help()
	die(err)
	panic("BUG")
}

func makeMemoryImage(cmd *cobra.Command, params *Parameters) (api.MemoryImage, error) {
	cacheType := *params.String(__MOCK_SERVER_CACHETYPE_FLAG)
	switch cacheType {
	case __MEMORY_CACHE_TYPE:
		return makeResidentMemoryImage()
	case __BOLT_CACHE_TYPE:
		return makeBoltMemoryImage(params)
	}

	err := fmt.Errorf("Unknown cache: '%s'", cacheType)
	cmd.Help()
	die(err)
	panic("BUG")
}

func makeMemoryCache(params *Parameters) (api.Cache, error) {
	log.Info("Using in-memory cache")
	bufferLength := *params.Int(__SERVER_BUFFER_FLAG)
	memCache := cache.MakeResidentMemoryCache(bufferLength, bufferLength)
	return memCache, nil
}

func makeResidentMemoryImage() (api.MemoryImage, error) {
	log.Info("Using in-memory MemoryImage")
	return cache.MakeResidentMemoryImage(), nil
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

const __TEMP_ROOT = "/tmp"
const __TEMP_DB_NAME = "godless.bolt"
