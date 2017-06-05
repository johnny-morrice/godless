package godless

import "math/rand"

func genIndex(rand *rand.Rand, size int) RemoteNamespaceIndex {
	index := EmptyRemoteNamespaceIndex()
	const ADDR_SCALE = 1
	const KEY_SCALE = 0.5
	const PATH_SCALE = 0.5

	for i := 0; i < size; i++ {
		keyCount := genCountRange(rand, 1, size, KEY_SCALE)
		indexKey := TableName(randPoint(rand, keyCount))
		addrCount := genCountRange(rand, 1, size, ADDR_SCALE)
		addrs := make([]RemoteStoreAddress, addrCount)
		for j := 0; j < addrCount; j++ {
			pathCount := genCountRange(rand, 1, size, PATH_SCALE)
			a := randPoint(rand, pathCount)
			addrs[j] = IPFSPath(a)
		}

		index.Index[indexKey] = addrs
	}

	return index
}
