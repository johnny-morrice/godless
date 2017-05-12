package godless

import (
	"encoding/json"
)

type APIService interface {
	APICloserService
	APIQueryService
	APIReflectService
}

type APICloserService interface {
	CloseAPI()
}

type APIReflectService interface {
	Reflect(APIReflectRequest) (<-chan APIResponse, error)
}

type APIQueryService interface {
	RunQuery(*Query) (<-chan APIResponse, error)
}

type APIResponder interface {
	RunQuery() APIResponse
}

type APIResponderFunc func() APIResponse

func (arf APIResponderFunc) RunQuery() APIResponse {
	return arf()
}

type APIQueryResponse struct {
	Rows []Row
}

type APIReflectionType uint16

const (
	REFLECT_NOOP = APIReflectionType(iota)
	REFLECT_READ_REMOTE_PATH
	REFLECT_READ_ALL_NAMESPACES
	REFLECT_READ_INDEX
)

type APIReflectRequest struct {
	Command APIReflectionType
}

// FIXME dubious type screams design flaw.
type APIRemoteIndex struct {
	Index map[string][]string
}

type APIReflectResponse struct {
	Namespace Namespace      `json:",omitEmpty"`
	Path      string         `json:",omitEmpty"`
	Index     APIRemoteIndex `json:",omitEmpty"`
}

type APIMessageType uint8

const (
	API_MESSAGE_NOOP = APIMessageType(iota)
	API_QUERY
	API_REFLECT
)

type APIResponse struct {
	Err             error
	Msg             string
	Type            APIMessageType
	QueryResponse   APIQueryResponse   `json:",omitEmpty"`
	ReflectResponse APIReflectResponse `json:",omitEmpty"`
}

func (response APIResponse) RenderJSON() (string, error) {
	bs, err := json.MarshalIndent(response, "", "\t")

	if err != nil {
		return "", err
	}

	return string(bs), nil
}

var RESPONSE_OK APIResponse = APIResponse{Msg: "ok"}
var RESPONSE_FAIL APIResponse = APIResponse{Msg: "error"}
var RESPONSE_QUERY APIResponse = APIResponse{Msg: "ok", Type: API_QUERY}
var RESPONSE_REFLECT APIResponse = APIResponse{Msg: "ok", Type: API_REFLECT}
