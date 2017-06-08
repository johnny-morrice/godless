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
)

// storeCmd represents the store command
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "The godless data store",
	Long: `Godless store commands to serve and inspect godless data.  To run a godless server, do:

	godless store serve`,
	Run: func(cmd *cobra.Command, args []string) {
		readPeers()

		err := cmd.Help()

		if err != nil {
			die(err)
		}
	},
}

var indexHash string
var peerNames []string
var peers []lib.RemoteStoreAddress
var ipfsService string

func init() {
	RootCmd.AddCommand(storeCmd)

	storeCmd.PersistentFlags().StringVar(&indexHash, "index", "", "IPFS hash of godless index")
	storeCmd.PersistentFlags().StringSliceVar(&peerNames, "peers", []string{}, "Comma separated list of IPNS peer names")
	storeCmd.PersistentFlags().StringVar(&ipfsService, "ipfs", "http://localhost:5001", "IPFS webservice URL")
}

func readPeers() {
	peers = make([]lib.RemoteStoreAddress, len(peerNames))

	for i, n := range peerNames {
		peers[i] = lib.IPFSPath(n)
	}
}
