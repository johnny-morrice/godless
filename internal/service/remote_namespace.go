package service

import (
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/eval"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

type remoteNamespace struct {
	NamespaceUpdate crdt.Namespace
	IndexUpdate     crdt.Index
	Store           api.RemoteStore
	Addr            crdt.RemoteStoreAddress
}

func LoadRemoteNamespace(store api.RemoteStore, addr crdt.RemoteStoreAddress) (api.RemoteNamespaceTree, error) {
	rn := &remoteNamespace{}
	rn.Store = store
	rn.Addr = addr
	rn.NamespaceUpdate = crdt.EmptyNamespace()

	_, err := rn.loadCurrentIndex()

	// We don't use the index for anything at this point.
	log.Debug("crdt.Index found at '%v'", addr)

	if err != nil {
		return nil, errors.Wrap(err, "Error loading new namespace")
	}

	return rn, nil
}

func MakeRemoteNamespace(store api.RemoteStore, namespace crdt.Namespace) api.RemoteNamespaceTree {
	rn := &remoteNamespace{}
	rn.Store = store
	rn.NamespaceUpdate = namespace
	return rn
}

func PersistNewRemoteNamespace(store api.RemoteStore, namespace crdt.Namespace) (api.RemoteNamespaceTree, error) {
	rn := MakeRemoteNamespace(store, namespace)

	kv, err := rn.Persist()

	if err != nil {
		return nil, err
	}

	return kv.(*remoteNamespace), nil
}

func (rn *remoteNamespace) Reset() {
	rn.NamespaceUpdate = crdt.EmptyNamespace()
	rn.IndexUpdate = crdt.EmptyIndex()
}

func (rn *remoteNamespace) Replicate(peerAddr crdt.RemoteStoreAddress, kvq api.KvQuery) {
	runner := api.APIResponderFunc(func() api.APIResponse { return rn.joinPeerIndex(peerAddr) })
	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

func (rn *remoteNamespace) joinPeerIndex(peerAddr crdt.RemoteStoreAddress) api.APIResponse {
	const failMsg = "remoteNamespace.joinPeerIndex failed"

	failResponse := api.RESPONSE_FAIL
	failResponse.Type = api.API_REPLICATE

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

	return api.RESPONSE_REPLICATE
}

func (rn *remoteNamespace) loadIndex(indexAddr crdt.RemoteStoreAddress) (crdt.Index, error) {
	return rn.Store.CatIndex(indexAddr)
}

// TODO there are likely to be many reflection features.  Replace switches with polymorphism.
func (rn *remoteNamespace) RunKvReflection(reflect api.APIReflectionType, kvq api.KvQuery) {
	var runner api.APIResponder
	switch reflect {
	case api.REFLECT_HEAD_PATH:
		runner = api.APIResponderFunc(rn.getReflectHead)
	case api.REFLECT_INDEX:
		runner = api.APIResponderFunc(rn.getReflectIndex)
	case api.REFLECT_DUMP_NAMESPACE:
		runner = api.APIResponderFunc(rn.dumpReflectNamespaces)
	default:
		panic("Unknown reflection command")
	}

	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

// TODO Not sure if best place for these to live.
func (rn *remoteNamespace) getReflectHead() api.APIResponse {
	response := api.RESPONSE_REFLECT
	response.ReflectResponse.Type = api.REFLECT_HEAD_PATH

	if rn.Addr == nil {
		response.Err = errors.New("No index available")
	} else {
		response.ReflectResponse.Path = rn.Addr.Path()
	}

	return response
}

func (rn *remoteNamespace) getReflectIndex() api.APIResponse {
	const failMsg = "remoteNamespace.getReflectIndex failed"
	response := api.RESPONSE_REFLECT

	index, err := rn.loadCurrentIndex()

	if err != nil {
		response.Msg = api.RESPONSE_FAIL_MSG
		response.Err = errors.Wrap(err, failMsg)
		return response
	}

	response.ReflectResponse.Index = index
	response.ReflectResponse.Type = api.REFLECT_INDEX

	return response
}

func (rn *remoteNamespace) dumpReflectNamespaces() api.APIResponse {
	const failMsg = "remoteNamespace.dumpReflectNamespace failed"
	response := api.RESPONSE_REFLECT
	response.ReflectResponse.Type = api.REFLECT_DUMP_NAMESPACE

	index, err := rn.loadCurrentIndex()

	if err != nil {
		response.Msg = api.RESPONSE_FAIL_MSG
		response.Err = errors.Wrap(err, failMsg)
		response.Type = api.API_REFLECT
		return response
	}

	tables := index.AllTables()
	everything := crdt.EmptyNamespace()

	lambda := api.NamespaceTreeLambda(func(ns crdt.Namespace) (bool, error) {
		everything = everything.JoinNamespace(ns)
		return false, nil
	})
	traversal := api.AddTableHints(tables, lambda)
	rn.LoadTraverse(traversal)
	return response
}

// RunKvQuery will block until the result can be written to kvq.
func (rn *remoteNamespace) RunKvQuery(q *query.Query, kvq api.KvQuery) {
	var runner api.APIResponder

	switch q.OpCode {
	case query.JOIN:
		visitor := eval.MakeNamespaceTreeJoin(rn)
		q.Visit(visitor)
		runner = visitor
	case query.SELECT:
		visitor := eval.MakeNamespaceTreeSelect(rn)
		q.Visit(visitor)
		runner = visitor
	default:
		q.OpCodePanic()
	}

	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

func (rn *remoteNamespace) IsChanged() bool {
	return !(rn.NamespaceUpdate.IsEmpty() && rn.IndexUpdate.IsEmpty())
}

func (rn *remoteNamespace) JoinTable(tableKey crdt.TableName, table crdt.Table) error {
	joined := rn.NamespaceUpdate.JoinTable(tableKey, table)
	rn.NamespaceUpdate = joined
	return nil
}

func (rn *remoteNamespace) LoadTraverse(nttr api.NamespaceTreeTableReader) error {
	const failMsg = "remoteNamespace.LoadTraverse failed"

	index, indexerr := rn.loadCurrentIndex()

	if indexerr != nil {
		return errors.Wrap(indexerr, failMsg)
	}

	tableAddrs := rn.findTableAddrs(index, nttr)

	return rn.traverseTableNamespaces(tableAddrs, nttr)
}

func (rn *remoteNamespace) traverseTableNamespaces(tableAddrs []crdt.RemoteStoreAddress, f api.NamespaceTreeReader) error {
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
func (rn *remoteNamespace) namespaceLoader(addrs []crdt.RemoteStoreAddress) (<-chan crdt.Namespace, chan<- struct{}) {
	nsch := make(chan crdt.Namespace)
	cancelch := make(chan struct{}, 1)

	go func() {
		defer close(nsch)
		for _, a := range addrs {
			namespace, err := rn.Store.CatNamespace(a)

			if err != nil {
				log.Error("remoteNamespace.namespaceLoader failed: %v", err)
				return
			}

			log.Info("Catted namespace from: %v", a)
		LOOP:
			for {
				select {
				case <-cancelch:
					return
				case nsch <- namespace:
					break LOOP
				}
			}
		}
	}()

	return nsch, cancelch
}

func (rn *remoteNamespace) findTableAddrs(index crdt.Index, tableHints api.TableHinter) []crdt.RemoteStoreAddress {
	out := []crdt.RemoteStoreAddress{}
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
func (rn *remoteNamespace) loadCurrentIndex() (crdt.Index, error) {
	return rn.loadIndex(rn.Addr)
}

// Write pending changes to IPFS and return the new parent namespace.
func (rn *remoteNamespace) Persist() (api.RemoteNamespace, error) {
	const failMsg = "remoteNamespace.Persist failed"

	var nsIndex crdt.Index
	var index crdt.Index
	var namespaceAddr crdt.RemoteStoreAddress
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
		NamespaceUpdate: crdt.EmptyNamespace(),
	}

	return out, nil
}

func (rn *remoteNamespace) persistNamespace(namespace crdt.Namespace) (crdt.RemoteStoreAddress, error) {
	const failMsg = "remoteNamespace.persistNamespace failed"

	namespaceAddr, err := rn.Store.AddNamespace(namespace)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	log.Info("Persisted crdt.Namespace at: %v", namespaceAddr)

	return namespaceAddr, nil
}

func (rn *remoteNamespace) indexNamespace(namespaceAddr crdt.RemoteStoreAddress, namespace crdt.Namespace) (crdt.Index, error) {
	const failMsg = "remoteNamespace.indexNamespace failed"

	tableNames := namespace.GetTableNames()

	indices := map[crdt.TableName]crdt.RemoteStoreAddress{}
	for _, t := range tableNames {
		indices[t] = namespaceAddr
	}

	return crdt.MakeIndex(indices), nil
}

func (rn *remoteNamespace) persistIndex(newIndex crdt.Index) (crdt.RemoteStoreAddress, error) {
	const failMsg = "remoteNamespace.persistIndex failed"
	addr, saveerr := rn.Store.AddIndex(newIndex)

	if saveerr != nil {
		return nil, errors.Wrap(saveerr, failMsg)
	}

	log.Info("Persisted crdt.Index at %v", addr)

	return addr, nil
}
