package query

//go:generate peg -switch -inline query.peg
//go:generate mockgen -package mock_godless -destination ../mock/mock_query.go -imports lib=github.com/johnny-morrice/godless/query -self_package lib github.com/johnny-morrice/godless/query QueryVisitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/function"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/proto"
	"github.com/pkg/errors"
)

type QueryOpCode uint16

const (
	QUERY_NOP = QueryOpCode(iota)
	SELECT
	JOIN
)

type Query struct {
	AST        *QueryAST    `json:"-"`
	Parser     *QueryParser `json:"-"`
	OpCode     QueryOpCode  `json:",omitempty"`
	TableKey   crdt.TableName
	Join       QueryJoin   `json:",omitempty"`
	Select     QuerySelect `json:",omitempty"`
	PublicKeys []crypto.PublicKeyHash
}

type whereVisitor interface {
	VisitWhere(int, *QueryWhere)
	LeaveWhere(*QueryWhere)
	VisitPredicate(*QueryPredicate)
}

type selectVisitor interface {
	VisitSelect(*QuerySelect)
	LeaveSelect(*QuerySelect)
}

type joinVisitor interface {
	VisitJoin(*QueryJoin)
	LeaveJoin(*QueryJoin)
	VisitRowJoin(int, *QueryRowJoin)
}

type cryptoVisitor interface {
	VisitPublicKeyHash(crypto.PublicKeyHash)
}

type debugVisitor interface {
	VisitAST(*QueryAST)
	VisitParser(*QueryParser)
}

type baseVisitor interface {
	VisitOpCode(QueryOpCode)
	VisitTableKey(crdt.TableName)
}

type QueryVisitor interface {
	whereVisitor
	selectVisitor
	cryptoVisitor
	joinVisitor
	debugVisitor
	baseVisitor
}

type QueryJoin struct {
	Rows []QueryRowJoin `json:",omitempty"`
}

func (join QueryJoin) IsEmpty() bool {
	return join.equals(QueryJoin{})
}

func (join QueryJoin) equals(other QueryJoin) bool {
	if len(join.Rows) != len(other.Rows) {
		return false
	}

	for i, myJoin := range join.Rows {
		theirJoin := other.Rows[i]
		if !myJoin.equals(theirJoin) {
			return false
		}
	}

	return true
}

type QueryRowJoin struct {
	RowKey crdt.RowName
	// TODO would this be clearer/more performant as a slice of pair structures?
	Entries map[crdt.EntryName]crdt.PointText `json:",omitempty"`
}

func (join QueryRowJoin) equals(other QueryRowJoin) bool {
	ok := join.RowKey == other.RowKey
	ok = ok && len(join.Entries) == len(other.Entries)

	if !ok {
		return false
	}

	keys := make([]string, 0, len(join.Entries))

	for k, _ := range join.Entries {
		keys = append(keys, string(k))
	}

	sort.Strings(keys)

	for _, k := range keys {
		ename := crdt.EntryName(k)
		if join.Entries[ename] != other.Entries[ename] {
			return false
		}
	}

	return true
}

type QuerySelect struct {
	Where QueryWhere `json:",omitempty"`
	Limit uint32     `json:",omitempty"`
}

func (querySelect QuerySelect) IsEmpty() bool {
	return 0 == querySelect.Limit && querySelect.Where.IsEmpty()
}

type QueryWhereOpCode uint16

const (
	WHERE_NOOP = QueryWhereOpCode(iota)
	AND
	OR
	PREDICATE
)

type QueryWhere struct {
	OpCode    QueryWhereOpCode `json:",omitempty"`
	Clauses   []QueryWhere     `json:",omitempty"`
	Predicate QueryPredicate   `json:",omitempty"`
}

func (where QueryWhere) IsEmpty() bool {
	var emptyOpCode QueryWhereOpCode
	empty := emptyOpCode == where.OpCode
	empty = empty && len(where.Clauses) == 0
	empty = empty && where.Predicate.IsEmpty()
	return empty
}

func (where QueryWhere) shallowEquals(other QueryWhere) bool {
	ok := where.OpCode == other.OpCode
	ok = ok && len(where.Clauses) == len(other.Clauses)

	if !ok {
		return false
	}

	if !where.Predicate.equals(other.Predicate) {
		return false
	}

	return true
}

type QueryPredicateOpCode uint16

const (
	PREDICATE_NOP = QueryPredicateOpCode(iota)
	STR_EQ
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
	FunctionName  string
	Keys          []crdt.EntryName `json:",omitempty"`
	Literals      []string         `json:",omitempty"`
	IncludeRowKey bool             `json:",omitempty"`
}

