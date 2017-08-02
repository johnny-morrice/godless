package query

import (
	"fmt"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/proto"
)

type whereMessageVisitor interface {
	VisitWhere(int, *proto.QueryWhereMessage)
	LeaveWhere(*proto.QueryWhereMessage)
	VisitPredicate(*proto.QueryPredicateMessage)
}

type queryMessageVisitor interface {
	whereMessageVisitor
	VisitOpCode(uint32)
	VisitTableKey(string)
	VisitPublicKeyHash(string)
	VisitJoin(*proto.QueryJoinMessage)
	LeaveJoin(*proto.QueryJoinMessage)
	VisitRowJoin(int, *proto.QueryRowJoinMessage)
	VisitSelect(*proto.QuerySelectMessage)
	LeaveSelect(*proto.QuerySelectMessage)
}

type whereMessageBuilder struct {
	message *proto.QueryWhereMessage
	stack   whereBuilderFrameStack
}

func (builder *whereMessageBuilder) VisitWhere(pos int, where *QueryWhere) {
	message := &proto.QueryWhereMessage{}

	if builder.message == nil {
		builder.message = message
	}

	frame := whereBuilderFrame{
		message: message,
		where:   where,
	}

	if len(builder.stack.stk) > 0 {
		tip := builder.stack.peek()
		tip.message.Clauses[pos] = message
	}

	builder.stack.push(frame)
	configureWhereMessage(where, message)
}

func (builder *whereMessageBuilder) LeaveWhere(where *QueryWhere) {
	frame := builder.stack.pop()

	if frame.where != where {
		panic("messageStack corruption")
	}
}

func (builder *whereMessageBuilder) VisitPredicate(*QueryPredicate) {

}

type whereBuilderFrameStack struct {
	stk []whereBuilderFrame
}

func makeWhereBuilderFrameStack() whereBuilderFrameStack {
	return whereBuilderFrameStack{stk: []whereBuilderFrame{}}
}

func (stack *whereBuilderFrameStack) push(frame whereBuilderFrame) {
	stack.stk = append(stack.stk, frame)
}

func (stack *whereBuilderFrameStack) peek() whereBuilderFrame {
	return stack.stk[stack.lastIndex()]
}

func (stack *whereBuilderFrameStack) pop() whereBuilderFrame {
	frame := stack.peek()
	stack.stk = stack.stk[:stack.lastIndex()]
	return frame
}

func (stack *whereBuilderFrameStack) lastIndex() int {
	return len(stack.stk) - 1
}

type whereBuilderFrame struct {
	message *proto.QueryWhereMessage
	where   *QueryWhere
}

func configureWhereMessage(queryWhere *QueryWhere, message *proto.QueryWhereMessage) {
	message.OpCode = uint32(queryWhere.OpCode)
	message.Predicate = MakeQueryPredicateMessage(queryWhere.Predicate)
	message.Clauses = make([]*proto.QueryWhereMessage, len(queryWhere.Clauses))
}

func MakeQueryMessage(query *Query) *proto.QueryMessage {
	hashTexts := make([]string, len(query.PublicKeys))

	for i, hash := range query.PublicKeys {
		hashTexts[i] = string(hash)
	}

	return &proto.QueryMessage{
		OpCode:    uint32(query.OpCode),
		Table:     string(query.TableKey),
		Join:      MakeQueryJoinMessage(query.Join),
		Select:    MakeQuerySelectMessage(query.Select),
		KeyHashes: hashTexts,
	}
}

func MakeQuerySelectMessage(querySelect QuerySelect) *proto.QuerySelectMessage {
	message := &proto.QuerySelectMessage{
		Limit: querySelect.Limit,
		Where: MakeQueryWhereMessage(querySelect.Where),
	}

	return message
}

func MakeQueryWhereMessage(queryWhere QueryWhere) *proto.QueryWhereMessage {
	builder := &whereMessageBuilder{}
	builder.stack = makeWhereBuilderFrameStack()
	whereStack := MakeWhereStack(&queryWhere)
	whereStack.Visit(builder)

	return builder.message
}

func ReadQueryMessage(message *proto.QueryMessage) (*Query, error) {
	unpb := makeQueryMessageDecoder()
	err := visitMessage(message, unpb)

	if err != nil {
		return nil, err
	}

	if unpb.Error() != nil {
		return nil, unpb.Error()
	}

	return unpb.Query, nil

}

