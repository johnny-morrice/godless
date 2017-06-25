package api

import (
	"github.com/johnny-morrice/godless/crdt"
)

type MemoryImage interface {
	PushIndex(index crdt.Index) error
	ForeachIndex(func(index crdt.Index)) error
	JoinAllIndices() (crdt.Index, error)
}
