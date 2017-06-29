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

	"github.com/johnny-morrice/godless/crypto"
)

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage godless keys",
	Long: `Godless signs your data using strong cryptography.

	These signatures are used to maintain data consistency, rather than via
	physically isolating the data within a host machine.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()

		if err != nil {
			die(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(keyCmd)
}

func generateKey() crypto.PublicKeyHash {
	priv, pub, genErr := crypto.GenerateKey()

	if genErr != nil {
		die(genErr)
	}

	hash, hashErr := pub.Hash()

	if hashErr != nil {
		die(hashErr)
	}

	cacheErr := keyStore.PutPrivateKey(priv)

	if cacheErr != nil {
		die(cacheErr)
	}

	return hash
}
