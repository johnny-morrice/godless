package godless

type RemoteStore interface {
	Connect() error
	Add(RemoteNamespaceRecord) (RemoteStoreIndex, error)
	Cat(RemoteStoreIndex) (RemoteNamespaceRecord, error)
	Disconnect() error
}

type RemoteStoreIndex interface {}

type RemoteNamespaceRecord struct {
	Namespace *Namespace
	Children []RemoteStoreIndex
}

var EMPTY_RECORD RemoteNamespaceRecord
