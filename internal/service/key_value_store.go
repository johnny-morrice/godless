package service

import (
	"fmt"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/log"
)

type keyValueStore struct {
	Debug     bool
	namespace api.RemoteNamespace
	queue     api.RequestPriorityQueue
	semaphore chan struct{}
	stopch    chan struct{}
}

func LaunchKeyValueStore(ns api.RemoteNamespace, queue api.RequestPriorityQueue, queryLimit int) (api.APIService, <-chan error) {
	errch := make(chan error, 1)

	kv := &keyValueStore{
		namespace: ns,
		queue:     queue,
		stopch:    make(chan struct{}),
	}

	if queryLimit > 0 {
		kv.semaphore = make(chan struct{}, queryLimit)
	}

	go kv.executeLoop(errch)

	return kv, errch
}

func (kv *keyValueStore) executeLoop(errch chan<- error) {
	defer close(errch)
	drainch := kv.queue.Drain()
	for {
		select {
		case anything, ok := <-drainch:
			if ok {
				thing := anything
				kv.fork(func() { kv.runQueueItem(errch, thing) })
			}
		case <-kv.stopch:
			return
		}
	}
}

func (kv *keyValueStore) fork(f func()) {
	if kv.Debug {
		f()
		return
	}

	go f()
}

func (kv *keyValueStore) runQueueItem(errch chan<- error, thing interface{}) {
	kv.lockResource()
	defer kv.unlockResource()

	log.Info("API executing request, %d remain in queue", kv.queue.Len())
	kvq, ok := thing.(api.KvQuery)

	if !ok {
		log.Error("Corrupt queue")
		errch <- fmt.Errorf("Corrupt queue")
	}

	kv.run(kvq)
}

func (kv *keyValueStore) run(kvq api.KvQuery) {
	go kvq.Run(kv.namespace)
}

func (kv *keyValueStore) lockResource() {
	if kv.semaphore == nil {
		return
	}

	log.Debug("API waiting for resource...")
	kv.semaphore <- struct{}{}
	log.Debug("API found resource")
}

func (kv *keyValueStore) unlockResource() {
	if kv.semaphore == nil {
		return
	}

	log.Debug("API releasing resource...")
	<-kv.semaphore
	log.Debug("API released resource")
}

func (kv *keyValueStore) Call(request api.APIRequest) (<-chan api.APIResponse, error) {
	switch request.Type {
	case api.API_QUERY:
		return kv.runQuery(request)
	case api.API_REPLICATE:
		return kv.replicate(request)
	case api.API_REFLECT:
		return kv.reflect(request)
	default:
		return nil, fmt.Errorf("Unknown request.Type: %v", request.Type)
	}
}

func (kv *keyValueStore) writeResponse(respch chan<- api.APIResponse, resp api.APIResponse) {
	select {
	case <-kv.stopch:
		return
	case respch <- resp:
		return
	}
}

func (kv *keyValueStore) enqueue(kvq api.KvQuery) {
	log.Info("Enqueing request...")
	kv.fork(func() {
		err := kv.queue.Enqueue(kvq.Request, kvq)
		if err != nil {
			log.Error("Failed to enqueue request: %s", err.Error())
			fail := api.RESPONSE_FAIL
			fail.Err = err
			go kv.writeResponse(kvq.Response, fail)
		}
	})
}

func (kv *keyValueStore) replicate(request api.APIRequest) (<-chan api.APIResponse, error) {
	log.Info("api.APIService Replicating...")
	kvq := api.MakeKvReplicate(request)
	kv.enqueue(kvq)

	return kvq.Response, nil
}

func (kv *keyValueStore) runQuery(request api.APIRequest) (<-chan api.APIResponse, error) {
	query := request.Query
	if log.CanLog(log.LOG_INFO) {
		text, err := query.PrettyText()
		if err == nil {
			log.Info("api.APIService running query.Query:\n%s", text)
		} else {
			log.Debug("Failed to pretty print query: %s", err.Error())
		}

	}

	if err := query.Validate(); err != nil {
		log.Warn("Invalid query.Query")
		return nil, err
	}
	kvq := api.MakeKvQuery(request)

	kv.enqueue(kvq)

	return kvq.Response, nil
}

func (kv *keyValueStore) reflect(request api.APIRequest) (<-chan api.APIResponse, error) {
	log.Info("api.APIService running reflect request...")
	kvq := api.MakeKvReflect(request)

	kv.enqueue(kvq)

	return kvq.Response, nil
}

func (kv *keyValueStore) CloseAPI() {
	close(kv.stopch)
	kv.namespace.Close()
	kv.queue.Close()
	log.Info("API closed")
}
