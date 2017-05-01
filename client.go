package godless

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

func (client *Client) SendRawQuery(source string) (*APIResponse, error) {
	return client.Post(MIME_QUERY, strings.NewReader(source))
}

func (client *Client) SendQuery(query *Query) (*APIResponse, error) {
	validerr := query.Validate()

	if validerr != nil {
		return nil, errors.Wrap(validerr, fmt.Sprintf("Cowardly refusing to send invalid query: %v", query))
	}

	buff := &bytes.Buffer{}
	encerr := togob(query, buff)

	if encerr != nil {
		return nil, errors.Wrap(encerr, "Gob encode failed")
	}

	return client.Post(MIME_GOB, buff)
}

func (client *Client) Post(bodyType string, body io.Reader) (*APIResponse, error) {
	addr := fmt.Sprintf("http://%s/api/query/run", client.Addr)
	logdbg("HTTP POST to %v", addr)

	resp, err := client.Http.Post(addr, bodyType, body)

	if err != nil {
		return nil, errors.Wrap(err, "HTTP POST failed")
	}

	defer resp.Body.Close()

	var apiresp *APIResponse
	ct := resp.Header[CONTENT_TYPE]
	if resp.StatusCode == 200 {
		if linearContains(ct, MIME_GOB) {
			apiresp = &APIResponse{}
			err = degob(apiresp, resp.Body)
		} else {
			return nil, incorrectContentType(resp.StatusCode, ct)
		}
	} else if resp.StatusCode == 500 {
		if linearContains(ct, MIME_GOB) {
			apiresp = &APIResponse{}
			err = degob(apiresp, resp.Body)
		} else {
			return nil, incorrectContentType(resp.StatusCode, ct)
		}
	} else {
		if linearContains(ct, "text/plain; charset=utf-8") {
			all, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				logwarn("Failed to read response body")
				return nil, fmt.Errorf("Unexpected API response (%v): %v", resp.StatusCode, ct)
			}

			return nil, fmt.Errorf("Unexpected API response (%v): \n\n%v", resp.StatusCode, string(all))
		} else {
			return nil, fmt.Errorf("Unexpected API response (%v): %v", resp.StatusCode, ct)
		}
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error decoding API response")
	}

	if apiresp.Err == nil {
		return apiresp, nil
	} else {
		return nil, errors.Wrap(apiresp.Err, "API returned error")
	}
}

func incorrectContentType(status int, ct []string) error {
	return fmt.Errorf("%v response had incorrect content type, was: %v", status, ct)
}
