package godless

type QueryJoinVisitor struct {
	Namespace *IpfsNamespace
}

func (visitor *QueryJoinVisitor) WriteQueryResult(kv KvQuery) {

}

func (visitor *QueryJoinVisitor) VisitOpCode(QueryOpCode) {
}

func (visitor *QueryJoinVisitor) VisitAST(*QueryAST) {
}

func (visitor *QueryJoinVisitor) VisitParser(*QueryParser) {
}

func (visitor *QueryJoinVisitor) VisitTableKey(string) {
}

func (visitor *QueryJoinVisitor) VisitJoin(*QueryJoin) {
}

func (visitor *QueryJoinVisitor) VisitRowJoin(int, *QueryRowJoin) {
}

func (visitor *QueryJoinVisitor) VisitSelect(*QuerySelect) {
}

func (visitor *QueryJoinVisitor) VisitWhere(int, *QueryWhere) {
}

func (visitor *QueryJoinVisitor) VisitWhereOpCode(QueryWhereOpCode) {
}

func (visitor *QueryJoinVisitor) LeaveWhere(*QueryWhere) {
}

func (visitor *QueryJoinVisitor) VisitPredicate(*QueryPredicate) {
}
