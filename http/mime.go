package http

import (
	"fmt"
	gohttp "net/http"

	"github.com/johnny-morrice/godless/internal/util"
)

const MIME_PROTO_TEXT = "text/plain"
const MIME_PROTO = "application/octet-stream"

// TODO That we have a MIME_EMPTY indicates design flaw.
const MIME_EMPTY = "text/plain"
const CONTENT_TYPE = "Content-Type"

func HasContentType(header gohttp.Header, contentType string) bool {
	headers, ok := header[CONTENT_TYPE]

	if !ok {
		return false
	}

	return util.LinearContains(headers, contentType)
}

func incorrectContentType(header gohttp.Header) error {
	return fmt.Errorf("Incorrect content type, was: %v", header[CONTENT_TYPE])
}
