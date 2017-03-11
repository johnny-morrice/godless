package godless

import (
	"bytes"
	"encoding/json"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type KeyValueService struct {
	Query chan<- KvQuery
}

type WireMapJoin struct {
	Val map[string][]string
}

func (service *KeyValueService) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/query", service.query).Methods("POST")
	return nil
}

func (service *KeyValueService) query(rw http.ResponseWriter, req *http.Request) {
	queryText := req.FormValue("query")
	query, err := CompileQuery(queryText)

	if err != nil {
		invalidRequest(rw, err)
		return
	}

	kvq := MakeKvQuery(query)
	service.handleKvQuery(rw, kvq)
}

func invalidRequest(rw http.ResponseWriter, err error) {
	log.Printf("Invalid Request details: %v", err)
	reportErr := sendErr(rw, err)
	log.Printf("Error sending JSON error report: '%v'", reportErr)
}

func (service *KeyValueService) handleKvQuery(rw http.ResponseWriter, kvq KvQuery) {
	service.Query<- kvq
	resp := <-kvq.Response
	if resp.Err == nil {
		err := sendGob(rw, resp.Val)
		log.Printf("Error sending gob: '%v'", err)
	} else {
		err := sendErr(rw, resp.Err)
		log.Printf("Error sending JSON error report: '%v'", err)
	}
}

func sendErr(rw http.ResponseWriter, err error) error {
	message := struct{
		ErrorMessage string
	}{
		ErrorMessage: err.Error(),
	}

	buff := bytes.Buffer{}
	enc := json.NewEncoder(&buff)
	encerr := enc.Encode(&message)

	if encerr != nil {
		panic(fmt.Sprintf("Bug encoding json error message: '%v'; ", err, encerr))
	}

	rw.WriteHeader(400)
	rw.Header()["Content-Type"] = []string{"application/json"}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendErr failed")
	}

	return nil
}

func sendGob(rw http.ResponseWriter, gobber interface{}) error {
	// Encode gob into buffer first to check for encoding errors.
	// TODO is that actually a good idea?
	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	encerr := enc.Encode(gobber)

	if encerr != nil {
		panic(fmt.Sprintf("BUG encoding gob: %v", encerr))
	}

	rw.Header()["Content-Type"] = []string{"application/octet-stream"}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendGob failed")
	}

	return nil
}
