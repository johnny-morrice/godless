package godless

import (
	"encoding/gob"
	"encoding/json"
	"io"
)

func degob(gobbee interface{}, r io.Reader) error {
	dec := gob.NewDecoder(r)
	return dec.Decode(gobbee)
}

func togob(gobbee interface{}, w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(gobbee)
}

func tojson(jsonee interface{}, w io.Writer) error {
	enc := json.NewEncoder(w)
 	return enc.Encode(jsonee)
}

func dejson(jsonee interface{}, r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(jsonee)
}
