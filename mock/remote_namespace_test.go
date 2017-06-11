package mock_godless

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	lib "github.com/johnny-morrice/godless"
	"github.com/pkg/errors"
)

func TestLoadRemoteNamespaceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockRemoteStore(ctrl)
	addr := lib.IPFSPath("The Index")

	mock.EXPECT().CatIndex(addr).Return(lib.EmptyIndex(), nil)

	ns, err := lib.LoadRemoteNamespace(mock, addr)

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
	addr := lib.IPFSPath("The Index")

	mock.EXPECT().CatIndex(addr).Return(lib.EmptyIndex(), errors.New("expected error"))

	ns, err := lib.LoadRemoteNamespace(mock, addr)

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

	addr := lib.IPFSPath("The Index")
	index := lib.MakeIndex(map[lib.TableName]lib.RemoteStoreAddress{
		MAIN_TABLE_KEY: addr,
	})
	namespace := lib.MakeNamespace(map[lib.TableName]lib.Table{
		MAIN_TABLE_KEY: lib.MakeTable(map[lib.RowName]lib.Row{
			"A row": lib.MakeRow(map[lib.EntryName]lib.Entry{
				"A thing": lib.MakeEntry([]lib.Point{"hi"}),
			}),
		}),
	})
	record := lib.RemoteNamespaceRecord{
		Namespace: namespace,
	}

	mock.EXPECT().AddNamespace(mtchrd(record)).Return(addr, nil)
	mock.EXPECT().AddIndex(index).Return(addr, nil)

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
	namespace := lib.EmptyNamespace()

	mock.EXPECT().AddIndex(gomock.Any()).Return(nil, errors.New("Expected error"))

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
	mockReader := NewMockNamespaceTreeTableReader(ctrl)
	addrA := lib.IPFSPath("Addr A")
	addrB := lib.IPFSPath("Addr B")
	addrC := lib.IPFSPath("Addr C")
	addrIndex := lib.IPFSPath("Addr Index")

	empty := lib.EmptyNamespace()
	tableA := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row A": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]lib.Point{"Point A"}),
		}),
	})
	tableB := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row B": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry B": lib.MakeEntry([]lib.Point{"Point B"}),
		}),
	})
	tableC := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row C": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry C": lib.MakeEntry([]lib.Point{"Point C"}),
		}),
	})

	const tableAName = "Table A"
	const tableBName = "Table B"
	const tableCName = "Table C"

	namespaceA := empty.JoinTable(tableAName, tableA)
	namespaceB := empty.JoinTable(tableBName, tableB)
	namespaceC := empty.JoinTable(tableCName, tableC)

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespaceA,
	}
	recordB := lib.RemoteNamespaceRecord{
		Namespace: namespaceB,
	}
	recordC := lib.RemoteNamespaceRecord{
		Namespace: namespaceC,
	}

	index := lib.MakeIndex(map[lib.TableName]lib.RemoteStoreAddress{
		tableAName: addrA,
		tableBName: addrB,
		tableCName: addrC,
	})

	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).Times(2)

	mockStore.EXPECT().CatNamespace(addrA).Return(recordA, nil)
	mockStore.EXPECT().CatNamespace(addrB).Return(recordB, nil)
	mockStore.EXPECT().CatNamespace(addrC).Return(recordC, nil)

	mockReader.EXPECT().ReadsTables().Return([]lib.TableName{tableAName, tableBName, tableCName})
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceA)).Return(false, nil)
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceB)).Return(false, nil)
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceC)).Return(false, nil)

	ns, err := lib.LoadRemoteNamespace(mockStore, addrIndex)

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
	indexAddr := lib.IPFSPath("Addr Index")
	namespaceAddr := lib.IPFSPath("Addr A")

	empty := lib.EmptyNamespace()
	tableA := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row A": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]lib.Point{"Point A"}),
		}),
	})

	namespaceA := empty.JoinTable("Table A", tableA)

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespaceA,
	}
	tableName := lib.TableName("Table A")
	index := lib.MakeIndex(map[lib.TableName]lib.RemoteStoreAddress{
		tableName: namespaceAddr,
	})

	mockStore.EXPECT().CatNamespace(namespaceAddr).Return(recordA, nil)
	mockStore.EXPECT().CatIndex(indexAddr).Return(index, nil).Times(2)
	mockReader.EXPECT().ReadsTables().Return([]lib.TableName{tableName})
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceA)).Return(false, errors.New("Expected error"))

	ns, err := lib.LoadRemoteNamespace(mockStore, indexAddr)

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
	addrIndex := lib.IPFSPath("Addr Index")
	addrA := lib.IPFSPath("Addr A")

	empty := lib.EmptyNamespace()
	tableA := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row A": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry A": lib.MakeEntry([]lib.Point{"Point A"}),
		}),
	})

	namespaceA := empty.JoinTable("Table A", tableA)

	recordA := lib.RemoteNamespaceRecord{
		Namespace: namespaceA,
	}

	tableName := lib.TableName("Table A")
	index := lib.MakeIndex(map[lib.TableName]lib.RemoteStoreAddress{
		tableName: addrA,
	})

	mockStore.EXPECT().CatNamespace(addrA).Return(recordA, nil)
	mockStore.EXPECT().CatIndex(addrIndex).Return(index, nil).Times(2)
	mockReader.EXPECT().ReadsTables().Return([]lib.TableName{tableName})
	mockReader.EXPECT().ReadNamespace(mtchns(namespaceA)).Return(true, nil)

	ns, err := lib.LoadRemoteNamespace(mockStore, addrIndex)

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

	// addrA := lib.RemoteStoreAddress(lib.IPFSPath("Addr A"))
	addrB := lib.RemoteStoreAddress(lib.IPFSPath("Addr B"))
	addrIndexA := lib.RemoteStoreAddress(lib.IPFSPath("Addr Index A"))
	addrIndexB := lib.RemoteStoreAddress(lib.IPFSPath("Addr Index B"))
	tableB := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row B": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry B": lib.MakeEntry([]lib.Point{"Point B"}),
		}),
	})

	// TODO not fully decided whether this test should do the initial persist too.

	// tableAName := lib.TableName("Table A")
	// namespaceA := lib.MakeNamespace(map[lib.TableName]lib.Table{
	// 	tableAName: lib.MakeTable(map[lib.RowName]lib.Row{
	// 		"Row Q": lib.MakeRow(map[lib.EntryName]lib.Entry{
	// 			"Entry Q": lib.MakeEntry([]lib.Point{"hi"}),
	// 		}),
	// 	}),
	// })
	namespaceA := lib.EmptyNamespace()
	tableBName := lib.TableName("Table B")
	namespaceB := namespaceA.JoinTable(tableBName, tableB)

	// recordA := lib.RemoteNamespaceRecord{
	// 	Namespace: namespaceA,
	// }

	recordB := lib.RemoteNamespaceRecord{
		Namespace: namespaceB,
	}

	// indexA := lib.MakeIndex(map[lib.TableName]lib.RemoteStoreAddress{
	// 	tableAName: addrA,
	// })
	indexA := lib.EmptyIndex()

	indexB := indexA.JoinNamespace(addrB, namespaceB)

	// mock.EXPECT().AddNamespace(mtchrd(recordA)).Return(addrA, nil)
	mock.EXPECT().AddIndex(gomock.Any()).Return(addrIndexA, nil)
	mock.EXPECT().AddNamespace(mtchrd(recordB)).Return(addrB, nil)
	mock.EXPECT().CatIndex(addrIndexA).Return(indexA, nil)
	mock.EXPECT().AddIndex(indexB).Return(addrIndexB, nil)

	ns1, perr1 := lib.PersistNewRemoteNamespace(mock, namespaceA)

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

	table := lib.MakeTable(map[lib.RowName]lib.Row{
		"Row Key": lib.MakeRow(map[lib.EntryName]lib.Entry{
			"Entry Key": lib.MakeEntry([]lib.Point{"Entry Point"}),
		}),
	})
	namespace := lib.EmptyNamespace()
	nextNamespace := namespace.JoinTable("Table Key", table)

	recordB := lib.RemoteNamespaceRecord{
		Namespace: nextNamespace,
	}

	mock.EXPECT().AddIndex(lib.EmptyIndex())
	mock.EXPECT().AddNamespace(mtchrd(recordB)).Return(nil, errors.New("Expected error"))

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

func mtchns(ns lib.Namespace) gomock.Matcher {
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

	return rdm.rd.Namespace.Equals(other.Namespace)
}

type nsmatcher struct {
	ns lib.Namespace
}

func (nsm nsmatcher) Matches(v interface{}) bool {
	other, ok := v.(lib.Namespace)

	if !ok {
		return false
	}

	return nsm.ns.Equals(other)
}

func (nsm nsmatcher) String() string {
	return fmt.Sprintf("matches Namespace: %v", nsm.ns)
}
