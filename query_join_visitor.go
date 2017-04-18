package godless

import (
	"fmt"
	"github.com/pkg/errors"
)

// Specific to IPFS.  We can add more visitors later.
type QueryJoinVisitor struct {
	noSelectVisitor
	noDebugVisitor
	errorCollectVisitor
	Namespace RemoteNamespaceTree
	tableKey string
	table Table
}

func MakeQueryJoinVisitor(ns *IpfsNamespace) *QueryJoinVisitor {
	return &QueryJoinVisitor{Namespace: ns}
}

func (visitor *QueryJoinVisitor) RunQuery(kv KvQuery) {
	if visitor.hasError() {
		visitor.reportError(kv)
		return
	}

	if visitor.tableKey == "" {
		panic("Expected table key")
	}

	err := visitor.Namespace.JoinTable(visitor.tableKey, visitor.table)

	if err != nil {
		kv.reportError(errors.Wrap(err, "Run query failed"))
		return
	}

	kv.reportSuccess(&RESPONSE_OK)
}

func (visitor *QueryJoinVisitor) VisitOpCode(opCode QueryOpCode) {
	if opCode != JOIN {
		panic("Expected JOIN OpCode")
	}
}

func (visitor *QueryJoinVisitor) VisitTableKey(tableKey string) {
	if visitor.hasError() {
		return
	}

	visitor.tableKey = tableKey
}

func (visitor *QueryJoinVisitor) VisitJoin(*QueryJoin) {
}

func (visitor *QueryJoinVisitor) VisitRowJoin(position int, rowJoin *QueryRowJoin) {
	if visitor.hasError() {
		return
	}

	row := Row{Entries: map[string][]string{}}

	for k, entry := range rowJoin.Entries {
		row.Entries[k] = []string{entry}
	}

	joined, err := visitor.table.JoinRow(rowJoin.RowKey, row)

	if err != nil {
		visitor.collectError(errors.Wrap(err, fmt.Sprintf("VisitRowJoin failed at position %v", position)))
		return
	}

	visitor.table = joined
}
