package http

import (
	"github.com/johnny-morrice/godless/api"
)

// setDefaultEndpoints overwrites empty endpoints with Defaults.
// It does not overwrite a non-empty endpoints.
func setDefaultEndpoints(endpoint *api.Endpoints) {
	if endpoint.CommandEndpoint == "" {
		endpoint.CommandEndpoint = API_ROOT
	}
}
