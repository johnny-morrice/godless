package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/function"
	"github.com/johnny-morrice/godless/internal/eval"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

type RemoteNamespaceCoreOptions struct {
	Store          api.RemoteStore
	HeadCache      api.HeadCache
	MemoryImage    api.MemoryImage
	IndexCache     api.IndexCache
	NamespaceCache api.NamespaceCache
	KeyStore       api.KeyStore
	IsPublicIndex  bool
	Pulse          time.Duration
	Debug          bool
	Functions      function.FunctionNamespace
}

func checkOptions(options RemoteNamespaceCoreOptions) {
	requiredOptions := map[string]interface{}{
		"Store":       options.Store,
		"HeadCache":   options.HeadCache,
		"MemoryImage": options.MemoryImage,
		"IndexCache":  options.IndexCache,
		"KeyStore":    options.KeyStore,
	}

	for name, req := range requiredOptions {
		if req == nil {
			panic(fmt.Sprintf("required RemoteNamespaceOption %v was nil", name))
		}
	}
}

type remoteNamespace struct {
	RemoteNamespaceCoreOptions
	namespaceTube chan addNamespaceRequest
	indexTube     chan addIndexRequest
	pulser        *time.Ticker
	stopch        chan struct{}
	wg            *sync.WaitGroup
	memImgTracker dirtyTracker
}

func MakeRemoteNamespaceCore(options RemoteNamespaceCoreOptions) api.RemoteNamespaceCore {
	pulseInterval := options.Pulse
	if pulseInterval == 0 {
		pulseInterval = __DEFAULT_PULSE
	}

	checkOptions(options)

	remote := &remoteNamespace{
		RemoteNamespaceCoreOptions: options,
		namespaceTube:              make(chan addNamespaceRequest),
		indexTube:                  make(chan addIndexRequest),
		pulser:                     time.NewTicker(pulseInterval),
		stopch:                     make(chan struct{}),
		wg:                         &sync.WaitGroup{},
		memImgTracker:              makeDirtyTracker(),
	}

	remote.wg.Add(__REMOTE_NAMESPACE_PROCESS_COUNT)
	initWait := remote.initializeMemoryImage()
	go remote.addNamespaces()
	go remote.addIndices()
	go remote.memoryImageWriteLoop()

	if initWait != nil {
		<-initWait
	}

	return remote
}

func (rn *remoteNamespace) initializeMemoryImage() <-chan struct{} {
	donech := make(chan struct{})
	head, err := rn.getHead()

	if err != nil {
		log.Error("remoteNamespace initialization failed to get HEAD: %s", err.Error())
		return nil
	}

	if crdt.IsNilPath(head) {
		return nil
	}

	index, err := rn.loadIndex(head)

	if err != nil {
		log.Error("remoteNamespace initialization failed to load Index: %s", err.Error())
		return nil
	}

	go func() {
		_, err := rn.insertIndex(index)

		if err == nil {
			log.Info("Initialized remoteNamespace with Index at: %s", head)
		} else {
			log.Error("Failed to initialize remoteNamespace with Index (%s): %s", head, err.Error())
		}

		close(donech)
	}()

	return donech
}

func (rn *remoteNamespace) memoryImageWriteLoop() {
	defer rn.wg.Done()
	for {
		select {
		case <-rn.stopch:
			return
		case <-rn.pulser.C:
			rn.writeDirtyMemoryImage()
		}
	}
}

func (rn *remoteNamespace) writeDirtyMemoryImage() {
	select {
	case <-rn.stopch:
		return
	case <-rn.memImgTracker.dirt:
		err := rn.WriteMemoryImage()
		if err != nil {
			log.Error("Error writing MemoryImage: %v", err)
		}
		return
	}
}

func (rn *remoteNamespace) WriteMemoryImage() error {
	index, err := rn.MemoryImage.GetIndex()

	if err != nil {
		return errors.Wrap(err, "Error joining MemoryImage indices")
	}

	path, err := rn.persistIndex(index)

	if err != nil {
		return errors.Wrap(err, "Error saving MemoryImage to IPFS")
	}

	log.Info("Added MemoryImage Index to IPFS at: %s", path)
	err = rn.setHead(path)

	if err != nil {
		return errors.Wrap(err, "Failed to update HEAD cache")
	}

	return nil
}

