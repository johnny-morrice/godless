package query

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/function"
)

// Visitor outline to help with editor macros.
//
// VisitAST(*QueryAST)
// VisitParser(*QueryParser)
// VisitOpCode(QueryOpCode)
// VisitTableKey(crdt.TableName)
// VisitJoin(*QueryJoin)
// VisitRowJoin(int, *QueryRowJoin)
// VisitSelect(*QuerySelect)
// VisitWhere(int, *QueryWhere)
// LeaveWhere(*QueryWhere)
// VisitPredicate(*QueryPredicate)

type queryPrinter struct {
	NoDebugVisitor
	ErrorCollectVisitor
	output    io.Writer
	tabIndent int
}

func (printer *queryPrinter) VisitPublicKeyHash(hash crypto.PublicKeyHash) {
	printer.write(" signed \"")
	printer.write(string(hash))
	printer.write("\"")
}

func (printer *queryPrinter) VisitOpCode(opCode QueryOpCode) {
	switch opCode {
	case SELECT:
		printer.write("select")
	case JOIN:
		printer.write("join")
	default:
		printer.CollectError(fmt.Errorf("Unknown "))
		return
	}
}

func (printer *queryPrinter) VisitTableKey(table crdt.TableName) {
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
	if join.IsEmpty() {
		return
	}

	printer.indent(-1)
}

func (printer *queryPrinter) VisitRowJoin(position int, row *QueryRowJoin) {
	if position > 0 {
		printer.write(",")
	}

	printer.indentWhitespace()
	// TODO shorthand key syntax for simple names.
	printer.write("(")
	printer.indent(1)
	printer.indentWhitespace()
	printer.write("@key")
	printer.write("=")

	printer.write("@\"")
	printer.writeText(string(row.RowKey))
	printer.write("\"")

	keys := make([]string, 0, len(row.Entries))
	for entry, _ := range row.Entries {
		keys = append(keys, string(entry))
	}
	sort.Strings(keys)

	for _, k := range keys {
		entry := crdt.EntryName(k)
		point := row.Entries[entry]
		printer.write(", ")
		printer.indentWhitespace()
		printer.writeKey(k)
		printer.write("=")
		printer.write("\"")
		printer.writeText(string(point))
		printer.write("\"")
	}
	printer.indent(-1)
	printer.indentWhitespace()
	printer.write(")")
}

func (printer *queryPrinter) VisitSelect(querySelect *QuerySelect) {
	if querySelect.IsEmpty() {
		return
	}

	printer.write(" where ")

}

func (printer *queryPrinter) LeaveSelect(querySelect *QuerySelect) {
	if querySelect.IsEmpty() {
		return
	}

	printer.indent(1)
	printer.indentWhitespace()
	printer.write("limit ")
	printer.write(querySelect.Limit)
	printer.indent(-1)
}

func (printer *queryPrinter) VisitWhere(position int, where *QueryWhere) {
	if where.IsEmpty() {
		return
	}

	if position > 0 {
		printer.write(", ")
	}

	printer.indent(1)
	printer.indentWhitespace()

	switch where.OpCode {
	case AND:
		printer.write("and(")
	case OR:
		printer.write("or(")
	case PREDICATE:
	default:
		printer.BadWhereOpCode(position, where)
	}

}

func (printer *queryPrinter) LeaveWhere(where *QueryWhere) {
	if where.IsEmpty() {
		return
	}
	printer.indentWhitespace()
	printer.write(")")
	printer.indent(-1)
}

func (printer *queryPrinter) VisitPredicate(pred *QueryPredicate) {
	if pred.IsEmpty() {
		return
	}

	printer.write(pred.FunctionName)
	printer.write("(")

	printer.indent(1)

	first := true
	if pred.IncludeRowKey {
		printer.indentWhitespace()
		printer.write("@key")
		first = false
	}

	for _, val := range pred.Values {
		if !first {
			printer.write(", ")
		}

		if val.IsKey {
			printer.predicateKey(val)
		} else {
			printer.predicateLiteral(val)
		}

		first = false
	}

	printer.indent(-1)
}

func (printer *queryPrinter) predicateKey(key PredicateValue) {
	printer.indentWhitespace()
	printer.writeKey(string(key.Key))
}

