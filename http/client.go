package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type ClientOptions struct {
	Endpoints
	ServerAddr string
	Http       *gohttp.Client
}

type client struct {
	ClientOptions
}

func MakeClient(options ClientOptions) (api.Client, error) {
	client := &client{ClientOptions: options}

	if client.ServerAddr == "" {
		return nil, errors.New("Expected ServerAddr")
	}

	client.UseDefaultEndpoints()

	if client.Http == nil {
		client.Http = defaultHttpClient()
	}

	return client, nil
}

func (client *client) Send(request api.Request) (api.Response, error) {
	err := request.Validate()

	if err != nil {
		return api.RESPONSE_FAIL, errors.Wrap(err, fmt.Sprintf("Cowardly refusing to send invalid Request: %v", request))
	}

	buff := &bytes.Buffer{}
	err = api.EncodeRequest(request, buff)

	if err != nil {
		return api.RESPONSE_FAIL, errors.Wrap(err, "SendQuery failed")
	}

	return client.Post(client.CommandEndpoint, MIME_PROTO, buff)
}

func (client *client) Post(path, bodyType string, body io.Reader) (api.Response, error) {
	addr := client.ServerAddr + path
	log.Info("HTTP POST to %s", addr)

	resp, err := client.Http.Post(addr, bodyType, body)

	if err != nil {
		return api.RESPONSE_FAIL, errors.Wrap(err, "HTTP POST failed")
	}

	defer resp.Body.Close()

	apiresp, err := client.decodeHttpResponse(resp)

	if err != nil {
		return api.RESPONSE_FAIL, errors.Wrap(err, "Error decoding API response")
	}

	if apiresp.Err == nil {
		return apiresp, nil
	} else {
		return apiresp, errors.Wrap(apiresp.Err, "API returned error")
	}
}

func (client *client) decodeHttpResponse(resp *gohttp.Response) (api.Response, error) {
	if resp.StatusCode == WEB_API_SUCCESS {
		return client.decodeSuccessResponse(resp)
	} else if resp.StatusCode == WEB_API_ERROR {
		return client.decodeFailureResponse(resp)
	} else {
		return client.decodeUnexpectedResponse(resp)
	}
}

func (client *client) decodeFailureResponse(resp *gohttp.Response) (api.Response, error) {
	ct := resp.Header[CONTENT_TYPE]

	if util.LinearContains(ct, MIME_PROTO) {
		return api.DecodeResponseText(resp.Body)
	} else {
		return api.RESPONSE_FAIL, incorrectContentType(resp.StatusCode, ct)
	}
}

func (client *client) decodeUnexpectedResponse(resp *gohttp.Response) (api.Response, error) {
	ct := resp.Header[CONTENT_TYPE]

	if util.LinearContains(ct, "text/plain; charset=utf-8") {
		all, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Warn("Failed to read response body")
			return api.RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%d): %v", resp.StatusCode, ct)
		}

		return api.RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%d): \n\n%s", resp.StatusCode, string(all))
	} else {
		return api.RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%d): %v", resp.StatusCode, ct)
	}
}

func (client *client) decodeSuccessResponse(resp *gohttp.Response) (api.Response, error) {
	ct := resp.Header[CONTENT_TYPE]
	if util.LinearContains(ct, MIME_PROTO) {
		return api.DecodeResponse(resp.Body)
	} else {
		return api.RESPONSE_FAIL, incorrectContentType(resp.StatusCode, ct)
	}
}

func incorrectContentType(status int, ct []string) error {
	return fmt.Errorf("%d response had incorrect content type, was: %v", status, ct)
}

var __frontendClient *gohttp.Client

func defaultHttpClient() *gohttp.Client {
	if __frontendClient == nil {
		__frontendClient = &gohttp.Client{
			Timeout: time.Duration(__FRONTEND_TIMEOUT),
		}
	}

	return __frontendClient
}

const __FRONTEND_TIMEOUT = 1 * time.Minute