func (rn *remoteNamespace) addNamespaces() {
	defer rn.wg.Done()
	for {
		select {
		case tubeItem := <-rn.namespaceTube:
			addRequest := tubeItem
			rn.fork(func() { rn.addNamespace(addRequest) })
		case <-rn.stopch:
			return
		}
	}
}

func (rn *remoteNamespace) addIndices() {
	defer rn.wg.Done()
	for {
		select {
		case tubeItem := <-rn.indexTube:
			addRequest := tubeItem
			rn.fork(func() { rn.addIndex(addRequest) })
		case <-rn.stopch:
			return
		}
	}
}

func (rn *remoteNamespace) addNamespace(addRequest addNamespaceRequest) {
	path, err := rn.Store.AddNamespace(addRequest.namespace)
	addRequest.reply(path, err)
}

func (rn *remoteNamespace) addIndex(addRequest addIndexRequest) {
	index := addRequest.index

	errch := make(chan error, 1)

	go func() {
		err := rn.MemoryImage.JoinIndex(index)
		rn.memImgTracker.markDirty()
		errch <- err
	}()

	path, ipfsErr := rn.persistIndex(index)
	persistErr := <-errch

	if persistErr == nil {
		persistErr = ipfsErr
	} else if ipfsErr != nil {
		messages := []string{persistErr.Error(), ipfsErr.Error()}
		msg := strings.Join(messages, " and ")
		persistErr = errors.New(msg)
	}

	addRequest.reply(path, persistErr)
}

func (rn *remoteNamespace) persistIndex(index crdt.Index) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.persistIndex failed"
	indexAddr, addErr := rn.Store.AddIndex(index)

	if addErr != nil {
		return crdt.NIL_PATH, errors.Wrap(addErr, failMsg)
	}

	cacheErr := rn.IndexCache.SetIndex(indexAddr, index)

	if cacheErr != nil {
		log.Error("Failed to write index cache for: %s (%s)", indexAddr, cacheErr.Error())
	}

	return indexAddr, nil
}

func (rn *remoteNamespace) Close() {
	close(rn.stopch)
	rn.wg.Wait()
	log.Info("Closed remoteNamespace")
}

