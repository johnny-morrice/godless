package api

import (
	"io"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"
)

func MakeAPIResponseMessage(resp Response) *proto.APIResponseMessage {
	message := &proto.APIResponseMessage{
		Message: resp.Msg,
		Type:    uint32(resp.Type),
		Path:    string(resp.Path),
	}

	if resp.Err != nil {
		message.Error = resp.Err.Error()
		return message
	}

	nsMsg, nsInvalid := crdt.MakeNamespaceMessage(resp.Namespace)
	message.Namespace = nsMsg
	logInvalidNamespace(nsInvalid)

	indexMsg, indexInvalid := crdt.MakeIndexMessage(resp.Index)
	message.Index = indexMsg

	logInvalidIndex(indexInvalid)

	return message
}

func ReadAPIResponseMessage(message *proto.APIResponseMessage) Response {
	resp := Response{
		Msg:  message.Message,
		Type: MessageType(message.Type),
		Path: crdt.IPFSPath(message.Path),
	}

	if message.Error != "" {
		resp.Err = errors.New(message.Error)
		return resp
	}

	ns, nsInvalid := crdt.ReadNamespaceMessage(message.Namespace)
	resp.Namespace = ns
	logInvalidNamespace(nsInvalid)

	index, indexInvalid := crdt.ReadIndexMessage(message.Index)
	resp.Index = index
	logInvalidIndex(indexInvalid)

	return resp
}

func EncodeAPIResponse(resp Response, w io.Writer) error {
	const failMsg = "EncodeAPIResponse failed"

	message := MakeAPIResponseMessage(resp)

	err := util.Encode(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeAPIResponse(r io.Reader) (Response, error) {
	const failMsg = "DecodeAPIResponse failed"

	message := &proto.APIResponseMessage{}

	err := util.Decode(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

func EncodeAPIResponseText(resp Response, w io.Writer) error {
	const failMsg = "EncodeAPIResponseText failed"

	message := MakeAPIResponseMessage(resp)

	err := util.EncodeText(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeAPIResponseText(r io.Reader) (Response, error) {
	const failMsg = "DecodeAPIResponseText failed"

	message := &proto.APIResponseMessage{}

	err := util.DecodeText(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

func logInvalidNamespace(invalid []crdt.InvalidNamespaceEntry) {
	if len(invalid) > 0 {
		log.Error("%d invalid Namespace entries", len(invalid))
	}
}

func logInvalidIndex(invalid []crdt.InvalidIndexEntry) {
	if len(invalid) > 0 {
		log.Error("%d invalid Index entries", len(invalid))
	}
}
