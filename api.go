package godless

type QueryAPIService interface {
	RunQuery(*Query) (<-chan APIResponse, error)
}

type APIResponse struct {
	Err error
	Msg string
	QueryId string
}

var RESPONSE_OK APIResponse = APIResponse{Msg: "ok"}
