package mock_godless

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

func TestRemoteNamespaceReplicate(t *testing.T) {
	// TODO integration test as design stands.
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
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point A")}),
		}),
	})
	tableB := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row B": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point B")}),
		}),
	})
	tableC := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row C": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry C": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point C")}),
		}),
	})

	const tableAName = "Table A"
	const tableBName = "Table B"
	const tableCName = "Table C"

	namespaceA := empty.JoinTable(tableAName, tableA)
	namespaceB := empty.JoinTable(tableBName, tableB)
	namespaceC := empty.JoinTable(tableCName, tableC)

	index := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		tableAName: crdt.UnsignedLink(addrA),
		tableBName: crdt.UnsignedLink(addrB),
		tableCName: crdt.UnsignedLink(addrC),
	})

	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).MinTimes(1)
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

func TestRunKvQuery(t *testing.T) {
	// TODO integration test as design stands.
}

func TestLoadTraverseSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockSearcher := NewMockNamespaceSearcher(ctrl)
	addrA := crdt.IPFSPath("Addr A")
	addrB := crdt.IPFSPath("Addr B")
	addrC := crdt.IPFSPath("Addr C")
	addrIndex := crdt.IPFSPath("Addr Index")

	signedAddrA := crdt.UnsignedLink(addrA)
	signedAddrB := crdt.UnsignedLink(addrB)
	signedAddrC := crdt.UnsignedLink(addrC)

	empty := crdt.EmptyNamespace()
	tableA := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point A")}),
		}),
	})
	tableB := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row B": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point B")}),
		}),
	})
	tableC := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row C": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry C": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point C")}),
		}),
	})

	const tableAName = "Table A"
	const tableBName = "Table B"
	const tableCName = "Table C"

	namespaceA := empty.JoinTable(tableAName, tableA)
	namespaceB := empty.JoinTable(tableBName, tableB)
	namespaceC := empty.JoinTable(tableCName, tableC)

	index := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		tableAName: crdt.UnsignedLink(addrA),
		tableBName: crdt.UnsignedLink(addrB),
		tableCName: crdt.UnsignedLink(addrC),
	})

	keepReading := api.TraversalUpdate{More: true}

	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).MinTimes(1)

	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatNamespace(addrB).Return(namespaceB, nil)
	mockStore.EXPECT().CatNamespace(addrC).Return(namespaceC, nil)

	mockSearcher.EXPECT().Search(index).Return([]crdt.Link{signedAddrA, signedAddrB, signedAddrC})
	mockSearcher.EXPECT().ReadNamespace(matchNamespace(namespaceA)).Return(keepReading)
	mockSearcher.EXPECT().ReadNamespace(matchNamespace(namespaceB)).Return(keepReading)
	mockSearcher.EXPECT().ReadNamespace(matchNamespace(namespaceC)).Return(keepReading)

	remote := loadRemote(mockStore, addrIndex)

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockSearcher)

	if lterr != nil {
		t.Error(lterr)
	}
}

func TestLoadTraverseFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockSearcher := NewMockNamespaceSearcher(ctrl)
	const indexAddr = crdt.IPFSPath("Addr Index")
	const namespaceAddr = crdt.IPFSPath("Addr A")
	signedNamespaceAddr := crdt.UnsignedLink(namespaceAddr)

	empty := crdt.EmptyNamespace()
	tableA := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point A")}),
		}),
	})

	namespaceA := empty.JoinTable("Table A", tableA)

	tableName := crdt.TableName("Table A")
	index := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		tableName: signedNamespaceAddr,
	})

	badTraverse := api.TraversalUpdate{More: true, Error: errors.New("Expected error")}
	mockStore.EXPECT().CatNamespace(namespaceAddr).Return(namespaceA, nil)
	mockStore.EXPECT().CatIndex(indexAddr).Return(index, nil).MinTimes(1)
	mockSearcher.EXPECT().Search(index).Return([]crdt.Link{signedNamespaceAddr})
	mockSearcher.EXPECT().ReadNamespace(matchNamespace(namespaceA)).Return(badTraverse)

	remote := loadRemote(mockStore, indexAddr)

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockSearcher)

	if lterr == nil {
		t.Error("lterr was nil")
	}
}

