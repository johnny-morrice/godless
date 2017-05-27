package godless

import (
	"fmt"
	"io"
	"sort"
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
	output    io.Writer
	tabIndent int
}

func (printer *queryPrinter) VisitOpCode(opCode QueryOpCode) {
	switch opCode {
	case SELECT:
		printer.write("select")
	case JOIN:
		printer.write("join")
	default:
		printer.collectError(fmt.Errorf("Unknown "))
		return
	}
}

func (printer *queryPrinter) VisitTableKey(table TableName) {
	printer.write(" ")
	printer.write(table)
}

func (printer *queryPrinter) VisitJoin(join *QueryJoin) {
	if join.IsEmpty() {
		return
	}

	printer.write(" rows")
	printer.indent(1)
}

func (printer *queryPrinter) LeaveJoin(join *QueryJoin) {
	printer.indent(-1)
}

func (printer *queryPrinter) VisitRowJoin(position int, row *QueryRowJoin) {
	if position > 0 {
		printer.write(",")
	}
	printer.newline()
	printer.tabs()
	// TODO shorthand key syntax for simple names.
	printer.write(" (@key")
	printer.write("=")

	printer.write("@'")
	printer.write(row.RowKey)
	printer.write("'")

	keys := make([]string, len(row.Entries))
	i := 0
	for entry, _ := range row.Entries {
		keys[i] = string(entry)
		i++
	}
	sort.Strings(keys)

	for _, k := range keys {
		entry := EntryName(k)
		point := row.Entries[entry]
		printer.write(", @'")
		printer.write(k)
		printer.write("'=")
		printer.write("'")
		printer.write(point)
		printer.write("'")
	}

	printer.write(")")
}

func (printer *queryPrinter) VisitSelect(querySelect *QuerySelect) {
	if querySelect.IsEmpty() {
		return
	}

	printer.write(" where ")

}

func (printer *queryPrinter) LeaveSelect(*QuerySelect) {
}

func (printer *queryPrinter) VisitWhere(position int, where *QueryWhere) {
	if position > 0 {
		printer.write(", ")
	}

	switch where.OpCode {
	case AND:
		printer.write("and(")
	case OR:
		printer.write("or(")
	case PREDICATE:
	default:
		printer.badWhereOpCode(position, where)
	}
}

func (printer *queryPrinter) LeaveWhere(*QueryWhere) {
	printer.write(")")
}

func (printer *queryPrinter) VisitPredicate(pred *QueryPredicate) {
	if pred.IsEmpty() {
		return
	}

	switch pred.OpCode {
	case STR_EQ:
		printer.write("str_eq(")
	case STR_NEQ:
		printer.write("str_neq(")
	default:
		printer.badPredicateOpCode(pred)
	}

	first := true
	if pred.IncludeRowKey {
		printer.write("@key")
		first = false
	}

	for _, k := range pred.Keys {
		if !first {
			printer.write(", ")
		}
		printer.write("@'")
		printer.write(string(k))
		printer.write("'")

		first = false
	}

	for _, l := range pred.Literals {
		if !first {
			printer.write(", ")
		}
		printer.write("'")
		printer.write(l)
		printer.write("'")

		first = false
	}

	printer.write(")")
}

func (printer *queryPrinter) tabs() {
	for i := 0; i < printer.tabIndent; i++ {
		printer.write("\t")
	}
}

func (printer *queryPrinter) indent(indent int) {
	printer.tabIndent += indent

	if printer.tabIndent < 0 {
		panic(fmt.Sprintf("Invalid tabIndent: %v", printer.tabIndent))
	}
}

func (printer *queryPrinter) write(token interface{}) {
	fmt.Fprintf(printer.output, "%v", token)
}

func (printer *queryPrinter) newline() {
	printer.write("\n")
}

func (printer *queryPrinter) space() {
	printer.write(" ")
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

func (visitor *queryFlattener) LeaveSelect(*QuerySelect) {
}

func (visitor *queryFlattener) VisitJoin(join *QueryJoin) {
	visitor.join = *join
}

func (visitor *queryFlattener) LeaveJoin(*QueryJoin) {
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

func (visitor *queryValidator) LeaveSelect(*QuerySelect) {

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