func (pred QueryPredicate) IsEmpty() bool {
	return pred.equals(QueryPredicate{})
}

func (pred QueryPredicate) equals(other QueryPredicate) bool {
	ok := pred.FunctionName == other.FunctionName
	ok = ok && pred.IncludeRowKey == other.IncludeRowKey
	ok = ok && len(pred.Keys) == len(other.Keys)
	ok = ok && len(pred.Literals) == len(other.Literals)

	if !ok {
		return false
	}

	for i, myKey := range pred.Keys {
		theirKey := other.Keys[i]
		if myKey != theirKey {
			return false
		}
	}

	for i, myLit := range pred.Literals {
		theirLit := other.Literals[i]
		if myLit != theirLit {
			return false
		}
	}

	return true
}

func Compile(source string) (*Query, error) {
	parser := &QueryParser{Buffer: source}
	parser.Pretty = true
	parser.Init()

	if err := parser.Parse(); err != nil {
		if log.CanLog(log.LOG_DEBUG) {
			parser.PrintSyntaxTree()
		}

		return nil, errors.Wrap(err, "Query parse failed")
	}

	parser.Execute()

	query, err := parser.QueryAST.Compile()

	if err != nil {
		if log.CanLog(log.LOG_DEBUG) {
			log.Debug("AST:\n\n%s\n\n", prettyPrintJson(parser.QueryAST))
			parser.PrintSyntaxTree()
		}

		return nil, errors.Wrap(err, "Query compile failed")
	}

	query.Parser = parser

	return query, nil
}

func EncodeQuery(query *Query, w io.Writer) error {
	const failMsg = "EncodeQuery failed"

	message := MakeQueryMessage(query)

	err := util.Encode(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeQuery(r io.Reader) (*Query, error) {
	const failMsg = "DecodeQuery failed"
	message := &proto.QueryMessage{}

	err := util.Decode(message, r)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	return ReadQueryMessage(message)
}

func (query *Query) PrettyPrint(w io.Writer) error {
	printer := &queryPrinter{output: w}

	query.Visit(printer)

	return printer.Error()
}

func (query *Query) Analyse() string {
	jsonQuery := prettyPrintJson(query)
	jsonAST := prettyPrintJson(query.AST)
	return fmt.Sprintf("Compiled:\n\n%s\n\nAST:\n\n%s", jsonQuery, jsonAST)
}

type ValidationContext struct {
	Functions function.FunctionNamespace
}

func (query *Query) Validate(context ValidationContext) error {
	validator := &queryValidator{
		Functions: context.Functions,
	}

	query.Visit(validator)

	return validator.err
}

func (query *Query) Equals(other *Query) bool {
	flattenMe := &queryFlattener{}
	flattenThem := &queryFlattener{}

	query.Visit(flattenMe)
	other.Visit(flattenThem)

	return flattenMe.Equals(flattenThem)
}

func (query *Query) Visit(visitor QueryVisitor) {
	visitor.VisitAST(query.AST)
	visitor.VisitParser(query.Parser)
	visitor.VisitOpCode(query.OpCode)
	visitor.VisitTableKey(query.TableKey)

	for _, hash := range query.PublicKeys {
		visitor.VisitPublicKeyHash(hash)
	}

	switch query.OpCode {
	case JOIN:
		queryJoin := &query.Join
		visitor.VisitJoin(queryJoin)
		for i, row := range query.Join.Rows {
			visitor.VisitRowJoin(i, &row)
		}
		visitor.LeaveJoin(queryJoin)
	case SELECT:
		querySelect := &query.Select
		visitor.VisitSelect(querySelect)

		stack := MakeWhereStack(&query.Select.Where)
		stack.Visit(visitor)
		visitor.LeaveSelect(querySelect)
	case QUERY_NOP:
		// Do nothing.
	default:
		query.OpCodePanic()
	}
}

func (query *Query) OpCodePanic() {
	panic(fmt.Sprintf("Unknown Query OpCode: %v", query.OpCode))
}

func (query *Query) PrettyText() (string, error) {
	buff := &bytes.Buffer{}
	err := query.PrettyPrint(buff)

	return buff.String(), err
}

func prettyPrintJson(jsonable interface{}) string {
	bs, err := json.MarshalIndent(jsonable, "", " ")

	if err != nil {
		panic(errors.Wrap(err, "BUG prettyPrintJson failed"))
	}
	return string(bs)
}
