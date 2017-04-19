package godless
//go:generate peg -switch -inline query.peg
//go:generate mockgen -destination mock/mock_query.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless QueryVisitor

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/pkg/errors"
)

type QueryOpCode uint16

const (
	QUERY_NOP = QueryOpCode(iota)
	SELECT
	JOIN
)

type Query struct {
	AST *QueryAST `json:"-"`
	Parser *QueryParser `json:"-"`
	OpCode QueryOpCode `json:",omitempty"`
	TableKey string
	Join QueryJoin `json:",omitempty"`
	Select QuerySelect `json:",omitempty"`
}

type whereVisitor interface {
	VisitWhere(int, *QueryWhere)
	LeaveWhere(*QueryWhere)
	VisitPredicate(*QueryPredicate)
}

type QueryVisitor interface {
	whereVisitor
	VisitOpCode(QueryOpCode)
	VisitAST(*QueryAST)
	VisitParser(*QueryParser)
	VisitTableKey(string)
	VisitJoin(*QueryJoin)
	VisitRowJoin(int, *QueryRowJoin)
	VisitSelect(*QuerySelect)
}

type QueryRun interface {
	RunQuery(KvQuery)
}

type QueryJoin struct {
	Rows []QueryRowJoin `json:",omitempty"`
}

type QueryRowJoin struct {
	RowKey string
	Entries map[string]string `json:",omitempty"`
}

type QuerySelect struct {
	Where QueryWhere `json:",omitempty"`
	Limit uint `json:",omitempty"`
}

type QueryWhereOpCode uint16

const (
	WHERE_NOOP = QueryWhereOpCode(iota)
	AND
	OR
	PREDICATE
)

type QueryWhere struct {
	OpCode QueryWhereOpCode `json:",omitempty"`
	Clauses []QueryWhere `json:",omitempty"`
	Predicate QueryPredicate `json:",omitempty"`
}

type QueryPredicateOpCode uint16

const (
	PREDICATE_NOP = QueryPredicateOpCode(iota)
	STR_EQ
	STR_NEQ
	// TODO flesh these out
	// STR_EMPTY
	// STR_NEMPTY
	// STR_GT
	// STR_LT
	// STR_GTE
	// STR_LTE
	// NUM_EQ
	// NUM_NEQ
	// NUM_GT
	// NUM_LT
	// NUM_GTE
	// NUM_LTE
	// TIME_EQ
	// TIME_NEQ
	// TIME_GT
	// TIME_LT
	// TIME_GTE
	// TIME_LTE
)

type QueryPredicate struct {
	OpCode QueryPredicateOpCode `json:",omitempty"`
	Keys []string `json:",omitempty"`
	Literals []string `json:",omitempty"`
	IncludeRowKey bool `json:",omitempty"`
}

func (pred QueryPredicate) match(tableKey string, r Row) bool {
	// TODO
	return false;
}

func CompileQuery(source string) (*Query, error) {
	parser := &QueryParser{Buffer: source}
	parser.Pretty = true
	parser.Init()

	if err := parser.Parse(); err != nil {
		if __DEBUG {
			parser.PrintSyntaxTree()
		}

		return nil, errors.Wrap(err, "Query parse failed")
	}

	parser.Execute()

	query, err := parser.QueryAST.Compile()

	if err != nil {
		if __DEBUG {
			fmt.Fprintf(os.Stderr, "AST:\n\n%s\n\n", prettyPrintJson(parser.QueryAST))
			parser.PrintSyntaxTree()
		}

		return nil, errors.Wrap(err, "Query compile failed")
	}

	query.Parser = parser

	return query, nil
}

func (query *Query) Analyse() string {
	return fmt.Sprintf("Compiled:\n\n%v\n\nAST:\n\n%v", prettyPrintJson(query), prettyPrintJson(query.AST))
}

func (query *Query) Validate() error {
	validator := &queryValidator{}

	query.Visit(validator)

	return validator.err
}

func (query *Query) Visit(visitor QueryVisitor) {
	visitor.VisitAST(query.AST)
	visitor.VisitParser(query.Parser)
	visitor.VisitOpCode(query.OpCode)
	visitor.VisitTableKey(query.TableKey)

	switch query.OpCode {
	case JOIN:
		visitor.VisitJoin(&query.Join)
		for i, row := range query.Join.Rows {
			visitor.VisitRowJoin(i, &row)
		}
	case SELECT:
		visitor.VisitSelect(&query.Select)

		stack := makeWhereStack(&query.Select.Where)
		stack.visit(visitor)
	case QUERY_NOP:
		// Do nothing.
	default:
		query.opcodePanic()
	}
}

func (query *Query) opcodePanic() {
	panic(fmt.Sprintf("Unknown Query OpCode: %v", query.OpCode))
}

func prettyPrintJson(jsonable interface{}) string {
	bs, err := json.MarshalIndent(jsonable, "", " ")

	if err != nil {
		panic(errors.Wrap(err, "BUG prettyPrintJson failed"))
	}
	return string(bs)
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
		visitor.collectError(errors.New(fmt.Sprintf("Invalid Query OpCode: %v", opCode)))
	}
}

func (visitor *queryValidator) VisitTableKey(tableKey string) {
	if tableKey == "" {
		visitor.collectError(errors.New("Empty table key"))
	}
}

func (visitor *queryValidator) VisitSelect(*QuerySelect) {
}

func (visitor *queryValidator) VisitWhere(position int, where *QueryWhere) {
	switch where.OpCode {
	case AND:
	case OR:
	case PREDICATE:
		// Okay!
	default:
		visitor.collectError(errors.New(fmt.Sprintf("Unknown Where OpCode at position %v: %v", position, where)))
	}
}

func (visitor *queryValidator) LeaveWhere(*QueryWhere) {
}

func (visitor *queryValidator) VisitPredicate(predicate *QueryPredicate) {
	switch predicate.OpCode {
	case STR_EQ:
	case STR_NEQ:
		// Okay!
	default:
		visitor.collectError(errors.New(fmt.Sprintf("Unknown Predicate OpCode: %v",predicate)))
	}
}
