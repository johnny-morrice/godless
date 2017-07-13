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
	t.FailNow()
}

func TestApiReplicateFailure(t *testing.T) {
	t.FailNow()
}

func TestApiReflect(t *testing.T) {
	t.FailNow()
}

func TstApiReflectFailure(t *testing.T) {
	t.FailNow()
}

func TestApiQuerySuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// TODO test join query
	t.FailNow()

	mock := NewMockCore(ctrl)
	query, err := query.Compile("select things where str_eq(stuff, \"Hello\")")
	testutil.AssertNil(t, err)

	mock.EXPECT().RunQuery(query, kvqmatcher{}).Do(writeStubResponse)
	mock.EXPECT().Close()

	api, errch := launchAPI(mock)
	respch, err := runQuery(api, query)

	if err != nil {
		t.Error(err)
	}

	if respch == nil {
		t.Error("Response channel was nil")
	}

	validateResponseCh(t, respch)

	api.CloseAPI()

	for err := range errch {
		t.Error(err)
	}
}

func TestAPIQueryFailure(t *testing.T) {
	t.FailNow()
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

func writeStubResponse(q *query.Query, kvq api.Command) {
	kvq.Response <- api.RESPONSE_QUERY
}

func TestApiQueryJoinSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)
	query := &query.Query{
		OpCode:   query.JOIN,
		TableKey: "Table Key",
		Join: query.QueryJoin{
			Rows: []query.QueryRowJoin{
				query.QueryRowJoin{
					RowKey: "Row thing",
					Entries: map[crdt.EntryName]crdt.PointText{
						"Hello": "world",
					},
				},
			},
		},
	}

	mock.EXPECT().RunQuery(query, kvqmatcher{}).Do(writeStubResponse)
	mock.EXPECT().Close()

	api, errch := launchAPI(mock)
	actualRespch, err := runQuery(api, query)

	if err != nil {
		t.Error(err)
	}

	if actualRespch == nil {
		t.Error("Response channel was nil")
	}

	validateResponseCh(t, actualRespch)

	api.CloseAPI()

	for err := range errch {
		t.Error(err)
	}
}

func TestApiQueryJoinFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)
	query := &query.Query{
		OpCode:   query.JOIN,
		TableKey: "Table Key",
		Join: query.QueryJoin{
			Rows: []query.QueryRowJoin{
				query.QueryRowJoin{
					RowKey: "Row thing",
					Entries: map[crdt.EntryName]crdt.PointText{
						"Hello": "world",
					},
				},
			},
		},
	}

	mock.EXPECT().RunQuery(query, kvqmatcher{}).Do(writeStubResponse)
	mock.EXPECT().Close()

	api, errch := launchAPI(mock)
	resp, qerr := runQuery(api, query)

	if qerr != nil {
		t.Error(qerr)
	}

	if resp == nil {
		t.Error("Response channel was nil")
	}

	r := validateResponseCh(t, resp)

	api.CloseAPI()

	if err := <-errch; err != nil {
		t.Error("err was not nil")
	}

	if r.Err != nil {
		t.Error("Non failure APIResponse")
	}
}

// No EXPECT but still valid mock: verifies no calls.
func TestApiQueryInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockCore(ctrl)
	query := &query.Query{}

	mock.EXPECT().Close()

	api, _ := launchAPI(mock)
	resp, err := runQuery(api, query)

	if err == nil {
		t.Error("err was nil")
	}

	if resp != nil {
		t.Error("Response channel was not nil")
	}

	api.CloseAPI()
}

func runQuery(service api.RequestService, query *query.Query) (<-chan api.Response, error) {
	return service.Call(api.Request{Type: api.API_QUERY, Query: query})
}

func launchAPI(remote api.Core) (api.Service, <-chan error) {
	const queryLimit = 1
	return launchConcurrentAPI(remote, queryLimit)
}

func launchConcurrentAPI(remote api.Core, queryLimit int) (api.Service, <-chan error) {
	queue := cache.MakeResidentBufferQueue(__UNKNOWN_CACHE_SIZE)
	options := service.QueuedApiServiceOptions{
		Core:       remote,
		Queue:      queue,
		QueryLimit: queryLimit,
	}
	return service.LaunchQueuedApiService(options)
}

type kvqmatcher struct {
}

func (kvqmatcher) String() string {
	return "any KvQuery"
}

func (kvqmatcher) Matches(v interface{}) bool {
	_, ok := v.(api.Command)

	return ok
}

const __TEST_TIMEOUT = time.Second * 1
