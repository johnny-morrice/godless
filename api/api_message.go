package api

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/log"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"
)

func MakeAPIResponseMessage(resp APIResponse) *proto.APIResponseMessage {
	message := &proto.APIResponseMessage{
		Message: resp.Msg,
		Type:    uint32(resp.Type),
	}

	if resp.Err != nil {
		message.Error = resp.Err.Error()
		return message
	}

	switch resp.Type {
	case API_QUERY:
		message.QueryResponse = makeAPIQueryResponseMessage(resp.QueryResponse)
	case API_REFLECT:
		message.ReflectResponse = makeAPIReflectMessage(resp.ReflectResponse)
	case API_REPLICATE:
	default:
		panic(fmt.Sprintf("Unknown APIResponse.Type: %v", resp))
	}

	return message
}

func ReadAPIResponseMessage(message *proto.APIResponseMessage) APIResponse {
	resp := APIResponse{
		Msg:  message.Message,
		Type: APIMessageType(message.Type),
	}

	if message.Error != "" {
		resp.Err = errors.New(message.Error)
		return resp
	}

	switch resp.Type {
	case API_QUERY:
		resp.QueryResponse = readAPIQueryResponse(message.QueryResponse)
	case API_REFLECT:
		resp.ReflectResponse = readAPIReflectResponse(message.ReflectResponse)
	case API_REPLICATE:
	default:
		// TODO dupe code
		panic(fmt.Sprintf("Unknown APIResponse.Type: %v", message))
	}

	return resp
}

func EncodeAPIResponse(resp APIResponse, w io.Writer) error {
	const failMsg = "EncodeAPIResponse failed"

	message := MakeAPIResponseMessage(resp)

	err := util.Encode(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeAPIResponse(r io.Reader) (APIResponse, error) {
	const failMsg = "DecodeAPIResponse failed"

	message := &proto.APIResponseMessage{}

	err := util.Decode(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

func EncodeAPIResponseText(resp APIResponse, w io.Writer) error {
	const failMsg = "EncodeAPIResponseText failed"

	message := MakeAPIResponseMessage(resp)

	err := util.EncodeText(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeAPIResponseText(r io.Reader) (APIResponse, error) {
	const failMsg = "DecodeAPIResponseText failed"

	message := &proto.APIResponseMessage{}

	err := util.DecodeText(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

func makeAPIQueryResponseMessage(resp APIQueryResponse) *proto.APIQueryResponseMessage {
	ns := crdt.MakeNamespaceStreamMessage(resp.Entries)
	message := &proto.APIQueryResponseMessage{Namespace: ns}
	return message
}

func makeAPIReflectMessage(resp APIReflectResponse) *proto.APIReflectResponseMessage {
	message := &proto.APIReflectResponseMessage{Type: uint32(resp.Type)}

	switch resp.Type {
	case REFLECT_HEAD_PATH:
		message.Path = string(resp.Path)
	case REFLECT_INDEX:
		indexMessage, invalid := crdt.MakeIndexMessage(resp.Index)

		invalidCount := len(invalid)

		if invalidCount > 0 {
			log.Error("makeAPIReflectMessage: %v invalid Index entries", invalidCount)
		}

		message.Index = indexMessage
	case REFLECT_DUMP_NAMESPACE:
		namespace, invalid := crdt.MakeNamespaceMessage(resp.Namespace)

		invalidCount := len(invalid)

		if invalidCount > 0 {
			log.Error("makeAPIReflectMessage: %v invalid Namespace entries", invalidCount)
		}

		message.Namespace = namespace
	default:
		panic(fmt.Sprintf("Unknown APIReflectResponse.Type: %v", resp))
	}

	return message
}

func readAPIQueryResponse(message *proto.APIQueryResponseMessage) APIQueryResponse {
	resp := APIQueryResponse{}
	resp.Entries = crdt.ReadNamespaceStreamMessage(message.Namespace)
	return resp
}

func readAPIReflectResponse(message *proto.APIReflectResponseMessage) APIReflectResponse {
	resp := APIReflectResponse{Type: APIReflectionType(message.Type)}

	switch resp.Type {
	case REFLECT_HEAD_PATH:
		resp.Path = crdt.IPFSPath(message.Path)
	case REFLECT_INDEX:
		index, invalid := crdt.ReadIndexMessage(message.Index)

		// TODO Should write those details to disk.
		if len(invalid) > 0 {
			log.Warn("Invalid index stream details")
		}

		resp.Index = index
	case REFLECT_DUMP_NAMESPACE:
		// namespace, invalid, err
		crdt.ReadNamespaceMessage(message.Namespace)

		// FIXME implement error handling
		panic("not implemented")
	default:
		panic(fmt.Sprintf("Unknown APIReflectResponse.Type: %v", message))
	}

	return resp
}
