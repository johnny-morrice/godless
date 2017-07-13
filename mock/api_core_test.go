package mock_godless

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/cache"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/internal/service"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

func TestRemoteNamespaceCoreReplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := NewMockRemoteStore(ctrl)

	namespace := crdt.MakeNamespace(map[crdt.TableName]crdt.Table{
		"cars": crdt.MakeTable(map[crdt.RowName]crdt.Row{
			"car10": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
				"driver": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Mr Blogs")}),
			}),
		}),
	})

	const namespaceAddr = crdt.IPFSPath("Namespace addr")
	const indexAddr = crdt.IPFSPath("Index addr")

	index := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		"cars": crdt.UnsignedLink(namespaceAddr),
	})

	mockStore.EXPECT().AddIndex(matchIndex(index)).Return(indexAddr, nil).MinTimes(1)
	mockStore.EXPECT().CatIndex(indexAddr).Return(index, nil).MinTimes(1)
	mockStore.EXPECT().CatNamespace(namespaceAddr).Return(namespace, nil)

	core := makeRemote(mockStore)
	defer core.Close()

	resp := makeReplicateRequest(core, indexAddr)
	testutil.AssertNil(t, resp.Err)
	err := core.WriteMemoryImage()
	testutil.AssertNil(t, err)

	testReflectHead(t, core, indexAddr)
	testReflectIndex(t, core, index)
	testReflectNamespace(t, core, namespace)
}

func makeReplicateRequest(core api.Core, path crdt.IPFSPath) api.Response {
	links := []crdt.Link{crdt.UnsignedLink(path)}
	request := api.Request{Type: api.API_REPLICATE, Replicate: links}
	command, err := request.MakeCommand()
	panicOnBadInit(err)
	go command.Run(core)
	return readApiResponse(command)
}

func TestRemoteNamespaceCoreRunQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := NewMockRemoteStore(ctrl)

	joinQuery, err := query.Compile("join cars rows (@key=car10, driver=\"Mr Blogs\")")
	testutil.AssertNil(t, err)
	selectQuery, err := query.Compile("select cars")
	testutil.AssertNil(t, err)

	// TODO would be nice to get a DSL for writing namespaces.  JSON?
	namespace := crdt.MakeNamespace(map[crdt.TableName]crdt.Table{
		"cars": crdt.MakeTable(map[crdt.RowName]crdt.Row{
			"car10": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
				"driver": crdt.MakeEntry([]crdt.Point{crdt.UnsignedPoint("Mr Blogs")}),
			}),
		}),
	})

	const namespaceAddr = crdt.IPFSPath("Namespace addr")
	const indexAddr = crdt.IPFSPath("Index addr")

	index := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		"cars": crdt.UnsignedLink(namespaceAddr),
	})

	mockStore.EXPECT().AddNamespace(matchNamespace(namespace)).Return(namespaceAddr, nil)
	mockStore.EXPECT().AddIndex(matchIndex(index)).Return(indexAddr, nil).MinTimes(1)
	mockStore.EXPECT().CatIndex(indexAddr).Return(index, nil).MinTimes(1)
	mockStore.EXPECT().CatNamespace(namespaceAddr).Return(namespace, nil)

	remote := loadRemote(mockStore, indexAddr)
	defer remote.Close()

	log.Debug("running join")
	joinResponse := makeQueryRequest(remote, joinQuery)
	testutil.AssertNil(t, joinResponse.Err)
	err = remote.WriteMemoryImage()
	testutil.AssertNil(t, err)
	log.Debug("running select")
	selectResponse := makeQueryRequest(remote, selectQuery)
	log.Debug("completed queries")
	testutil.AssertNil(t, joinResponse.Err)
	testutil.Assert(t, "Unexpected namespace", namespace.Equals(selectResponse.Namespace))
}

func makeQueryRequest(core api.Core, query *query.Query) api.Response {
	request := api.Request{Type: api.API_QUERY, Query: query}
	command, err := request.MakeCommand()
	panicOnBadInit(err)
	go command.Run(core)
	return readApiResponse(command)
}

func TestRemoteNamespaceCoreReflect(t *testing.T) {
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

	mockStore.EXPECT().AddIndex(matchIndex(index)).AnyTimes()
	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).MinTimes(1)
	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatNamespace(addrB).Return(namespaceB, nil)
	mockStore.EXPECT().CatNamespace(addrC).Return(namespaceC, nil)

	joinedNamespace := namespaceA.JoinNamespace(namespaceB)
	joinedNamespace = joinedNamespace.JoinNamespace(namespaceC)

	remote := loadRemote(mockStore, addrIndex)
	defer remote.Close()

	testReflectHead(t, remote, addrIndex)
	testReflectIndex(t, remote, index)
	testReflectNamespace(t, remote, joinedNamespace)
}

