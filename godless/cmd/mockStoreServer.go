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

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
)

var mockStoreServeCmd = &cobra.Command{
	Use:   "server",
	Short: "Mock version of `store server`",
	Run: func(cmd *cobra.Command, args []string) {
		panic("mock store server not implemented")
	},
}

func makeMockCache(cmd *cobra.Command) (api.Cache, error) {
	switch cacheType {
	case __MEMORY_CACHE_TYPE:
		return makeMemoryCache()
	case __BOLT_CACHE_TYPE:
		return makeBoltCache()
	default:
		err := fmt.Errorf("Unknown cache: '%s'", cacheType)
		cmd.Help()
		die(err)
	}

	return nil, fmt.Errorf("Bug in makeServerCache")
}

func makeMemoryCache() (api.Cache, error) {
	memCache := cache.MakeResidentMemoryCache(bufferLength, bufferLength)
	return memCache, nil
}

var cacheType string

func init() {
	mockStoreCmd.AddCommand(mockStoreServeCmd)

	// TODO implement temp path
	tempDbPath := "/tmp/GODLESS_MOCK_DB"

	mockStoreServeCmd.PersistentFlags().StringVar(&addr, "address", __DEFAULT_LISTEN_ADDR, "Listen address for server")
	mockStoreServeCmd.PersistentFlags().DurationVar(&interval, "synctime", __DEFAULT_REPLICATION_INTERVAL, "Interval between peer replications")
	mockStoreServeCmd.PersistentFlags().DurationVar(&pulse, "pulse", __DEFAULT_PULSE, "Interval between writes to IPFS")
	mockStoreServeCmd.PersistentFlags().BoolVar(&earlyConnect, "early", __DEFAULT_EARLY_CONNECTION, "Early check on IPFS API access")
	mockStoreServeCmd.PersistentFlags().IntVar(&apiQueryLimit, "concurrent", defaultConcurrency(), "Number of simulataneous queries run by the API. limit < 0 for no restrictions.")
	mockStoreServeCmd.PersistentFlags().BoolVar(&publicServer, "public", __DEFAULT_SERVER_PUBLIC_STATUS, "Don't limit pubsub updates to the public key list")
	mockStoreServeCmd.PersistentFlags().DurationVar(&serverTimeout, "timeout", __DEFAULT_SERVER_TIMEOUT, "Timeout for serverside HTTP queries")
	mockStoreServeCmd.PersistentFlags().IntVar(&apiQueueLength, "qlength", __DEFAULT_QUEUE_LENGTH, "API Priority queue length")
	mockStoreServeCmd.PersistentFlags().IntVar(&bufferLength, "buffer", __DEFAULT_BUFFER_LENGTH, "Buffer length")
	mockStoreServeCmd.PersistentFlags().StringVar(&databaseFilePath, "dbpath", tempDbPath, "Embedded database file path")
}
