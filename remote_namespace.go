package godless

import (
	"github.com/pkg/errors"
)

type remoteNamespace struct {
	NamespaceUpdate Namespace
	IndexUpdate     Index
	Store           RemoteStore
	Addr            RemoteStoreAddress
}

func LoadRemoteNamespace(store RemoteStore, addr RemoteStoreAddress) (KvNamespaceTree, error) {
	rn := &remoteNamespace{}
	rn.Store = store
	rn.Addr = addr
	rn.NamespaceUpdate = EmptyNamespace()

	_, err := rn.loadCurrentIndex()

	// We don't use the index for anything at this point.
	logdbg("Index found at '%v'", addr)

	if err != nil {
		return nil, errors.Wrap(err, "Error loading new namespace")
	}

	return rn, nil
}

func MakeRemoteNamespace(store RemoteStore, namespace Namespace) KvNamespaceTree {
	rn := &remoteNamespace{}
	rn.Store = store
	rn.NamespaceUpdate = namespace
	return rn
}

func PersistNewRemoteNamespace(store RemoteStore, namespace Namespace) (KvNamespaceTree, error) {
	rn := MakeRemoteNamespace(store, namespace)

	kv, err := rn.Persist()

	if err != nil {
		return nil, err
	}

	return kv.(*remoteNamespace), nil
}

func (rn *remoteNamespace) Reset() {
	rn.NamespaceUpdate = EmptyNamespace()
	rn.IndexUpdate = __EMPTY_INDEX
}

func (rn *remoteNamespace) Replicate(peerAddr RemoteStoreAddress, kvq KvQuery) {
	runner := APIResponderFunc(func() APIResponse { return rn.joinPeerIndex(peerAddr) })
	response := runner.RunQuery()
	kvq.writeResponse(response)
}

func (rn *remoteNamespace) joinPeerIndex(peerAddr RemoteStoreAddress) APIResponse {
	const failMsg = "remoteNamespace.joinPeerIndex failed"

	failResponse := RESPONSE_FAIL
	failResponse.Type = API_REPLICATE

	myIndex, myErr := rn.loadCurrentIndex()

	if myErr != nil {
		failResponse.Err = errors.Wrap(myErr, failMsg)
		return failResponse
	}

	theirIndex, theirErr := rn.loadIndex(peerAddr)

	if theirErr != nil {
		failResponse.Err = errors.Wrap(theirErr, failMsg)
		return failResponse
	}

	rn.IndexUpdate = myIndex.JoinIndex(theirIndex)

	return RESPONSE_REPLICATE
}

func (rn *remoteNamespace) loadIndex(indexAddr RemoteStoreAddress) (Index, error) {
	return rn.Store.CatIndex(indexAddr)
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
	response.ReflectResponse.Type = REFLECT_HEAD_PATH

	if rn.Addr == nil {
		response.Err = errors.New("No index available")
	} else {
		response.ReflectResponse.Path = rn.Addr.Path()
	}

	return response
}

func (rn *remoteNamespace) getReflectIndex() APIResponse {
	const failMsg = "remoteNamespace.getReflectIndex failed"
	response := RESPONSE_REFLECT

	index, err := rn.loadCurrentIndex()

	if err != nil {
		response.Msg = RESPONSE_FAIL_MSG
		response.Err = errors.Wrap(err, failMsg)
		return response
	}

	response.ReflectResponse.Index = index
	response.ReflectResponse.Type = REFLECT_INDEX

	return response
}

func (rn *remoteNamespace) dumpReflectNamespaces() APIResponse {
	const failMsg = "remoteNamespace.dumpReflectNamespace failed"
	response := RESPONSE_REFLECT
	response.ReflectResponse.Type = REFLECT_DUMP_NAMESPACE

	index, err := rn.loadCurrentIndex()

	if err != nil {
		response.Msg = RESPONSE_FAIL_MSG
		response.Err = errors.Wrap(err, failMsg)
		response.Type = API_REFLECT
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
	return !(rn.NamespaceUpdate.IsEmpty() && rn.IndexUpdate.IsEmpty())
}

func (rn *remoteNamespace) JoinTable(tableKey TableName, table Table) error {
	joined := rn.NamespaceUpdate.JoinTable(tableKey, table)
	rn.NamespaceUpdate = joined
	return nil
}

func (rn *remoteNamespace) LoadTraverse(nttr NamespaceTreeTableReader) error {
	const failMsg = "remoteNamespace.LoadTraverse failed"

	index, indexerr := rn.loadCurrentIndex()

	if indexerr != nil {
		return errors.Wrap(indexerr, failMsg)
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
				logerr("remoteNamespace.namespaceLoader failed: %v", err)
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

func (rn *remoteNamespace) findTableAddrs(index Index, tableHints TableHinter) []RemoteStoreAddress {
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
func (rn *remoteNamespace) loadCurrentIndex() (Index, error) {
	return rn.loadIndex(rn.Addr)
}

// Write pending changes to IPFS and return the new parent namespace.
func (rn *remoteNamespace) Persist() (KvNamespace, error) {
	const failMsg = "remoteNamespace.Persist failed"

	var nsIndex Index
	var index Index
	var namespaceAddr RemoteStoreAddress
	var nsErr error

	// TODO tidy up
	if !rn.NamespaceUpdate.IsEmpty() {
		namespaceAddr, nsErr = rn.persistNamespace(rn.NamespaceUpdate)

		if nsErr != nil {
			return nil, errors.Wrap(nsErr, failMsg)
		}

		var updateErr error
		nsIndex, updateErr = rn.indexNamespace(namespaceAddr, rn.NamespaceUpdate)

		if updateErr != nil {
			return nil, errors.Wrap(updateErr, failMsg)
		}
	}

	if rn.Addr != nil {
		var loadErr error
		index, loadErr = rn.loadCurrentIndex()

		if loadErr != nil {
			return nil, errors.Wrap(loadErr, failMsg)
		}
	}

	index = index.JoinIndex(nsIndex)
	index = index.JoinIndex(rn.IndexUpdate)

	indexAddr, indexErr := rn.persistIndex(index)

	if indexErr != nil {
		return nil, errors.Wrap(indexErr, failMsg)
	}

	out := &remoteNamespace{
		Addr:            indexAddr,
		Store:           rn.Store,
		NamespaceUpdate: EmptyNamespace(),
	}

	return out, nil
}

func (rn *remoteNamespace) persistNamespace(namespace Namespace) (RemoteStoreAddress, error) {
	const failMsg = "remoteNamespace.persistNamespace failed"
	part := RemoteNamespaceRecord{Namespace: namespace}

	namespaceAddr, err := rn.Store.AddNamespace(part)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	loginfo("Persisted Namespace at: %v", namespaceAddr)

	return namespaceAddr, nil
}

func (rn *remoteNamespace) indexNamespace(namespaceAddr RemoteStoreAddress, namespace Namespace) (Index, error) {
	const failMsg = "remoteNamespace.indexNamespace failed"

	tableNames := namespace.GetTableNames()

	indices := map[TableName]RemoteStoreAddress{}
	for _, t := range tableNames {
		indices[t] = namespaceAddr
	}

	return MakeIndex(indices), nil
}

func (rn *remoteNamespace) persistIndex(newIndex Index) (RemoteStoreAddress, error) {
	const failMsg = "remoteNamespace.persistIndex failed"
	addr, saveerr := rn.Store.AddIndex(newIndex)

	if saveerr != nil {
		return nil, errors.Wrap(saveerr, failMsg)
	}

	loginfo("Persisted Index at %v", addr)

	return addr, nil
}
