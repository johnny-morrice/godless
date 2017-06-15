package service

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	gohttp "net/http"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/internal/http"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

type Client struct {
	addr string
	web  *gohttp.Client
}

func MakeClient(addr string) *Client {
	return &Client{
		addr: addr,
		web:  http.DefaultFrontentClient(),
	}
}

func MakeClientWithHttp(addr string, webClient *gohttp.Client) *Client {
	return &Client{
		addr: addr,
		web:  webClient,
	}
}

func (client *Client) SendReflection(command api.APIReflectionType) (api.APIResponse, error) {
	var part string
	switch command {
	case api.REFLECT_HEAD_PATH:
		part = "head"
	case api.REFLECT_DUMP_NAMESPACE:
		part = "namespace"
	case api.REFLECT_INDEX:
		part = "index"
	default:
		return api.RESPONSE_FAIL, fmt.Errorf("Unknown api.APIReflectionType: %v", command)
	}

	path := fmt.Sprintf("%v/%v", REFLECT_API_ROOT, part)
	return client.Post(path, http.MIME_EMPTY, &bytes.Buffer{})
}

func (client *Client) SendQuery(q *query.Query) (api.APIResponse, error) {
	validerr := q.Validate()

	if validerr != nil {
		return api.RESPONSE_FAIL, errors.Wrap(validerr, fmt.Sprintf("Cowardly refusing to send invalid query: %v", q))
	}

	buff := &bytes.Buffer{}
	encerr := query.EncodeQuery(q, buff)

	if encerr != nil {
		return api.RESPONSE_FAIL, errors.Wrap(encerr, "SendQuery failed")
	}

	return client.Post(QUERY_API_ROOT, http.MIME_PROTO, buff)
}

func (client *Client) Post(path, bodyType string, body io.Reader) (api.APIResponse, error) {
	addr := fmt.Sprintf("http://%s%s%s", client.addr, API_ROOT, path)
	log.Info("HTTP POST to %v", addr)

	resp, err := client.web.Post(addr, bodyType, body)

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

func (client *Client) decodeHttpResponse(resp *gohttp.Response) (api.APIResponse, error) {
	if resp.StatusCode == WEB_API_SUCCESS {
		return client.decodeSuccessResponse(resp)
	} else if resp.StatusCode == WEB_API_ERROR {
		return client.decodeFailureResponse(resp)
	} else {
		return client.decodeUnexpectedResponse(resp)
	}
}

func (client *Client) decodeFailureResponse(resp *gohttp.Response) (api.APIResponse, error) {
	ct := resp.Header[http.CONTENT_TYPE]

	if util.LinearContains(ct, http.MIME_PROTO) {
		return api.DecodeAPIResponseText(resp.Body)
	} else {
		return api.RESPONSE_FAIL, incorrectContentType(resp.StatusCode, ct)
	}
}

func (client *Client) decodeUnexpectedResponse(resp *gohttp.Response) (api.APIResponse, error) {
	ct := resp.Header[http.CONTENT_TYPE]

	if util.LinearContains(ct, "text/plain; charset=utf-8") {
		all, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Warn("Failed to read response body")
			return api.RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%v): %v", resp.StatusCode, ct)
		}

		return api.RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%v): \n\n%v", resp.StatusCode, string(all))
	} else {
		return api.RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%v): %v", resp.StatusCode, ct)
	}
}

func (client *Client) decodeSuccessResponse(resp *gohttp.Response) (api.APIResponse, error) {
	ct := resp.Header[http.CONTENT_TYPE]
	if util.LinearContains(ct, http.MIME_PROTO) {
		return api.DecodeAPIResponse(resp.Body)
	} else {
		return api.RESPONSE_FAIL, incorrectContentType(resp.StatusCode, ct)
	}
}

func incorrectContentType(status int, ct []string) error {
	return fmt.Errorf("%v response had incorrect content type, was: %v", status, ct)
}
