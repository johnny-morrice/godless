package eval

import (
	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

type NamespaceTreeJoin struct {
	query.NoSelectVisitor
	query.NoDebugVisitor
	query.ErrorCollectVisitor
	Namespace api.NamespaceTree
	tableKey  crdt.TableName
	table     crdt.Table
}

func MakeNamespaceTreeJoin(ns api.NamespaceTree) *NamespaceTreeJoin {
	return &NamespaceTreeJoin{Namespace: ns}
}

func (visitor *NamespaceTreeJoin) RunQuery() api.APIResponse {
	fail := api.RESPONSE_FAIL
	fail.Type = api.API_QUERY

	err := visitor.Error()
	if err != nil {
		fail.Err = visitor.Error()
		return fail
	}

	if visitor.tableKey == "" {
		panic("Expected table key")
	}

	err = visitor.Namespace.JoinTable(visitor.tableKey, visitor.table)

	if err != nil {
		fail.Err = errors.Wrap(err, "NamespaceTreeJoin failed")
		return fail
	}

	return api.RESPONSE_QUERY
}

func (visitor *NamespaceTreeJoin) VisitOpCode(opCode query.QueryOpCode) {
	if opCode != query.JOIN {
		visitor.CollectError(errors.New("Expected JOIN OpCode"))
	}
}

func (visitor *NamespaceTreeJoin) VisitTableKey(tableKey crdt.TableName) {
	if visitor.Error() != nil {
		return
	}

	visitor.tableKey = tableKey
}

func (visitor *NamespaceTreeJoin) VisitJoin(*query.QueryJoin) {
}

func (visitor *NamespaceTreeJoin) LeaveJoin(*query.QueryJoin) {
}

func (visitor *NamespaceTreeJoin) VisitRowJoin(position int, rowJoin *query.QueryRowJoin) {
	if visitor.Error() != nil {
		return
	}

	row := crdt.Row{}

	for k, entryValue := range rowJoin.Entries {
		entry := crdt.MakeEntry([]crdt.Point{entryValue})
		row = row.JoinEntry(k, entry)
	}

	joined := visitor.table.JoinRow(rowJoin.RowKey, row)

	visitor.table = joined
}
