package http

import (
	"net/url"
)

type Endpoints struct {
	CommandEndpoint string
}

func (endpoint *Endpoints) IsCommandEndpoint(url *url.URL) bool {
	return endpoint.CommandEndpoint == url.Path
}

// UseDefaultEndpoints overwrites empty endpoints with Defaults.
// It does not overwrite a non-empty endpoints.
func (endpoint *Endpoints) UseDefaultEndpoints() {
	if endpoint.CommandEndpoint == "" {
		endpoint.CommandEndpoint = API_ROOT
	}
}
