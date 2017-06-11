package crdt

import "sort"

type RemoteStoreAddress interface {
	Path() string
}

type IPFSPath string

func (path IPFSPath) Path() string {
	return string(path)
}

type byPath []RemoteStoreAddress

func (addrs byPath) Len() int {
	return len(addrs)
}

func (addrs byPath) Swap(i, j int) {
	addrs[i], addrs[j] = addrs[j], addrs[i]
}

func (addrs byPath) Less(i, j int) bool {
	return addrs[i].Path() < addrs[j].Path()
}

func normalStoreAddress(addrs []RemoteStoreAddress) []RemoteStoreAddress {
	uniq := uniqStoreAddress(addrs)
	sort.Sort(byPath(uniq))
	return uniq
}

func uniqStoreAddress(addrs []RemoteStoreAddress) []RemoteStoreAddress {
	dedupe := map[string]RemoteStoreAddress{}

	for _, a := range addrs {
		path := a.Path()
		if _, present := dedupe[path]; !present {
			dedupe[path] = a
		}
	}

	uniq := make([]RemoteStoreAddress, len(dedupe))

	i := 0
	for _, a := range dedupe {
		uniq[i] = a
		i++
	}

	return uniq
}