func TestLoadTraverseAbort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockSearcher := NewMockNamespaceSearcher(ctrl)
	addrIndex := crdt.IPFSPath("Addr Index")
	addrA := crdt.IPFSPath("Addr A")
	signedAddrA := crdt.UnsignedLink(addrA)

	empty := crdt.EmptyNamespace()
	tableA := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row A": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry A": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point A")}),
		}),
	})

	namespaceA := empty.JoinTable("Table A", tableA)

	tableName := crdt.TableName("Table A")
	index := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		tableName: signedAddrA,
	})

	abort := api.TraversalUpdate{}
	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).MinTimes(1)
	mockSearcher.EXPECT().Search(index).Return([]crdt.Link{signedAddrA})
	mockSearcher.EXPECT().ReadNamespace(matchNamespace(namespaceA)).Return(abort)

	remote := loadRemote(mockStore, addrIndex)

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockSearcher)

	if lterr != nil {
		t.Error(lterr)
	}
}

func TestJoinTableSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)

	// addrA := crdt.IPFSPath(crdt.IPFSPath("Addr A"))
	addrB := crdt.IPFSPath(crdt.IPFSPath("Addr B"))
	addrIndexA := crdt.IPFSPath(crdt.IPFSPath("Addr Index A"))
	addrIndexB := crdt.IPFSPath(crdt.IPFSPath("Addr Index B"))
	tableB := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row B": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry B": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Point B")}),
		}),
	})

	signedAddrB := crdt.UnsignedLink(addrB)

	tableBName := crdt.TableName("Table B")
	namespaceB := crdt.EmptyNamespace().JoinTable(tableBName, tableB)

	indexA := crdt.EmptyIndex()

	indexB := indexA.JoinNamespace(signedAddrB, namespaceB)

	mock.EXPECT().AddNamespace(matchNamespace(namespaceB)).Return(addrB, nil)
	mock.EXPECT().CatIndex(addrIndexA).Return(indexA, nil).MinTimes(1)
	mock.EXPECT().AddIndex(matchIndex(indexB)).Return(addrIndexB, nil)

	remote := loadRemote(mock, addrIndexA)

	if remote == nil {
		t.Error("remote was nil")
	}

	jerr := remote.JoinTable(tableBName, tableB)

	if jerr != nil {
		t.Error(jerr)
	}
}

func TestJoinTableFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)

	const namespaceAddr = crdt.IPFSPath("NS Addr 1")
	table := crdt.MakeTable(map[crdt.RowName]crdt.Row{
		"Row Key": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
			"Entry Key": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Entry Point")}),
		}),
	})

	signedAddr := crdt.UnsignedLink(namespaceAddr)

	namespace := crdt.EmptyNamespace().JoinTable("Table Key", table)
	index := crdt.EmptyIndex().JoinNamespace(signedAddr, namespace)

	mock.EXPECT().AddNamespace(matchNamespace(namespace)).Return(namespaceAddr, nil)
	mock.EXPECT().AddIndex(matchIndex(index)).Return(crdt.NIL_PATH, errors.New("Expected error"))

	remote := makeRemote(mock)

	testutil.AssertNonNil(t, remote)

	jerr := remote.JoinTable("Table Key", table)

	testutil.AssertNonNil(t, jerr)
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

func makeRemote(store api.RemoteStore) api.RemoteNamespaceTree {
	headCache := cache.MakeResidentHeadCache()
	options := remoteOptions(store, headCache)
	return service.MakeRemoteNamespace(options)
}

func loadRemote(store api.RemoteStore, addr crdt.IPFSPath) api.RemoteNamespaceTree {
	headCache := cache.MakeResidentHeadCache()
	err := headCache.SetHead(addr)
	panicOnBadInit(err)

	options := remoteOptions(store, headCache)
	return service.MakeRemoteNamespace(options)
}

func remoteOptions(store api.RemoteStore, headCache api.HeadCache) service.RemoteNamespaceOptions {
	options := service.RemoteNamespaceOptions{
		Store:      store,
		HeadCache:  headCache,
		IndexCache: fakeIndexCache{},
		KeyStore:   &crypto.KeyStore{},
	}

	return options
}

type fakeIndexCache struct {
}

func (cache fakeIndexCache) GetIndex(crdt.IPFSPath) (crdt.Index, error) {
	return crdt.EmptyIndex(), cache.err()
}

func (cache fakeIndexCache) SetIndex(crdt.IPFSPath, crdt.Index) error {
	return cache.err()
}

func (cache fakeIndexCache) err() error {
	return errors.New("Not a real index cache")
}

func panicOnBadInit(err error) {
	if err != nil {
		panic(err)
	}
}

const __UNKNOWN_CACHE_SIZE = -1
