package api

import "github.com/johnny-morrice/godless/crdt"

type HeadCache interface {
	SetHead(crdt.IPFSPath)
	GetHead(crdt.IPFSPath)
}

type RequestPriorityQueue interface {
}
