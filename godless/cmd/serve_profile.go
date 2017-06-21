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
	"time"

	"github.com/spf13/cobra"
)

// serveProfileCmd represents the serve_profile command
var serveProfileCmd = &cobra.Command{
	Use:    "profile",
	Hidden: true,
	Short:  "Profile a godless server",
	Long:   `Run a CPU profile for the specified time and save to the specified file`,
	Run: func(cmd *cobra.Command, args []string) {
		cpuProf, err := cpuProfOutput()

		if err != nil {
			die(err)
		}

		runProfiler(cpuProf)
		readKeysFromViper()
		serve()
	},
}

var profileTime time.Duration
var cpuprof string

func init() {
	serveCmd.AddCommand(serveProfileCmd)

	serveProfileCmd.Flags().DurationVar(&profileTime, "time", time.Minute, "Duration of profile run")
	serveProfileCmd.Flags().StringVar(&cpuprof, "cpuprof", "cpu.prof", "CPU Profile output file")
}

func runProfiler(cpuProf *os.File) {
	pprof.StartCPUProfile(cpuProf)

	go func() {
		defer func() {
			pprof.StopCPUProfile()
			cpuProf.Close()
			os.Exit(0)
		}()
		timer := time.NewTimer(profileTime)
		<-timer.C
	}()
}

func cpuProfOutput() (*os.File, error) {
	return os.Create(cpuprof)
}
