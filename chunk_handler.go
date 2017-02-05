package godless

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/mux"
)

type ChunkMux struct {
	Sets map[string]*ChunkSet
}

func (chmux ChunkMux) Handler() http.Handler {
	r := mux.NewRouter()

	for name, set := range chmux.Sets {
		r.Handle(name, ChunkHandler{Set: set})
	}

	return r
}

type ChunkHandler struct {
	Set *ChunkSet
}

func (h ChunkHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	enc := gob.NewEncoder(rw)

	headers := rw.Header()
	headers["content-type"] = []string{"application/octet-stream"}

	for _, ch := range h.Set.Chunks {
		enc.Encode(ch)
	}
}
