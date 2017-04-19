package godless

import (
	"github.com/pkg/errors"
)

type remoteNamespace struct {
	loaded bool
	Update *Namespace

	Store RemoteStore
	Index RemoteStoreIndex
	// Breaking 12 factor rule in caching namespace...
	Namespace *Namespace
	Children []*remoteNamespace
}

func LoadRemoteNamespace(Store RemoteStore, Index RemoteStoreIndex) (KvNamespace, error) {
	ns := &remoteNamespace{}
	ns.Store = Store
	ns.Index = Index
	ns.Update = MakeNamespace()
	ns.Namespace = MakeNamespace()
	ns.Children = []*remoteNamespace{}

	err := ns.load()

	if err != nil {
		return nil, errors.Wrap(err, "Error loading new namespace")
	}

	return ns, nil
}

func PersistNewRemoteNamespace(Store RemoteStore, namespace *Namespace) (KvNamespace, error) {
	ns := &remoteNamespace{}
	ns.Store = Store
	ns.Update = namespace
	ns.Namespace = MakeNamespace()
	ns.Children = []*remoteNamespace{}

	return ns.Persist()
}

func (ns *remoteNamespace) RunKvQuery(kvq KvQuery) {
	query := kvq.Query
	var runner QueryRun

	switch query.OpCode {
	case JOIN:
		visitor := MakeQueryJoinVisitor(ns)
		query.Visit(visitor)
		runner = visitor
	case SELECT:
		visitor := MakeQuerySelectVisitor(ns)
		query.Visit(visitor)
		runner = visitor
	default:
		query.opcodePanic()
	}

	runner.RunQuery(kvq)
}

func (ns *remoteNamespace) IsChanged() bool {
	return !ns.Namespace.IsEmpty()
}

func (ns *remoteNamespace) JoinTable(tableKey string, table Table) error {
	joined, joinerr := ns.Update.JoinTable(tableKey, table)

	if joinerr != nil {
		return errors.Wrap(joinerr, "remoteNamespace.JoinTable failed")
	}

	ns.Update = joined

	return nil
}

func (ns *remoteNamespace) NamespaceLeaf() *Namespace {
	return ns.Namespace
}

func (ns *remoteNamespace) LoadTraverse(f NamespaceTreeReader) error {
	stack := make([]*remoteNamespace, 1)
	stack[0] = ns

	for i := 0 ; i < len(stack); i++ {
		current := stack[i]
		err := current.load()

		if err != nil {
			return errors.Wrap(err, "Error in remoteNamespace loadTraverse")
		}

		leaf := current.NamespaceLeaf()
		abort, visiterr := f.ReadNamespace(leaf)

		if visiterr != nil {
			return errors.Wrap(err, "Error in remoteNamespace traversal")
		}

		if abort {
			return nil
		}

		stack = append(stack, current.Children...)
	}

	return nil
}

// Load chunks over IPFS
// TODO opportunity to query IPFS in parallel?
func (ns *remoteNamespace) load() error {
	if ns.Index == "" {
		logdie("tried to load remoteNamespace with empty Index")
	}

	if ns.loaded {
		logwarn("remoteNamespace already loaded from: '%v'", ns.Index)
		return nil
	}

	part, err := ns.Store.Cat(ns.Index)

	if err != nil {
		return errors.Wrap(err, "Error in remoteNamespace Cat")
	}

	ns.Namespace = part.Namespace
	ns.Children = make([]*remoteNamespace, len(part.Children))

	for i, file := range part.Children {
		child := &remoteNamespace{}
		child.Index = file
		child.Store = ns.Store
		ns.Children[i] = child
	}

	ns.loaded = true
	return nil
}

// Write pending changes to IPFS and return the new parent namespace.
func (ns *remoteNamespace) Persist() (KvNamespace, error) {
	part := RemoteNamespaceRecord{}
	part.Namespace = ns.Update

	// If this is the first namespace in the chain, don't save children.
	// TODO become parent of multiple children.
	if ns.Index != "" {
		part.Children[0] = []RemoteStoreIndex{ns.Index}
	}

	addr, err := ns.Store.Add(part)

	if err != nil {
		return nil, errors.Wrap(err, "Error adding remoteNamespace to Store")
	}

	logdbg("Persisted Namespace at: %v", addr)

	out := &remoteNamespace{}
	out.loaded = true
	out.Store = ns.Store
	out.Index = addr
	out.Namespace = ns.Update
	out.Update = MakeNamespace()
	out.Children = []*remoteNamespace{ns}

	return out, nil
}