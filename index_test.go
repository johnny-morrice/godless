package godless

import (
	"math/rand"
	"reflect"
	"testing/quick"
)

func genIndex(rand *rand.Rand, size int) RemoteNamespaceIndex {
	mapType := reflect.TypeOf(map[string][]string{})
	value, ok := quick.Value(mapType, rand)

	if !ok {
		panic("Could not generate index")
	}

	textMap := value.Interface().(map[string][]string)

	index := EmptyRemoteNamespaceIndex()

	for k, vs := range textMap {
		addrs := make([]RemoteStoreAddress, len(vs))
		for i, v := range vs {
			addrs[i] = IPFSPath(v)
		}

		indexKey := TableName(k)
		index.Index[indexKey] = addrs
	}

	return index
}
