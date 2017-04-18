package godless

type RemoteNamespaceTree interface {
	NamespaceLeaf() *Namespace
	JoinTable(string, Table) error
	LoadTraverse(RemoteNamespaceReader) error
}

type RemoteNamespaceReader interface {
	ReadNamespace(*Namespace) (bool, error)
}

// RemoteNamespaceReader functions return true when they have finished reading
// the tree.
type RemoteNamespaceLambda func(ns *Namespace) (bool, error)

func (rnrf RemoteNamespaceLambda) ReadNamespace(ns *Namespace) (bool, error) {
	f := (func(ns *Namespace) (bool, error))(rnrf)
	return f(ns)
}
