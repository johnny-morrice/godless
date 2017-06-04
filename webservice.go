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

	topLevel.HandleFunc(QUERY_API_ROOT, service.queryRun)

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

func (service *WebService) queryRun(rw http.ResponseWriter, req *http.Request) {
	var query *Query
	var err error
	var isText bool

	isText, err = needsParse(req)

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	if isText {
		buff := bytes.Buffer{}
		_, err = buff.ReadFrom(req.Body)

		if err != nil {
			queryText := buff.String()
			query, err = CompileQuery(queryText)
		}
	} else {
		query = &Query{}
		degob(query, req.Body)
	}

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	service.runQuery(rw, query)
}

func invalidRequest(rw http.ResponseWriter, err error) {
	logdbg("Invalid Request details: %v", err)
	reportErr := sendErr(rw, err)
	if reportErr != nil {
		logerr("Error sending JSON error report: '%v'", reportErr)
	}
}

func (service *WebService) runQuery(rw http.ResponseWriter, query *Query) {
	respch, err := service.API.RunQuery(query)
	service.respond(rw, respch, err)
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

func sendErr(rw http.ResponseWriter, err error) error {
	message := APIResponse{
		Err: err,
	}

	buff := bytes.Buffer{}
	encerr := tojson(&message, &buff)

	if encerr != nil {
		panic(fmt.Sprintf("Bug encoding json error message: '%v'; ", encerr))
	}

	rw.WriteHeader(400)
	rw.Header()["Content-Type"] = []string{MIME_JSON}
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

func needsParse(req *http.Request) (bool, error) {
	ct := req.Header[CONTENT_TYPE]
	if linearContains(ct, MIME_QUERY) {
		return true, nil
	} else if linearContains(ct, MIME_PROTO) {
		return false, nil
	} else {
		return false, fmt.Errorf("No suitable MIME in request: '%v'", ct)
	}
}
