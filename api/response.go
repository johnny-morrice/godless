package api

import (
	"bytes"
	"io"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/proto"
)

type Response struct {
	Msg       string
	Err       error
	Type      MessageType
	Path      crdt.IPFSPath
	Namespace crdt.Namespace
	Index     crdt.Index
}

func (resp Response) IsEmpty() bool {
	return resp.Equals(Response{})
}

func (resp Response) AsText() (string, error) {
	const failMsg = "AsText failed"

	w := &bytes.Buffer{}
	err := EncodeResponseText(resp, w)

	if err != nil {
		return "", errors.Wrap(err, failMsg)
	}

	return w.String(), nil
}

func (resp Response) Equals(other Response) bool {
	ok := resp.Msg == other.Msg
	ok = ok && resp.Type == other.Type
	ok = ok && resp.Path == other.Path

	if !ok {
		return false
	}

	if resp.Err != nil {
		if other.Err == nil {
			return false
		} else if resp.Err.Error() != other.Err.Error() {
			return false
		}
	}

	if !resp.Namespace.Equals(other.Namespace) {
		return false
	}

	if !resp.Index.Equals(other.Index) {
		return false
	}

	return true
}

func EncodeResponse(resp Response, w io.Writer) error {
	const failMsg = "EncodeResponse failed"

	message := MakeAPIResponseMessage(resp)

	err := util.Encode(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeResponse(r io.Reader) (Response, error) {
	const failMsg = "DecodeResponse failed"

	message := &proto.APIResponseMessage{}

	err := util.Decode(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

func EncodeResponseText(resp Response, w io.Writer) error {
	const failMsg = "EncodeResponseText failed"

	message := MakeAPIResponseMessage(resp)

	err := util.EncodeText(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeResponseText(r io.Reader) (Response, error) {
	const failMsg = "DecodeResponseText failed"

	message := &proto.APIResponseMessage{}

	err := util.DecodeText(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

var RESPONSE_FAIL_MSG = "error"
var RESPONSE_OK_MSG = "ok"
var RESPONSE_OK Response = Response{Msg: RESPONSE_OK_MSG}
var RESPONSE_FAIL Response = Response{Msg: RESPONSE_FAIL_MSG}
var RESPONSE_QUERY Response = Response{Msg: RESPONSE_OK_MSG, Type: API_QUERY}
var RESPONSE_REPLICATE Response = Response{Msg: RESPONSE_OK_MSG, Type: API_REPLICATE}
var RESPONSE_REFLECT Response = Response{Msg: RESPONSE_OK_MSG, Type: API_REFLECT}
