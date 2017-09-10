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
	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
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
	params := mockStoreParams.Merge(mockStoreServerParams)
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

	queue := makePriorityQueue(mockStoreServerParams)
	memimg, err := makeMemoryImage(cmd, mockStoreServerParams)

	if err != nil {
		die(err)
	}

	cache, err := makeCache(cmd, mockStoreServerParams)

	if err != nil {
		die(err)
	}

	peer, err := makeDataPeer(cmd, mockStoreServerParams)

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

var mockStoreServerParams *Parameters = &Parameters{}

func init() {
	mockStoreCmd.AddCommand(mockStoreServeCmd)

	// TODO implement temp path
	tempDbPath := "/tmp/GODLESS_MOCK_DB.bolt"

	mockStoreServeCmd.PersistentFlags().BoolVar(mockStoreServerParams.Bool(__MOCK_SERVER_TESTMODE_FLAG), __MOCK_SERVER_TESTMODE_FLAG, true, "Fake it till you make it")
	mockStoreServeCmd.PersistentFlags().StringVar(mockStoreServerParams.String(__MOCK_SERVER_CACHETYPE_FLAG), __MOCK_SERVER_CACHETYPE_FLAG, __MEMORY_CACHE_TYPE, "Cache type")
	addServerParams(mockStoreCmd, mockStoreServerParams, tempDbPath)
}

const __MOCK_SERVER_TESTMODE_FLAG = "testmode"
const __MOCK_SERVER_CACHETYPE_FLAG = "cachetype"
const __MEMORY_CACHE_TYPE = "memory"
const __BOLT_CACHE_TYPE = "disk"
