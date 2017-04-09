package godless

type QuerySelectVisitor struct {
	Namespace *IpfsNamespace
}

func (visitor *QuerySelectVisitor) RunQuery(kv KvQuery) {

}

func (visitor *QuerySelectVisitor) VisitOpCode(QueryOpCode) {
}

func (visitor *QuerySelectVisitor) VisitAST(*QueryAST) {
}

func (visitor *QuerySelectVisitor) VisitParser(*QueryParser) {
}

func (visitor *QuerySelectVisitor) VisitTableKey(string) {
}

func (visitor *QuerySelectVisitor) VisitJoin(*QueryJoin) {
}

func (visitor *QuerySelectVisitor) VisitRowJoin(int, *QueryRowJoin) {
}

func (visitor *QuerySelectVisitor) VisitSelect(*QuerySelect) {
}

func (visitor *QuerySelectVisitor) VisitWhere(int, *QueryWhere) {
}

func (visitor *QuerySelectVisitor) VisitWhereOpCode(QueryWhereOpCode) {
}

func (visitor *QuerySelectVisitor) LeaveWhere(*QueryWhere) {
}

func (visitor *QuerySelectVisitor) VisitPredicate(*QueryPredicate) {
}
