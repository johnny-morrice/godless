package godless

import (
	"crypto/sha256"
	"github.com/pkg/errors"
)

type SemiLattice interface {
	Join(other SemiLattice) (SemiLattice, error)
}

// Semi-lattice type that implements our storage
type Namespace struct {
	Objects map[string]Object
}

func NewNamespace() *Namespace {
	return &Namespace{
		Objects: map[string]Object{},
	}
}

func (ns *Namespace) JoinNamespace(other *Namespace) (*Namespace, error) {
	build := map[string]Object{}

	for k, v := range ns.Objects {
		build[k] = v
	}

	for k, otherv := range other.Objects {
		v, present := ns.Objects[k]

		if present {
			joined, err := v.JoinObject(otherv)

			if err != nil {
				return nil, errors.Wrap(err, "Error in Namespace join")
			}

			build[k] = joined
		} else {
			build[k] = otherv
		}
	}

	return &Namespace{Objects: build}, nil
}

func (ns *Namespace) Join(other SemiLattice) (SemiLattice, error) {
	if ons, ok := other.(*Namespace); ok {
		return ns.JoinNamespace(ons)
	}

	return nil, errors.New("Expected *Namespace in Join")
}

func (ns *Namespace) JoinObject(key string, obj Object) error {
	joined, err := obj.JoinObject(ns.Objects[key])

	if err != nil {
		return errors.Wrap(err, "Namespace JoinObject failed")
	}

	ns.Objects[key] = joined

	return nil
}

// type ObjType uint8
//
// const (
// 	SET = ObjType(iota)
// 	MAP
// )

// TODO improved type validation
type Object struct {
	// Type ObjType
	Obj SemiLattice
}

func (o Object) JoinObject(other Object) (Object, error) {
	// if o.Type != other.Type {
	// 	return Object{}, fmt.Errorf("Expected Object of type '%v' but got '%v'", o.Type, other.Type)
	// }

	// Zero value makes join easy
	zero := Object{}
	if other.Obj == zero {
		return o, nil
	}

	joined, err := o.Obj.Join(other.Obj)

	if err != nil {
		return Object{}, err
	}

	out := Object{
		// Type: o.Type,
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
	// Handle zero value
	if len(other.Members) == 0 {
		return set
	}

	build := Set{
		Members: append(set.Members, other.Members...),
	}
	return build.uniq()
}

func (set Set) Join(other SemiLattice) (SemiLattice, error) {
	if os, ok := other.(Set); ok {
		return set.JoinSet(os), nil
	}

	return nil, errors.New("Expected Set in Join")
}

func (set Set) uniq() Set {
	return Set{Members: uniq256(set.Members)}
}

type Map struct {
	Members map[string][]string
}

func (m Map) JoinMap(other Map) Map {
	// Handle zero value
	if len(other.Members) == 0 {
		return m
	}

	build := map[string][]string{}

	for k, v := range m.Members {
		build[k] = v
	}

	for k, v := range other.Members {
		initv, present := m.Members[k]

		if present {
			build[k] = append(initv, v...)
		} else {
			build[k] = v
		}

	}

	ret := Map{Members: build}
	return ret.uniq()
}

func (m Map) Join(other SemiLattice) (SemiLattice, error) {
	if om, ok := other.(Map); ok {
		return m.JoinMap(om), nil
	}

	return nil, errors.New("Expected Map in Join")
}

func (m Map) uniq() Map {
	build := map[string][]string{}
	for k, vs := range m.Members {
		build[k] = uniq256(vs)
	}

	return Map{Members: build}
}

// uniq256 deduplicates a slice of strings using sha256.
func uniq256(dups []string) []string {
	dedup := map[[sha256.Size]byte]string{}

	for _, s := range dups {
		bs := []byte(s)
		k := sha256.Sum256(bs)
		if _, present := dedup[k]; !present {
			dedup[k] = s
		}
	}

	out := []string{}

	for _, v := range dedup {
		out = append(out, v)
	}

	return out
}