func (printer *queryPrinter) predicateLiteral(lit PredicateValue) {
	printer.indentWhitespace()
	printer.write("\"")
	printer.writeText(string(lit.Literal))
	printer.write("\"")
}

func (printer *queryPrinter) indentWhitespace() {
	printer.newline()
	printer.tabs()
}

func (printer *queryPrinter) tabs() {
	for i := 0; i < printer.tabIndent; i++ {
		printer.write("\t")
	}
}

func (printer *queryPrinter) indent(indent int) {
	printer.tabIndent += indent

	if printer.tabIndent < 0 {
		panic(fmt.Sprintf("Invalid tabIndent: %d", printer.tabIndent))
	}
}

func (printer *queryPrinter) write(token interface{}) {
	fmt.Fprintf(printer.output, "%v", token)
}

func (printer *queryPrinter) writeKey(token string) {
	if len(token) > 0 && printer.isEasyKey(token) {
		printer.write(token)
	} else {
		printer.write("@\"")
		printer.writeText(token)
		printer.write("\"")
	}
}

func (printer *queryPrinter) isEasyKey(token string) bool {
	parts := strings.Split(token, "")
	for _, p := range parts {
		if !strings.Contains(__KEY_SYMS, p) {
			return false
		}
	}

	return true
}

func (printer *queryPrinter) writeText(token string) {
	quoted := quote(token)
	printer.write(quoted)
}

func (printer *queryPrinter) newline() {
	printer.write("\n")
}

func (printer *queryPrinter) space() {
	printer.write(" ")
}

type queryFlattener struct {
	NoDebugVisitor
	ErrorCollectVisitor

	tableName  crdt.TableName
	publicKeys []crypto.PublicKeyHash
	opCode     QueryOpCode
	join       QueryJoin
	slct       QuerySelect
	allClauses []QueryWhere
}

func (visitor *queryFlattener) VisitPublicKeyHash(hash crypto.PublicKeyHash) {
	visitor.publicKeys = append(visitor.publicKeys, hash)
}

func (visitor *queryFlattener) VisitOpCode(opCode QueryOpCode) {
	visitor.opCode = opCode
}

func (visitor *queryFlattener) VisitTableKey(tableKey crdt.TableName) {
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
	NoDebugVisitor
	ErrorCollectVisitor
	NoJoinVisitor
	Functions  function.FunctionNamespace
	whereStack []*QueryWhere
}

func (visitor *queryValidator) VisitPublicKeyHash(hash crypto.PublicKeyHash) {

}

func (visitor *queryValidator) VisitOpCode(opCode QueryOpCode) {
	switch opCode {
	case SELECT:
	case JOIN:
		// Okay!
	default:
		visitor.BadOpcode(opCode)
	}
}

func (visitor *queryValidator) VisitTableKey(tableKey crdt.TableName) {
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
		visitor.BadWhereOpCode(position, where)
	}

	visitor.pushWhere(where)
}

func (visitor *queryValidator) LeaveWhere(where *QueryWhere) {
	tip := visitor.popWhere()
	if tip != where {
		visitor.panicOnBadStack()
	}
}

func (visitor *queryValidator) pushWhere(where *QueryWhere) {
	visitor.whereStack = append(visitor.whereStack, where)
}

func (visitor *queryValidator) popWhere() *QueryWhere {
	tipIndex := visitor.getTipIndex()
	tip := visitor.whereStack[tipIndex]
	visitor.whereStack = visitor.whereStack[:tipIndex]
	return tip
}

func (visitor *queryValidator) getTipIndex() int {
	count := len(visitor.whereStack)

	if count == 0 {
		visitor.panicOnBadStack()
	}

	return count - 1
}

func (visitor *queryValidator) panicOnBadStack() {
	panic("where clause stack corruption in validator")
}

func (visitor *queryValidator) VisitPredicate(predicate *QueryPredicate) {
	if visitor.Functions == nil {
		visitor.CollectError(errors.New("No FunctionNamespace for validator"))
		return
	}

	tip := visitor.whereStack[visitor.getTipIndex()]

	if tip.OpCode == PREDICATE {
		_, err := visitor.Functions.GetFunction(predicate.FunctionName)

		if err != nil {
			visitor.CollectError(err)
		}
	}

}