func MakeQueryPredicateMessage(predicate QueryPredicate) *proto.QueryPredicateMessage {
	message := &proto.QueryPredicateMessage{
		FunctionName: predicate.FunctionName,
		Userow:       predicate.IncludeRowKey,
		Values:       make([]*proto.PredicateValue, len(predicate.Values)),
	}

	for i, val := range predicate.Values {
		messageVal := &proto.PredicateValue{
			IsKey: val.IsKey,
		}

		if val.IsKey {
			messageVal.Text = string(val.Key)
		} else {
			messageVal.Text = string(val.Literal)
		}

		message.Values[i] = messageVal
	}

	return message
}

func MakeQueryJoinMessage(join QueryJoin) *proto.QueryJoinMessage {
	message := &proto.QueryJoinMessage{
		Rows: make([]*proto.QueryRowJoinMessage, len(join.Rows)),
	}

	for i, r := range join.Rows {
		message.Rows[i] = MakeQueryRowJoinMessage(r)
	}

	return message
}

func MakeQueryRowJoinMessage(row QueryRowJoin) *proto.QueryRowJoinMessage {
	message := &proto.QueryRowJoinMessage{
		Row:     string(row.RowKey),
		Entries: make([]*proto.QueryRowJoinEntryMessage, 0, len(row.Entries)),
	}

	// We don't store these in IPFS so no need for stable order.
	for e, p := range row.Entries {
		rowJoin := &proto.QueryRowJoinEntryMessage{
			Entry: string(e),
			Point: string(p),
		}
		message.Entries = append(message.Entries, rowJoin)
	}

	return message
}

type queryMessageDecoder struct {
	ErrorCollectVisitor
	Query *Query
	stack whereBuilderFrameStack
}

func makeQueryMessageDecoder() *queryMessageDecoder {
	return &queryMessageDecoder{
		Query: &Query{},
		stack: makeWhereBuilderFrameStack(),
	}
}

func (decoder *queryMessageDecoder) VisitWhere(position int, message *proto.QueryWhereMessage) {
	where := decoder.createChildWhere(position)

	frame := whereBuilderFrame{
		message: message,
		where:   where,
	}

	decoder.decodeWhere(frame)
	decoder.stack.push(frame)
}

func (decoder *queryMessageDecoder) VisitPublicKeyHash(publicKeyHash string) {
	decoder.Query.PublicKeys = append(decoder.Query.PublicKeys, crypto.PublicKeyHash(publicKeyHash))
}

func (decoder *queryMessageDecoder) createChildWhere(position int) *QueryWhere {
	if len(decoder.stack.stk) == 0 {
		return &decoder.Query.Select.Where
	}

	tip := decoder.stack.peek()
	return &tip.where.Clauses[position]
}

func (decoder *queryMessageDecoder) LeaveWhere(message *proto.QueryWhereMessage) {
	frame := decoder.stack.pop()
	if frame.message != message {
		panic("queryMessageDecoder.stack corruption")
	}
}

func (decoder *queryMessageDecoder) VisitPredicate(*proto.QueryPredicateMessage) {
}

func (decoder *queryMessageDecoder) VisitOpCode(opCode uint32) {
	switch opCode {
	case MESSAGE_NOOP:
		fallthrough
	case MESSAGE_SELECT:
		fallthrough
	case MESSAGE_JOIN:
		decoder.Query.OpCode = QueryOpCode(opCode)
	}
}

func (decoder *queryMessageDecoder) VisitTableKey(table string) {
	decoder.Query.TableKey = crdt.TableName(table)
}

func (decoder *queryMessageDecoder) VisitJoin(message *proto.QueryJoinMessage) {
	decoder.Query.Join.Rows = make([]QueryRowJoin, len(message.Rows))
}

func (decoder *queryMessageDecoder) LeaveJoin(*proto.QueryJoinMessage) {
}

func (decoder *queryMessageDecoder) VisitRowJoin(position int, message *proto.QueryRowJoinMessage) {
	row := &decoder.Query.Join.Rows[position]
	decoder.decodeRowJoin(row, message)
}

func (decoder *queryMessageDecoder) VisitSelect(message *proto.QuerySelectMessage) {
	decoder.Query.Select.Limit = message.Limit
}

func (decoder *queryMessageDecoder) LeaveSelect(*proto.QuerySelectMessage) {
}

func (decoder *queryMessageDecoder) decodeWhere(frame whereBuilderFrame) {
	msg := frame.message
	where := frame.where
	switch msg.OpCode {
	case MESSAGE_AND:
		fallthrough
	case MESSAGE_OR:
		fallthrough
	case MESSAGE_NOOP:
		fallthrough
	case MESSAGE_PREDICATE:
		where.OpCode = QueryWhereOpCode(msg.OpCode)
	default:
		decoder.badWhereMessageOpCode(msg)
	}

	where.Clauses = make([]QueryWhere, len(msg.Clauses))
	decoder.decodePredicate(&where.Predicate, msg.Predicate)
}

