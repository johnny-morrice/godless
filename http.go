package godless

import (
	"net/http"
	"time"
)

var __client *http.Client

func defaultHttpClient() *http.Client {
	if __client == nil {
		__client = &http.Client{
			Timeout: time.Duration(__DEFAULT_TIMEOUT),
		}
	}

	return __client
}

const __DEFAULT_TIMEOUT = 10 * time.Minute
