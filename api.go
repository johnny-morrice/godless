package godless

import "io"

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
	Rows []NamespaceStreamEntry
}

type APIReflectionType uint16

const (
	REFLECT_NOOP = APIReflectionType(iota)
	REFLECT_HEAD_PATH
	REFLECT_DUMP_NAMESPACE
	REFLECT_INDEX
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
	Msg             string
	Err             error
	Type            APIMessageType
	QueryResponse   APIQueryResponse   `json:",omitEmpty"`
	ReflectResponse APIReflectResponse `json:",omitEmpty"`
}

func (resp APIResponse) Equals(other APIResponse) bool {
	ok := resp.Msg == other.Msg
	ok = ok && resp.Err == other.Err
	ok = ok && resp.Type == other.Type
	ok = ok && resp.ReflectResponse.Path == other.ReflectResponse.Path

	if resp.Type == API_QUERY {
		if len(resp.QueryResponse.Rows) != len(other.QueryResponse.Rows) {
			return false
		}

		if !StreamEquals(resp.QueryResponse.Rows, other.QueryResponse.Rows) {
			return false
		}
	} else if resp.Type == API_REFLECT {
		// FIXME
		logerr("TODO Equals not implemented for API_REFLECT type")
		return false
	}

	return true
}

func EncodeAPIResponse(resp APIResponse, w io.Writer) error {
	return nil
}

func DecodeAPIResponse(r io.Reader) (APIResponse, error) {
	return RESPONSE_FAIL, nil
}

func EncodeAPIResponseText(resp APIResponse, w io.Writer) error {
	return nil
}

func DecodeAPIResponseText(r io.Reader) (APIResponse, error) {
	return RESPONSE_FAIL, nil
}

var RESPONSE_FAIL_MSG = "error"
var RESPONSE_OK_MSG = "ok"
var RESPONSE_OK APIResponse = APIResponse{Msg: RESPONSE_OK_MSG}
var RESPONSE_FAIL APIResponse = APIResponse{Msg: RESPONSE_FAIL_MSG}
var RESPONSE_QUERY APIResponse = APIResponse{Msg: RESPONSE_OK_MSG, Type: API_QUERY}
var RESPONSE_REFLECT APIResponse = APIResponse{Msg: RESPONSE_OK_MSG, Type: API_REFLECT}
