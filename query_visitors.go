package godless

import (
	"fmt"
	"io"
)

// Visitor outline to help with editor macros.
//
// VisitAST(*QueryAST)
// VisitParser(*QueryParser)
// VisitOpCode(QueryOpCode)
// VisitTableKey(TableName)
// VisitJoin(*QueryJoin)
// VisitRowJoin(int, *QueryRowJoin)
// VisitSelect(*QuerySelect)
// VisitWhere(int, *QueryWhere)
// LeaveWhere(*QueryWhere)
// VisitPredicate(*QueryPredicate)

type queryPrinter struct {
	noDebugVisitor
	errorCollectVisitor
	output io.Writer
}

func (printer *queryPrinter) VisitWhere(int, *QueryWhere) {
}

func (printer *queryPrinter) LeaveWhere(*QueryWhere) {
}

func (printer *queryPrinter) VisitPredicate(*QueryPredicate) {
}

func (printer *queryPrinter) VisitOpCode(opCode QueryOpCode) {
	switch opCode {
	case SELECT:
		printer.write("select")
	case JOIN:
		printer.write("join")
	default:
		printer.collectError(fmt.Errorf("Unknown "))
	}
}

func (printer *queryPrinter) VisitTableKey(TableName) {
}

func (printer *queryPrinter) VisitJoin(*QueryJoin) {
}

func (printer *queryPrinter) VisitRowJoin(int, *QueryRowJoin) {
}

func (printer *queryPrinter) VisitSelect(*QuerySelect) {
}

func (printer *queryPrinter) write(token interface{}) {

}

type queryFlattener struct {
	noDebugVisitor
	errorCollectVisitor

	tableName  TableName
	opCode     QueryOpCode
	join       QueryJoin
	slct       QuerySelect
	allClauses []QueryWhere
}

func (visitor *queryFlattener) VisitOpCode(opCode QueryOpCode) {
	visitor.opCode = opCode
}

func (visitor *queryFlattener) VisitTableKey(tableKey TableName) {
	visitor.tableName = tableKey
}

func (visitor *queryFlattener) VisitSelect(slct *QuerySelect) {
	visitor.slct = *slct
}

func (visitor *queryFlattener) VisitJoin(join *QueryJoin) {
	visitor.join = *join
}

func (visitor *queryFlattener) VisitRowJoin(int, *QueryRowJoin) {
}

func (visitor *queryFlattener) VisitWhere(position int, where *QueryWhere) {
	visitor.allClauses = append(visitor.allClauses, *where)
}

func (visitor *queryFlattener) LeaveWhere(*QueryWhere) {
}

func (visitor *queryFlattener) VisitPredicate(*QueryPredicate) {
}

func (visitor *queryFlattener) Equals(other *queryFlattener) bool {
	ok := visitor.opCode == other.opCode
	ok = ok && visitor.tableName == other.tableName
	ok = ok && visitor.slct.Limit == other.slct.Limit
	ok = ok && len(visitor.allClauses) == len(other.allClauses)

	if !ok {
		return false
	}

	if !visitor.join.equals(other.join) {
		return false
	}

	for i, myClause := range visitor.allClauses {
		theirClause := other.allClauses[i]

		if !myClause.shallowEquals(theirClause) {
			return false
		}
	}

	return true
}

type queryValidator struct {
	noDebugVisitor
	errorCollectVisitor
	noJoinVisitor
}

func (visitor *queryValidator) VisitOpCode(opCode QueryOpCode) {
	switch opCode {
	case SELECT:
	case JOIN:
		// Okay!
	default:
		visitor.badOpcode(opCode)
	}
}

func (visitor *queryValidator) VisitTableKey(tableKey TableName) {
	if tableKey == "" {
		visitor.badTableName(tableKey)
	}
}

func (visitor *queryValidator) VisitSelect(*QuerySelect) {
}

func (visitor *queryValidator) VisitWhere(position int, where *QueryWhere) {
	switch where.OpCode {
	case WHERE_NOOP:
	case AND:
	case OR:
	case PREDICATE:
		// Okay!
	default:
		visitor.badWhereOpCode(position, where)
	}
}

func (visitor *queryValidator) LeaveWhere(*QueryWhere) {
}

func (visitor *queryValidator) VisitPredicate(predicate *QueryPredicate) {
	switch predicate.OpCode {
	case PREDICATE_NOP:
	case STR_EQ:
	case STR_NEQ:
		// Okay!
	default:
		visitor.badPredicateOpCode(predicate)
	}
}
