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
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	ipfs "github.com/ipfs/go-ipfs-api"
	lib "github.com/johnny-morrice/godless"
)

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "Read godless data from the remote store or filesystem.",
	Long: `Read a namespace, index, query, or API response from the remote store or filesystem.

The default behaviour is to read binary data and output text.`,
	Run: func(cmd *cobra.Command, args []string) {
		validateStoreCatArgs(cmd)

		input, openErr := catOpen()

		if openErr != nil {
			die(openErr)
		}

		defer drainInput(input)

		streamer := makeCatStreamer()
		message, decodeErr := streamer.decode(input)

		if decodeErr != nil {
			die(decodeErr)
		}

		encodeErr := catEncode(message)

		if encodeErr != nil {
			die(encodeErr)
		}
	},
}

var filePath string
var isNamespace bool
var isIndex bool
var isQuery bool
var isAPIResponse bool
var catBinaryOut bool
var catBinaryIn bool

func drainInput(r io.ReadCloser) {
	ioutil.ReadAll(r)

	err := r.Close()

	if err != nil {
		die(err)
	}
}

type catStreamer interface {
	decode(io.Reader) (proto.Message, error)
}

func catEncode(message proto.Message) error {
	if catBinaryOut {
		bs, err := proto.Marshal(message)

		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(bs)

		return err
	} else {
		return proto.MarshalText(os.Stdout, message)
	}
}

type namespaceStreamer struct{}

type indexStreamer struct{}

type apiStreamer struct{}

type queryStreamer struct{}

func (streamer namespaceStreamer) decode(r io.Reader) (proto.Message, error) {
	pb := &lib.NamespaceMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func (streamer indexStreamer) decode(r io.Reader) (proto.Message, error) {
	pb := &lib.IndexMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func (streamer apiStreamer) decode(r io.Reader) (proto.Message, error) {
	pb := &lib.APIResponseMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func (streamer queryStreamer) decode(r io.Reader) (proto.Message, error) {
	pb := &lib.QueryMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func catDecode(r io.Reader, message proto.Message) error {
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return err
	}

	if catBinaryIn {
		return proto.Unmarshal(bs, message)
	}

	return proto.UnmarshalText(string(bs), message)
}

func makeCatStreamer() catStreamer {
	if isNamespace {
		return namespaceStreamer{}
	} else if isAPIResponse {
		return apiStreamer{}
	} else if isQuery {
		return queryStreamer{}
	} else if isIndex {
		return indexStreamer{}
	} else {
		panic("Unknown streamer type")
	}
}

func catOpen() (io.ReadCloser, error) {
	if hash != "" {
		// I imagined I'd use the godless library for this, but cat
		// is a low level tool so I guess ok to call IPFS directly.
		shell := ipfs.NewShell(ipfsService)
		return shell.Cat(hash)
	} else if filePath != "" {
		return os.Open(filePath)
	} else {
		panic("no input source")
	}
	return nil, nil
}

func validateStoreCatArgs(cmd *cobra.Command) {
	var hint error

	sourceCount := countNonEmpty([]string{filePath, hash})
	dataTypeCount := countTrue([]bool{isNamespace, isIndex, isQuery, isAPIResponse})

	if sourceCount == 0 && dataTypeCount == 0 {
		err := cmd.Help()

		if err != nil {
			die(err)
		}

		return
	}

	if sourceCount != 1 {
		hint = joinError(hint, "Must specify one of --file or --hash")
	}

	if dataTypeCount != 1 {
		hint = joinError(hint, "Must specify one data type (e.g. --namespace)")
	}

	if hint != nil {
		die(hint)
	}
}

func countNonEmpty(args []string) int {
	count := 0
	for _, a := range args {
		if a != "" {
			count++
		}
	}

	return count
}

func countTrue(args []bool) int {
	count := 0
	for _, a := range args {
		if a {
			count++
		}
	}

	return count
}

func joinError(err error, msg string) error {
	if err == nil {
		return errors.New(msg)
	}

	return fmt.Errorf("%v\n%s", err.Error(), msg)
}

func init() {
	storeCmd.AddCommand(catCmd)

	catCmd.Flags().StringVar(&filePath, "file", "", "Filename")
	catCmd.Flags().BoolVar(&isNamespace, "namespace", false, "Cat a Namespace")
	catCmd.Flags().BoolVar(&isIndex, "index", false, "Cat an Index")
	catCmd.Flags().BoolVar(&isQuery, "query", false, "Cat a Query")
	catCmd.Flags().BoolVar(&isAPIResponse, "api", false, "Cat an APIResponse")
	catCmd.Flags().BoolVar(&catBinaryIn, "binaryInput", true, "Input is binary")
	catCmd.Flags().BoolVar(&catBinaryOut, "binaryOutput", false, "Output is binary")
}
