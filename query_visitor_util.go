package godless

type noSelectVisitor struct {}

func (visitor *noSelectVisitor) VisitSelect(*QuerySelect) {
}

func (visitor *noSelectVisitor) VisitWhere(int, *QueryWhere) {
}

func (visitor *noSelectVisitor) LeaveWhere(*QueryWhere) {
}

func (visitor *noSelectVisitor) VisitPredicate(*QueryPredicate) {
}

type noJoinVisitor struct {}

func (visitor *noJoinVisitor) VisitJoin(*QueryJoin) {
}

func (visitor *noJoinVisitor) VisitRowJoin(int, *QueryRowJoin) {
}


type noDebugVisitor struct {}

func (visitor *noDebugVisitor) VisitAST(*QueryAST) {
}

func (visitor *noDebugVisitor) VisitParser(*QueryParser) {
}

type errorCollectVisitor struct {
	err error
}

func (visitor *errorCollectVisitor) hasError() bool {
	return visitor.err != nil
}

func (visitor *errorCollectVisitor) collectError(err error) {
		visitor.err = err
}

func (visitor *errorCollectVisitor) reportError(kv KvQuery) {
	if visitor.err != nil {
		panic("No error too report")
	}

	kv.reportError(visitor.err)
}

type whereStack struct {
	stk []whereFrame
}

func makeWhereStack(where *QueryWhere) *whereStack{
	return 	&whereStack{
		stk: []whereFrame{whereFrame{where: where}},
	}
}

func (stack *whereStack) visit(visitor whereVisitor) {
	for i := 0; len(stack.stk) > 0; {
		tip := &stack.stk[len(stack.stk) - 1]
		tipWhere := tip.where

		if stack.isMarked() {
			visitor.LeaveWhere(tipWhere)
			stack.pop()
			i--
		} else {
			visitor.VisitWhere(tip.position, tipWhere)
			visitor.VisitPredicate(&tipWhere.Predicate)
			clauses := tipWhere.Clauses;
			clauseCount := len(clauses)
			for j := clauseCount - 1; j >= 0; j-- {
				next := whereFrame{
					where: &clauses[j],
					position: j,
				}
				stack.push(next)
			}
			i += clauseCount
			stack.mark()
		}
	}
}

func (stack *whereStack) pop() whereFrame {
	tip := stack.stk[len(stack.stk) - 1]
	stack.stk = stack.stk[:len(stack.stk) - 1]
	return tip
}

func (stack *whereStack) push(frame whereFrame) {
	stack.stk = append(stack.stk, frame)
}

func (stack *whereStack) mark() {
	tip := &stack.stk[len(stack.stk) - 1]
	tip.mark = true
}

func (stack *whereStack) isMarked() bool {
	return stack.stk[len(stack.stk) - 1].mark
}

type whereFrame struct {
	mark bool
	position int
	where *QueryWhere
}
