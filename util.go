package godless

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

func imin(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func linearContains(sl []string, term string) bool {
	for _, s := range sl {
		if s == term {
			return true
		}
	}

	return false
}

func writeBytes(bs []byte, w io.Writer) error {
	written, err := w.Write(bs)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("write failed after %v bytes", written))
	}

	return nil
}

func encode(message proto.Message, w io.Writer) error {
	const failMsg = "encode failed"
	bs, err := proto.Marshal(message)
	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	err = writeBytes(bs, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func decode(message proto.Message, r io.Reader) error {
	const failMsg = "decode failed"
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return proto.Unmarshal(bs, message)
}

func encodeText(message proto.Message, w io.Writer) error {
	const failMsg = "encodeText failed"

	err := proto.MarshalText(w, message)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func decodeText(message proto.Message, r io.Reader) error {
	const failMsg = "decodeText failed"
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	text := string(bs)
	return proto.UnmarshalText(text, message)
}
