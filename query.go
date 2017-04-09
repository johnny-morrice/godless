package godless

//go:generate peg -switch -inline query.peg

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

type QueryVisitor interface {
	VisitOpCode(QueryOpCode)
	VisitAST(*QueryAST)
	VisitParser(*QueryParser)
	VisitTableKey(string)
	VisitJoin(*QueryJoin)
	VisitRowJoin(int, *QueryRowJoin)
	VisitSelect(*QuerySelect)
	VisitWhere(int, *QueryWhere)
	VisitWhereOpCode(QueryWhereOpCode)
	LeaveWhere(*QueryWhere)
	VisitPredicate(*QueryPredicate)
}

type QueryResult interface {
	WriteQueryResult(KvQuery)
}

type QueryJoin struct {
	Rows []QueryRowJoin `json:",omitempty"`
}

type QueryRowJoin struct {
	RowKey string
	Values map[string]string `json:",omitempty"`
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

func (query *Query) Run(kvq KvQuery, ns *IpfsNamespace) {
	var result QueryResult

	switch query.OpCode {
	case JOIN:
		visitor := &QueryJoinVisitor{Namespace: ns}
		query.Visit(visitor)
		result = visitor
	case SELECT:
		visitor := &QuerySelectVisitor{Namespace: ns}
		query.Visit(visitor)
		result = visitor
	default:
		query.opcodePanic()
	}

	result.WriteQueryResult(kvq)
}

type whereFrame struct {
	visited bool
	position int
	where *QueryWhere
}

func (query *Query) Visit(visitor QueryVisitor) {
	visitor.VisitOpCode(query.OpCode)
	visitor.VisitAST(query.AST)
	visitor.VisitParser(query.Parser)
	visitor.VisitTableKey(query.TableKey)

	switch query.OpCode {
	case JOIN:
		visitor.VisitJoin(&query.Join)
		for i, row := range query.Join.Rows {
			visitor.VisitRowJoin(i, &row)
		}
	case SELECT:
		visitor.VisitSelect(&query.Select)

		whereStack := []whereFrame{
			whereFrame{where: &query.Select.Where},
		}

		for i := 0; len(whereStack) > 0; {
			top := whereStack[len(whereStack) - 1]
			topWhere := top.where

			if top.visited {
				visitor.LeaveWhere(topWhere)
				whereStack = whereStack[:len(whereStack) - 1]
				i--
			} else {
				visitor.VisitWhere(top.position, topWhere)
				visitor.VisitWhereOpCode(topWhere.OpCode)
				visitor.VisitPredicate(&topWhere.Predicate)
				clauses := topWhere.Clauses;
				clauseCount := len(clauses)
				for j := clauseCount - 1; j >= 0; j-- {
					next := whereFrame{
						where: &clauses[j],
						position: j,
					}
					whereStack = append(whereStack, next)
				}
				top.visited = true
				i += clauseCount
			}
		}
	default:
		query.opcodePanic()
	}
}

func (query *Query) opcodePanic() {
	panic(fmt.Sprintf("Unknown QueryOpcode: %v", query.OpCode))
}

func prettyPrintJson(jsonable interface{}) string {
	bs, err := json.MarshalIndent(jsonable, "", " ")

	if err != nil {
		panic(errors.Wrap(err, "BUG prettyPrintJson failed"))
	}
	return string(bs)
}