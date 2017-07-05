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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/johnny-morrice/godless/crypto"
)

// keyListCmd represents the key_list command
var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List godless key hashes",
	Run: func(cmd *cobra.Command, args []string) {
		readKeysFromViper()

		if listPrivateOnly {
			for _, priv := range keyStore.GetAllPrivateKeys() {
				pub := priv.GetPublicKey()
				printKeyHash(pub)
			}

			return
		}

		for _, pub := range keyStore.GetAllPublicKeys() {
			printKeyHash(pub)
		}
	},
}

func printKeyHash(pub crypto.PublicKey) {
	hash, err := pub.Hash()

	// TODO could print the other keys
	if err != nil {
		die(err)
	}

	fmt.Println(string(hash))
}

var listPrivateOnly bool

func init() {
	keyCmd.AddCommand(keyListCmd)

	keyListCmd.PersistentFlags().BoolVar(&listPrivateOnly, "private", false, "List private keys only")
}
