package godless

import (
	"github.com/pkg/errors"
)

type KvQuery struct {
	Query *Query
	Response chan KvResponse
}

func MakeKvQuery(query *Query) KvQuery {
	return KvQuery{
		Query: query,
		Response: make(chan KvResponse),
	}
}

func (kvq KvQuery) writeResponse(val interface{}, err error) {
	kvq.Response<- KvResponse{
		Err: err,
		Val: val,
	}
}

type KvResponse struct {
	Err error
	// Figure out a proper interface type for this.
	Val interface{}
}

func LaunchKeyValueStore(ns *IpfsNamespace) (chan<-KvQuery, <-chan error) {
	interact := make(chan KvQuery)
	errch := make(chan error, 1)

	kv := &keyValueStore{
		Namespace: ns,
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

	return interact, errch
}


type keyValueStore struct {
	Namespace *IpfsNamespace
}

func (kv *keyValueStore) transact(kvq KvQuery) error {
	kvq.Query.Run(kvq, kv.Namespace)

	if kv.Namespace.dirty {
		next, err := kv.Namespace.Persist()

		if err != nil {
			return errors.Wrap(err, "KeyValueStore persist failed")
		}

		kv.Namespace = next
	}

	return nil
}