// FIXME test error path
func testReflectHead(t *testing.T, remote api.Core, expected crdt.IPFSPath) {
	resp := reflectOnRemote(remote, api.REFLECT_HEAD_PATH)

	testutil.AssertNil(t, resp.Err)
	testutil.AssertEquals(t, "Unexpected HEAD path", expected, resp.Path)
}

func testReflectIndex(t *testing.T, remote api.Core, expected crdt.Index) {
	resp := reflectOnRemote(remote, api.REFLECT_INDEX)
	actual := resp.Index

	testutil.AssertNil(t, resp.Err)
	testutil.Assert(t, "Unexpected index", expected.Equals(actual))
}

func testReflectNamespace(t *testing.T, remote api.Core, expected crdt.Namespace) {
	resp := reflectOnRemote(remote, api.REFLECT_DUMP_NAMESPACE)
	actual := resp.Namespace

	testutil.AssertNil(t, resp.Err)
	testutil.Assert(t, "Unexpected namespace", expected.Equals(actual))
}

func reflectOnRemote(remote api.Core, reflection api.ReflectionType) api.Response {
	request := api.Request{Type: api.API_REFLECT, Reflection: reflection}
	command, err := request.MakeCommand()
	panicOnBadInit(err)
	go command.Run(remote)

	resp := readApiResponse(command)

	return resp
}

func TestRemoteNamespaceCoreLoadTraverseSuccess(t *testing.T) {
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

	mockStore.EXPECT().AddIndex(index).Return(addrIndex, nil).AnyTimes()
	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).AnyTimes()

	mockStore.EXPECT().CatNamespace(addrA).Return(namespaceA, nil)
	mockStore.EXPECT().CatNamespace(addrB).Return(namespaceB, nil)
	mockStore.EXPECT().CatNamespace(addrC).Return(namespaceC, nil)

	mockSearcher.EXPECT().Search(index).Return([]crdt.Link{signedAddrA, signedAddrB, signedAddrC})
	mockSearcher.EXPECT().ReadSearchResult(matchNamespaceResult(namespaceA)).Return(keepReading)
	mockSearcher.EXPECT().ReadSearchResult(matchNamespaceResult(namespaceB)).Return(keepReading)
	mockSearcher.EXPECT().ReadSearchResult(matchNamespaceResult(namespaceC)).Return(keepReading)

	remote := loadRemote(mockStore, addrIndex)
	defer remote.Close()

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockSearcher)

	if lterr != nil {
		t.Error(lterr)
	}
}

func TestRemoteNamespaceCoreLoadTraverseFailure(t *testing.T) {
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

	mockStore.EXPECT().AddIndex(index).AnyTimes().Return(indexAddr, nil)
	mockStore.EXPECT().CatIndex(indexAddr).Return(index, nil).MinTimes(1)

	mockSearcher.EXPECT().Search(index).Return([]crdt.Link{signedNamespaceAddr})
	mockSearcher.EXPECT().ReadSearchResult(matchNamespaceResult(namespaceA)).Return(badTraverse)

	remote := loadRemote(mockStore, indexAddr)
	defer remote.Close()

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockSearcher)

	if lterr == nil {
		t.Error("lterr was nil")
	}
}

func TestRemoteNamespaceCoreLoadTraverseAbort(t *testing.T) {
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

	mockStore.EXPECT().AddIndex(index).Return(addrIndex, nil).AnyTimes()
	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).MinTimes(1)

	mockSearcher.EXPECT().Search(index).Return([]crdt.Link{signedAddrA})
	mockSearcher.EXPECT().ReadSearchResult(matchNamespaceResult(namespaceA)).Return(abort)

	remote := loadRemote(mockStore, addrIndex)
	defer remote.Close()

	if remote == nil {
		t.Error("remote was nil")
	}

	lterr := remote.LoadTraverse(mockSearcher)

	if lterr != nil {
		t.Error(lterr)
	}
}

