package mock_godless

import (
	"reflect"
	"testing"

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
	mock.EXPECT().RunKvQuery(query, mtchkvqq(query))

	api, errch := lib.LaunchKeyValueStore(mock)
	resp, err := api.RunQuery(query)

	if err != nil {
		t.Error(err)
	}

	if resp == nil {
		t.Error("Response channel was nil")
	}

	api.CloseAPI()

	for err := range errch {
		t.Error(err)
	}
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
	mock.EXPECT().RunKvQuery(query, mtchkvqq(query))
	mock.EXPECT().Persist().Return(mock, nil)

	api, errch := lib.LaunchKeyValueStore(mock)
	resp, err := api.RunQuery(query)

	if err != nil {
		t.Error(err)
	}

	if resp == nil {
		t.Error("Response channel was nil")
	}

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
	mock.EXPECT().RunKvQuery(query, mtchkvqq(query))
	mock.EXPECT().Persist().Return(nil, errors.New("Expected error"))

	api, errch := lib.LaunchKeyValueStore(mock)
	resp, qerr := api.RunQuery(query)

	if qerr != nil {
		t.Error(qerr)
	}

	if resp == nil {
		t.Error("Response channel was nil")
	}

	api.CloseAPI()

	if err := <-errch; err == nil {
		t.Error("err was nil")
	}

	empty := lib.APIResponse{}
	if r := <-resp; !reflect.DeepEqual(r, empty) {
		t.Error("Non zero APIResponse")
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

func mtchkvqq(q *lib.Query) gomock.Matcher {
	return kvqqmatcher{q}
}

type kvqqmatcher struct {
	q *lib.Query
}

func (kvqqm kvqqmatcher) String() string {
	return "is matching KvQuery"
}

func (kvqqm kvqqmatcher) Matches(v interface{}) bool {
	other, ok := v.(lib.KvQuery)

	if !ok {
		return false
	}

	return reflect.DeepEqual(*kvqqm.q, *other.Query)
}
