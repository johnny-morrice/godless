package godless

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

var RESPONSE_OK APIResponse = APIResponse{Msg: "ok"}
var RESPONSE_FAIL APIResponse = APIResponse{Msg: "error"}
