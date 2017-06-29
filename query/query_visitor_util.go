package query

import (
	"fmt"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/pkg/errors"
)

type NoSelectVisitor struct{}

func (visitor *NoSelectVisitor) VisitSelect(*QuerySelect) {
}

func (visitor *NoSelectVisitor) LeaveSelect(*QuerySelect) {
}

func (visitor *NoSelectVisitor) VisitWhere(int, *QueryWhere) {
}

func (visitor *NoSelectVisitor) LeaveWhere(*QueryWhere) {
}

func (visitor *NoSelectVisitor) VisitPredicate(*QueryPredicate) {
}

type NoJoinVisitor struct{}

func (visitor *NoJoinVisitor) VisitJoin(*QueryJoin) {
}

func (visitor *NoJoinVisitor) LeaveJoin(*QueryJoin) {
}

func (visitor *NoJoinVisitor) VisitRowJoin(int, *QueryRowJoin) {
}

type NoDebugVisitor struct{}

func (visitor *NoDebugVisitor) VisitAST(*QueryAST) {
}

func (visitor *NoDebugVisitor) VisitParser(*QueryParser) {
}

type ErrorCollectVisitor struct {
	err error
}

func (visitor *ErrorCollectVisitor) BadPublicKey(hash crypto.PublicKeyHash) {
	visitor.CollectError(fmt.Errorf("Bad PublicKeyHash: %v", hash))
}

func (visitor *ErrorCollectVisitor) BadOpcode(opCode QueryOpCode) {
	visitor.CollectError(fmt.Errorf("Unknown Query OpCode: %v", opCode))
}

func (visitor *ErrorCollectVisitor) badTableName(table crdt.TableName) {
	var err error
	if table == "" {
		err = errors.New("Empty table key")
	} else {
		err = fmt.Errorf("Bad table name: %v", table)
	}

	visitor.CollectError(err)
}

func (visitor *ErrorCollectVisitor) BadWhereOpCode(position int, where *QueryWhere) {
	err := fmt.Errorf("Unknown Where OpCode at position %v: %v", position, where)
	visitor.CollectError(err)
}

func (visitor *ErrorCollectVisitor) BadPredicateOpCode(predicate *QueryPredicate) {
	err := fmt.Errorf("Unknown Predicate OpCode: %v", predicate)
	visitor.CollectError(err)
}

func (visitor *ErrorCollectVisitor) CollectError(err error) {
	if visitor.err == nil {
		visitor.err = err
	} else {
		visitor.err = errors.Wrapf(err, "%v, and", visitor.err)
	}

}

func (visitor *ErrorCollectVisitor) Error() error {
	return visitor.err
}

type WhereStack struct {
	stk []whereFrame
}

func MakeWhereStack(where *QueryWhere) *WhereStack {
	return &WhereStack{
		stk: []whereFrame{whereFrame{where: where}},
	}
}

func (stack *WhereStack) Visit(visitor whereVisitor) {
	for i := 0; len(stack.stk) > 0; {
		head := &stack.stk[len(stack.stk)-1]
		headWhere := head.where

		if stack.isMarked() {
			visitor.LeaveWhere(headWhere)
			stack.pop()
			i--
		} else {
			visitor.VisitWhere(head.position, headWhere)
			visitor.VisitPredicate(&headWhere.Predicate)
			stack.mark()
			clauses := headWhere.Clauses
			clauseCount := len(clauses)
			for j := clauseCount - 1; j >= 0; j-- {
				next := whereFrame{
					where:    &clauses[j],
					position: j,
				}
				stack.push(next)
			}
			i += clauseCount
		}
	}
}

func (stack *WhereStack) pop() whereFrame {
	head := stack.stk[len(stack.stk)-1]
	stack.stk = stack.stk[:len(stack.stk)-1]
	return head
}

func (stack *WhereStack) push(frame whereFrame) {
	stack.stk = append(stack.stk, frame)
}

func (stack *WhereStack) mark() {
	head := &stack.stk[len(stack.stk)-1]
	head.mark = true
}

func (stack *WhereStack) isMarked() bool {
	return stack.stk[len(stack.stk)-1].mark
}

type whereFrame struct {
	mark     bool
	position int
	where    *QueryWhere
}
