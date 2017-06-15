package mock_godless

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/internal/testutil"
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

	remote := service.MakeRemoteNamespace(mockStore)
	assertUnchanged(t, remote)

	err := remote.JoinTable(tableName, table)
	testutil.AssertNil(t, err)
	assertChanged(t, remote)

	remote.Reset()
	assertUnchanged(t, remote)
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

	remote, err := service.LoadRemoteNamespace(mockStore, addrIndex)

	testutil.AssertNil(t, err)

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
	request := api.MakeKvReflect(reflection)
	go request.Run(remote)

	resp := readApiResponse(request)

	return resp
}

func TestIsChanged(t *testing.T) {
	TestRemoteNamespaceReset(t)
}

func TestLoadRemoteNamespaceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)
	addr := crdt.IPFSPath("The Index")

	mock.EXPECT().CatIndex(addr).Return(crdt.EmptyIndex(), nil)

	ns, err := service.LoadRemoteNamespace(mock, addr)

	if ns == nil {
		t.Error("ns was nil")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestLoadRemoteNamespaceFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)
	addr := crdt.IPFSPath("The Index")

	mock.EXPECT().CatIndex(addr).Return(crdt.EmptyIndex(), errors.New("expected error"))

	ns, err := service.LoadRemoteNamespace(mock, addr)

	if ns != nil {
		t.Error("ns was not nil")
	}

	if err == nil {
		t.Error("err was nil")
	}
}

func TestPersistNewRemoteNamespaceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)

	addr := crdt.IPFSPath("The Index")
	index := crdt.MakeIndex(map[crdt.TableName]crdt.IPFSPath{
		MAIN_TABLE_KEY: addr,
	})
	namespace := crdt.MakeNamespace(map[crdt.TableName]crdt.Table{
		MAIN_TABLE_KEY: crdt.MakeTable(map[crdt.RowName]crdt.Row{
			"A row": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
				"A thing": crdt.MakeEntry([]crdt.Point{"hi"}),
			}),
		}),
	})

	mock.EXPECT().AddNamespace(matchns(namespace)).Return(addr, nil)
	mock.EXPECT().AddIndex(index).Return(addr, nil)

	ns, err := service.PersistNewRemoteNamespace(mock, namespace)

	if ns == nil {
		t.Error("ns was nil")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestPersistNewRemoteNamespaceFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)
	namespace := crdt.EmptyNamespace()

	mock.EXPECT().AddIndex(gomock.Any()).Return(crdt.NIL_PATH, errors.New("Expected error"))

	ns, err := service.PersistNewRemoteNamespace(mock, namespace)

	if ns != nil {
		t.Error("ns was not nil")
	}

	if err == nil {
		t.Error("err was nil")
	}
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

	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).Times(2)

	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatNamespace(addrB).Return(namespaceB, nil)
	mockStore.EXPECT().CatNamespace(addrC).Return(namespaceC, nil)

	mockReader.EXPECT().ReadsTables().Return([]crdt.TableName{tableAName, tableBName, tableCName})
	mockReader.EXPECT().ReadNamespace(matchns(namespaceA)).Return(keepReading)
	mockReader.EXPECT().ReadNamespace(matchns(namespaceB)).Return(keepReading)
	mockReader.EXPECT().ReadNamespace(matchns(namespaceC)).Return(keepReading)

	ns, err := service.LoadRemoteNamespace(mockStore, addrIndex)

	if ns == nil {
		t.Error("ns was nil")
	}

	if err != nil {
		t.Error(err)
	}

	lterr := ns.LoadTraverse(mockReader)

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
	mockStore.EXPECT().CatIndex(indexAddr).Return(index, nil).Times(2)
	mockReader.EXPECT().ReadsTables().Return([]crdt.TableName{tableName})
	mockReader.EXPECT().ReadNamespace(matchns(namespaceA)).Return(badTraverse)

	ns, err := service.LoadRemoteNamespace(mockStore, indexAddr)

	if ns == nil {
		t.Error("ns was nil")
	}

	if err != nil {
		t.Error(err)
	}

	lterr := ns.LoadTraverse(mockReader)

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
	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).Times(2)
	mockReader.EXPECT().ReadsTables().Return([]crdt.TableName{tableName})
	mockReader.EXPECT().ReadNamespace(matchns(namespaceA)).Return(abort)

	ns, err := service.LoadRemoteNamespace(mockStore, addrIndex)

	if ns == nil {
		t.Error("ns was nil")
	}

	if err != nil {
		t.Error(err)
	}

	lterr := ns.LoadTraverse(mockReader)

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

	// TODO not fully decided whether this test should do the initial persist too.

	// tableAName := crdt.TableName("Table A")
	// namespaceA := crdt.MakeNamespace(map[crdt.TableName]crdt.Table{
	// 	tableAName: crdt.MakeEntry(map[crdt.RowName]crdt.Row{
	// 		"Row Q": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
	// 			"Entry Q": crdt.MakeEntry([]crdt.Point{"hi"}),
	// 		}),
	// 	}),
	// })
	namespaceA := crdt.EmptyNamespace()
	tableBName := crdt.TableName("Table B")
	namespaceB := namespaceA.JoinTable(tableBName, tableB)

	// recordA := lib.RemoteNamespaceRecord{
	// 	Namespace: namespaceA,
	// }

	// indexA := crdt.MakeIndex(map[crdt.TableName]crdt.IPFSPath{
	// 	tableAName: addrA,
	// })
	indexA := crdt.EmptyIndex()

	indexB := indexA.JoinNamespace(addrB, namespaceB)

	// mock.EXPECT().AddNamespace(mtchrd(recordA)).Return(addrA, nil)
	mock.EXPECT().AddIndex(gomock.Any()).Return(addrIndexA, nil)
	mock.EXPECT().AddNamespace(matchns(namespaceB)).Return(addrB, nil)
	mock.EXPECT().CatIndex(addrIndexA).Return(indexA, nil)
	mock.EXPECT().AddIndex(indexB).Return(addrIndexB, nil)

	ns1, perr1 := service.PersistNewRemoteNamespace(mock, namespaceA)

	if perr1 != nil {
		t.Error(perr1)
	}

	if ns1 == nil {
		t.Error("ns1 was nil")
	}

	jerr := ns1.JoinTable(tableBName, tableB)

	if jerr != nil {
		t.Error(jerr)
	}

	ns2, perr2 := ns1.Persist()

	if perr2 != nil {
		t.Error(perr2)
	}

	if ns2 == nil {
		t.Error("ns2 was nil")
	}
}

func TestPersistFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)

	table := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row Key": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry Key": crdt.MakeEntry([]crdt.Point{"Entry Point"}),
		}),
	})
	namespace := crdt.EmptyNamespace()
	nextNamespace := namespace.JoinTable("Table Key", table)

	mock.EXPECT().AddIndex(crdt.EmptyIndex())
	mock.EXPECT().AddNamespace(matchns(nextNamespace)).Return(crdt.NIL_PATH, errors.New("Expected error"))

	ns1, perr1 := service.PersistNewRemoteNamespace(mock, namespace)

	if perr1 != nil {
		t.Error(perr1)
	}

	if ns1 == nil {
		t.Error("ns1 was nil")
	}

	jerr := ns1.JoinTable("Table Key", table)

	if jerr != nil {
		t.Error(jerr)
	}

	ns2, perr2 := ns1.Persist()

	if perr2 == nil {
		t.Error("perr2 was nil")
	}

	if ns2 != nil {
		t.Error("ns2 was not nil")
	}
}

func matchns(ns crdt.Namespace) gomock.Matcher {
	return nsmatcher{ns}
}

type nsmatcher struct {
	ns crdt.Namespace
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
