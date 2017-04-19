package godless

type NamespaceTree interface {
	NamespaceLeaf() *Namespace
	JoinTable(string, Table) error
	LoadTraverse(NamespaceTreeReader) error
}

type NamespaceTreeReader interface {
	ReadNamespace(*Namespace) (bool, error)
}

// NamespaceTreeReader functions return true when they have finished reading
// the tree.
type NamespaceTreeLambda func(ns *Namespace) (bool, error)

func (ntl NamespaceTreeLambda) ReadNamespace(ns *Namespace) (bool, error) {
	f := (func(ns *Namespace) (bool, error))(ntl)
	return f(ns)
}
