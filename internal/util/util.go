package util

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

func Imin(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func LinearContains(sl []string, term string) bool {
	for _, s := range sl {
		if s == term {
			return true
		}
	}

	return false
}

func WriteBytes(bs []byte, w io.Writer) error {
	written, err := w.Write(bs)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("write failed after %v bytes", written))
	}

	return nil
}

func Encode(message proto.Message, w io.Writer) error {
	const failMsg = "encode failed"
	bs, err := proto.Marshal(message)
	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	err = WriteBytes(bs, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func Decode(message proto.Message, r io.Reader) error {
	const failMsg = "decode failed"
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return proto.Unmarshal(bs, message)
}

func EncodeText(message proto.Message, w io.Writer) error {
	const failMsg = "encodeText failed"

	err := proto.MarshalText(w, message)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeText(message proto.Message, r io.Reader) error {
	const failMsg = "decodeText failed"
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	text := string(bs)
	return proto.UnmarshalText(text, message)
}
