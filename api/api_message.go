package api

import (
	"fmt"
	"io"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/log"
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
		makeAPIQueryResponseMessage(resp.QueryResponse, message)
	case API_REFLECT:
		makeAPIReflectMessage(resp.ReflectResponse, message)
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
		readAPIQueryResponse(message.QueryResponse, &resp)
	case API_REFLECT:
		readAPIReflectResponse(message.ReflectResponse, &resp)
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

func makeAPIQueryResponseMessage(resp APIQueryResponse, message *proto.APIResponseMessage) {
	ns := crdt.MakeNamespaceStreamMessage(resp.Entries)
	message.QueryResponse = &proto.APIQueryResponseMessage{Namespace: ns}
}

func makeAPIReflectMessage(resp APIReflectResponse, message *proto.APIResponseMessage) {
	reflectMessage := &proto.APIReflectResponseMessage{Type: uint32(resp.Type)}

	switch resp.Type {
	case REFLECT_HEAD_PATH:
		reflectMessage.Path = string(resp.Path)
	case REFLECT_INDEX:
		indexMessage, invalid := crdt.MakeIndexMessage(resp.Index)

		invalidCount := len(invalid)

		if invalidCount > 0 {
			log.Error("makeAPIReflectMessage: %d invalid Index entries", invalidCount)
		}

		reflectMessage.Index = indexMessage
	case REFLECT_DUMP_NAMESPACE:
		namespace, invalid := crdt.MakeNamespaceMessage(resp.Namespace)

		invalidCount := len(invalid)

		if invalidCount > 0 {
			log.Error("makeAPIReflectMessage: %d invalid Namespace entries", invalidCount)
		}

		reflectMessage.Namespace = namespace
	default:
		panic(fmt.Sprintf("Unknown APIReflectResponse.Type: %v", resp))
	}

	message.ReflectResponse = reflectMessage
}

func readAPIQueryResponse(message *proto.APIQueryResponseMessage, resp *APIResponse) {
	ns := crdt.ReadNamespaceStreamMessage(message.Namespace)
	resp.QueryResponse.Entries = ns
}

func readAPIReflectResponse(message *proto.APIReflectResponseMessage, resp *APIResponse) {
	reflectResp := APIReflectResponse{Type: APIReflectionType(message.Type)}

	switch reflectResp.Type {
	case REFLECT_HEAD_PATH:
		reflectResp.Path = crdt.IPFSPath(message.Path)
	case REFLECT_INDEX:
		index, invalid := crdt.ReadIndexMessage(message.Index)

		invalidCount := len(invalid)

		if invalidCount > 0 {
			log.Error("readAPIReflectResponse: %d invalid Index entries", invalidCount)
		}

		reflectResp.Index = index
	case REFLECT_DUMP_NAMESPACE:
		// namespace, invalid, err
		namespace, invalid, err := crdt.ReadNamespaceMessage(message.Namespace)

		invalidCount := len(invalid)

		if invalidCount > 0 {
			log.Error("readAPIReflectResponse: %d invalid Index entries", invalidCount)
		}

		if err == nil {
			reflectResp.Namespace = namespace
		} else {
			resp.Err = err
			resp.Msg = RESPONSE_FAIL_MSG
		}
	default:
		panic(fmt.Sprintf("Unknown APIReflectResponse.Type: %v", message))
	}

	resp.ReflectResponse = reflectResp
}
