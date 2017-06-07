package godless

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type Client struct {
	Addr string
	Http *http.Client
}

func MakeClient(addr string) *Client {
	return &Client{
		Addr: addr,
		Http: defaultHttpClient(),
	}
}

func (client *Client) SendReflection(command APIReflectionType) (APIResponse, error) {
	var part string
	switch command {
	case REFLECT_HEAD_PATH:
		part = "head"
	case REFLECT_DUMP_NAMESPACE:
		part = "namespace"
	case REFLECT_INDEX:
		part = "index"
	default:
		return RESPONSE_FAIL, fmt.Errorf("Unknown APIReflectionType: %v", command)
	}

	path := fmt.Sprintf("%v/%v", REFLECT_API_ROOT, part)
	return client.Post(path, MIME_EMPTY, &bytes.Buffer{})
}

func (client *Client) SendQuery(query *Query) (APIResponse, error) {
	validerr := query.Validate()

	if validerr != nil {
		return RESPONSE_FAIL, errors.Wrap(validerr, fmt.Sprintf("Cowardly refusing to send invalid query: %v", query))
	}

	buff := &bytes.Buffer{}
	encerr := EncodeQuery(query, buff)

	if encerr != nil {
		return RESPONSE_FAIL, errors.Wrap(encerr, "SendQuery failed")
	}

	return client.Post(QUERY_API_ROOT, MIME_PROTO, buff)
}

func (client *Client) Post(path, bodyType string, body io.Reader) (APIResponse, error) {
	addr := fmt.Sprintf("http://%s%s%s", client.Addr, API_ROOT, path)
	logdbg("HTTP POST to %v", addr)

	resp, err := client.Http.Post(addr, bodyType, body)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, "HTTP POST failed")
	}

	defer resp.Body.Close()

	apiresp, err := client.decodeHttpResponse(resp)

	if err != nil {
		return RESPONSE_FAIL, errors.Wrap(err, "Error decoding API response")
	}

	if apiresp.Err == nil {
		return apiresp, nil
	} else {
		return apiresp, errors.Wrap(apiresp.Err, "API returned error")
	}
}

func (client *Client) decodeHttpResponse(resp *http.Response) (APIResponse, error) {
	var apiresp APIResponse
	var err error

	// TODO this is a bit horrible.
	if resp.StatusCode == 200 {
		return client.decodeSuccessResponse(resp)
	} else if resp.StatusCode == 500 {
		return client.decodeFailureResponse(resp)
	} else {
		return client.decodeUnexpectedResponse(resp)
	}

	return apiresp, err
}

func (client *Client) decodeFailureResponse(resp *http.Response) (APIResponse, error) {
	ct := resp.Header[CONTENT_TYPE]

	if linearContains(ct, MIME_PROTO) {
		return DecodeAPIResponseText(resp.Body)
	} else {
		return RESPONSE_FAIL, incorrectContentType(resp.StatusCode, ct)
	}
}

func (client *Client) decodeUnexpectedResponse(resp *http.Response) (APIResponse, error) {
	ct := resp.Header[CONTENT_TYPE]

	if linearContains(ct, "text/plain; charset=utf-8") {
		all, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			logwarn("Failed to read response body")
			return RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%v): %v", resp.StatusCode, ct)
		}

		return RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%v): \n\n%v", resp.StatusCode, string(all))
	} else {
		return RESPONSE_FAIL, fmt.Errorf("Unexpected API response (%v): %v", resp.StatusCode, ct)
	}
}

func (client *Client) decodeSuccessResponse(resp *http.Response) (APIResponse, error) {
	ct := resp.Header[CONTENT_TYPE]
	if linearContains(ct, MIME_PROTO) {
		return DecodeAPIResponse(resp.Body)
	} else {
		return RESPONSE_FAIL, incorrectContentType(resp.StatusCode, ct)
	}
}

func incorrectContentType(status int, ct []string) error {
	return fmt.Errorf("%v response had incorrect content type, was: %v", status, ct)
}
