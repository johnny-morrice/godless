package godless

import (
	"fmt"
	"github.com/pkg/errors"
)

type QueryJoinVisitor struct {
	Namespace *IpfsNamespace
	tableKey string
	table Table
	err error
}

func (visitor *QueryJoinVisitor) RunQuery(kv KvQuery) {
	if visitor.err != nil {
		kv.reportError(visitor.err)
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

func (visitor *QueryJoinVisitor) VisitAST(*QueryAST) {
}

func (visitor *QueryJoinVisitor) VisitParser(*QueryParser) {
}

func (visitor *QueryJoinVisitor) VisitTableKey(tableKey string) {
	if visitor.err != nil {
		return
	}

	visitor.tableKey = tableKey
}

func (visitor *QueryJoinVisitor) VisitJoin(*QueryJoin) {
}

func (visitor *QueryJoinVisitor) VisitRowJoin(position int, rowJoin *QueryRowJoin) {
	if visitor.err != nil {
		return
	}

	row := Row{Entries: map[string][]string{}}

	for k, entry := range rowJoin.Values {
		row.Entries[k] = []string{entry}
	}

	joined, err := visitor.table.JoinRow(rowJoin.RowKey, row)

	if err != nil {
		visitor.err = errors.Wrap(err, fmt.Sprintf("VisitRowJoin failed at position %v", position))
		return
	}

	visitor.table = joined
}

func (visitor *QueryJoinVisitor) VisitSelect(*QuerySelect) {
}

func (visitor *QueryJoinVisitor) VisitWhere(int, *QueryWhere) {
}

func (visitor *QueryJoinVisitor) VisitWhereOpCode(QueryWhereOpCode) {
}

func (visitor *QueryJoinVisitor) LeaveWhere(*QueryWhere) {
}

func (visitor *QueryJoinVisitor) VisitPredicate(*QueryPredicate) {
}
