package godless

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
)

type kvTransform func(*IpfsNamespace)

type KvOpCode uint8

const (
	GET_SET = KvOpCode(iota)
	JOIN_SET
	GET_MAP
	GET_MAP_VALUES
	JOIN_MAP
)

type KvQuery struct {
	OpCode KvOpCode
	Val interface{}
	Response chan KvResponse
}

type KvSetJoin struct {
	NamespaceKey string
	Values []string
}

type KvSetQuery struct {
	NamespaceKey string
}

type KvMapJoin struct {
	NamespaceKey string
	Values map[string][]string
}

type KvMapQuery struct {
	NamespaceKey string
}

type KvMapValuesQuery struct {
	NamespaceKey string
	MapKey string
}

func (kvq KvQuery) transform() kvTransform {
	switch kvq.OpCode {
	case GET_SET:
		return kvq.getSet
	case JOIN_SET:
		return kvq.joinSet
	case GET_MAP:
		return kvq.getMap
	case GET_MAP_VALUES:
		return kvq.getMapValues
	case JOIN_MAP:
		return kvq.joinMap
	default:
		panic(fmt.Sprintf("BUG Unknown KvOpCode: %v", kvq.OpCode))
	}
}

func (kvq KvQuery) getSet(ns *IpfsNamespace) {
	sq, ok := kvq.Val.(KvSetQuery)

	if !ok {
		panic("BUG expected KvSetQuery")
	}

	kvq.writeResponse(ns.GetSet(sq.NamespaceKey))
}

func (kvq KvQuery) joinSet(ns *IpfsNamespace) {
	sj, ok := kvq.Val.(KvSetJoin)

	if !ok {
		panic("BUG expected KvSetJoin")
	}

	err := ns.JoinSet(sj.NamespaceKey, sj.Values)
	kvq.writeResponse(nil, err)
}

func (kvq KvQuery) getMap(ns *IpfsNamespace) {
	mq, ok := kvq.Val.(KvMapQuery)

	if !ok {
		panic("BUG expected KvMapQuery")
	}

	kvq.writeResponse(ns.GetMap(mq.NamespaceKey))
}

func (kvq KvQuery) joinMap(ns *IpfsNamespace) {
	mj, ok := kvq.Val.(KvMapJoin)

	if !ok {
		panic("BUG expected KvMapJoin")
	}

	err := ns.JoinMap(mj.NamespaceKey, mj.Values)
	kvq.writeResponse(nil, err)
}

func (kvq KvQuery) getMapValues(ns *IpfsNamespace) {
	mvq, ok := kvq.Val.(KvMapValuesQuery)

	if !ok {
		panic("BUG expected KvMapValuesQuery")
	}

	kvq.writeResponse(ns.GetMapValues(mvq.NamespaceKey, mvq.MapKey))
}

func (kvq KvQuery) writeResponse(val interface{}, err error) {
	kvq.Response<- KvResponse{
		Err: err,
		Val: val,
	}
}

type KvResponse struct {
	Err error
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
			err := kv.transact(kvq.transform())

			if err != nil {
				log.Printf("ERROR key value store died with: %v")
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

func (kv *keyValueStore) transact(f kvTransform) error {
	f(kv.Namespace)

	if kv.Namespace.dirty {
		next, err := kv.Namespace.Persist()

		if err != nil {
			return errors.Wrap(err, "KeyValueStore persist failed")
		}

		kv.Namespace = next
	}

	return nil
}