func TestRemoteNamespaceCoreJoinTableSuccess(t *testing.T) {
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

	mock.EXPECT().CatIndex(addrIndexA).Return(indexA, nil).AnyTimes()
	mock.EXPECT().AddIndex(indexA).Return(addrIndexA, nil).MinTimes(1)
	mock.EXPECT().AddIndex(matchIndex(indexB)).Return(addrIndexB, nil).AnyTimes()

	// No index Catting with MemoryImage
	// mock.EXPECT().CatIndex(addrIndexA).Return(indexA, nil).MinTimes(1)

	remote := loadRemote(mock, addrIndexA)
	defer remote.Close()

	if remote == nil {
		t.Error("remote was nil")
	}

	path, err := remote.JoinTable(tableBName, tableB)
	testutil.AssertNil(t, err)
	testutil.AssertEquals(t, "Unexpected index address", addrIndexB, path)

	err = remote.WriteMemoryImage()
	testutil.AssertNil(t, err)
}

func TestRemoteNamespaceCoreJoinTableFailure(t *testing.T) {
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
	mock.EXPECT().AddIndex(matchIndex(index)).Return(crdt.NIL_PATH, errors.New("Expected error")).MinTimes(1)

	remote := makeRemote(mock)
	defer remote.Close()

	testutil.AssertNonNil(t, remote)

	path, err := remote.JoinTable("Table Key", table)
	testutil.AssertNonNil(t, err)
	testutil.AssertEquals(t, "Expected empty index address", crdt.NIL_PATH, path)
	err = remote.WriteMemoryImage()
	testutil.AssertNonNil(t, err)
}

type resultNamespaceMatcher struct {
	ns crdt.Namespace
}

func matchNamespaceResult(ns crdt.Namespace) gomock.Matcher {
	return resultNamespaceMatcher{ns: ns}
}

func (matcher resultNamespaceMatcher) Matches(v interface{}) bool {
	other, ok := v.(api.SearchResult)

	if !ok {
		return false
	}

	return matcher.ns.Equals(other.Namespace)
}

func (matcher resultNamespaceMatcher) String() string {
	return fmt.Sprintf("matches SearchResult with Namespace: %v", matcher.ns)
}

type nsmatcher struct {
	ns crdt.Namespace
}

func matchNamespace(ns crdt.Namespace) gomock.Matcher {
	return nsmatcher{ns}
}

func (matcher nsmatcher) Matches(v interface{}) bool {
	other, ok := v.(crdt.Namespace)

	if !ok {
		return false
	}

	return matcher.ns.Equals(other)
}

func (matcher nsmatcher) String() string {
	return fmt.Sprintf("matches Namespace: %v", matcher.ns)
}

type indexmatcher struct {
	index crdt.Index
}

func matchIndex(index crdt.Index) indexmatcher {
	return indexmatcher{index: index}
}

func (matcher indexmatcher) Matches(v interface{}) bool {
	other, ok := v.(crdt.Index)

	if !ok {
		log.Debug("Not an Index for: %v vs %v", matcher.index, v)
		return false
	}

	same := matcher.index.Equals(other)

	return same
}

func (matcher indexmatcher) String() string {
	return fmt.Sprintf("matches Index: %v", matcher.index)
}

func readApiResponse(command api.Command) api.Response {
	const duration = time.Second * 1
	timeout := time.NewTimer(duration)
	select {
	case resp := <-command.Response:
		return resp
	case <-timeout.C:
		panic("Timed out")
	}
}

func makeRemote(store api.RemoteStore) api.RemoteNamespaceCore {
	headCache := cache.MakeResidentHeadCache()
	options := remoteOptions(store, headCache)
	return service.MakeRemoteNamespaceCore(options)
}

func loadRemote(store api.RemoteStore, addr crdt.IPFSPath) api.RemoteNamespaceCore {
	headCache := cache.MakeResidentHeadCache()
	err := headCache.SetHead(addr)
	panicOnBadInit(err)

	options := remoteOptions(store, headCache)
	return service.MakeRemoteNamespaceCore(options)
}

func remoteOptions(store api.RemoteStore, headCache api.HeadCache) service.RemoteNamespaceCoreOptions {
	options := service.RemoteNamespaceCoreOptions{
		Store:          store,
		HeadCache:      headCache,
		IndexCache:     fakeIndexCache{},
		NamespaceCache: fakeNamespaceCache{},
		KeyStore:       &crypto.KeyStore{},
		MemoryImage:    cache.MakeResidentMemoryImage(),
		Debug:          true,
	}

	return options
}

type fakeNamespaceCache struct {
}

func (cache fakeNamespaceCache) GetNamespace(crdt.IPFSPath) (crdt.Namespace, error) {
	return crdt.EmptyNamespace(), cache.err()
}

func (cache fakeNamespaceCache) SetNamespace(crdt.IPFSPath, crdt.Namespace) error {
	return cache.err()
}

func (cache fakeNamespaceCache) err() error {
	return errors.New("Not a real namespace cache")
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
