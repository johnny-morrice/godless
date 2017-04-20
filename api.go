package godless

import (
	"encoding/json"
)

type QueryAPIService interface {
	RunQuery(*Query) (<-chan APIResponse, error)
	Close()
}

type APIResponse struct {
	Err error
	Msg string
	Rows []Row
	QueryId string
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
