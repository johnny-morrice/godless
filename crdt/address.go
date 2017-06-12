package crdt

import "sort"

type IPFSPath string

const NIL_PATH IPFSPath = ""

func IsNilPath(path IPFSPath) bool {
	return path == NIL_PATH
}

type byPath []IPFSPath

func (addrs byPath) Len() int {
	return len(addrs)
}

func (addrs byPath) Swap(i, j int) {
	addrs[i], addrs[j] = addrs[j], addrs[i]
}

func (addrs byPath) Less(i, j int) bool {
	return addrs[i] < addrs[j]
}

func normalStoreAddress(addrs []IPFSPath) []IPFSPath {
	uniq := uniqStoreAddress(addrs)
	sort.Sort(byPath(uniq))
	return uniq
}

func uniqStoreAddress(addrs []IPFSPath) []IPFSPath {
	dedupe := map[IPFSPath]IPFSPath{}

	for _, a := range addrs {
		path := a
		if _, present := dedupe[path]; !present {
			dedupe[path] = a
		}
	}

	uniq := make([]IPFSPath, len(dedupe))

	i := 0
	for _, a := range dedupe {
		uniq[i] = a
		i++
	}

	return uniq
}
