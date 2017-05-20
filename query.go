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
	AST      *QueryAST    `json:"-"`
	Parser   *QueryParser `json:"-"`
	OpCode   QueryOpCode  `json:",omitempty"`
	TableKey TableName
	Join     QueryJoin   `json:",omitempty"`
	Select   QuerySelect `json:",omitempty"`
}

func MakeQueryMessage(query *Query) *QueryMessage {
	return &QueryMessage{
		Opcode: uint32(query.OpCode),
		Table:  string(query.TableKey),
		Join:   MakeQueryJoinMessage(query.Join),
		Select: MakeQuerySelectMessage(query.Select),
	}
}

func MakeQuerySelectMessage(querySelect QuerySelect) *QuerySelectMessage {
	message := &QuerySelectMessage{
		Limit: querySelect.Limit,
		Where: MakeQueryWhereMessage(querySelect.Where),
	}

	return message
}

func MakeQueryWhereMessage(queryWhere QueryWhere) *QueryWhereMessage {
	builder := &whereMessageBuilder{}
	whereStack := makeWhereStack(&queryWhere)
	whereStack.visit(builder)

	return builder.message
}

type whereMessageBuilder struct {
	message      *QueryWhereMessage
	messageStack []whereBuilderFrame
}

func (builder *whereMessageBuilder) VisitWhere(pos int, where *QueryWhere) {
	message := &QueryWhereMessage{}

	if builder.message == nil {
		builder.message = message
	}

	frame := whereBuilderFrame{
		message: message,
		where:   where,
	}

	tip := builder.peek()
	tip.message.Clauses[pos] = message

	builder.push(frame)
	configureWhereMessage(where, message)
}

func (builder *whereMessageBuilder) LeaveWhere(where *QueryWhere) {
	frame := builder.pop()

	if frame.where != where {
		panic("messageStack corruption")
	}
}

func (builder *whereMessageBuilder) VisitPredicate(*QueryPredicate) {

}

func (builder *whereMessageBuilder) push(frame whereBuilderFrame) {
	builder.messageStack = append(builder.messageStack, frame)
}

func (builder *whereMessageBuilder) peek() whereBuilderFrame {
	return builder.messageStack[builder.lastIndex()]
}

func (builder *whereMessageBuilder) pop() whereBuilderFrame {
	frame := builder.peek()
	builder.messageStack = builder.messageStack[:builder.lastIndex()]
	return frame
}

func (builder *whereMessageBuilder) lastIndex() int {
	return len(builder.messageStack) - 1
}

type whereBuilderFrame struct {
	message *QueryWhereMessage
	where   *QueryWhere
}

func configureWhereMessage(queryWhere *QueryWhere, message *QueryWhereMessage) {
	message.Opcode = uint32(queryWhere.OpCode)
	message.Predicate = MakeQueryPredicateMessage(queryWhere.Predicate)
	message.Clauses = make([]*QueryWhereMessage, len(queryWhere.Clauses))
}

func MakeQueryPredicateMessage(predicate QueryPredicate) *QueryPredicateMessage {
	message := &QueryPredicateMessage{
		Opcode:   uint32(predicate.OpCode),
		Userow:   predicate.IncludeRowKey,
		Literals: make([]string, len(predicate.Literals)),
		Keys:     make([]string, len(predicate.Keys)),
	}

	for i, l := range predicate.Literals {
		message.Literals[i] = string(l)
	}

	for i, k := range predicate.Keys {
		message.Keys[i] = string(k)
	}

	return message
}

func MakeQueryJoinMessage(join QueryJoin) *QueryJoinMessage {
	message := &QueryJoinMessage{
		Rows: make([]*QueryRowJoinMessage, len(join.Rows)),
	}

	for i, r := range join.Rows {
		message.Rows[i] = MakeQueryRowJoinMessage(r)
	}

	return message
}

func MakeQueryRowJoinMessage(row QueryRowJoin) *QueryRowJoinMessage {
	message := &QueryRowJoinMessage{
		Row:     string(row.RowKey),
		Entries: make([]*QueryRowJoinEntryMessage, len(row.Entries)),
	}

	// We don't store these in IPFS so no need for stable order.
	i := 0
	for e, p := range row.Entries {
		message.Entries[i] = &QueryRowJoinEntryMessage{
			Entry: string(e),
			Point: string(p),
		}
		i++
	}

	return message
}

func ReadQueryMessage(message *QueryMessage) *Query {
	return nil
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
	VisitTableKey(TableName)
	VisitJoin(*QueryJoin)
	VisitRowJoin(int, *QueryRowJoin)
	VisitSelect(*QuerySelect)
}

type QueryJoin struct {
	Rows []QueryRowJoin `json:",omitempty"`
}

type QueryRowJoin struct {
	RowKey  RowName
	Entries map[EntryName]Point `json:",omitempty"`
}

type QuerySelect struct {
	Where QueryWhere `json:",omitempty"`
	Limit uint32     `json:",omitempty"`
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
	OpCode        QueryPredicateOpCode `json:",omitempty"`
	Keys          []EntryName          `json:",omitempty"`
	Literals      []string             `json:",omitempty"`
	IncludeRowKey bool                 `json:",omitempty"`
}

func (pred QueryPredicate) match(tableKey string, r Row) bool {
	// TODO
	return false
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

func (visitor *queryValidator) VisitTableKey(tableKey TableName) {
	if tableKey == "" {
		visitor.collectError(errors.New("Empty table key"))
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
		visitor.collectError(errors.New(fmt.Sprintf("Unknown Where OpCode at position %v: %v", position, where)))
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
		visitor.collectError(errors.New(fmt.Sprintf("Unknown Predicate OpCode: %v", predicate)))
	}
}
