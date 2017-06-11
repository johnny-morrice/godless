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
	"io"
	"io/ioutil"
	"os"

	pb "github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	ipfs "github.com/ipfs/go-ipfs-api"
	"github.com/johnny-morrice/godless/proto"
)

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:   "cat",
	Short: "Read godless data from the remote store or filesystem.",
	Long: `Read a namespace, index, query, or API response from the remote store or filesystem.

The default behaviour is to read binary data and output text.`,
}

func catMessage(cmd *cobra.Command, streamer catStreamer) {
	validateStoreCatArgs(cmd)

	input, openErr := catOpen()

	if openErr != nil {
		die(openErr)
	}

	defer drainInput(input)

	message, decodeErr := streamer.decode(input)

	if decodeErr != nil {
		die(decodeErr)
	}

	encodeErr := catEncode(message)

	if encodeErr != nil {
		die(encodeErr)
	}
}

var filePath string
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
	decode(io.Reader) (pb.Message, error)
}

func catEncode(message pb.Message) error {
	if catBinaryOut {
		bs, err := pb.Marshal(message)

		if err != nil {
			return err
		}

		_, err = os.Stdout.Write(bs)

		return err
	} else {
		return pb.MarshalText(os.Stdout, message)
	}
}

type namespaceStreamer struct{}

type indexStreamer struct{}

type apiStreamer struct{}

type queryStreamer struct{}

func (streamer namespaceStreamer) decode(r io.Reader) (pb.Message, error) {
	pb := &proto.NamespaceMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func (streamer indexStreamer) decode(r io.Reader) (pb.Message, error) {
	pb := &proto.IndexMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func (streamer apiStreamer) decode(r io.Reader) (pb.Message, error) {
	pb := &proto.APIResponseMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func (streamer queryStreamer) decode(r io.Reader) (pb.Message, error) {
	pb := &proto.QueryMessage{}
	err := catDecode(r, pb)
	return pb, err
}

func catDecode(r io.Reader, message pb.Message) error {
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return err
	}

	if catBinaryIn {
		return pb.Unmarshal(bs, message)
	}

	return pb.UnmarshalText(string(bs), message)
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
	sourceCount := countNonEmpty([]string{filePath, hash})

	if sourceCount == 0 {
		err := cmd.Help()

		if err != nil {
			die(err)
		}

		return
	}

	if sourceCount != 1 {
		hint := errors.New("Must specify one of --file or --hash")
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

func init() {
	storeCmd.AddCommand(catCmd)

	catCmd.PersistentFlags().StringVar(&filePath, "file", "", "Filename")
	catCmd.PersistentFlags().BoolVar(&catBinaryIn, "binaryInput", true, "Input is binary")
	catCmd.PersistentFlags().BoolVar(&catBinaryOut, "binaryOutput", false, "Output is binary")
}
