package godless

import (
	"fmt"
	"github.com/pkg/errors"
)

type SemiLattice interface {
	Join(other SemiLattice) (SemiLattice, error)
}

// Semi-lattice type that implements our storage
type Namespace struct {
	Objects map[string]Object
}

func (ns *Namespace) JoinNamespace(other *Namespace) (*Namespace, error) {
	out := map[string]Object{}

	for k, v := range ns.Objects {
		otherv, present := other.Objects[k]

		if present {
			joined, err := v.JoinObject(otherv)

			if err != nil {
				return nil, errors.Wrap(err, "Error in Namespace join")
			}

			out[k] = joined
		} else {
			out[k] = v
		}
	}

	return &Namespace{Objects: out}, nil
}

func (ns *Namespace) Join(other SemiLattice) (SemiLattice, error) {
	if ons, ok := other.(*Namespace); ok {
		return ns.JoinNamespace(ons)
	}

	return nil, errors.New("Expected *Namespace in Join")
}

type ObjType uint8

const (
	SET = ObjType(iota)
	MAP
)

// TODO improved type validation
type Object struct {
	Type ObjType
	Obj SemiLattice
}

func (o Object) JoinObject(other Object) (Object, error) {
	if o.Type != other.Type {
		return Object{}, fmt.Errorf("Expected Object of type '%v' but got '%v'", o.Type, other.Type)
	}

	joined, err := o.Obj.Join(other.Obj)

	if err != nil {
		return Object{}, err
	}

	out := Object{
		Type: o.Type,
		Obj: joined,
	}

	return out, nil
}

func (o Object) Join(other SemiLattice) (SemiLattice, error) {
	otherobj, ok := other.(Object)

	if !ok {
		return nil, errors.New("Expected Object in Join")
	}

	return o.JoinObject(otherobj)
}

type Set struct {
	Members []string
}

func (set Set) JoinSet(other Set) Set {
	return Set{
		Members: append(set.Members, other.Members...),
	}
}

func (set Set) Join(other SemiLattice) (SemiLattice, error) {
	if os, ok := other.(Set); ok {
		return set.JoinSet(os), nil
	}

	return nil, errors.New("Expected Set in Join")
}

type Map struct {
	Members map[string][]string
}

func (m Map) JoinMap(other Map) Map {
	out := map[string][]string{}

	for k, v := range m.Members {
		out[k] = v
	}

	for k, v := range other.Members {
		initv, present := m.Members[k]

		if present {
			out[k] = append(initv, v...)
		} else {
			out[k] = v
		}

	}

	return Map{Members: out}
}

func (m Map) Join(other SemiLattice) (SemiLattice, error) {
	if om, ok := other.(Map); ok {
		return m.JoinMap(om), nil
	}

	return nil, errors.New("Expected Map in Join")
}
