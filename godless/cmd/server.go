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

	lib "github.com/johnny-morrice/godless"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a Godless server",
	Long:  `A godless server listens to queries over HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		var kvNamespace lib.KvNamespace

		store := lib.MakeIPFSPeer(ipfsService, ipfsOffline)
		err := store.Connect()

		defer disconnect(store)

		if err != nil {
			die(err)
		}

		kvNamespace, err = makeKvNamespace(store)

		if err != nil {
			die(err)
		}

		err = serve(kvNamespace, store)

		if err != nil {
			die(err)
		}
	},
}

var addr string
var interval time.Duration

func disconnect(store lib.RemoteStore) {
	err := store.Disconnect()

	if err != nil {
		die(err)
	}
}

func serve(kvNamespace lib.KvNamespace, store lib.RemoteStore) error {
	api, apiErrCh := lib.LaunchKeyValueStore(kvNamespace)

	webService := &lib.WebService{
		API: api,
	}

	webStopCh, webErr := lib.Serve(addr, webService.Handler())

	if webErr != nil {
		return webErr
	}

	replicateStopCh, peerErrCh := lib.Replicate(api, store, interval, pubsubTopics)

	var procErr error
LOOP:
	for {
		select {
		case apiErr, ok := <-apiErrCh:
			procErr = apiErr
			break LOOP
		case peerErr := <-peerErrCh:
			procErr = peerErr
			break LOOP
		}
	}

	webStopCh <- nil
	replicateStopCh <- nil

	return procErr
}

func makeKvNamespace(store lib.RemoteStore) (lib.KvNamespace, error) {
	if hash == "" {
		namespace := lib.EmptyNamespace()
		return lib.PersistNewRemoteNamespace(store, namespace)
	} else {
		index := lib.IPFSPath(hash)
		return lib.LoadRemoteNamespace(store, index)
	}
}

func init() {
	storeCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&addr, "address", "localhost:8085", "Listen address for server")
	serveCmd.Flags().DurationVar(&interval, "interval", time.Second*1, "Interval between replications")
}
