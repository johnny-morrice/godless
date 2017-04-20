package godless

import (
	"fmt"
	"github.com/pkg/errors"
)

type NamespaceTreeJoin struct {
	noSelectVisitor
	noDebugVisitor
	errorCollectVisitor
	Namespace NamespaceTree
	tableKey string
	table Table
}

func MakeNamespaceTreeJoin(ns NamespaceTree) *NamespaceTreeJoin {
	return &NamespaceTreeJoin{Namespace: ns}
}

func (visitor *NamespaceTreeJoin) RunQuery() APIResponse {
	fail := RESPONSE_FAIL

	if visitor.hasError() {
		fail.Err = visitor.reportError()
		return fail
	}

	if visitor.tableKey == "" {
		panic("Expected table key")
	}

	err := visitor.Namespace.JoinTable(visitor.tableKey, visitor.table)

	if err != nil {
		fail.Err = errors.Wrap(err, "NamespaceTreeJoin failed")
		return fail
	}

	return RESPONSE_OK
}

func (visitor *NamespaceTreeJoin) VisitOpCode(opCode QueryOpCode) {
	if opCode != JOIN {
		visitor.collectError(errors.New("Expected JOIN OpCode"))
	}
}

func (visitor *NamespaceTreeJoin) VisitTableKey(tableKey string) {
	if visitor.hasError() {
		return
	}

	visitor.tableKey = tableKey
}

func (visitor *NamespaceTreeJoin) VisitJoin(*QueryJoin) {
}

func (visitor *NamespaceTreeJoin) VisitRowJoin(position int, rowJoin *QueryRowJoin) {
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
