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
	"github.com/spf13/cobra"
	lib "github.com/johnny-morrice/godless"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run a Godless server",
	Long: `A godless server listens to queries over HTTP.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var kvNamespace lib.KvNamespace
		var stopch chan<- interface{}

		peer := lib.MakeIPFSPeer(peerUrl)
		err = peer.Connect()
		defer peer.Disconnect()

		if err != nil {
			die(err)
		}

		if hash == "" {
			namespace := lib.MakeNamespace()
			kvNamespace, err = lib.PersistNewRemoteNamespace(peer, namespace)
		} else {
			index := lib.IPFSPath(hash)
			kvNamespace, err = lib.LoadRemoteNamespace(peer, index)
		}

		if err != nil {
			die(err)
		}

		api, errch := lib.LaunchKeyValueStore(kvNamespace)

		service := &lib.KeyValueService{
			API: api,
		}

		stopch, err = lib.Serve(addr, service.Handler())

		if err != nil {
			die(err)
		}

		err = <-errch
		stopch<- nil

		die(err)
	},
}

var addr string
var hash string
var peerUrl string

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&addr, "address", "localhost:8085", "Listen address for server")
	serveCmd.Flags().StringVar(&hash, "hash", "", "IPFS hash of namespace head")
	serveCmd.Flags().StringVar(&peerUrl, "peer", "http://localhost:5001", "IPFS peer URL")
}
