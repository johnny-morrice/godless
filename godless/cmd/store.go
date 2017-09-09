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

import "github.com/spf13/cobra"

// storeCmd represents the store command
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "The godless data store",
	Long: `Godless store commands to serve and inspect godless data.  To run a godless server, do:

	godless store serve`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()

		if err != nil {
			die(err)
		}
	},
}

var storeParams *Parameters = &Parameters{}

func init() {
	RootCmd.AddCommand(storeCmd)

	addStoreParams(storeCmd, storeParams)
}

func addStoreParams(cmd *cobra.Command, params *Parameters) {
	cmd.PersistentFlags().StringVar(params.String(__STORE_HASH_FLAG), __STORE_HASH_FLAG, "", "IPFS hash")
	cmd.PersistentFlags().StringSliceVar(params.StringSlice(__STORE_TOPICS_FLAG), __STORE_TOPICS_FLAG, []string{}, "Comma separated list of pubsub topics")
	cmd.PersistentFlags().StringVar(params.String(__STORE_IPFS_FLAG), __STORE_IPFS_FLAG, "http://localhost:5001", "IPFS webservice URL")
}

const __STORE_HASH_FLAG = "hash"
const __STORE_TOPICS_FLAG = "topics"
const __STORE_IPFS_FLAG = "ipfs"