func (decoder *queryMessageDecoder) decodeRowJoin(row *QueryRowJoin, message *proto.QueryRowJoinMessage) {
	row.RowKey = crdt.RowName(message.Row)
	row.Entries = map[crdt.EntryName]crdt.PointText{}

	for _, messageEntry := range message.Entries {
		entry := crdt.EntryName(messageEntry.Entry)
		point := crdt.PointText(messageEntry.Point)
		row.Entries[entry] = point
	}
}

func (decoder *queryMessageDecoder) decodePredicate(pred *QueryPredicate, message *proto.QueryPredicateMessage) {
	pred.FunctionName = message.FunctionName

	pred.Values = make([]PredicateValue, len(message.Values))

	for i, msgVal := range message.Values {
		val := PredicateValue{
			IsKey: msgVal.IsKey,
		}

		if msgVal.IsKey {
			val.Key = crdt.EntryName(msgVal.Text)
		} else {
			val.Literal = crdt.PointText(msgVal.Text)
		}

		pred.Values[i] = val
	}

	pred.IncludeRowKey = message.Userow
}

func (decoder *queryMessageDecoder) badWhereMessageOpCode(message *proto.QueryWhereMessage) {
	err := fmt.Errorf("Bad queryWhereMessageOpCode: %v", message)
	decoder.CollectError(err)
}

func (decoder *queryMessageDecoder) badPredicateMessageOpCode(message *proto.QueryPredicateMessage) {
	err := fmt.Errorf("Bad queryPredicateMessageOpCode: %v", message)
	decoder.CollectError(err)
}

func visitMessage(message *proto.QueryMessage, visitor queryMessageVisitor) error {
	visitor.VisitOpCode(message.OpCode)
	visitor.VisitTableKey(message.Table)

	for _, pub := range message.KeyHashes {
		visitor.VisitPublicKeyHash(pub)
	}

	switch message.OpCode {
	case MESSAGE_JOIN:
		visitor.VisitJoin(message.Join)
		for i, row := range message.Join.Rows {
			visitor.VisitRowJoin(i, row)
		}
		visitor.LeaveJoin(message.Join)
	case MESSAGE_SELECT:
		visitor.VisitSelect(message.Select)

		stack := makeWhereMessageStack(message.Select.Where)
		stack.visit(visitor)
		visitor.LeaveSelect(message.Select)
	case MESSAGE_NOOP:
		// Do nothing.
	default:
		return fmt.Errorf("Bad QueryMessage.OpCode: %v", message.OpCode)
	}

	return nil
}

type whereMessageStack struct {
	stk []whereMessageFrame
}

func makeWhereMessageStack(where *proto.QueryWhereMessage) *whereMessageStack {
	return &whereMessageStack{
		stk: []whereMessageFrame{whereMessageFrame{where: where}},
	}
}

func (stack *whereMessageStack) visit(visitor whereMessageVisitor) {
	for i := 0; len(stack.stk) > 0; {
		head := &stack.stk[len(stack.stk)-1]
		headWhere := head.where

		if stack.isMarked() {
			visitor.LeaveWhere(headWhere)
			stack.pop()
			i--
		} else {
			visitor.VisitWhere(head.position, headWhere)
			visitor.VisitPredicate(headWhere.Predicate)
			stack.mark()
			clauses := headWhere.Clauses
			clauseCount := len(clauses)
			for j := clauseCount - 1; j >= 0; j-- {
				next := whereMessageFrame{
					where:    clauses[j],
					position: j,
				}
				stack.push(next)
			}
			i += clauseCount
		}
	}
}

func (stack *whereMessageStack) pop() whereMessageFrame {
	head := stack.stk[len(stack.stk)-1]
	stack.stk = stack.stk[:len(stack.stk)-1]
	return head
}

func (stack *whereMessageStack) push(frame whereMessageFrame) {
	stack.stk = append(stack.stk, frame)
}

func (stack *whereMessageStack) mark() {
	head := &stack.stk[len(stack.stk)-1]
	head.mark = true
}

func (stack *whereMessageStack) isMarked() bool {
	return stack.stk[len(stack.stk)-1].mark
}

type whereMessageFrame struct {
	mark     bool
	position int
	where    *proto.QueryWhereMessage
}

func messageOpCodePanic(opCode uint32) {

}

const (
	MESSAGE_NOOP = uint32(iota)
	MESSAGE_SELECT
	MESSAGE_JOIN
)

const (
	MESSAGE_WHERE_NOOP = uint32(iota)
	MESSAGE_AND
	MESSAGE_OR
	MESSAGE_PREDICATE
)

const (
	MESSAGE_PREDICATE_NOOP = uint32(iota)
	MESSAGE_STR_EQ
	MESSAGE_STR_NEQ
)
