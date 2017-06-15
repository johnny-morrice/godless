package service

import (
	"fmt"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

func LaunchKeyValueStore(ns api.RemoteNamespace) (api.APIService, <-chan error) {
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
	namespace api.RemoteNamespace
	input     chan<- api.KvQuery
}

func (kv *keyValueStore) Call(request api.APIRequest) (<-chan api.APIResponse, error) {
	switch request.Type {
	case api.API_QUERY:
		return kv.runQuery(request.Query)
	case api.API_REPLICATE:
		return kv.replicate(request.Replicate)
	case api.API_REFLECT:
		return kv.reflect(request.Reflection)
	default:
		return nil, fmt.Errorf("Unknown request.Type: %v", request.Type)
	}
}

func (kv *keyValueStore) replicate(peerAddr crdt.IPFSPath) (<-chan api.APIResponse, error) {
	log.Info("api.APIService Replicating: %v", peerAddr)
	kvq := api.MakeKvReplicate(peerAddr)
	kv.input <- kvq

	return kvq.TrasactionResult, nil
}

func (kv *keyValueStore) runQuery(query *query.Query) (<-chan api.APIResponse, error) {
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

func (kv *keyValueStore) reflect(request api.APIReflectionType) (<-chan api.APIResponse, error) {
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
		err := kv.namespace.Persist()

		if err == nil {
			log.Info("API transaction OK")
			commitFailure := kv.namespace.Commit()

			if commitFailure != nil {
				log.Error("Commit failed")
			}
		} else {
			log.Error("API transaction failed: %v", err)
			log.Info("Rollback failed persist")
			rollbackFailure := kv.namespace.Rollback()

			if rollbackFailure != nil {
				log.Error("Rollback failure: %v", rollbackFailure)
			}

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
