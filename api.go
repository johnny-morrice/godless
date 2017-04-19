package godless

type QueryAPIService interface {
	RunQuery(*Query) (<-chan APIResponse, error)
	Close()
}

type APIResponse struct {
	Err error
	Msg string
	QueryId string
}

var RESPONSE_OK APIResponse = APIResponse{Msg: "ok"}
