package mock_godless

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	lib "github.com/johnny-morrice/godless"
)

func TestLoadRemoteNamespaceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)
	index := lib.IPFSPath("The Index")

	namespace := lib.MakeNamespace()
	record := lib.RemoteNamespaceRecord{
		Namespace: namespace,
		Children: []lib.RemoteStoreIndex{},
	}

	mock.EXPECT().Cat(index).Return(record, nil)

	ns, err := lib.LoadRemoteNamespace(mock, index)

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
	index := lib.IPFSPath("The Index")

	mock.EXPECT().Cat(index).Return(lib.EMPTY_RECORD, errors.New("expected error"))

	ns, err := lib.LoadRemoteNamespace(mock, index)

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
	namespace := lib.MakeNamespace()
	record := lib.RemoteNamespaceRecord{
		Namespace: namespace,
		Children: []lib.RemoteStoreIndex{},
	}

	index := lib.IPFSPath("The Index")

	mock.EXPECT().Add(mtchrd(record)).Return(index, nil)

	ns, err := lib.PersistNewRemoteNamespace(mock, namespace)

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
	namespace := lib.MakeNamespace()
	record := lib.RemoteNamespaceRecord{
		Namespace: namespace,
		Children: []lib.RemoteStoreIndex{},
	}

	mock.EXPECT().Add(mtchrd(record)).Return(nil, errors.New("Expected error"))

	ns, err := lib.PersistNewRemoteNamespace(mock, namespace)

	if ns != nil {
		t.Error("ns was not nil")
	}

	if err == nil {
		t.Error("err was nil")
	}
}

func TestRunKvQuery(t *testing.T) {
	// Wait till query and visitors are tested
}

func TestLoadTraverseSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockRemoteStore(ctrl)
	mockReader := NewMockNamespaceTreeReader(ctrl)
	indexA := lib.IPFSPath("Index A")
	indexB := lib.IPFSPath("Index B")
	indexC := lib.IPFSPath("Index C")

	empty := lib.MakeNamespace()
	tableA := lib.Table{
		Rows: map[string]lib.Row{
			"Row A": lib.Row{
				Entries: map[string][]string{
					"Entry A": []string{"Value A"},
				},
			},
		},
	}
	tableB := lib.Table{
		Rows: map[string]lib.Row{
			"Row B": lib.Row{
				Entries: map[string][]string{
					"Entry B": []string{"Value B"},
				},
			},
		},
	}
	tableC := lib.Table{
		Rows: map[string]lib.Row{
			"Row C": lib.Row{
				Entries: map[string][]string{
					"Entry C": []string{"Value C"},
				},
			},
		},
	}

	namespaceA, errA := empty.JoinTable("Table A", tableA)
	namespaceB, errB := empty.JoinTable("Table B", tableB)
	namespaceC, errC := empty.JoinTable("Table C", tableC)

	if !(errA == nil && errB == nil && errC == nil) {
		t.Error("JoinTable failed", errA, errB, errC)
	}

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespaceA,
		Children: []lib.RemoteStoreIndex{indexB, indexC},
	}
	recordB := lib.RemoteNamespaceRecord{
		Namespace: namespaceB,
		Children: []lib.RemoteStoreIndex{},
	}
	recordC := lib.RemoteNamespaceRecord{
		Namespace: namespaceC,
		Children: []lib.RemoteStoreIndex{},
	}

	mockStore.EXPECT().Cat(indexA).Return(recordA, nil)
	mockStore.EXPECT().Cat(indexB).Return(recordB, nil)
	mockStore.EXPECT().Cat(indexC).Return(recordC, nil)
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceA)).Return(false, nil)
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceB)).Return(false, nil)
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceC)).Return(false, nil)

	ns, err := lib.LoadRemoteNamespace(mockStore, indexA)

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
	mockReader := NewMockNamespaceTreeReader(ctrl)
	indexA := lib.IPFSPath("Index A")
	indexB := lib.IPFSPath("Index B")
	indexC := lib.IPFSPath("Index C")

	empty := lib.MakeNamespace()
	tableA := lib.Table{
		Rows: map[string]lib.Row{
			"Row A": lib.Row{
				Entries: map[string][]string{
					"Entry A": []string{"Value A"},
				},
			},
		},
	}

	namespaceA, errA := empty.JoinTable("Table A", tableA)

	if !(errA == nil) {
		t.Error("JoinTable failed", errA)
	}

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespaceA,
		Children: []lib.RemoteStoreIndex{indexB, indexC},
	}

	mockStore.EXPECT().Cat(indexA).Return(recordA, nil)
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceA)).Return(false, errors.New("Expected error"))

	ns, err := lib.LoadRemoteNamespace(mockStore, indexA)

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
	mockReader := NewMockNamespaceTreeReader(ctrl)
	indexA := lib.IPFSPath("Index A")
	indexB := lib.IPFSPath("Index B")
	indexC := lib.IPFSPath("Index C")

	empty := lib.MakeNamespace()
	tableA := lib.Table{
		Rows: map[string]lib.Row{
			"Row A": lib.Row{
				Entries: map[string][]string{
					"Entry A": []string{"Value A"},
				},
			},
		},
	}

	namespaceA, errA := empty.JoinTable("Table A", tableA)

	if !(errA == nil) {
		t.Error("JoinTable failed", errA)
	}

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespaceA,
		Children: []lib.RemoteStoreIndex{indexB, indexC},
	}

	mockStore.EXPECT().Cat(indexA).Return(recordA, nil)
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceA)).Return(true, nil)

	ns, err := lib.LoadRemoteNamespace(mockStore, indexA)

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

	index := lib.RemoteStoreIndex(lib.IPFSPath("Index Thing"))
	anotherIndex := lib.RemoteStoreIndex(lib.IPFSPath("Another index"))
	table := lib.Table{
		Rows: map[string]lib.Row{
			"Row Key": lib.Row{
				Entries: map[string][]string{
					"Entry Key": []string{"Entry Value"},
				},
			},
		},
	}
	namespace := lib.MakeNamespace()
	nextNamespace, _ := namespace.JoinTable("Table Key", table)

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespace,
		Children: []lib.RemoteStoreIndex{},
	}

	recordB := lib.RemoteNamespaceRecord{
		Namespace: nextNamespace,
		Children: []lib.RemoteStoreIndex{index},
	}

	mock.EXPECT().Add(mtchrd(recordA)).Return(index, nil)
	mock.EXPECT().Add(mtchrd(recordB)).Return(anotherIndex, nil)

	ns1, perr1 := lib.PersistNewRemoteNamespace(mock, namespace)

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

	index := lib.RemoteStoreIndex(lib.IPFSPath("Index Thing"))
	table := lib.Table{
		Rows: map[string]lib.Row{
			"Row Key": lib.Row{
				Entries: map[string][]string{
					"Entry Key": []string{"Entry Value"},
				},
			},
		},
	}
	namespace := lib.MakeNamespace()
	nextNamespace, _ := namespace.JoinTable("Table Key", table)

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespace,
		Children: []lib.RemoteStoreIndex{},
	}

	recordB := lib.RemoteNamespaceRecord{
		Namespace: nextNamespace,
		Children: []lib.RemoteStoreIndex{index},
	}

	mock.EXPECT().Add(mtchrd(recordA)).Return(index, nil)
	mock.EXPECT().Add(mtchrd(recordB)).Return(nil, errors.New("Expected error"))

	ns1, perr1 := lib.PersistNewRemoteNamespace(mock, namespace)

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

func mtchns(ns *lib.Namespace) gomock.Matcher {
	return nsmatcher{ns}
}

func mtchrd(rd lib.RemoteNamespaceRecord) gomock.Matcher {
	return rdmatcher{rd}
}

type rdmatcher struct {
	rd lib.RemoteNamespaceRecord
}

func (rdm rdmatcher) String() string {
	return "is matching RemoteNamespaceRecord"
}

func (rdm rdmatcher) Matches(v interface{}) bool {
	other, ok := v.(lib.RemoteNamespaceRecord)

	if !ok {
		return false
	}

	if reflect.DeepEqual(*rdm.rd.Namespace, *other.Namespace){
		return true
	}

	return reflect.DeepEqual(rdm.rd.Children, other.Children)
}


type nsmatcher struct {
	ns *lib.Namespace
}

func (nsm nsmatcher) Matches(v interface{}) bool {
	other, ok := v.(*lib.Namespace)

	if !ok {
		return false
	}

	return reflect.DeepEqual(*nsm.ns, *other)
}

func (nsm nsmatcher) String() string {
	return "is matcing Namespace"
}
