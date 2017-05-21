package godless

//go:generate mockgen -destination mock/mock_key_value_store.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless KvNamespace

import (
	"github.com/pkg/errors"
)

type KvNamespace interface {
	RunKvQuery(*Query, KvQuery)
	RunKvReflection(APIReflectRequest, KvQuery)
	IsChanged() bool
	Persist() (KvNamespace, error)
}

type kvRunner interface {
	Run(KvNamespace, KvQuery)
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
	runner   kvRunner
	Response chan APIResponse
}

func MakeKvQuery(query *Query) KvQuery {
	return KvQuery{
		runner:   kvQueryRunner{query: query},
		Response: make(chan APIResponse),
	}
}

func MakeKvReflect(request APIReflectRequest) KvQuery {
	return KvQuery{
		runner:   kvReflectRunner{reflection: request},
		Response: make(chan APIResponse),
	}
}

// TODO these should make more general sense and be public.
func (kvq KvQuery) writeResponse(val APIResponse) {
	kvq.Response <- val
}

func (kvq KvQuery) visitError(err error) {
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

			logdbg("Key Value API received query")
			err := kv.transact(kvq)

			if err != nil {
				close(kvq.Response)
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

func (kv *keyValueStore) RunQuery(query *Query) (<-chan APIResponse, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	kvq := MakeKvQuery(query)

	kv.input <- kvq

	return kvq.Response, nil
}

func (kv *keyValueStore) Reflect(request APIReflectRequest) (<-chan APIResponse, error) {
	kvq := MakeKvReflect(request)

	kv.input <- kvq

	return kvq.Response, nil
}

func (kv *keyValueStore) CloseAPI() {
	close(kv.input)
}

func (kv *keyValueStore) transact(kvq KvQuery) error {
	kvq.Run(kv.namespace)

	if kv.namespace.IsChanged() {
		logdbg("Persisting new namespace")
		next, err := kv.namespace.Persist()

		if err != nil {
			return errors.Wrap(err, "KeyValueStore persist failed")
		}

		kv.namespace = next
	} else {
		logdbg("Namespace unchanged")
	}

	return nil
}
