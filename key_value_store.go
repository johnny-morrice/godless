package godless

//go:generate mockgen -destination mock/mock_key_value_store.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless KvNamespace

import (
	"github.com/pkg/errors"
)

type KvNamespace interface {
	RunKvQuery(*Query, KvQuery)
	RunKvReflection(APIReflectRequest, KvQuery)
	Replicate(RemoteStoreAddress, KvQuery)
	IsChanged() bool
	Persist() (KvNamespace, error)
	Reset()
}

type kvRunner interface {
	Run(KvNamespace, KvQuery)
}

type kvReplicator struct {
	peerAddr RemoteStoreAddress
}

func (replicator kvReplicator) Run(kvn KvNamespace, kvq KvQuery) {
	kvn.Replicate(replicator.peerAddr, kvq)
}

type kvQueryRunner struct {
	query *Query
}

func (kqr kvQueryRunner) Run(kvn KvNamespace, kvq KvQuery) {
	kvn.RunKvQuery(kqr.query, kvq)
}

type kvReflectRunner struct {
	reflection APIReflectRequest
}

func (krr kvReflectRunner) Run(kvn KvNamespace, kvq KvQuery) {
	kvn.RunKvReflection(krr.reflection, kvq)
}

type KvQuery struct {
	runner            kvRunner
	Response          chan APIResponse
	transactionResult chan APIResponse
}

func makeApiQuery(runner kvRunner) KvQuery {
	return KvQuery{
		runner:            runner,
		Response:          make(chan APIResponse),
		transactionResult: make(chan APIResponse),
	}
}

func MakeKvQuery(query *Query) KvQuery {
	return makeApiQuery(kvQueryRunner{query: query})
}

func MakeKvReflect(request APIReflectRequest) KvQuery {
	return makeApiQuery(kvReflectRunner{reflection: request})
}

func MakeKvReplicate(peerAddr RemoteStoreAddress) KvQuery {
	return makeApiQuery(kvReplicator{peerAddr: peerAddr})
}

// TODO these should make more general sense and be public.
func (kvq KvQuery) writeResponse(val APIResponse) {
	kvq.Response <- val
}

func (kvq KvQuery) Error(err error) {
	kvq.writeResponse(APIResponse{Err: err})
}

func (kvq KvQuery) reportSuccess(val APIResponse) {
	kvq.writeResponse(val)
}

func (kvq KvQuery) Run(kvn KvNamespace) {
	kvq.runner.Run(kvn, kvq)
}

func LaunchKeyValueStore(ns KvNamespace) (APIService, <-chan error) {
	interact := make(chan KvQuery)
	errch := make(chan error, 1)

	kv := &keyValueStore{
		namespace: ns,
		input:     interact,
	}
	go func() {
		defer close(errch)
		for kvq := range interact {
			loginfo("API received new request")

			err := kv.transact(kvq)

			if err != nil {
				logerr("key value store died with: %v", err)
				errch <- errors.Wrap(err, "Key value store died")
				return
			}
		}
	}()

	return kv, errch
}

type keyValueStore struct {
	namespace KvNamespace
	input     chan<- KvQuery
}

func (kv *keyValueStore) Replicate(peerAddr RemoteStoreAddress) (<-chan APIResponse, error) {
	loginfo("APIService Replicating: %v", peerAddr)
	kvq := MakeKvReplicate(peerAddr)
	kv.input <- kvq

	return kvq.transactionResult, nil
}

func (kv *keyValueStore) RunQuery(query *Query) (<-chan APIResponse, error) {
	if canLog(LOG_INFO) {
		text, err := query.PrettyText()
		if err == nil {
			loginfo("APIService running Query:\n%v", text)
		} else {
			logdbg("Failed to pretty print query: %v", err)
		}

	}

	if err := query.Validate(); err != nil {
		logwarn("Invalid Query")
		return nil, err
	}
	kvq := MakeKvQuery(query)

	kv.input <- kvq

	return kvq.transactionResult, nil
}

func (kv *keyValueStore) Reflect(request APIReflectRequest) (<-chan APIResponse, error) {
	loginfo("APIService running reflect request: %v", request)
	kvq := MakeKvReflect(request)

	kv.input <- kvq

	return kvq.transactionResult, nil
}

func (kv *keyValueStore) CloseAPI() {
	close(kv.input)
}

func (kv *keyValueStore) transact(kvq KvQuery) error {
	go kvq.Run(kv.namespace)

	resp := <-kvq.Response
	if kv.namespace.IsChanged() {
		next, err := kv.namespace.Persist()

		if err == nil {
			loginfo("API transaction OK")
			kv.namespace = next
		} else {
			logerr("API transaction failed: %v", err)
			loginfo("Rollback failed persist")
			kv.namespace.Reset()

			respText, reportErr := resp.AsText()

			if reportErr != nil {
				loginfo("Overridden response:\n%v", respText)
			} else {
				logwarn("Could not serialize overriden response: %v", reportErr)
			}

			respType := resp.Type
			resp = RESPONSE_FAIL
			resp.Type = respType
		}
	} else {
		loginfo("No API transaction required for read query")
	}

	kvq.transactionResult <- resp

	return nil
}
