package godless

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
)

type APIService interface {
	APICloserService
	APIQueryService
	APIReflectService
	APIPeerService
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

type APIPeerService interface {
	Replicate(peerAddr RemoteStoreAddress) (<-chan APIResponse, error)
}

type APIResponder interface {
	RunQuery() APIResponse
}

type APIResponderFunc func() APIResponse

func (arf APIResponderFunc) RunQuery() APIResponse {
	return arf()
}

type APIQueryResponse struct {
	Entries []NamespaceStreamEntry
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

type APIReflectResponse struct {
	Type      APIReflectionType
	Namespace Namespace            `json:",omitEmpty"`
	Path      string               `json:",omitEmpty"`
	Index     RemoteNamespaceIndex `json:",omitEmpty"`
}

type APIMessageType uint8

const (
	API_MESSAGE_NOOP = APIMessageType(iota)
	API_QUERY
	API_REFLECT
	API_REPLICATE
)

type APIResponse struct {
	Msg             string
	Err             error
	Type            APIMessageType
	QueryResponse   APIQueryResponse   `json:",omitEmpty"`
	ReflectResponse APIReflectResponse `json:",omitEmpty"`
}

func (resp APIResponse) AsText() (string, error) {
	const failMsg = "AsText failed"

	w := &bytes.Buffer{}
	err := EncodeAPIResponseText(resp, w)

	if err != nil {
		return "", errors.Wrap(err, failMsg)
	}

	return w.String(), nil
}

func (resp APIResponse) Equals(other APIResponse) bool {
	ok := resp.Msg == other.Msg
	ok = ok && resp.Type == other.Type

	if !ok {
		logwarn("not ok")
		return false
	}

	if resp.Err != nil {
		if other.Err == nil {
			logwarn("err not equal")
			return false
		} else if resp.Err.Error() != other.Err.Error() {
			logwarn("err text not equal")
			return false
		}
	}

	if resp.Type == API_QUERY {
		if len(resp.QueryResponse.Entries) != len(other.QueryResponse.Entries) {
			logwarn("rows have unequal length")
			logwarn("resp %v other %v", len(resp.QueryResponse.Entries), len(other.QueryResponse.Entries))
			return false
		}

		if !StreamEquals(resp.QueryResponse.Entries, other.QueryResponse.Entries) {
			logwarn("rows not equal")
			return false
		}
	} else if resp.Type == API_REFLECT {
		if resp.ReflectResponse.Path != other.ReflectResponse.Path {
			logwarn("path not equal")
			return false
		}

		if !resp.ReflectResponse.Index.Equals(other.ReflectResponse.Index) {
			logwarn("index not equal")
			return false
		}

		if !resp.ReflectResponse.Namespace.Equals(other.ReflectResponse.Namespace) {
			logwarn("namespace not equal")
			return false
		}
	}

	return true
}

func EncodeAPIResponse(resp APIResponse, w io.Writer) error {
	const failMsg = "EncodeAPIResponse failed"

	message := MakeAPIResponseMessage(resp)

	err := encode(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeAPIResponse(r io.Reader) (APIResponse, error) {
	const failMsg = "DecodeAPIResponse failed"

	message := &APIResponseMessage{}

	err := decode(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

func EncodeAPIResponseText(resp APIResponse, w io.Writer) error {
	const failMsg = "EncodeAPIResponseText failed"

	message := MakeAPIResponseMessage(resp)

	err := encodeText(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeAPIResponseText(r io.Reader) (APIResponse, error) {
	const failMsg = "DecodeAPIResponseText failed"

	message := &APIResponseMessage{}

	err := decodeText(message, r)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, failMsg)
	}

	return ReadAPIResponseMessage(message), nil
}

var RESPONSE_FAIL_MSG = "error"
var RESPONSE_OK_MSG = "ok"
var RESPONSE_OK APIResponse = APIResponse{Msg: RESPONSE_OK_MSG}
var RESPONSE_FAIL APIResponse = APIResponse{Msg: RESPONSE_FAIL_MSG}
var RESPONSE_QUERY APIResponse = APIResponse{Msg: RESPONSE_OK_MSG, Type: API_QUERY}
var RESPONSE_REPLICATE APIResponse = APIResponse{Msg: RESPONSE_OK_MSG, Type: API_REPLICATE}
var RESPONSE_REFLECT APIResponse = APIResponse{Msg: RESPONSE_OK_MSG, Type: API_REFLECT}
