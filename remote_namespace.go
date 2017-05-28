package godless

import (
	"github.com/pkg/errors"
)

type remoteNamespace struct {
	Update Namespace
	Store  RemoteStore
	Addr   RemoteStoreAddress
}

func LoadRemoteNamespace(store RemoteStore, addr RemoteStoreAddress) (KvNamespaceTree, error) {
	rn := &remoteNamespace{}
	rn.Store = store
	rn.Addr = addr
	rn.Update = EmptyNamespace()

	_, err := rn.loadIndex()

	// We don't use the index for anything at this point.
	logdbg("Index found at '%v'", addr)

	if err != nil {
		return nil, errors.Wrap(err, "Error loading new namespace")
	}

	return rn, nil
}

func PersistNewRemoteNamespace(store RemoteStore, namespace Namespace) (KvNamespaceTree, error) {
	rn := &remoteNamespace{}
	rn.Store = store
	rn.Update = namespace

	kv, err := rn.Persist()

	if err != nil {
		return nil, err
	}

	return kv.(*remoteNamespace), nil
}

// TODO there are likely to be many reflection features.  Replace switches with polymorphism.
func (rn *remoteNamespace) RunKvReflection(reflect APIReflectRequest, kvq KvQuery) {
	var runner APIResponder
	switch reflect.Command {
	case REFLECT_HEAD_PATH:
		runner = APIResponderFunc(rn.getReflectHead)
	case REFLECT_INDEX:
		runner = APIResponderFunc(rn.getReflectIndex)
	case REFLECT_DUMP_NAMESPACE:
		runner = APIResponderFunc(rn.dumpReflectNamespaces)
	default:
		panic("Unknown reflection command")
	}

	response := runner.RunQuery()
	kvq.writeResponse(response)
}

// TODO Not sure if best place for these to live.
func (rn *remoteNamespace) getReflectHead() APIResponse {
	response := RESPONSE_REFLECT
	response.ReflectResponse.Path = rn.Addr.Path()
	return response
}

func (rn *remoteNamespace) getReflectIndex() APIResponse {
	response := RESPONSE_REFLECT

	index, err := rn.loadIndex()

	if err != nil {
		response.Msg = RESPONSE_FAIL_MSG
		response.Err = errors.Wrap(err, "getReflectIndex failed")
		return response
	}

	response.ReflectResponse.Index = index.APIIndex()

	return response
}

func (rn *remoteNamespace) dumpReflectNamespaces() APIResponse {
	response := RESPONSE_REFLECT

	index, err := rn.loadIndex()

	if err != nil {
		response.Msg = RESPONSE_FAIL_MSG
		response.Err = errors.Wrap(err, "dumpReflectNamespace failed")
		return response
	}

	tables := index.AllTables()
	everything := EmptyNamespace()

	lambda := NamespaceTreeLambda(func(ns Namespace) (bool, error) {
		everything = everything.JoinNamespace(ns)
		return false, nil
	})
	traversal := AddTableHints(tables, lambda)
	rn.LoadTraverse(traversal)
	return response
}

// RunKvQuery will block until the result can be written to kvq.
func (rn *remoteNamespace) RunKvQuery(query *Query, kvq KvQuery) {
	var runner APIResponder

	logQuery(query)

	switch query.OpCode {
	case JOIN:
		visitor := MakeNamespaceTreeJoin(rn)
		query.Visit(visitor)
		runner = visitor
	case SELECT:
		visitor := MakeNamespaceTreeSelect(rn)
		query.Visit(visitor)
		runner = visitor
	default:
		query.opcodePanic()
	}

	response := runner.RunQuery()
	kvq.writeResponse(response)
}

func (rn *remoteNamespace) IsChanged() bool {
	return !rn.Update.IsEmpty()
}

func (rn *remoteNamespace) JoinTable(tableKey TableName, table Table) error {
	joined := rn.Update.JoinTable(tableKey, table)
	rn.Update = joined
	return nil
}

