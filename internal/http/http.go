package http

import (
	"net/http"
	"time"
)

var __backendClient *http.Client
var __frontendClient *http.Client
var __pingClient *http.Client

func DefaultFrontentClient() *http.Client {
	if __frontendClient == nil {
		__frontendClient = &http.Client{
			Timeout: time.Duration(__FRONTEND_TIMEOUT),
		}
	}

	return __frontendClient
}

func DefaultBackendClient() *http.Client {
	if __backendClient == nil {
		__backendClient = makeBackendClient(__BACKEND_TIMEOUT)
	}

	return __backendClient
}

func makeBackendClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: time.Duration(__BACKEND_TIMEOUT),
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
}

const __PING_TIMEOUT = 10 * time.Second
const __BACKEND_TIMEOUT = 10 * time.Minute
const __FRONTEND_TIMEOUT = 1 * time.Minute
