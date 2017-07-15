package http

import (
	"bytes"
	"fmt"
	gohttp "net/http"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

const API_ROOT = "/api"
const QUERY_API_ROOT = "/query"
const REFLECT_API_ROOT = "/reflect"

type WebServiceOptions struct {
	Endpoints
	Api api.RequestService
}

type WebService struct {
	WebServiceOptions
	stopch chan struct{}
}

func MakeWebService(options WebServiceOptions) api.WebService {
	service := &WebService{
		WebServiceOptions: options,
		stopch:            make(chan struct{}),
	}

	service.UseDefaultEndpoints()

	return service
}

func (service *WebService) Close() {
	close(service.stopch)
}

func (service *WebService) GetApiRequestHandler() gohttp.Handler {
	return gohttp.HandlerFunc(service.handleApiRequest)
}

func (service *WebService) handleApiRequest(rw gohttp.ResponseWriter, req *gohttp.Request) {
	if !service.IsCommandEndpoint(req.RequestURI) {
		log.Info("Bad URL for ApiRequestHandler")
		rw.WriteHeader(NOT_FOUND)
		return
	}

	log.Info("WebService runQuery at: %v", req.RequestURI)
	request, err := api.DecodeRequest(req.Body)

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	respch, err := service.Api.Call(request)
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
	encerr := api.EncodeResponseText(message, &buff)

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
	encerr := api.EncodeResponse(resp, buff)

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
	NOT_FOUND       = 404
	WEB_API_SUCCESS = 200
	WEB_API_ERROR   = 500
)
