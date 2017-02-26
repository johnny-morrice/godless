package godless

type SemiLattice interface {
	Join(other SemiLattice) SemiLattice
}

// Semi-lattice type that implements our storage
type Namespace struct {
	Objects map[string]Object
}

func (ns *Namespace) JoinNamespace(other *Namespace) *Namespace {
	out := map[string]Object{}

	for k, v := range ns.Objects {
		otherv, present := other.Objects[k]

		if present {
			out[k] = v.JoinObject(otherv)
		} else {
			out[k] = v
		}
	}

	return &Namespace{Objects: out}
}

func (ns *Namespace) Join(other SemiLattice) SemiLattice {
	if ons, ok := other.(*Namespace); ok {
		return ns.JoinNamespace(ons)
	}

	panic("expected Set")
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

func (o Object) JoinObject(other Object) Object {
	if o.Type != other.Type {
		panic("Unexpected Object types")
	}

	return Object{
		Type: o.Type,
		Obj: o.Obj.Join(other.Obj),
	}
}

func (o Object) Join(other SemiLattice) SemiLattice {
	otherobj, ok := other.(Object)

	if !ok {
		panic("Expected Object")
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

func (set Set) Join(other SemiLattice) SemiLattice {
	if os, ok := other.(Set); ok {
		return set.JoinSet(os)
	}

	panic("expected Set")
}

// TODO ensure is always grow-only
type Map struct {
	Members map[string]string
}

func (m Map) JoinMap(other Map) Map {
	out := map[string]string{}

	for k, v := range m.Members {
		out[k] = v
	}

	for k, v := range other.Members {
		out[k] = v
	}

	return Map{Members: out}
}

func (m Map) Join(other SemiLattice) SemiLattice {
	if om, ok := other.(Map); ok {
		return m.JoinMap(om)
	}

	panic("expected Map")
}
