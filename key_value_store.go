package godless

//go:generate mockgen -destination mock/mock_key_value_store.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless KvNamespace

import (
	"github.com/pkg/errors"
)

type KvNamespace interface {
	RunKvQuery(KvQuery)
	IsChanged() bool
	Persist() (KvNamespace, error)
}

type KvQuery struct {
	Query *Query
	Response chan APIResponse
}

func MakeKvQuery(query *Query) KvQuery {
	return KvQuery{
		Query: query,
		Response: make(chan APIResponse),
	}
}

func (kvq KvQuery) writeResponse(val APIResponse) {
	kvq.Response<- val
}

func (kvq KvQuery) reportError(err error) {
	kvq.writeResponse(APIResponse{Err: err})
}

func (kvq KvQuery) reportSuccess(val APIResponse) {
	kvq.writeResponse(val)
}

func LaunchKeyValueStore(ns KvNamespace) (QueryAPIService, <-chan error) {
	interact := make(chan KvQuery)
	errch := make(chan error, 1)

	kv := &keyValueStore{
		namespace: ns,
		input: interact,
	}
	go func() {
		for kvq := range interact {
			logdbg("interacting...")
			err := kv.transact(kvq)

			if err != nil {
				close(kvq.Response)
				logerr("key value store died with: %v", err)
				errch<- errors.Wrap(err, "Key value store died")
				return
			}
		}

		close(errch)
	}()

	return kv, errch
}


type keyValueStore struct {
	namespace KvNamespace
	input chan<- KvQuery
}

func (kv *keyValueStore) RunQuery(query *Query) (<-chan APIResponse, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	kvq := MakeKvQuery(query)

	kv.input<- kvq

	return kvq.Response, nil
}

func (kv *keyValueStore) Close() {
	close(kv.input)
}

func (kv *keyValueStore) transact(kvq KvQuery) error {
	kv.namespace.RunKvQuery(kvq)

	if kv.namespace.IsChanged() {
		next, err := kv.namespace.Persist()

		if err != nil {
			return errors.Wrap(err, "KeyValueStore persist failed")
		}

		kv.namespace = next
	}

	return nil
}
