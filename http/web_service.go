package http

import (
	"bytes"
	"fmt"
	gohttp "net/http"

	"github.com/gorilla/mux"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

const API_ROOT = "/api"
const QUERY_API_ROOT = "/query"
const REFLECT_API_ROOT = "/reflect"

// TODO WebService could support replication.
type WebService struct {
	apiService api.RequestService
	stopch     chan struct{}
}

func MakeWebService(apiService api.RequestService) *WebService {
	return &WebService{
		apiService: apiService,
		stopch:     make(chan struct{}),
	}
}

func (service *WebService) Close() {
	close(service.stopch)
}

func (service *WebService) Handler() gohttp.Handler {
	root := mux.NewRouter()
	topLevel := root.PathPrefix(API_ROOT).Subrouter()

	reflectMux := topLevel.PathPrefix(REFLECT_API_ROOT).Subrouter()
	reflectMux.HandleFunc("/head", service.HandleReflectHead)
	reflectMux.HandleFunc("/index", service.HandleReflectIndex)
	reflectMux.HandleFunc("/namespace", service.HandleReflectNamespace)

	topLevel.HandleFunc(QUERY_API_ROOT, service.HandleQuery)

	return root
}

func (service *WebService) HandleReflectHead(rw gohttp.ResponseWriter, req *gohttp.Request) {
	log.Info("WebService reflectHead at: %v", req.RequestURI)
	service.reflect(rw, api.REFLECT_HEAD_PATH)
}

func (service *WebService) HandleReflectIndex(rw gohttp.ResponseWriter, req *gohttp.Request) {
	log.Info("WebService reflectIndex at: %v", req.RequestURI)
	service.reflect(rw, api.REFLECT_INDEX)
}

func (service *WebService) HandleReflectNamespace(rw gohttp.ResponseWriter, req *gohttp.Request) {
	log.Info("WebService reflectDumpNamespace at: %v", req.RequestURI)
	service.reflect(rw, api.REFLECT_DUMP_NAMESPACE)
}

func (service *WebService) reflect(rw gohttp.ResponseWriter, reflection api.ReflectionType) {
	respch, err := service.apiService.Call(api.Request{Type: api.API_REFLECT, Reflection: reflection})
	service.respond(rw, respch, err)
}

func (service *WebService) HandleQuery(rw gohttp.ResponseWriter, req *gohttp.Request) {
	log.Info("WebService runQuery at: %v", req.RequestURI)
	q, err := query.DecodeQuery(req.Body)

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	respch, err := service.apiService.Call(api.Request{Type: api.API_QUERY, Query: q})
	service.respond(rw, respch, err)
}

func invalidRequest(rw gohttp.ResponseWriter, err error) {
	log.Info("Invalid Request details: %s", err.Error())
	reportErr := sendErr(rw, err)
	if reportErr != nil {
		log.Error("Error sending error report: '%s'", reportErr.Error())
	}
}

func (service *WebService) readResponse(respch <-chan api.Response) api.Response {
	select {
	case resp := <-respch:
		return resp
	case <-service.stopch:
		fail := api.RESPONSE_FAIL
		fail.Err = errors.New("WebService closed")
		return fail
	}
}

func (service *WebService) respond(rw gohttp.ResponseWriter, respch <-chan api.Response, err error) {
	if err != nil {
		invalidRequest(rw, err)
		return
	}

	log.Info("Webservice waiting for API...")
	resp := service.readResponse(respch)
	log.Info("Webservice received API response")

	err = sendMessage(rw, resp)

	if err != nil {
		log.Error("Error sending response: %v", err)
	}
}

// TODO why are we sending errors in plaintext again?
func sendErr(rw gohttp.ResponseWriter, err error) error {
	message := api.Response{
		Err: err,
	}

	buff := bytes.Buffer{}
	encerr := api.EncodeAPIResponseText(message, &buff)

	if encerr != nil {
		panic(fmt.Sprintf("Bug encoding json error message: '%v'; ", encerr.Error()))
	}

	log.Info("Sending error APIResponse (%v bytes) to HTTP client...", buff.Len())
	rw.WriteHeader(WEB_API_ERROR)
	rw.Header()[CONTENT_TYPE] = []string{MIME_PROTO_TEXT}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendErr failed")
	}

	log.Info("Sent error response to HTTP client")

	return nil
}

func sendMessage(rw gohttp.ResponseWriter, resp api.Response) error {
	// Encode gob into buffer first to check for encoding errors.
	// TODO is that actually a good idea?
	buff := &bytes.Buffer{}
	encerr := api.EncodeAPIResponse(resp, buff)

	if encerr != nil {
		panic(fmt.Sprintf("BUG encoding resp: %v", encerr))
	}

	log.Info("Sending APIResponse (%d bytes) to HTTP client...", buff.Len())
	rw.Header()[CONTENT_TYPE] = []string{MIME_PROTO}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendMessage failed")
	}

	log.Info("Sent response to HTTP client")

	return nil
}

const (
	WEB_API_SUCCESS = 200
	WEB_API_ERROR   = 500
)