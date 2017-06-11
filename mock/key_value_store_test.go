package mock_godless

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
	"github.com/pkg/errors"
)

func TestRunQueryReadSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockKvNamespace(ctrl)
	query := &lib.Query{
		OpCode:   lib.SELECT,
		TableKey: "Table Key",
		Select: lib.QuerySelect{
			Limit: 1,
			Where: lib.QueryWhere{
				OpCode: lib.PREDICATE,
				Predicate: lib.QueryPredicate{
					OpCode:   lib.STR_EQ,
					Literals: []string{"Hi"},
					Keys:     []lib.EntryName{"Entry A"},
				},
			},
		},
	}

	mock.EXPECT().IsChanged().Return(false)
	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)

	api, errch := lib.LaunchKeyValueStore(mock)
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

func validateResponseCh(t *testing.T, respch <-chan lib.APIResponse) lib.APIResponse {
	timeout := time.NewTimer(__TEST_TIMEOUT)

	for {
		select {
		case <-timeout.C:
			t.Error("Timeout reading response")
			t.FailNow()
			return lib.APIResponse{}
		case r := <-respch:
			timeout.Stop()
			return r
		}
	}
}

func writeStubResponse(q *lib.Query, kvq lib.KvQuery) {
	kvq.Response <- lib.RESPONSE_QUERY
}

func TestRunQueryWriteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockKvNamespace(ctrl)
	query := &lib.Query{
		OpCode:   lib.JOIN,
		TableKey: "Table Key",
		Join: lib.QueryJoin{
			Rows: []lib.QueryRowJoin{
				lib.QueryRowJoin{
					RowKey: "Row thing",
					Entries: map[lib.EntryName]lib.Point{
						"Hello": "world",
					},
				},
			},
		},
	}

	mock.EXPECT().IsChanged().Return(true)
	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)
	mock.EXPECT().Persist().Return(mock, nil)

	api, errch := lib.LaunchKeyValueStore(mock)
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
	query := &lib.Query{
		OpCode:   lib.JOIN,
		TableKey: "Table Key",
		Join: lib.QueryJoin{
			Rows: []lib.QueryRowJoin{
				lib.QueryRowJoin{
					RowKey: "Row thing",
					Entries: map[lib.EntryName]lib.Point{
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

	api, errch := lib.LaunchKeyValueStore(mock)
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
	query := &lib.Query{}

	api, _ := lib.LaunchKeyValueStore(mock)
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
	_, ok := v.(lib.KvQuery)

	return ok
}

const __TEST_TIMEOUT = time.Second * 1
