package godless

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const API_ROOT = "/api"
const QUERY_API_ROOT = "/query"
const REFLECT_API_ROOT = "/reflect"

type WebService struct {
	API APIService
}

func (service *WebService) Handler() http.Handler {
	root := mux.NewRouter()
	topLevel := root.PathPrefix(API_ROOT).Subrouter()

	reflectMux := topLevel.PathPrefix(REFLECT_API_ROOT).Subrouter()
	reflectMux.HandleFunc("/head", service.reflectHead)
	reflectMux.HandleFunc("/index", service.reflectIndex)
	reflectMux.HandleFunc("/namespace", service.reflectDumpNamespace)

	topLevel.HandleFunc(QUERY_API_ROOT, service.runQuery)

	return root
}

func (service *WebService) reflectHead(rw http.ResponseWriter, req *http.Request) {
	service.reflect(rw, APIReflectRequest{Command: REFLECT_HEAD_PATH})
}

func (service *WebService) reflectIndex(rw http.ResponseWriter, req *http.Request) {
	service.reflect(rw, APIReflectRequest{Command: REFLECT_INDEX})
}

func (service *WebService) reflectDumpNamespace(rw http.ResponseWriter, req *http.Request) {
	service.reflect(rw, APIReflectRequest{Command: REFLECT_DUMP_NAMESPACE})
}

func (service *WebService) reflect(rw http.ResponseWriter, reflection APIReflectRequest) {
	respch, err := service.API.Reflect(reflection)
	service.respond(rw, respch, err)
}

func (service *WebService) runQuery(rw http.ResponseWriter, req *http.Request) {
	query, err := DecodeQuery(req.Body)

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	respch, err := service.API.RunQuery(query)
	service.respond(rw, respch, err)
}

func invalidRequest(rw http.ResponseWriter, err error) {
	logdbg("Invalid Request details: %v", err)
	reportErr := sendErr(rw, err)
	if reportErr != nil {
		logerr("Error sending JSON error report: '%v'", reportErr)
	}
}

// TODO more coherency.
func (service *WebService) respond(rw http.ResponseWriter, respch <-chan APIResponse, err error) {
	if err != nil {
		invalidRequest(rw, err)
		return
	}

	resp := <-respch

	err = sendMessage(rw, resp)

	if err != nil {
		logerr("Error sending response: %v", err)
	}
}

// TODO why are we sending errors in plaintext again?
func sendErr(rw http.ResponseWriter, err error) error {
	message := APIResponse{
		Err: err,
	}

	buff := bytes.Buffer{}
	encerr := EncodeAPIResponseText(message, &buff)

	if encerr != nil {
		panic(fmt.Sprintf("Bug encoding json error message: '%v'; ", encerr))
	}

	rw.WriteHeader(WEB_API_ERROR)
	rw.Header()["Content-Type"] = []string{MIME_PROTO_TEXT}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendErr failed")
	}

	return nil
}

func sendMessage(rw http.ResponseWriter, resp APIResponse) error {
	// Encode gob into buffer first to check for encoding errors.
	// TODO is that actually a good idea?
	buff := &bytes.Buffer{}
	encerr := EncodeAPIResponse(resp, buff)

	if encerr != nil {
		panic(fmt.Sprintf("BUG encoding resp: %v", encerr))
	}

	rw.Header()[CONTENT_TYPE] = []string{MIME_PROTO}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendMessage failed")
	}

	return nil
}

const (
	WEB_API_SUCCESS = 200
	WEB_API_ERROR   = 500
)
