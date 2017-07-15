package api

import (
	gohttp "net/http"
)

type Client interface {
	Send(request Request) (Response, error)
}

type WebService interface {
	GetApiRequestHandler() gohttp.Handler
	Close()
}
