package http

import (
	gohttp "net/http"
	"time"
)

func MakeBackendHttpClient(timeout time.Duration) *gohttp.Client {
	return &gohttp.Client{
		Timeout: time.Duration(timeout),
		Transport: &gohttp.Transport{
			DisableKeepAlives: true,
		},
	}
}
