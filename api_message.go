package godless

import "fmt"

func MakeAPIResponseMessage(resp APIResponse) *APIResponseMessage {
	message := &APIResponseMessage{
		Error:   resp.Err.Error(),
		Message: resp.Msg,
		Type:    uint32(resp.Type),
	}

	switch resp.Type {
	case API_QUERY:
		message.QueryResponse = makeAPIQueryResponseMessage(resp.QueryResponse)
	case API_REFLECT:
		message.RefelectResponse = makeAPIReflectMessage(resp.ReflectResponse)
	default:
		panic(fmt.Sprintf("Unknown APIResponse.Type: %v", resp))
	}

	return message
}

func ReadAPIResponseMessage(message *APIResponseMessage) APIResponse {
	return RESPONSE_FAIL
}

func makeAPIQueryResponseMessage(response APIQueryResponse) *APIQueryResponseMessage {
	ns := MakeNamespaceStreamMessage(response.Rows)
	message := &APIQueryResponseMessage{Namespace: ns}
	return message
}

func makeAPIReflectMessage(response APIReflectResponse) *APIReflectResponseMessage {
	message := &APIReflectResponseMessage{Type: uint32(response.Type)}

	switch response.Type {
	case REFLECT_HEAD_PATH:
		message.Path = response.Path
	case REFLECT_INDEX:
		message.Index = MakeIndexMessage(response.Index)
	case REFLECT_DUMP_NAMESPACE:
		message.Namespace = MakeNamespaceMessage(response.Namespace)
	}

	return message
}