func (rn *remoteNamespace) LoadTraverse(nttr NamespaceTreeTableReader) error {
	index, indexerr := rn.loadIndex()

	if indexerr != nil {
		return errors.Wrap(indexerr, "LoadTraverse failed")
	}

	tableAddrs := rn.findTableAddrs(index, nttr)

	return rn.traverseTableNamespaces(tableAddrs, nttr)
}

func (rn *remoteNamespace) traverseTableNamespaces(tableAddrs []RemoteStoreAddress, f NamespaceTreeReader) error {
	nsch, cancelch := rn.namespaceLoader(tableAddrs)
	defer close(cancelch)
	for ns := range nsch {
		stop, err := f.ReadNamespace(ns)

		if stop || err != nil {
			cancelch <- struct{}{}
		}

		if err != nil {
			return errors.Wrap(err, "traverseTableNamespaces failed")
		}
	}

	return nil
}

// Preload namespaces while the previous is analysed.
func (rn *remoteNamespace) namespaceLoader(addrs []RemoteStoreAddress) (<-chan Namespace, chan<- struct{}) {
	nsch := make(chan Namespace)
	cancelch := make(chan struct{}, 1)

	go func() {
		defer close(nsch)
		for _, a := range addrs {
			nsr, err := rn.Store.CatNamespace(a)

			if err != nil {
				logerr("namespaceLoader failed: %v", err)
				return
			}

			logdbg("Catted namespace from: %v", a)
		LOOP:
			for {
				select {
				case <-cancelch:
					return
				case nsch <- nsr.Namespace:
					break LOOP
				}
			}
		}
	}()

	return nsch, cancelch
}

func (rn *remoteNamespace) findTableAddrs(index RemoteNamespaceIndex, tableHints TableHinter) []RemoteStoreAddress {
	out := []RemoteStoreAddress{}
	for _, t := range tableHints.ReadsTables() {
		addrs, err := index.GetTableAddrs(t)

		if err == nil {
			out = append(out, addrs...)
		}
	}

	return out
}

// Load chunks over IPFS
// TODO opportunity to query IPFS in parallel?
func (rn *remoteNamespace) loadIndex() (RemoteNamespaceIndex, error) {
	if rn.Addr == nil {
		// panic("tried to load remoteNamespace with empty Addr")
		return EMPTY_INDEX, nil
	}

	index, err := rn.Store.CatIndex(rn.Addr)

	if err != nil {
		return RemoteNamespaceIndex{}, errors.Wrap(err, "Error in remoteNamespace CatNamespace")
	}

	return index, nil
}

// Write pending changes to IPFS and return the new parent namespace.
func (rn *remoteNamespace) Persist() (KvNamespace, error) {
	namespace := rn.Update
	namespaceAddr, nserr := rn.persistNamespace(namespace)

	if nserr != nil {
		return nil, nserr
	}

	indexAddr, indexerr := rn.persistIndex(namespaceAddr, namespace)

	if indexerr != nil {
		return nil, indexerr
	}

	out := &remoteNamespace{
		Addr:   indexAddr,
		Store:  rn.Store,
		Update: EmptyNamespace(),
	}

	return out, nil
}

func (rn *remoteNamespace) persistNamespace(namespace Namespace) (RemoteStoreAddress, error) {
	part := RemoteNamespaceRecord{Namespace: namespace}

	namespaceAddr, err := rn.Store.AddNamespace(part)

	if err != nil {
		return nil, errors.Wrap(err, "Error adding remoteNamespace to Store")
	}

	logdbg("Persisted Namespace at: %v", namespaceAddr)

	return namespaceAddr, nil
}

func (rn *remoteNamespace) persistIndex(addr RemoteStoreAddress, namespace Namespace) (RemoteStoreAddress, error) {
	index, loaderr := rn.loadIndex()

	if loaderr != nil {
		return nil, errors.Wrap(loaderr, "persistIndex failed")
	}

	newIndex := index.JoinNamespace(addr, namespace)
	addr, saveerr := rn.Store.AddIndex(newIndex)

	// TODO duplicate code.
	if saveerr != nil {
		return nil, errors.Wrap(saveerr, "persistIndex failed")
	}

	return addr, nil
}
