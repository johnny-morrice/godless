package godless

import (
	"bytes"
	"fmt"
	"net/http"
	"io"
	"strings"

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

func (client *Client) SendRawQuery(source string) (*ApiResponse, error) {
	return client.Post(MIME_QUERY, strings.NewReader(source))
}

func (client *Client) SendQuery(query *Query) (*ApiResponse, error) {
	buff := &bytes.Buffer{}
	encerr := togob(query, buff)

	if encerr != nil {
		return nil, errors.Wrap(encerr, "Gob encode failed")
	}

	return client.Post(MIME_GOB, buff)
}

func (client *Client) Post(bodyType string, body io.Reader) (*ApiResponse, error) {
	addr := fmt.Sprintf("%s/query/run", client.Addr)
	resp, err := client.Http.Post(addr, bodyType, body)

	if err != nil {
		return nil, errors.Wrap(err, "HTTP POST failed")
	}

	defer resp.Body.Close()

	var apiresp *ApiResponse
	var apierr *ApiError
	ct := resp.Header[CONTENT_TYPE]
	if resp.StatusCode == 200 {
		if linearContains(ct, MIME_GOB) {
			apiresp = &ApiResponse{}
			err = degob(apiresp, resp.Body)
		} else {
			return nil, incorrectContentType(resp.StatusCode, ct)
		}
	} else {
		if linearContains(ct, MIME_GOB) {
			apierr = &ApiError{}
			err = dejson(apierr, resp.Body)
		} else {
			return nil, incorrectContentType(resp.StatusCode, ct)
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error decoding API response")
	}

	if apiresp != nil {
		return apiresp, nil
	} else if apierr != nil {
		return nil, fmt.Errorf("API error: %v", apierr.Err)
	} else {
		panic("bug")
	}
}

func incorrectContentType(status int, ct []string) error {
	return fmt.Errorf("%v response had incorrect content type, was: %v", status,  ct)
}
