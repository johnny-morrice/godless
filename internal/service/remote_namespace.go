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
	Cache           api.HeadCache
}

func MakeRemoteNamespace(store api.RemoteStore, cache api.HeadCache) api.RemoteNamespaceTree {
	return &remoteNamespace{Store: store, Cache: cache}
}

func (rn *remoteNamespace) Rollback() error {
	const failMsg = "remoteNamespace.Rollback failed"
	rn.NamespaceUpdate = crdt.EmptyNamespace()
	rn.IndexUpdate = crdt.EmptyIndex()
	err := rn.Cache.Rollback()

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func (rn *remoteNamespace) Commit() error {
	const failMsg = "remoteNamespace.Commit failed"

	err := rn.Cache.Commit()

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func (rn *remoteNamespace) Replicate(peerAddr crdt.IPFSPath, kvq api.KvQuery) {
	runner := api.APIResponderFunc(func() api.APIResponse { return rn.joinPeerIndex(peerAddr) })
	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

func (rn *remoteNamespace) joinPeerIndex(peerAddr crdt.IPFSPath) api.APIResponse {
	const failMsg = "remoteNamespace.joinPeerIndex failed"

	failResponse := api.RESPONSE_FAIL
	failResponse.Type = api.API_REPLICATE

	myAddr, cacheErr := rn.getHeadTransaction()

	if cacheErr != nil {
		failResponse.Err = errors.Wrap(cacheErr, failMsg)
		return failResponse
	}

	if peerAddr == myAddr {
		return api.RESPONSE_REPLICATE
	}

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

func (rn *remoteNamespace) loadIndex(indexAddr crdt.IPFSPath) (crdt.Index, error) {
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

	myAddr, err := rn.getHeadTransaction()

	if err != nil {
		response.Err = errors.Wrap(err, "remoteNamespace.getReflectHead failed")
		response.Msg = api.RESPONSE_FAIL_MSG
	} else if crdt.IsNilPath(myAddr) {
		response.Err = errors.New("No index available")
		response.Msg = api.RESPONSE_FAIL_MSG
	} else {
		response.ReflectResponse.Path = myAddr
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
		response = api.RESPONSE_FAIL
		response.Err = errors.Wrap(err, failMsg)
		response.Type = api.API_REFLECT
		return response
	}

	tables := index.AllTables()
	everything := crdt.EmptyNamespace()

	lambda := api.NamespaceTreeLambda(func(ns crdt.Namespace) api.TraversalUpdate {
		everything = everything.JoinNamespace(ns)
		return api.TraversalUpdate{More: true}
	})
	traversal := api.AddTableHints(tables, lambda)
	err = rn.LoadTraverse(traversal)

	if err != nil {
		response = api.RESPONSE_FAIL
		response.Err = errors.Wrap(err, failMsg)
		response.Type = api.API_REFLECT
	}

	response.ReflectResponse.Namespace = everything
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

func (rn *remoteNamespace) traverseTableNamespaces(tableAddrs []crdt.IPFSPath, f api.NamespaceTreeReader) error {
	nsch, cancelch := rn.namespaceLoader(tableAddrs)
	defer close(cancelch)
	for ns := range nsch {
		update := f.ReadNamespace(ns)

		if !(update.More && update.Error == nil) {
			cancelch <- struct{}{}
		}

		if update.Error != nil {
			return errors.Wrap(update.Error, "traverseTableNamespaces failed")
		}
	}

	return nil
}

// Preload namespaces while the previous is analysed.
func (rn *remoteNamespace) namespaceLoader(addrs []crdt.IPFSPath) (<-chan crdt.Namespace, chan<- struct{}) {
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

func (rn *remoteNamespace) findTableAddrs(index crdt.Index, tableHints api.TableHinter) []crdt.IPFSPath {
	out := []crdt.IPFSPath{}
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
	myAddr, err := rn.getHeadTransaction()

	if err != nil {
		return crdt.EmptyIndex(), errors.Wrap(err, "remoteNamespace.loadCurrentIndex failed")
	} else if crdt.IsNilPath(myAddr) {
		return crdt.EmptyIndex(), errors.New("No current index")
	}

	return rn.loadIndex(myAddr)
}

// Write pending changes to IPFS and return the new parent namespace.
func (rn *remoteNamespace) Persist() error {
	const failMsg = "remoteNamespace.Persist failed"

	var nsIndex crdt.Index
	var index crdt.Index
	var namespaceAddr crdt.IPFSPath
	var nsErr error

	log.Info("Persisting RemoteNamespace...")

	// TODO tidy up
	if !rn.NamespaceUpdate.IsEmpty() {
		namespaceAddr, nsErr = rn.persistNamespace(rn.NamespaceUpdate)

		if nsErr != nil {
			return errors.Wrap(nsErr, failMsg)
		}

		var updateErr error
		nsIndex, updateErr = rn.indexNamespace(namespaceAddr, rn.NamespaceUpdate)

		if updateErr != nil {
			return errors.Wrap(updateErr, failMsg)
		}
	}

	newIndex := nsIndex.JoinIndex(rn.IndexUpdate)

	cacheErr := rn.Cache.BeginWriteTransaction()

	if cacheErr != nil {
		return errors.Wrap(cacheErr, failMsg)
	}

	myAddr, getErr := rn.Cache.GetHead()

	if getErr != nil {
		return errors.Wrap(getErr, failMsg)
	}

	if !crdt.IsNilPath(myAddr) {
		var loadErr error
		index, loadErr = rn.loadIndex(myAddr)

		if loadErr != nil {
			return errors.Wrap(loadErr, failMsg)
		}
	}

	index = index.JoinIndex(newIndex)

	indexAddr, indexErr := rn.persistIndex(index)

	if indexErr != nil {
		return errors.Wrap(indexErr, failMsg)
	}

	setErr := rn.Cache.SetHead(indexAddr)

	if setErr != nil {
		return errors.Wrap(setErr, failMsg)
	}

	log.Info("Persisted RemoteNamespace")

	return nil
}

func (rn *remoteNamespace) persistNamespace(namespace crdt.Namespace) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.persistNamespace failed"

	namespaceAddr, err := rn.Store.AddNamespace(namespace)

	if err != nil {
		return crdt.NIL_PATH, errors.Wrap(err, failMsg)
	}

	log.Info("Persisted crdt.Namespace at: %v", namespaceAddr)

	return namespaceAddr, nil
}

func (rn *remoteNamespace) indexNamespace(namespaceAddr crdt.IPFSPath, namespace crdt.Namespace) (crdt.Index, error) {
	const failMsg = "remoteNamespace.indexNamespace failed"

	tableNames := namespace.GetTableNames()

	indices := map[crdt.TableName]crdt.IPFSPath{}
	for _, t := range tableNames {
		indices[t] = namespaceAddr
	}

	return crdt.MakeIndex(indices), nil
}

func (rn *remoteNamespace) persistIndex(newIndex crdt.Index) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.persistIndex failed"
	addr, saveerr := rn.Store.AddIndex(newIndex)

	if saveerr != nil {
		return crdt.NIL_PATH, errors.Wrap(saveerr, failMsg)
	}

	log.Info("Persisted crdt.Index at %v", addr)

	return addr, nil
}

func (rn *remoteNamespace) getHeadTransaction() (crdt.IPFSPath, error) {
	var path crdt.IPFSPath
	err := rn.Cache.BeginReadTransaction()

	if err != nil {
		return crdt.NIL_PATH, err
	}

	path, err = rn.Cache.GetHead()

	if err != nil {
		return crdt.NIL_PATH, err
	}

	err = rn.Cache.Commit()

	if err != nil {
		return crdt.NIL_PATH, err
	}

	return path, nil
}
