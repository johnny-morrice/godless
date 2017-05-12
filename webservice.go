package godless

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type KeyValueService struct {
	API QueryAPIService
}

func (service *KeyValueService) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/api/query/run", service.queryRun)
	return r
}

func (service *KeyValueService) queryRun(rw http.ResponseWriter, req *http.Request) {
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

func (service *KeyValueService) runQuery(rw http.ResponseWriter, query *Query) {
	respch, err := service.API.RunQuery(query)

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	resp := <-respch

	err = sendGob(rw, resp)

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

func sendGob(rw http.ResponseWriter, gobber interface{}) error {
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
	if linearContains(ct, MIME_QUERY) {
		return true, nil
	} else if linearContains(ct, MIME_GOB) {
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
