package mock_godless

import (
	"bytes"
	"fmt"
	gohttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/http"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
)

func TestWebServiceGetApiRequestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockService(ctrl)

	const SIZE = 50
	requestA := api.GenRequest(testutil.Rand(), SIZE)
	requestB := api.GenRequest(testutil.Rand(), SIZE)
	expected := api.GenResponse(testutil.Rand(), SIZE)

	respch := make(chan api.Response, 1)
	respch <- expected

	mock.EXPECT().Call(matchRequest(requestA)).Return(respch, nil)
	mock.EXPECT().Call(matchRequest(requestB)).Return(nil, expectedError())

	options := api.WebServiceOptions{
		Api: mock,
		Endpoints: api.Endpoints{
			CommandEndpoint: TEST_COMMAND_ENDPOINT,
		},
	}

	webService := http.MakeWebService(options)
	defer webService.Close()

	handler := webService.GetApiRequestHandler()
	actual, err := webApiCall(handler, requestA)
	testutil.AssertNil(t, err)
	testutil.Assert(t, "Unexpected api.Response", expected.Equals(actual))

	resp, err := webApiCall(handler, requestB)
	testutil.AssertNil(t, err)
	testutil.AssertNonNil(t, resp.Err)

	resp, err = webApiHttpCall(handler, httptest.NewRequest("GET", TEST_SERVER_ADDR, nil))
	testutil.AssertNonNil(t, err)
}

func webApiCall(handler gohttp.Handler, request api.Request) (api.Response, error) {
	buff := &bytes.Buffer{}
	err := api.EncodeRequest(request, buff)
	panicOnBadInit(err)

	httpRequest := httptest.NewRequest("POST", WEB_SERVICE_URL, buff)

	return webApiHttpCall(handler, httpRequest)
}

func webApiHttpCall(handler gohttp.Handler, request *gohttp.Request) (api.Response, error) {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	recorder.Flush()
	resp := recorder.Result()

	log.Debug("HTTP response: %v", resp)

	statusOk := resp.StatusCode == http.WEB_API_ERROR
	statusOk = statusOk || resp.StatusCode == http.WEB_API_SUCCESS

	var apiResp api.Response
	var err error
	if http.HasContentType(resp.Header, http.MIME_PROTO) {
		apiResp, err = api.DecodeResponse(resp.Body)
	} else if http.HasContentType(resp.Header, http.MIME_PROTO_TEXT) {
		apiResp, err = api.DecodeResponseText(resp.Body)
	} else {
		log.Debug("failed to decode api response")
	}

	if !statusOk && err == nil {
		return api.Response{}, fmt.Errorf("Bad response status: %v", resp.StatusCode)
	}

	return apiResp, err
}

func matchRequest(request api.Request) gomock.Matcher {
	return requestMatcher{request: request}
}

type requestMatcher struct {
	request api.Request
}

func (matcher requestMatcher) String() string {
	return fmt.Sprintf("matches request: %v", matcher.request)
}

func (matcher requestMatcher) Matches(any interface{}) bool {
	other, ok := any.(api.Request)

	if !ok {
		return false
	}

	return matcher.request.Equals(other)
}

const WEB_SERVICE_URL = TEST_SERVER_ADDR + TEST_COMMAND_ENDPOINT
