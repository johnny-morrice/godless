package mock_godless

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/query"
)

func TestApiReplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)

	const path = "Hello"
	links := []crdt.Link{crdt.UnsignedLink(path)}

	mock.EXPECT().Replicate(links, commandMatcher{}).Do(replicateStub)
	mock.EXPECT().Close()

	api, errch := launchAPI(mock)
	defer tidyApi(t, api, errch)

	respch, err := runReplicate(api, links)

	testutil.AssertNil(t, err)
	testutil.AssertNonNil(t, respch)
	validateResponseCh(t, respch)
}

func TestApiReflect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)

	reflection := api.REFLECT_HEAD_PATH

	mock.EXPECT().Reflect(reflection, commandMatcher{}).Do(reflectStub)
	mock.EXPECT().Close()

	api, errch := launchAPI(mock)
	defer tidyApi(t, api, errch)

	respch, err := runReflect(api, reflection)

	testutil.AssertNil(t, err)
	testutil.AssertNonNil(t, respch)
	validateResponseCh(t, respch)
}

func TestApiQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)

	query, err := query.Compile("select things where str_eq(stuff, \"Hello\")")
	testutil.AssertNil(t, err)

	mock.EXPECT().RunQuery(query, commandMatcher{}).Do(runQueryStub)
	mock.EXPECT().Close()

	api, errch := launchAPI(mock)
	defer tidyApi(t, api, errch)

	respch, err := runQuery(api, query)

	testutil.AssertNil(t, err)
	testutil.AssertNonNil(t, respch)
	validateResponseCh(t, respch)
}

func validateResponseCh(t *testing.T, respch <-chan api.Response) api.Response {
	timeout := time.NewTimer(__TEST_TIMEOUT)

	select {
	case <-timeout.C:
		t.Error("Timeout reading response")
		t.FailNow()
		return api.Response{}
	case r := <-respch:
		timeout.Stop()
		return r
	}
}

func replicateStub(links []crdt.Link, command api.Command) {
	command.Response <- api.RESPONSE_QUERY
}

func reflectStub(reflection api.ReflectionType, command api.Command) {
	command.Response <- api.RESPONSE_QUERY
}

func runQueryStub(q *query.Query, command api.Command) {
	command.Response <- api.RESPONSE_QUERY
}

func TestApiInvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)

	mock.EXPECT().Close()

	request := api.Request{}

	api, errch := launchAPI(mock)

	defer tidyApi(t, api, errch)

	resp, err := api.Call(request)

	if resp != nil {
		t.Error("Expected nil response")
	}

	testutil.AssertNonNil(t, err)
}

// No EXPECT but still valid mock: verifies no calls.
func TestApiInvalidQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)
	mock.EXPECT().Close()

	query := &query.Query{}

	api, errch := launchAPI(mock)

	defer tidyApi(t, api, errch)

	resp, err := runQuery(api, query)

	if resp != nil {
		t.Error("Expected nil response")
	}

	testutil.AssertNonNil(t, err)
}

func tidyApi(t *testing.T, api api.Service, errch <-chan error) {
	api.CloseAPI()

	for err := range errch {
		t.Error(err)
	}
}

func runReplicate(service api.RequestService, links []crdt.Link) (<-chan api.Response, error) {
	return service.Call(api.MakeReplicateRequest(links))
}

func runReflect(service api.RequestService, reflection api.ReflectionType) (<-chan api.Response, error) {
	return service.Call(api.MakeReflectRequest(reflection))
}

func runQuery(service api.RequestService, query *query.Query) (<-chan api.Response, error) {
	return service.Call(api.MakeQueryRequest(query))
}

func launchAPI(core api.Core) (api.Service, <-chan error) {
	const queryLimit = 1
	return launchConcurrentAPI(core, queryLimit)
}

func launchConcurrentAPI(core api.Core, queryLimit int) (api.Service, <-chan error) {
	queue := cache.MakeResidentBufferQueue(__UNKNOWN_CACHE_SIZE)
	options := service.QueuedApiServiceOptions{
		Core:       core,
		Queue:      queue,
		QueryLimit: queryLimit,
		Validator:  api.StandardRequestValidator(),
	}
	return service.LaunchQueuedApiService(options)
}

type commandMatcher struct {
}

func (commandMatcher) String() string {
	return "any KvQuery"
}

func (commandMatcher) Matches(v interface{}) bool {
	_, ok := v.(api.Command)

	return ok
}

const __TEST_TIMEOUT = time.Second * 1
