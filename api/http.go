package api

import (
	"net/url"

	gohttp "net/http"
)

type Client interface {
	Send(request Request) (Response, error)
}

type WebServiceOptions struct {
	Endpoints
	Api RequestService
}

type WebService interface {
	SetOptions(options WebServiceOptions)
	GetApiRequestHandler() gohttp.Handler
	Close()
}

type Endpoints struct {
	CommandEndpoint string
}

func (endpoint *Endpoints) IsCommandEndpoint(url *url.URL) bool {
	return endpoint.CommandEndpoint == url.Path
}