func (rn *remoteNamespace) Replicate(links []crdt.Link, kvq api.Command) {
	runner := api.ResponderLambda(func() api.Response { return rn.joinPeerIndex(links) })
	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

func (rn *remoteNamespace) joinPeerIndex(links []crdt.Link) api.Response {
	const failMsg = "remoteNamespace.joinPeerIndex failed"
	failResponse := api.RESPONSE_FAIL

	log.Info("Replicating peer indices...")

	keys := rn.KeyStore.GetAllPublicKeys()

	joined := crdt.EmptyIndex()

	someFailed := false
	for _, link := range links {
		if !rn.IsPublicIndex {
			log.Info("Verifying link...")
			isVerified := link.IsVerifiedByAny(keys)
			if !isVerified {
				log.Warn("Skipping unverified Index Link")
				someFailed = true
				continue
			}
			log.Info("Verified link: %s", link.Path())
		}

		peerAddr := link.Path()

		theirIndex, theirErr := rn.loadIndex(peerAddr)

		if theirErr != nil {
			log.Error("Failed to replicate Index at: %s", peerAddr)
			someFailed = true
			continue
		}

		joined = joined.JoinIndex(theirIndex)
	}

	indexAddr, perr := rn.insertIndex(joined)

	if perr != nil {
		log.Error("Index replication failed")
		failResponse.Err = errors.Wrap(perr, failMsg)
		return failResponse
	}

	resp := api.RESPONSE_REPLICATE

	if someFailed {
		resp.Msg = "Update ok with load failures"
	}

	log.Info("Index replicated to: %s", indexAddr)

	return resp
}

func (rn *remoteNamespace) loadIndex(indexAddr crdt.IPFSPath) (crdt.Index, error) {
	const failMsg = "remoteNamespace.loadIndex failed"
	cached, cacheErr := rn.IndexCache.GetIndex(indexAddr)

	if cacheErr == nil {
		return cached, nil
	} else {
		log.Warn("Index cache miss for: %s", indexAddr)
	}

	index, err := rn.Store.CatIndex(indexAddr)

	if err != nil {
		return crdt.EmptyIndex(), errors.Wrap(err, failMsg)
	}

	go rn.updateIndexCache(indexAddr, index)

	return index, nil
}

// TODO there are likely to be many reflection features.  Replace switches with polymorphism.
func (rn *remoteNamespace) Reflect(reflect api.ReflectionType, kvq api.Command) {
	var runner api.Responder
	switch reflect {
	case api.REFLECT_HEAD_PATH:
		runner = api.ResponderLambda(rn.getReflectHead)
	case api.REFLECT_INDEX:
		runner = api.ResponderLambda(rn.getReflectIndex)
	case api.REFLECT_DUMP_NAMESPACE:
		runner = api.ResponderLambda(rn.dumpReflectNamespaces)
	default:
		panic("Unknown reflection command")
	}

	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

// TODO Not sure if best place for these to live.
func (rn *remoteNamespace) getReflectHead() api.Response {
	response := api.RESPONSE_REFLECT

	myAddr, err := rn.getHead()

	if err != nil {
		response.Err = errors.Wrap(err, "remoteNamespace.getReflectHead failed")
		response.Msg = api.RESPONSE_FAIL_MSG
	} else if crdt.IsNilPath(myAddr) {
		response.Err = errors.New("No index available")
		response.Msg = api.RESPONSE_FAIL_MSG
	} else {
		response.Path = myAddr
	}

	return response
}

func (rn *remoteNamespace) getReflectIndex() api.Response {
	const failMsg = "remoteNamespace.getReflectIndex failed"
	response := api.RESPONSE_REFLECT

	index, err := rn.loadCurrentIndex()

	if err != nil {
		response.Msg = api.RESPONSE_FAIL_MSG
		response.Err = errors.Wrap(err, failMsg)
		return response
	}

	response.Index = index

	return response
}

func (rn *remoteNamespace) dumpReflectNamespaces() api.Response {
	const failMsg = "remoteNamespace.dumpReflectNamespace failed"
	response := api.RESPONSE_REFLECT

	index, err := rn.loadCurrentIndex()

	if err != nil {
		response = api.RESPONSE_FAIL
		response.Err = errors.Wrap(err, failMsg)
		response.Type = api.API_REFLECT
		return response
	}

	everything := crdt.EmptyNamespace()

	namespaceError := false
	indexError := false
	lambda := api.SearchResultLambda(func(result api.SearchResult) api.TraversalUpdate {
		more := api.TraversalUpdate{More: true}

		if result.NamespaceLoadFailure {
			namespaceError = true
			return more
		}

		if result.IndexLoadFailure {
			indexError = true
			return api.TraversalUpdate{More: false}
		}

		everything = everything.JoinNamespace(result.Namespace)
		return more
	})

	searcher := api.SignedTableSearcher{
		Keys:   rn.KeyStore.GetAllPublicKeys(),
		Reader: lambda,
		Tables: index.AllTables(),
	}

	err = rn.LoadTraverse(searcher)

	if err != nil {
		response = api.RESPONSE_FAIL
		response.Err = errors.Wrap(err, failMsg)
		response.Type = api.API_REFLECT
	}

	if indexError {
		log.Error("Index load errors should be handled already")
		err := errors.New("Index load failure")
		response = api.RESPONSE_FAIL
		response.Err = errors.Wrap(err, failMsg)
		response.Type = api.API_REFLECT
	}

	response.Namespace = everything

	if namespaceError {
		response.Msg = "ok with load errors"
	}

	return response
}

// RunQuery will block until the result can be written to kvq.
func (rn *remoteNamespace) RunQuery(q *query.Query, kvq api.Command) {
	var runner api.Responder

	switch q.OpCode {
	case query.JOIN:
		log.Info("Running join...")
		visitor := eval.MakeNamespaceTreeJoin(rn, rn.KeyStore)
		q.Visit(visitor)
		runner = visitor
	case query.SELECT:
		log.Info("Running select...")
		options := eval.SelectOptions{
			Namespace: rn,
			KeyStore:  rn.KeyStore,
			Functions: rn.Functions,
		}
		visitor := eval.MakeNamespaceTreeSelect(options)
		q.Visit(visitor)
		runner = visitor
	default:
		q.OpCodePanic()
	}

	response := runner.RunQuery()
	kvq.WriteResponse(response)
}

// TODO there should be more clarity on who locks and when.
func (rn *remoteNamespace) JoinTable(tableKey crdt.TableName, table crdt.Table) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.JoinTable failed"

	joined := crdt.EmptyNamespace().JoinTable(tableKey, table)

	addr, nsErr := rn.insertNamespace(joined)

	if nsErr != nil {
		return crdt.NIL_PATH, errors.Wrap(nsErr, failMsg)
	}

	signed, signErr := crdt.SignedLink(addr, rn.KeyStore.GetAllPrivateKeys())

	if signErr != nil {
		return crdt.NIL_PATH, errors.Wrap(signErr, failMsg)
	}

	index := crdt.EmptyIndex().JoinTable(tableKey, signed)

	indexAddr, indexErr := rn.insertIndex(index)

	if indexErr != nil {
		return crdt.NIL_PATH, errors.Wrap(indexErr, failMsg)
	}

	return indexAddr, nil
}

func (rn *remoteNamespace) LoadTraverse(searcher api.NamespaceSearcher) error {
	const failMsg = "remoteNamespace.LoadTraverse failed"

	index, indexerr := rn.loadCurrentIndex()

	if indexerr != nil {
		indexLoadFailure := api.SearchResult{IndexLoadFailure: true}
		searcher.ReadSearchResult(indexLoadFailure)
		return errors.Wrap(indexerr, failMsg)
	}

	tableAddrs := searcher.Search(index)

	return rn.traverseTableNamespaces(tableAddrs, searcher)
}

func (rn *remoteNamespace) traverseTableNamespaces(tableAddrs []crdt.Link, f api.SearchResultTraverser) error {
	resultch, cancelch := rn.namespaceLoader(tableAddrs)
	defer close(cancelch)
	for result := range resultch {
		update := f.ReadSearchResult(result)

		if !update.More && update.Error == nil {
			log.Info("Cancelling traverse...")
			return nil
		}

		if update.Error != nil {
			log.Info("Aborting traverse with error: %s", update.Error.Error())
			return errors.Wrap(update.Error, "traverseTableNamespaces failed")
		}
	}

	return nil
}

// Preload namespaces while the previous is analysed.
func (rn *remoteNamespace) namespaceLoader(addrs []crdt.Link) (<-chan api.SearchResult, chan<- struct{}) {
	resultch := make(chan api.SearchResult)
	cancelch := make(chan struct{}, 1)

	go func() {
		defer close(resultch)
		for _, a := range addrs {
			namespace, err := rn.loadNamespace(a.Path())

			if err != nil {
				log.Error("remoteNamespace.namespaceLoader: %s", err.Error())
				resultch <- api.SearchResult{NamespaceLoadFailure: true}
				continue
			}

			log.Info("Catted namespace from: %s", a.Path())
			successResult := api.SearchResult{Namespace: namespace}
			select {
			case <-rn.stopch:
				return
			case <-cancelch:
				return
			case resultch <- successResult:
				break
			}
		}
	}()

	return resultch, cancelch
}

func (rn *remoteNamespace) loadNamespace(namespaceAddr crdt.IPFSPath) (crdt.Namespace, error) {
	const failMsg = "remoteNamespace.loadNamespace failed"

	ns, cacheErr := rn.NamespaceCache.GetNamespace(namespaceAddr)

	if cacheErr == nil {
		return ns, nil
	}

	log.Info("Cache miss for namespace at: %s", namespaceAddr)
	ns, remoteErr := rn.Store.CatNamespace(namespaceAddr)

	if remoteErr != nil {
		return crdt.EmptyNamespace(), errors.Wrap(remoteErr, failMsg)
	}

	return ns, nil
}

func (rn *remoteNamespace) loadCurrentIndex() (crdt.Index, error) {
	const failMsg = "remoteNamespace.loadCurrentIndex failed"

	index, err := rn.MemoryImage.GetIndex()

	// TODO fall back to GetHead on memory image failure.

	return index, errors.Wrap(err, failMsg)
}

func (rn *remoteNamespace) fork(f func()) {
	if rn.Debug {
		f()
		return
	}

	go f()
}

func (rn *remoteNamespace) writeNamespaceTube(addRequest addNamespaceRequest) {
	select {
	case <-rn.stopch:
		return
	case rn.namespaceTube <- addRequest:
		return
	}
}

func (rn *remoteNamespace) writeIndexTube(addRequest addIndexRequest) {
	select {
	case <-rn.stopch:
		return
	case rn.indexTube <- addRequest:
		return
	}
}

func (rn *remoteNamespace) readAddResponse(respch chan addResponse) addResponse {
	select {
	case <-rn.stopch:
		return addResponse{err: errors.New("remote namespace stopped")}
	case resp := <-respch:
		return resp
	}
}

func (rn *remoteNamespace) insertNamespace(namespace crdt.Namespace) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.persistNamespace failed"
	resultChan := make(chan addResponse)
	addRequest := addNamespaceRequest{
		namespace: namespace,
		addResponder: addResponder{
			result: resultChan,
			stopch: rn.stopch,
		},
	}
	rn.writeNamespaceTube(addRequest)

	result := rn.readAddResponse(resultChan)

	if result.err != nil {
		return crdt.NIL_PATH, errors.Wrap(result.err, failMsg)
	}

	cacheErr := rn.NamespaceCache.SetNamespace(result.path, namespace)

	if cacheErr != nil {
		log.Error("Failed to write to namespace cache at: %s (%s)", result.path, cacheErr.Error())
	}

	return result.path, nil
}

