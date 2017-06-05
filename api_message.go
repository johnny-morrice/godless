package godless

import (
	"errors"
	"fmt"
)

func MakeAPIResponseMessage(resp APIResponse) *APIResponseMessage {
	message := &APIResponseMessage{
		Message: resp.Msg,
		Type:    uint32(resp.Type),
	}

	if resp.Err != nil {
		message.Error = resp.Err.Error()
		return message
	}

	switch resp.Type {
	case API_QUERY:
		logdbg("making query response: %v", resp.QueryResponse)
		message.QueryResponse = makeAPIQueryResponseMessage(resp.QueryResponse)
	case API_REFLECT:
		message.ReflectResponse = makeAPIReflectMessage(resp.ReflectResponse)
	default:
		panic(fmt.Sprintf("Unknown APIResponse.Type: %v", resp))
	}

	return message
}

func ReadAPIResponseMessage(message *APIResponseMessage) APIResponse {
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
	default:
		// TODO dupe code
		panic(fmt.Sprintf("Unknown APIResponse.Type: %v", message))
	}

	return resp
}

func makeAPIQueryResponseMessage(resp APIQueryResponse) *APIQueryResponseMessage {
	ns := MakeNamespaceStreamMessage(resp.Rows)
	message := &APIQueryResponseMessage{Namespace: ns}
	return message
}

func makeAPIReflectMessage(resp APIReflectResponse) *APIReflectResponseMessage {
	message := &APIReflectResponseMessage{Type: uint32(resp.Type)}

	switch resp.Type {
	case REFLECT_HEAD_PATH:
		message.Path = resp.Path
	case REFLECT_INDEX:
		message.Index = MakeIndexMessage(resp.Index)
	case REFLECT_DUMP_NAMESPACE:
		message.Namespace = MakeNamespaceMessage(resp.Namespace)
	default:
		panic(fmt.Sprintf("Unknown APIReflectResponse.Type: %v", resp))
	}

	return message
}

func readAPIQueryResponse(message *APIQueryResponseMessage) APIQueryResponse {
	resp := APIQueryResponse{}
	resp.Rows = ReadNamespaceStreamMessage(message.Namespace)
	return resp
}

func readAPIReflectResponse(message *APIReflectResponseMessage) APIReflectResponse {
	resp := APIReflectResponse{Type: APIReflectionType(message.Type)}

	switch resp.Type {
	case REFLECT_HEAD_PATH:
		resp.Path = message.Path
	case REFLECT_INDEX:
		resp.Index = ReadIndexMessage(message.Index)
	case REFLECT_DUMP_NAMESPACE:
		resp.Namespace = ReadNamespaceMessage(message.Namespace)
	default:
		panic(fmt.Sprintf("Unknown APIReflectResponse.Type: %v", message))
	}

	return resp
}
