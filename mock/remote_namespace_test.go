package mock_godless

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

func TestRemoteNamespaceJoinTable(t *testing.T) {
	TestRemoteNamespaceReset(t)
}

func TestRemoteNamespaceReplicate(t *testing.T) {
	// TODO integration test as design stands.
}

func TestRemoteNamespaceReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := NewMockRemoteStore(ctrl)

	const tableName = "ATableName"
	const rowName = "ARowName"
	const entryName = "AEntryName"
	const point = "APoint"
	table := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		rowName: crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			entryName: crdt.MakeEntry([]crdt.Point{point}),
		}),
	})

	remote := makeRemote(mockStore)
	assertUnchanged(t, remote)

	err := remote.JoinTable(tableName, table)
	testutil.AssertNil(t, err)
	assertChanged(t, remote)
}

func assertChanged(t *testing.T, remote api.RemoteNamespace) {
	if !remote.IsChanged() {
		t.Error("Expected remote change")
	}
}

func assertUnchanged(t *testing.T, remote api.RemoteNamespace) {
	if remote.IsChanged() {
		t.Error("Unexpected remote change")
	}
}

func TestRemoteNamespaceRunKvReflection(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := NewMockRemoteStore(ctrl)

	addrA := crdt.IPFSPath("Addr A")
	addrB := crdt.IPFSPath("Addr B")
	addrC := crdt.IPFSPath("Addr C")
	addrIndex := crdt.IPFSPath("Addr Index")

	empty := crdt.EmptyNamespace()
	tableA := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{"Point A"}),
		}),
	})
	tableB := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row B": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{"Point B"}),
		}),
	})
	tableC := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row C": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry C": crdt.MakeEntry([]crdt.Point{"Point C"}),
		}),
	})

	const tableAName = "Table A"
	const tableBName = "Table B"
	const tableCName = "Table C"

	namespaceA := empty.JoinTable(tableAName, tableA)
	namespaceB := empty.JoinTable(tableBName, tableB)
	namespaceC := empty.JoinTable(tableCName, tableC)

	index := crdt.MakeIndex(map[crdt.TableName]crdt.IPFSPath{
		tableAName: addrA,
		tableBName: addrB,
		tableCName: addrC,
	})

	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).AnyTimes()
	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatNamespace(addrB).Return(namespaceB, nil)
	mockStore.EXPECT().CatNamespace(addrC).Return(namespaceC, nil)

	joinedNamespace := namespaceA.JoinNamespace(namespaceB)
	joinedNamespace = joinedNamespace.JoinNamespace(namespaceC)

	remote := loadRemote(mockStore, addrIndex)

	testReflectHead(t, remote, addrIndex)
	testReflectIndex(t, remote, index)
	testReflectNamespace(t, remote, joinedNamespace)
}

// FIXME test error path
func testReflectHead(t *testing.T, remote api.RemoteNamespace, expected crdt.IPFSPath) {
	resp := reflectOnRemote(remote, api.REFLECT_HEAD_PATH)

	testutil.AssertNil(t, resp.Err)
	testutil.AssertEquals(t, "Unexpected HEAD path", expected, resp.ReflectResponse.Path)
}

func testReflectIndex(t *testing.T, remote api.RemoteNamespace, expected crdt.Index) {
	resp := reflectOnRemote(remote, api.REFLECT_INDEX)

	testutil.AssertNil(t, resp.Err)
	testutil.Assert(t, "Unexpected index", expected.Equals(resp.ReflectResponse.Index))
}

func testReflectNamespace(t *testing.T, remote api.RemoteNamespace, expected crdt.Namespace) {
	resp := reflectOnRemote(remote, api.REFLECT_DUMP_NAMESPACE)

	testutil.AssertNil(t, resp.Err)
	testutil.Assert(t, "Unexpected namespace", expected.Equals(resp.ReflectResponse.Namespace))
}

func reflectOnRemote(remote api.RemoteNamespace, reflection api.APIReflectionType) api.APIResponse {
	request := api.MakeKvReflect(api.APIRequest{Type: api.API_REFLECT, Reflection: reflection})
	go request.Run(remote)

	resp := readApiResponse(request)

	return resp
}

func TestIsChanged(t *testing.T) {
	TestRemoteNamespaceReset(t)
}

func TestRunKvQuery(t *testing.T) {
	// TODO integration test as design stands.
}

func TestLoadTraverseSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockReader := NewMockNamespaceTreeTableReader(ctrl)
	addrA := crdt.IPFSPath("Addr A")
	addrB := crdt.IPFSPath("Addr B")
	addrC := crdt.IPFSPath("Addr C")
	addrIndex := crdt.IPFSPath("Addr Index")

	empty := crdt.EmptyNamespace()
	tableA := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{"Point A"}),
		}),
	})
	tableB := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row B": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{"Point B"}),
		}),
	})
	tableC := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row C": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry C": crdt.MakeEntry([]crdt.Point{"Point C"}),
		}),
	})

	const tableAName = "Table A"
	const tableBName = "Table B"
	const tableCName = "Table C"

	namespaceA := empty.JoinTable(tableAName, tableA)
	namespaceB := empty.JoinTable(tableBName, tableB)
	namespaceC := empty.JoinTable(tableCName, tableC)

	index := crdt.MakeIndex(map[crdt.TableName]crdt.IPFSPath{
		tableAName: addrA,
		tableBName: addrB,
		tableCName: addrC,
	})

	keepReading := api.TraversalUpdate{More: true}

	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).AnyTimes()

	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatNamespace(addrB).Return(namespaceB, nil)
	mockStore.EXPECT().CatNamespace(addrC).Return(namespaceC, nil)

	mockReader.EXPECT().ReadsTables().Return([]crdt.TableName{tableAName, tableBName, tableCName})
	mockReader.EXPECT().ReadNamespace(matchNamespace(namespaceA)).Return(keepReading)
	mockReader.EXPECT().ReadNamespace(matchNamespace(namespaceB)).Return(keepReading)
	mockReader.EXPECT().ReadNamespace(matchNamespace(namespaceC)).Return(keepReading)

	remote := loadRemote(mockStore, addrIndex)

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockReader)

	if lterr != nil {
		t.Error(lterr)
	}
}

func TestLoadTraverseFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockReader := NewMockNamespaceTreeTableReader(ctrl)
	indexAddr := crdt.IPFSPath("Addr Index")
	namespaceAddr := crdt.IPFSPath("Addr A")

	empty := crdt.EmptyNamespace()
	tableA := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{"Point A"}),
		}),
	})

	namespaceA := empty.JoinTable("Table A", tableA)

	tableName := crdt.TableName("Table A")
	index := crdt.MakeIndex(map[crdt.TableName]crdt.IPFSPath{
		tableName: namespaceAddr,
	})

	badTraverse := api.TraversalUpdate{More: true, Error: errors.New("Expected error")}
	mockStore.EXPECT().CatNamespace(namespaceAddr).Return(namespaceA, nil)
	mockStore.EXPECT().CatIndex(indexAddr).Return(index, nil).AnyTimes()
	mockReader.EXPECT().ReadsTables().Return([]crdt.TableName{tableName})
	mockReader.EXPECT().ReadNamespace(matchNamespace(namespaceA)).Return(badTraverse)

	remote := loadRemote(mockStore, indexAddr)

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockReader)

	if lterr == nil {
		t.Error("lterr was nil")
	}
}

func TestLoadTraverseAbort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockReader := NewMockNamespaceTreeTableReader(ctrl)
	addrIndex := crdt.IPFSPath("Addr Index")
	addrA := crdt.IPFSPath("Addr A")

	empty := crdt.EmptyNamespace()
	tableA := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{"Point A"}),
		}),
	})

	namespaceA := empty.JoinTable("Table A", tableA)

	tableName := crdt.TableName("Table A")
	index := crdt.MakeIndex(map[crdt.TableName]crdt.IPFSPath{
		tableName: addrA,
	})

	abort := api.TraversalUpdate{}
	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).AnyTimes()
	mockReader.EXPECT().ReadsTables().Return([]crdt.TableName{tableName})
	mockReader.EXPECT().ReadNamespace(matchNamespace(namespaceA)).Return(abort)

	remote := loadRemote(mockStore, addrIndex)

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockReader)

	if lterr != nil {
		t.Error(lterr)
	}
}

func TestPersistSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)

	// addrA := crdt.IPFSPath(crdt.IPFSPath("Addr A"))
	addrB := crdt.IPFSPath(crdt.IPFSPath("Addr B"))
	addrIndexA := crdt.IPFSPath(crdt.IPFSPath("Addr Index A"))
	addrIndexB := crdt.IPFSPath(crdt.IPFSPath("Addr Index B"))
	tableB := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row B": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{"Point B"}),
		}),
	})

	tableBName := crdt.TableName("Table B")
	namespaceB := crdt.EmptyNamespace().JoinTable(tableBName, tableB)

	indexA := crdt.EmptyIndex()

	indexB := indexA.JoinNamespace(addrB, namespaceB)

	mock.EXPECT().AddNamespace(matchNamespace(namespaceB)).Return(addrB, nil)
	mock.EXPECT().CatIndex(addrIndexA).Return(indexA, nil).AnyTimes()
	mock.EXPECT().AddIndex(matchIndex(indexB)).Return(addrIndexB, nil)

	remote := makeRemote(mock)

	if remote == nil {
		t.Error("remote was nil")
	}

	jerr := remote.JoinTable(tableBName, tableB)

	if jerr != nil {
		t.Error(jerr)
	}

	perr := remote.Persist()

	if perr != nil {
		t.Error(perr)
	}

	cerr := remote.Commit()

	if cerr != nil {
		t.Error(cerr)
	}
}

func TestPersistFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)

	const namespaceAddr = crdt.IPFSPath("NS Addr 1")
	table := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row Key": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry Key": crdt.MakeEntry([]crdt.Point{"Entry Point"}),
		}),
	})

	namespace := crdt.EmptyNamespace().JoinTable("Table Key", table)
	index := crdt.EmptyIndex().JoinNamespace(namespaceAddr, namespace)

	mock.EXPECT().AddNamespace(matchNamespace(namespace)).Return(namespaceAddr, nil)
	mock.EXPECT().AddIndex(matchIndex(index)).Return(crdt.NIL_PATH, errors.New("Expected error"))

	remote := makeRemote(mock)

	testutil.AssertNonNil(t, remote)

	jerr := remote.JoinTable("Table Key", table)

	if jerr != nil {
		t.Error(jerr)
	}

	perr := remote.Persist()

	testutil.AssertNonNil(t, perr)

	rerr := remote.Rollback()

	testutil.AssertNil(t, rerr)
}

type nsmatcher struct {
	ns crdt.Namespace
}

func matchNamespace(ns crdt.Namespace) gomock.Matcher {
	return nsmatcher{ns}
}

func (nsm nsmatcher) Matches(v interface{}) bool {
	other, ok := v.(crdt.Namespace)

	if !ok {
		return false
	}

	return nsm.ns.Equals(other)
}

func (nsm nsmatcher) String() string {
	return fmt.Sprintf("matches Namespace: %v", nsm.ns)
}

type indexmatcher struct {
	index crdt.Index
}

func matchIndex(index crdt.Index) indexmatcher {
	return indexmatcher{index: index}
}

func (imatcher indexmatcher) Matches(v interface{}) bool {
	other, ok := v.(crdt.Index)

	if !ok {
		log.Debug("Not an Index for: %v vs %v", imatcher.index, v)
		return false
	}

	same := imatcher.index.Equals(other)

	log.Debug("Match Index? %v: %v to %v", same, imatcher.index, other)

	return same
}

func (imatcher indexmatcher) String() string {
	return fmt.Sprintf("matches Index: %v", imatcher.index)
}

func readApiResponse(kvq api.KvQuery) api.APIResponse {
	const duration = time.Second * 1
	timeout := time.NewTimer(duration)
	select {
	case resp := <-kvq.Response:
		return resp
	case <-timeout.C:
		panic("Timed out")
	}
}

func makeRemote(mock *MockRemoteStore) api.RemoteNamespaceTree {
	cache := cache.MakeResidentHeadCache()
	return service.MakeRemoteNamespace(mock, cache)
}

func loadRemote(mock *MockRemoteStore, addr crdt.IPFSPath) api.RemoteNamespaceTree {
	cache := cache.MakeResidentHeadCache()
	err := cache.BeginWriteTransaction()
	panicOnBadInit(err)
	err = cache.SetHead(addr)
	panicOnBadInit(err)
	err = cache.Commit()
	panicOnBadInit(err)
	return service.MakeRemoteNamespace(mock, cache)
}

func panicOnBadInit(err error) {
	if err != nil {
		panic(err)
	}
}
