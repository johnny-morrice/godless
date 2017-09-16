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
	"runtime/pprof"

	"github.com/spf13/cobra"

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/datapeer"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/prof"
)

// benchStoreServerCmd represents the benchStoreServer command
var benchStoreServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Benchmark `store server` command",

	Run: func(cmd *cobra.Command, args []string) {
		readKeysFromViper()
		// TODO parameter validation.

		// FIXME Really messy how we're passing both the file and the profiler around...
		detailProfiler, detailFile := makeDetailProfiler(benchStoreServerParams)
		options := benchServeOptions(cmd, detailProfiler)
		benchServe(options, benchStoreServerParams, detailFile)
	},
}

func benchServe(options lib.Options, params *Parameters, detailFile *os.File) {
	godless, err := lib.New(options)
	defer shutdown(godless)

	if err != nil {
		die(err)
	}

	cpuProfileFile := startProfiling(params)

	shutdownBenchmarkServerOnTrap(godless, cpuProfileFile, detailFile)

	for runError := range godless.Errors() {
		log.Error("%s", runError.Error())
	}
}

func benchServeOptions(cmd *cobra.Command, profiler api.Profiler) lib.Options {
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

	queue := makePriorityQueue(params)

	webService := makeBenchWebService(params, profiler)

	memimg, err := makeBoltMemoryImage(params)

	if err != nil {
		die(err)
	}

	cache, err := makeBoltCache(params)

	if err != nil {
		die(err)
	}

	peer, err := makeBenchDataPeer(cmd, params, profiler)

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

func makeDetailProfiler(params *Parameters) (api.Profiler, *os.File) {
	filePath := *params.String(__BENCH_PROFILE_DETAIL_FILE_FLAG)
	file, err := os.Create(filePath)

	if err != nil {
		die(err)
	}

	profiler := prof.MakeWriterProfiler(file)
	return profiler, file
}

func startProfiling(params *Parameters) *os.File {
	file, err := createCPUProfileFile(params)

	if err != nil {
		die(err)
	}

	pprof.StartCPUProfile(file)

	return file
}

func createCPUProfileFile(params *Parameters) (*os.File, error) {
	return os.Create(*params.String(__BENCH_CPU_PROFILE_FILE_FLAG))
}

func shutdownBenchmarkServerOnTrap(godless *lib.Godless, cpuProfileFile *os.File, detailProfileFile *os.File) {
	installTrapHandler(func(signal os.Signal) {
		go func() {
			log.Warn("Caught signal: %s", signal.String())
			pprof.StopCPUProfile()
			// FIXME handle Close() errors
			detailProfileFile.Close()
			cpuProfileFile.Close()
			shutdown(godless)
		}()

	})
}

func makeBenchWebService(params *Parameters, profiler api.Profiler) api.WebService {
	webService := http.MakeWebService(api.WebServiceOptions{})
	profOptions := http.ProfilingWebServiceOptions{
		WS:   webService,
		Prof: profiler,
	}
	return http.MakeProfilingWebService(profOptions)
}

func makeBenchDataPeer(cmd *cobra.Command, params *Parameters, profiler api.Profiler) (api.DataPeer, error) {
	peer, err := makeIpfsDataPeer(params)

	if err != nil {
		return nil, err
	}

	profOptions := datapeer.ProfilingDataPeerOptions{Peer: peer, Prof: profiler}
	profPeer := datapeer.MakeProfilingDataPeer(profOptions)
	return profPeer, nil
}

var benchStoreServerParams *Parameters = &Parameters{}

func init() {
	benchStoreCmd.AddCommand(benchStoreServerCmd)

	dbPath := makeTempDbFile(__BENCH_DB_DIR_PREFIX)
	addServerParams(benchStoreServerCmd, benchStoreServerParams, dbPath)

	benchStoreServerCmd.PersistentFlags().StringVar(benchStoreServerParams.String(__BENCH_CPU_PROFILE_FILE_FLAG), __BENCH_CPU_PROFILE_FILE_FLAG, __DEFAULT_BENCH_CPU_PROFILE_FILE, "CPU profile output file")
	benchStoreServerCmd.PersistentFlags().StringVar(benchStoreServerParams.String(__BENCH_PROFILE_DETAIL_FILE_FLAG), __BENCH_PROFILE_DETAIL_FILE_FLAG, __DEFAULT_BENCH_PROFILE_DETAIL_FILE, "Detailed profile output file")
}

const __BENCH_CMD_NAME = "godless_bench_store_server"
const __BENCH_DB_DIR_PREFIX = __BENCH_CMD_NAME
const __BENCH_CPU_PROFILE_FILE_FLAG = "cpu-profile"
const __BENCH_PROFILE_DETAIL_FILE_FLAG = "detail-profile"
const __DEFAULT_BENCH_CPU_PROFILE_FILE = __BENCH_CMD_NAME + ".cpu.prof"
const __DEFAULT_BENCH_PROFILE_DETAIL_FILE = __BENCH_CMD_NAME + "_profile_detail.timelog"
