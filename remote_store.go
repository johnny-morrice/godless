package godless

//go:generate mockgen -destination mock/mock_remote_store.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless RemoteStore

type RemoteStore interface {
	Connect() error
	Add(RemoteNamespaceRecord) (RemoteStoreIndex, error)
	Cat(RemoteStoreIndex) (RemoteNamespaceRecord, error)
	Disconnect() error
}

type RemoteStoreIndex interface {
	Path() string
}

type RemoteNamespaceRecord struct {
	// TODO elsewhere we use pointer for Namespace but this is easier to test.
	Namespace *Namespace
	Children  []RemoteStoreIndex
}

var EMPTY_RECORD RemoteNamespaceRecord
