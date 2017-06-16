package service

import (
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/eval"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

type addResponse struct {
	path crdt.IPFSPath
	err  error
}

type addNamespace struct {
	namespace crdt.Namespace
	result    chan addResponse
}

type addIndex struct {
	index  crdt.Index
	result chan addResponse
}

type remoteNamespace struct {
	namespaceTube chan addNamespace
	indexTube     chan addIndex
	Store         api.RemoteStore
	HeadCache     api.HeadCache
	IndexCache    api.IndexCache
}

func MakeRemoteNamespace(store api.RemoteStore, headCache api.HeadCache, indexCache api.IndexCache) api.RemoteNamespaceTree {
	remote := &remoteNamespace{Store: store, HeadCache: headCache, IndexCache: indexCache}
	go remote.AddNamespaces()
	go remote.AddIndices()
	return remote
}

func (rn *remoteNamespace) AddNamespaces() {
	for tubeItem := range rn.namespaceTube {
		addNs := tubeItem
		go func() {
			defer close(addNs.result)
			path, err := rn.Store.AddNamespace(addNs.namespace)
			addNs.result <- addResponse{path: path, err: err}
		}()
	}
}

func (rn *remoteNamespace) AddIndices() {
	for tubeItem := range rn.indexTube {
		addIdx := tubeItem
		index, loadErr := rn.loadCurrentIndex()

		if loadErr != nil {
			go func() {
				defer close(addIdx.result)
				addIdx.result <- addResponse{err: loadErr}
			}()
			continue
		}

		nextIndex := index.JoinIndex(addIdx.index)
		path, addErr := rn.Store.AddIndex(nextIndex)
		go func() {
			defer close(addIdx.result)
			addIdx.result <- addResponse{path: path, err: addErr}
		}()
	}
}

func (rn *remoteNamespace) Close() {
	close(rn.namespaceTube)
	close(rn.indexTube)
}

func (rn *remoteNamespace) Replicate(peerAddr crdt.IPFSPath, kvq api.KvQuery) {
	runner := api.APIResponderFunc(func() api.APIResponse { return rn.joinPeerIndex(peerAddr) })
	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

func (rn *remoteNamespace) joinPeerIndex(peerAddr crdt.IPFSPath) api.APIResponse {
	const failMsg = "remoteNamespace.joinPeerIndex failed"
	failResponse := api.RESPONSE_FAIL

	theirIndex, theirErr := rn.loadIndex(peerAddr)

	if theirErr != nil {
		failResponse.Err = errors.Wrap(theirErr, failMsg)
		return failResponse
	}

	_, perr := rn.persistIndex(theirIndex)

	if perr != nil {
		failResponse.Err = errors.Wrap(perr, failMsg)
		return failResponse
	}

	return api.RESPONSE_REPLICATE
}

func (rn *remoteNamespace) loadIndex(indexAddr crdt.IPFSPath) (crdt.Index, error) {
	const failMsg = "remoteNamespace.loadIndex failed"
	cached, cacheErr := rn.IndexCache.GetIndex(indexAddr)

	if cacheErr == nil {
		return cached, nil
	} else {
		log.Warn("Index cache miss for: %v", indexAddr)
	}

	index, err := rn.Store.CatIndex(indexAddr)

	if err != nil {
		return crdt.EmptyIndex(), errors.Wrap(err, failMsg)
	}

	go rn.updateIndexCache(indexAddr, index)

	return index, nil
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

// TODO there should be more clarity on who locks and when.
func (rn *remoteNamespace) JoinTable(tableKey crdt.TableName, table crdt.Table) error {
	const failMsg = "remoteNamespace.JoinTable failed"

	joined := crdt.EmptyNamespace().JoinTable(tableKey, table)

	addr, nsErr := rn.persistNamespace(joined)

	if nsErr != nil {
		return errors.Wrap(nsErr, failMsg)
	}

	index := crdt.EmptyIndex().JoinTable(tableKey, addr)

	_, indexErr := rn.persistIndex(index)

	if indexErr != nil {
		return errors.Wrap(indexErr, failMsg)
	}

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

func (rn *remoteNamespace) persistNamespace(namespace crdt.Namespace) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.persistNamespace failed"
	resultChan := make(chan addResponse)
	rn.namespaceTube <- addNamespace{namespace: namespace, result: resultChan}

	result := <-resultChan

	if result.err != nil {
		return crdt.NIL_PATH, errors.Wrap(result.err, failMsg)
	}

	return result.path, nil
}

func (rn *remoteNamespace) persistIndex(index crdt.Index) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.persistIndex failed"

	resultChan := make(chan addResponse)
	rn.indexTube <- addIndex{index: index, result: resultChan}

	result := <-resultChan

	if result.err != nil {
		return crdt.NIL_PATH, errors.Wrap(result.err, failMsg)
	}

	return result.path, nil
}

func (rn *remoteNamespace) updateIndexCache(addr crdt.IPFSPath, index crdt.Index) {
	err := rn.IndexCache.SetIndex(addr, index)
	if err != nil {
		log.Error("Failed to update index cache: %v", err)
	}
}

func (rn *remoteNamespace) getHeadTransaction() (crdt.IPFSPath, error) {
	var path crdt.IPFSPath
	err := rn.HeadCache.BeginReadTransaction()

	if err != nil {
		return crdt.NIL_PATH, err
	}

	defer func() {
		err := rn.HeadCache.Commit()
		if err != nil {
			log.Error("Error commiting cache: %v", err)
		}
	}()

	path, err = rn.HeadCache.GetHead()

	if err != nil {
		return crdt.NIL_PATH, err
	}

	return path, nil
}
