package service

import (
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

func LaunchKeyValueStore(ns api.KvNamespace) (api.APIService, <-chan error) {
	interact := make(chan api.KvQuery)
	errch := make(chan error, 1)

	kv := &keyValueStore{
		namespace: ns,
		input:     interact,
	}
	go func() {
		defer close(errch)
		for kvq := range interact {
			log.Info("API received new request")

			err := kv.transact(kvq)

			if err != nil {
				log.Error("key value store died with: %v", err)
				errch <- errors.Wrap(err, "Key value store died")
				return
			}
		}
	}()

	return kv, errch
}

type keyValueStore struct {
	namespace api.KvNamespace
	input     chan<- api.KvQuery
}

func (kv *keyValueStore) Replicate(peerAddr crdt.RemoteStoreAddress) (<-chan api.APIResponse, error) {
	log.Info("api.APIService Replicating: %v", peerAddr)
	kvq := api.MakeKvReplicate(peerAddr)
	kv.input <- kvq

	return kvq.TrasactionResult, nil
}

func (kv *keyValueStore) RunQuery(query *query.Query) (<-chan api.APIResponse, error) {
	if log.CanLog(log.LOG_INFO) {
		text, err := query.PrettyText()
		if err == nil {
			log.Info("api.APIService running query.Query:\n%v", text)
		} else {
			log.Debug("Failed to pretty print query: %v", err)
		}

	}

	if err := query.Validate(); err != nil {
		log.Warn("Invalid query.Query")
		return nil, err
	}
	kvq := api.MakeKvQuery(query)

	kv.input <- kvq

	return kvq.TrasactionResult, nil
}

func (kv *keyValueStore) Reflect(request api.APIReflectionType) (<-chan api.APIResponse, error) {
	log.Info("api.APIService running reflect request: %v", request)
	kvq := api.MakeKvReflect(request)

	kv.input <- kvq

	return kvq.TrasactionResult, nil
}

func (kv *keyValueStore) CloseAPI() {
	close(kv.input)
}

func (kv *keyValueStore) transact(kvq api.KvQuery) error {
	go kvq.Run(kv.namespace)

	resp := <-kvq.Response
	if kv.namespace.IsChanged() {
		next, err := kv.namespace.Persist()

		if err == nil {
			log.Info("API transaction OK")
			kv.namespace = next
		} else {
			log.Error("API transaction failed: %v", err)
			log.Info("Rollback failed persist")
			kv.namespace.Reset()

			respText, reportErr := resp.AsText()

			if reportErr != nil {
				log.Info("Overridden response:\n%v", respText)
			} else {
				log.Warn("Could not serialize overriden response: %v", reportErr)
			}

			respType := resp.Type
			resp = api.RESPONSE_FAIL
			resp.Type = respType
		}
	} else {
		log.Info("No API transaction required for read query")
	}

	kvq.TrasactionResult <- resp

	return nil
}
