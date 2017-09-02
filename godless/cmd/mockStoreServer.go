// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/datapeer"
	"github.com/johnny-morrice/godless/http"
)

var mockStoreServeCmd = &cobra.Command{
	Use:   "server",
	Short: "Mock version of `store server`",
	Run: func(cmd *cobra.Command, args []string) {
		readKeysFromViper()
		serve(mockServeOptions(cmd))
	},
}

func mockServeOptions(cmd *cobra.Command) lib.Options {
	client := http.MakeBackendHttpClient(serverTimeout)

	queue := makePriorityQueue()
	memimg, err := makeMockMemoryImage(cmd)

	if err != nil {
		die(err)
	}

	cache, err := makeMockCache(cmd)

	if err != nil {
		die(err)
	}

	peer, err := makeMockDataPeer(cmd)

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
	}
}

func makeMockDataPeer(cmd *cobra.Command) (api.DataPeer, error) {
	if testMode {
		return makeMemoryDataPeer()
	} else {
		return makeIpfsDataPeer()
	}
}

func makeIpfsDataPeer() (api.DataPeer, error) {
	options := datapeer.IpfsWebServiceOptions{
		Url:  ipfsService,
		Http: makeBackendClient(),
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

func makeMockCache(cmd *cobra.Command) (api.Cache, error) {
	switch cacheType {
	case __MEMORY_CACHE_TYPE:
		return makeMemoryCache()
	case __BOLT_CACHE_TYPE:
		return makeBoltCache()
	}

	err := fmt.Errorf("Unknown cache: '%s'", cacheType)
	cmd.Help()
	die(err)
	panic("BUG")
}

func makeMockMemoryImage(cmd *cobra.Command) (api.MemoryImage, error) {
	switch cacheType {
	case __MEMORY_CACHE_TYPE:
		return makeBoltMemoryImage()
	case __BOLT_CACHE_TYPE:
		return makeResidentMemoryImage()
	}

	err := fmt.Errorf("Unknown cache: '%s'", cacheType)
	cmd.Help()
	die(err)
	panic("BUG")
}

func makeMemoryCache() (api.Cache, error) {
	memCache := cache.MakeResidentMemoryCache(bufferLength, bufferLength)
	return memCache, nil
}

func makeResidentMemoryImage() (api.MemoryImage, error) {
	return cache.MakeResidentMemoryImage(), nil
}

var cacheType string
var testMode bool
var testDatabaseFilePath string

func init() {
	mockStoreCmd.AddCommand(mockStoreServeCmd)

	// TODO implement temp path
	tempDbPath := "/tmp/GODLESS_MOCK_DB.bolt"

	mockStoreServeCmd.PersistentFlags().BoolVar(&testMode, "testmode", true, "Fake it till you make it")
	mockStoreServeCmd.PersistentFlags().StringVar(&cacheType, "cachetype", __DEFAULT_CACHE_TYPE, "Cache type")
	mockStoreServeCmd.PersistentFlags().StringVar(&addr, "address", __DEFAULT_LISTEN_ADDR, "Listen address for server")
	mockStoreServeCmd.PersistentFlags().DurationVar(&interval, "synctime", __DEFAULT_REPLICATION_INTERVAL, "Interval between peer replications")
	mockStoreServeCmd.PersistentFlags().DurationVar(&pulse, "pulse", __DEFAULT_PULSE, "Interval between writes to IPFS")
	mockStoreServeCmd.PersistentFlags().BoolVar(&earlyConnect, "early", __DEFAULT_EARLY_CONNECTION, "Early check on IPFS API access")
	mockStoreServeCmd.PersistentFlags().IntVar(&apiQueryLimit, "concurrent", defaultConcurrency(), "Number of simulataneous queries run by the API. limit < 0 for no restrictions.")
	mockStoreServeCmd.PersistentFlags().BoolVar(&publicServer, "public", __DEFAULT_SERVER_PUBLIC_STATUS, "Don't limit pubsub updates to the public key list")
	mockStoreServeCmd.PersistentFlags().DurationVar(&serverTimeout, "timeout", __DEFAULT_SERVER_TIMEOUT, "Timeout for serverside HTTP queries")
	mockStoreServeCmd.PersistentFlags().IntVar(&apiQueueLength, "qlength", __DEFAULT_QUEUE_LENGTH, "API Priority queue length")
	mockStoreServeCmd.PersistentFlags().IntVar(&bufferLength, "buffer", __DEFAULT_BUFFER_LENGTH, "Buffer length")
	// BUG reusing this variable.
	// Cobra doesn't wait till a command is run to insert the default value, it happens right away.
	// We should create a parameter data structure for each command.
	// There might be other variants of this bug lurking.
	mockStoreServeCmd.PersistentFlags().StringVar(&testDatabaseFilePath, "dbpath", tempDbPath, "Embedded database file path")
}
