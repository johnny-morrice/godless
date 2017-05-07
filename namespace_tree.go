package godless

//go:generate mockgen -destination mock/mock_namespace_tree.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless NamespaceTree,NamespaceTreeTableReader

type NamespaceTree interface {
	JoinTable(string, Table) error
	LoadTraverse(NamespaceTreeTableReader) error
}

type KvNamespaceTree interface {
	KvNamespace
	NamespaceTree
}

type NamespaceTreeReader interface {
	ReadNamespace(Namespace) (bool, error)
}

type TableHinter interface {
	ReadsTables() []string
}

type NamespaceTreeTableReader interface {
	TableHinter
	NamespaceTreeReader
}

func AddTableHints(tables []string, ntr NamespaceTreeReader) NamespaceTreeTableReader {
	return tableHinterWrapper{
		hints:  tables,
		reader: ntr,
	}
}

type tableHinterWrapper struct {
	reader NamespaceTreeReader
	hints  []string
}

func (thw tableHinterWrapper) ReadsTables() []string {
	return thw.hints
}

func (thw tableHinterWrapper) ReadNamespace(ns Namespace) (bool, error) {
	return thw.reader.ReadNamespace(ns)
}

// NamespaceTreeReader functions return true when they have finished reading
// the tree.
type NamespaceTreeLambda func(ns Namespace) (bool, error)

func (ntl NamespaceTreeLambda) ReadNamespace(ns Namespace) (bool, error) {
	f := (func(ns Namespace) (bool, error))(ntl)
	return f(ns)
}
