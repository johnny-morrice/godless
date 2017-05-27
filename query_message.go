package godless

import "fmt"

type whereMessageVisitor interface {
	VisitWhere(int, *QueryWhereMessage)
	LeaveWhere(*QueryWhereMessage)
	VisitPredicate(*QueryPredicateMessage)
}

type queryMessageVisitor interface {
	whereMessageVisitor
	VisitOpCode(uint32)
	VisitTableKey(string)
	VisitJoin(*QueryJoinMessage)
	LeaveJoin(*QueryJoinMessage)
	VisitRowJoin(int, *QueryRowJoinMessage)
	VisitSelect(*QuerySelectMessage)
	LeaveSelect(*QuerySelectMessage)
}

type whereMessageBuilder struct {
	message *QueryWhereMessage
	stack   whereBuilderFrameStack
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
	message *QueryWhereMessage
	where   *QueryWhere
}

func configureWhereMessage(queryWhere *QueryWhere, message *QueryWhereMessage) {
	message.OpCode = uint32(queryWhere.OpCode)
	message.Predicate = MakeQueryPredicateMessage(queryWhere.Predicate)
	message.Clauses = make([]*QueryWhereMessage, len(queryWhere.Clauses))
}

func MakeQueryMessage(query *Query) *QueryMessage {
	return &QueryMessage{
		OpCode: uint32(query.OpCode),
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
	builder.stack = makeWhereBuilderFrameStack()
	whereStack := makeWhereStack(&queryWhere)
	whereStack.visit(builder)

	return builder.message
}

func ReadQueryMessage(message *QueryMessage) (*Query, error) {
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

func MakeQueryPredicateMessage(predicate QueryPredicate) *QueryPredicateMessage {
	message := &QueryPredicateMessage{
		OpCode:   uint32(predicate.OpCode),
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

type queryMessageDecoder struct {
	errorCollectVisitor
	Query *Query
	stack whereBuilderFrameStack
}

func makeQueryMessageDecoder() *queryMessageDecoder {
	return &queryMessageDecoder{
		Query: &Query{},
		stack: makeWhereBuilderFrameStack(),
	}
}

func (decoder *queryMessageDecoder) VisitWhere(position int, message *QueryWhereMessage) {
	where := decoder.createChildWhere()

	frame := whereBuilderFrame{
		message: message,
		where:   where,
	}

	decoder.configureQueryWhere(frame)
	decoder.stack.push(frame)
}

func (decoder *queryMessageDecoder) createChildWhere() *QueryWhere {
	if len(decoder.stack.stk) == 0 {
		return &QueryWhere{}
	}

	tip := decoder.stack.peek()
	tip.where.Clauses = append(tip.where.Clauses, QueryWhere{})
	return &tip.where.Clauses[len(tip.where.Clauses)-1]
}

func (decoder *queryMessageDecoder) LeaveWhere(message *QueryWhereMessage) {
	frame := decoder.stack.pop()
	if frame.message != message {
		panic("queryMessageDecoder.stack corruption")
	}
}

func (decoder *queryMessageDecoder) VisitPredicate(*QueryPredicateMessage) {
}

func (decoder *queryMessageDecoder) VisitOpCode(uint32) {
}

func (decoder *queryMessageDecoder) VisitTableKey(string) {
}

func (decoder *queryMessageDecoder) VisitJoin(*QueryJoinMessage) {
}

func (decoder *queryMessageDecoder) LeaveJoin(*QueryJoinMessage) {
}

func (decoder *queryMessageDecoder) VisitRowJoin(int, *QueryRowJoinMessage) {
}

func (decoder *queryMessageDecoder) VisitSelect(*QuerySelectMessage) {
}

func (decoder *queryMessageDecoder) LeaveSelect(*QuerySelectMessage) {
}

func (decoder *queryMessageDecoder) configureQueryWhere(frame whereBuilderFrame) {
	msg := frame.message
	where := frame.where
	switch msg.OpCode {
	case MESSAGE_AND:
		fallthrough
	case MESSAGE_OR:
		fallthrough
	case MESSAGE_PREDICATE:
		where.OpCode = QueryWhereOpCode(msg.OpCode)
	default:
		decoder.badWhereMessageOpCode(msg)
	}
}

func (decoder *queryMessageDecoder) badWhereMessageOpCode(message *QueryWhereMessage) {
	err := fmt.Errorf("Bad queryWhereMessageOpCode: %v", message)
	decoder.collectError(err)
}

func visitMessage(message *QueryMessage, visitor queryMessageVisitor) error {
	visitor.VisitOpCode(message.OpCode)
	visitor.VisitTableKey(message.Table)

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

func makeWhereMessageStack(where *QueryWhereMessage) *whereMessageStack {
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
	where    *QueryWhereMessage
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
