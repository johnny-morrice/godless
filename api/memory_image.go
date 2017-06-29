package api

import (
	"github.com/johnny-morrice/godless/crdt"
)

type MemoryImage interface {
	JoinIndex(index crdt.Index) error
	GetIndex() (crdt.Index, error)
	CloseMemoryImage() error
}
