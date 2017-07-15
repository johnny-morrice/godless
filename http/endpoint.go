package http

type Endpoints struct {
	CommandEndpoint string
}

func (endpoint *Endpoints) IsCommandEndpoint(url string) bool {
	panic("not implemented")
}

// UseDefaultEndpoints overwrites empty endpoints with Defaults.
// It does not overwrite a non-empty endpoints.
func (endpoint *Endpoints) UseDefaultEndpoints() {
	if endpoint.CommandEndpoint == "" {
		endpoint.CommandEndpoint = API_ROOT
	}
}
