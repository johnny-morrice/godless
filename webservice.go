package godless

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type KeyValueService struct {
	Query chan<- KvQuery
}

type ApiResponse struct {
	Msg string
	QueryId string
}

type ApiError struct {
	Err string
}

func (service *KeyValueService) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/query/run", service.queryRun).Methods("POST")
	return nil
}

func (service *KeyValueService) queryRun(rw http.ResponseWriter, req *http.Request) {
	queryText := req.FormValue("query")

	var query *Query
	var err error
	var isText bool

	isText, err = needsParse(req)

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	if (isText) {
		query, err = CompileQuery(queryText)
	} else {
		query = &Query{}
		degob(query, strings.NewReader(queryText))
	}

	if err != nil {
		invalidRequest(rw, err)
		return
	}



	kvq := MakeKvQuery(query)
	service.handleKvQuery(rw, kvq)
}

func invalidRequest(rw http.ResponseWriter, err error) {
	logdbg("Invalid Request details: %v", err)
	reportErr := sendErr(rw, err)
	logerr("Error sending JSON error report: '%v'", reportErr)
}

func (service *KeyValueService) handleKvQuery(rw http.ResponseWriter, kvq KvQuery) {
	service.Query<- kvq
	resp := <-kvq.Response
	if resp.Err == nil {
		err := sendGob(rw, resp.Val)
		logerr("Error sending gob: '%v'", err)
	} else {
		err := sendErr(rw, resp.Err)
		logerr("Error sending JSON error report: '%v'", err)
	}
}

func sendErr(rw http.ResponseWriter, err error) error {
	message := ApiError{
		Err: err.Error(),
	}

	buff := bytes.Buffer{}
	encerr := tojson(&message, &buff)

	if encerr != nil {
		panic(fmt.Sprintf("Bug encoding json error message: '%v'; ", err, encerr))
	}

	rw.WriteHeader(400)
	rw.Header()["Content-Type"] = []string{MIME_JSON}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendErr failed")
	}

	return nil
}

func sendGob(rw http.ResponseWriter, gobber *ApiResponse) error {
	// Encode gob into buffer first to check for encoding errors.
	// TODO is that actually a good idea?
	buff := &bytes.Buffer{}
	encerr := togob(gobber, buff)

	if encerr != nil {
		panic(fmt.Sprintf("BUG encoding gob: %v", encerr))
	}

	rw.Header()[CONTENT_TYPE] = []string{MIME_GOB}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendGob failed")
	}

	return nil
}

func needsParse(req *http.Request) (bool, error) {
	ct := req.Header[CONTENT_TYPE]
	if linearContains(ct, "MIME_QUERY")  {
		return true, nil
	} else if linearContains(ct, "MIME_GOB") {
		return false, nil
	} else {
		return false, fmt.Errorf("No suitable MIME in request: '%v'", ct)
	}
}

func linearContains(sl []string, term string) bool {
	for _, s := range sl {
		if s == term {
			return true
		}
	}

	return false
}
