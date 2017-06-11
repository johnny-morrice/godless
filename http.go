package godless

import (
	"net/http"
	"time"
)

var __backendClient *http.Client
var __frontendClient *http.Client

func defaultFrontentClient() *http.Client {
	if __frontendClient == nil {
		__frontendClient = &http.Client{
			Timeout: time.Duration(__FRONTEND_TIMEOUT),
		}
	}

	return __frontendClient
}

func defaultBackendClient() *http.Client {
	if __backendClient == nil {
		__backendClient = &http.Client{
			Timeout: time.Duration(__BACKEND_TIMEOUT),
		}
	}

	return __backendClient
}

const __BACKEND_TIMEOUT = 10 * time.Minute
const __FRONTEND_TIMEOUT = 1 * time.Minute
