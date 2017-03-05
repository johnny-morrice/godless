package godless

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type IpfsNamespaceService struct {
	Query chan<- KvQuery
}

func (service *IpfsNamespaceService) Handler() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/{namespaceKey}/map", service.getMap).Methods("GET")
	r.HandleFunc("/{namespaceKey}/map", service.joinMap).Methods("POST")
	r.HandleFunc("/{namespaceKey}/map/{mapKey}", service.getMapValues).Methods("GET")
	r.HandleFunc("/{namespaceKey}/set", service.getSet).Methods("GET")
	r.HandleFunc("/{namespaceKey}/set", service.joinSet).Methods("POST")
	return nil
}

func (service *IpfsNamespaceService) getMap(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	namespaceKey := vars["namespaceKey"]
	kvq := MakeKvQuery(GET_MAP, KvMapQuery{NamespaceKey: namespaceKey})
	service.Query<- kvq
	resp := <-kvq.Response
	sendResponse(rw, resp)
}

func (service *IpfsNamespaceService) getMapValues(rw http.ResponseWriter, req *http.Request) {

}

func (service *IpfsNamespaceService) joinMap(rw http.ResponseWriter, req *http.Request) {

}

func (service *IpfsNamespaceService) getSet(rw http.ResponseWriter, req *http.Request) {

}

func (service *IpfsNamespaceService) joinSet(rw http.ResponseWriter, req *http.Request) {

}

func sendResponse(rw http.ResponseWriter, resp KvResponse) {
	if resp.Err != nil {
		err := sendGob(rw, resp.Val)
		log.Printf("Error sending gob: %v", err)
	}
}

func sendGob(rw http.ResponseWriter, gobber interface{}) error {
	// Encode gob into buffer first to check for encoding errors.
	// TODO is that actually a good idea?
	buff := bytes.Buffer{}
	enc := gob.NewEncoder(&buff)
	encerr := enc.Encode(gobber)

	if encerr != nil {
		panic(fmt.Sprintf("BUG encoding error: %v", encerr))
	}

	rw.Header()["Content-Type"] = []string{"application/octet-stream"}
	_, senderr := rw.Write(buff.Bytes())

	if senderr != nil {
		return errors.Wrap(senderr, "sendGob failed")
	}

	return nil
}
