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
	"path"
	"strings"

	lib "github.com/johnny-morrice/godless"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise your godless environment",
	Long: `Generate crypto keys and setup your config file.

	See 'godless key' for more control`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}

var keyStore api.KeyStore = lib.MakeKeyStore()

func flushKeysToViper() {
	privKeys := keyStore.GetAllPrivateKeys()
	pubKeys := keyStore.GetAllPublicKeys()

	privTexts := make([]string, 0, len(privKeys))
	pubTexts := make([]string, 0, len(pubKeys))

	for _, priv := range privKeys {
		text, err := crypto.SerializePrivateKey(priv)
		if err != nil {
			log.Error("Failed to serialize PrivateKey: %v", err.Error())
		}

		privTexts = append(privTexts, string(text))
	}

	for _, pub := range pubKeys {
		text, err := crypto.SerializePublicKey(pub)

		if err != nil {
			log.Error("Failed to serialize PublicKey: %v", err.Error())
		}

		pubTexts = append(pubTexts, string(text))
	}

	publicKeyConfigValue := strings.Join(pubTexts, ":")
	privateKeyConfigValue := strings.Join(privTexts, ":")

	viper.Set(__PRIVATE_KEY_CONFIG_KEY, privateKeyConfigValue)
	viper.Set(__PUBLIC_KEY_CONFIG_KEY, publicKeyConfigValue)
}

func writeKeys() {
	flushKeysToViper()

	configFileName := viper.ConfigFileUsed()

	if configFileName != "" {
		configFileName = homeConfigFilePath()
	}

	file, err := os.Create(configFileName)

	if err != nil {

	}

	defer file.Close()
}

func homeConfigFilePath() string {
	dir := os.Getenv("HOME")
	name := __CONFIG_FILE_NAME
	return path.Join(dir, name)
}

const __PRIVATE_KEY_CONFIG_KEY = "PrivateKeys"
const __PUBLIC_KEY_CONFIG_KEY = "PublicKeys"
