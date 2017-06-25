package mock_godless

import (
	"math/rand"
	"sync"
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

func TestKeyValueStoreITCase(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	const queryCount = 100
	const genSize = 50
	const addrIndex = crdt.IPFSPath("Index Addr")
	const namespaceAddr = crdt.IPFSPath("Namespace Addr")
	const queryLimit = 50

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)

	queries := make([]*query.Query, queryCount)
	tables := []crdt.TableName{}

	for i := 0; i < len(queries); i++ {
		gen := query.GenQuery(testutil.Rand(), genSize)

		if gen.OpCode == query.JOIN {
			tables = append(tables, gen.TableKey)
		} else {
			// Use a real table on occasion :)
			if rand.Float32() < 0.9 && len(tables) > 0 {
				randIndex := rand.Intn(len(tables))
				randTable := tables[randIndex]
				gen.TableKey = randTable
			}
		}

		queries[i] = gen
	}

	remote := makeRemote(mock)
	api, errch := launchConcurrentAPI(remote, queryLimit)

	index := crdt.EmptyIndex()

	signedNamespaceAddr := crdt.UnsignedLink(namespaceAddr)
	for _, t := range tables {
		index = index.JoinTable(t, signedNamespaceAddr)
	}

	mock.EXPECT().AddIndex(gomock.Any()).MinTimes(1).Return(addrIndex, nil)
	mock.EXPECT().AddNamespace(gomock.Any()).MinTimes(1).Return(namespaceAddr, nil)
	mock.EXPECT().CatNamespace(gomock.Any()).MinTimes(1).Return(crdt.EmptyNamespace(), nil)

	// No index catting with memoryImage
	// mock.EXPECT().CatIndex(gomock.Any()).MinTimes(1).Return(index, nil)

	go func() {
		defer api.CloseAPI()
		wg := &sync.WaitGroup{}

		for _, q := range queries {
			query := q
			if rand.Float32() > 0.5 {
				wg.Add(1)
				go func() {
					checkPlausibleResponse(t, api, query)
					wg.Done()
				}()
			} else {
				checkPlausibleResponse(t, api, query)
			}
		}

		wg.Wait()
	}()

	for err := range errch {
		testutil.AssertNonNil(t, err)
	}
}

func checkPlausibleResponse(t *testing.T, service api.APIRequestService, q *query.Query) {
	respch, _ := runQuery(service, q)
	resp := <-respch

	if resp.IsEmpty() {
		t.Error("Response should not be empty")
	}
}

func TestRunQueryReadSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteNamespace(ctrl)
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

	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)
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

func validateResponseCh(t *testing.T, respch <-chan api.APIResponse) api.APIResponse {
	timeout := time.NewTimer(__TEST_TIMEOUT)

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

func writeStubResponse(q *query.Query, kvq api.KvQuery) {
	kvq.Response <- api.RESPONSE_QUERY
}

func TestRunQueryWriteSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteNamespace(ctrl)
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

	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)
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

func TestRunQueryWriteFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteNamespace(ctrl)
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

	mock.EXPECT().RunKvQuery(query, kvqmatcher{}).Do(writeStubResponse)
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
func TestRunQueryInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteNamespace(ctrl)
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

func runQuery(service api.APIRequestService, query *query.Query) (<-chan api.APIResponse, error) {
	return service.Call(api.APIRequest{Type: api.API_QUERY, Query: query})
}

func launchAPI(remote api.RemoteNamespace) (api.APIService, <-chan error) {
	const queryLimit = 1
	return launchConcurrentAPI(remote, queryLimit)
}

func launchConcurrentAPI(remote api.RemoteNamespace, queryLimit int) (api.APIService, <-chan error) {
	queue := cache.MakeResidentBufferQueue(__UNKNOWN_CACHE_SIZE)
	return service.LaunchKeyValueStore(remote, queue, queryLimit)
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