func (rn *remoteNamespace) insertIndex(index crdt.Index) (crdt.IPFSPath, error) {
	const failMsg = "remoteNamespace.persistIndex failed"

	resultChan := make(chan addResponse)
	addRequest := addIndexRequest{
		index: index,
		addResponder: addResponder{
			result: resultChan,
			stopch: rn.stopch,
		},
	}
	rn.writeIndexTube(addRequest)

	result := rn.readAddResponse(resultChan)

	if result.err != nil {
		return crdt.NIL_PATH, errors.Wrap(result.err, failMsg)
	}

	return result.path, nil
}

func (rn *remoteNamespace) updateIndexCache(addr crdt.IPFSPath, index crdt.Index) {
	err := rn.IndexCache.SetIndex(addr, index)
	if err != nil {
		log.Error("Failed to update index cache: %s", err.Error())
	}
}

func (rn *remoteNamespace) getHead() (crdt.IPFSPath, error) {
	return rn.HeadCache.GetHead()
}

func (rn *remoteNamespace) setHead(head crdt.IPFSPath) error {
	return rn.HeadCache.SetHead(head)
}

type dirtyTracker struct {
	dirt chan struct{}
}

func makeDirtyTracker() dirtyTracker {
	return dirtyTracker{dirt: make(chan struct{}, 1)}
}

func (tracker *dirtyTracker) markDirty() {
	select {
	case tracker.dirt <- struct{}{}:
		return
	default:
		// Already written
		return
	}
}

type addResponse struct {
	path crdt.IPFSPath
	err  error
}

type addResponder struct {
	result chan addResponse
	stopch chan struct{}
}

type addNamespaceRequest struct {
	namespace crdt.Namespace
	addResponder
}

// TODO leaks goroutine after shutdown.
func (add addResponder) reply(path crdt.IPFSPath, err error) {
	go func() {
		defer close(add.result)
		select {
		case add.result <- addResponse{path: path, err: err}:
			return
		case <-add.stopch:
			return
		}

	}()
}

type addIndexRequest struct {
	index crdt.Index
	addResponder
}

const __DEFAULT_PULSE = time.Second * 10
const __REMOTE_NAMESPACE_PROCESS_COUNT = 3
