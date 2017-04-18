package godless

import (
	"github.com/pkg/errors"
)

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

func LaunchKeyValueStore(ns *IpfsNamespace) (QueryAPIService, <-chan error) {
	interact := make(chan KvQuery)
	errch := make(chan error, 1)

	kv := &keyValueStore{
		namespace: ns,
		input: interact,
	}
	go func() {
		for kvq := range interact {
			err := kv.transact(kvq)

			if err != nil {
				logerr("key value store died with: %v", err)
				errch<- errors.Wrap(err, "Key value store died")
				return
			}
		}
	}()

	return kv, errch
}


type keyValueStore struct {
	namespace *IpfsNamespace
	input chan<- KvQuery
}

func (kv *keyValueStore) RunQuery(query *Query) (<-chan APIResponse, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	kvq := MakeKvQuery(query)

	go func() {
		kv.input<- kvq
	}()

	return kvq.Response, nil
}

func (kv *keyValueStore) transact(kvq KvQuery) error {
	kvq.Query.Run(kvq, kv.namespace)

	if kv.namespace.Update.IsEmpty() {
		next, err := kv.namespace.Persist()

		if err != nil {
			return errors.Wrap(err, "KeyValueStore persist failed")
		}

		kv.namespace = next
	}

	return nil
}
