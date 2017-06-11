package mock_godless

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

func TestRunQueryReadSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockKvNamespace(ctrl)
	query := &query.Query{
		OpCode:   query.SELECT,
		TableKey: "Table Key",
		Select: query.QuerySelect{
			Limit: 1,
			Where: query.QueryWhere{
				OpCode: query.PREDICATE,
				Predicate: query.QueryPredicate{
					OpCode:   query.STR_EQ,
					Literals: []string{"Hi"},
					Keys:     []crdt.EntryName{"Entry A"},
				},
			},
		},
	}

	mock.EXPECT().IsChanged().Return(false)
	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)

	api, errch := service.LaunchKeyValueStore(mock)
	respch, err := api.RunQuery(query)

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

func validateResponseCh(t *testing.T, respch <-chan api.APIResponse) api.APIResponse {
	timeout := time.NewTimer(__TEST_TIMEOUT)

	for {
		select {
		case <-timeout.C:
			t.Error("Timeout reading response")
			t.FailNow()
			return api.APIResponse{}
		case r := <-respch:
			timeout.Stop()
			return r
		}
	}
}

func writeStubResponse(q *query.Query, kvq api.KvQuery) {
	kvq.Response <- api.RESPONSE_QUERY
}

func TestRunQueryWriteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockKvNamespace(ctrl)
	query := &query.Query{
		OpCode:   query.JOIN,
		TableKey: "Table Key",
		Join: query.QueryJoin{
			Rows: []query.QueryRowJoin{
				query.QueryRowJoin{
					RowKey: "Row thing",
					Entries: map[crdt.EntryName]crdt.Point{
						"Hello": "world",
					},
				},
			},
		},
	}

	mock.EXPECT().IsChanged().Return(true)
	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)
	mock.EXPECT().Persist().Return(mock, nil)

	api, errch := service.LaunchKeyValueStore(mock)
	actualRespch, err := api.RunQuery(query)

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

func TestRunQueryWriteFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockKvNamespace(ctrl)
	query := &query.Query{
		OpCode:   query.JOIN,
		TableKey: "Table Key",
		Join: query.QueryJoin{
			Rows: []query.QueryRowJoin{
				query.QueryRowJoin{
					RowKey: "Row thing",
					Entries: map[crdt.EntryName]crdt.Point{
						"Hello": "world",
					},
				},
			},
		},
	}

	mock.EXPECT().IsChanged().Return(true)
	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)
	mock.EXPECT().Reset()
	mock.EXPECT().Persist().Return(nil, errors.New("Expected error"))

	api, errch := service.LaunchKeyValueStore(mock)
	resp, qerr := api.RunQuery(query)

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
func TestRunQueryInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockKvNamespace(ctrl)
	query := &query.Query{}

	api, _ := service.LaunchKeyValueStore(mock)
	resp, err := api.RunQuery(query)

	if err == nil {
		t.Error("err was nil")
	}

	if resp != nil {
		t.Error("Response channel was not nil")
	}

	api.CloseAPI()
}

type kvqmatcher struct {
}

func (kvqmatcher) String() string {
	return "any KvQuery"
}

func (kvqmatcher) Matches(v interface{}) bool {
	_, ok := v.(api.KvQuery)

	return ok
}

const __TEST_TIMEOUT = time.Second * 1
