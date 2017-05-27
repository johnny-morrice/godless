package godless

//go:generate peg -switch -inline query.peg
//go:generate mockgen -destination mock/mock_query.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless QueryVisitor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/gogo/protobuf/proto"
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
	builder.messageStack = []whereBuilderFrame{}
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

	if len(builder.messageStack) > 0 {
		tip := builder.peek()
		tip.message.Clauses[pos] = message
	}

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

// FIXME not implemented
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
	LeaveJoin(*QueryJoin)
	VisitRowJoin(int, *QueryRowJoin)
	VisitSelect(*QuerySelect)
	LeaveSelect(*QuerySelect)
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
	RowKey  RowName
	Entries map[EntryName]Point `json:",omitempty"`
}

func (join QueryRowJoin) equals(other QueryRowJoin) bool {
	ok := join.RowKey == other.RowKey
	ok = ok && len(join.Entries) == len(other.Entries)

	if !ok {
		return false
	}

	keys := make([]string, len(join.Entries))
	i := 0
	for k, _ := range join.Entries {
		keys[i] = string(k)
		i++
	}

	sort.Strings(keys)

	for _, k := range keys {
		ename := EntryName(k)
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

func (pred QueryPredicate) IsEmpty() bool {
	return pred.equals(QueryPredicate{})
}

func (pred QueryPredicate) equals(other QueryPredicate) bool {
	ok := pred.OpCode == other.OpCode
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

func EncodeQuery(query *Query, w io.Writer) error {
	const failMsg = "EncodeQuery failed"

	message := MakeQueryMessage(query)

	bs, err := proto.Marshal(message)
	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	err = writeBytes(bs, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeQuery(r io.Reader) (*Query, error) {
	const failMsg = "DecodeQuery failed"
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	message := &QueryMessage{}
	err = proto.Unmarshal(bs, message)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	return ReadQueryMessage(message), nil
}

func (query *Query) PrettyPrint(w io.Writer) error {
	printer := &queryPrinter{output: w}

	query.Visit(printer)

	return printer.Error()
}

func (query *Query) Analyse() string {
	return fmt.Sprintf("Compiled:\n\n%v\n\nAST:\n\n%v", prettyPrintJson(query), prettyPrintJson(query.AST))
}

func (query *Query) Validate() error {
	validator := &queryValidator{}

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

		stack := makeWhereStack(&query.Select.Where)
		stack.visit(visitor)
		visitor.LeaveSelect(querySelect)
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
