package eval

import (
	"fmt"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

type NamespaceTreeSelect struct {
	query.NoJoinVisitor
	query.NoDebugVisitor
	query.ErrorCollectVisitor
	Namespace          api.RemoteNamespace
	crit               *rowCriteria
	keys               []crypto.PublicKey
	keyStore           api.KeyStore
	namespaceLoadError bool
	indexLoadError     bool
}

func MakeNamespaceTreeSelect(namespace api.RemoteNamespace, keyStore api.KeyStore) *NamespaceTreeSelect {
	return &NamespaceTreeSelect{
		Namespace: namespace,
		crit: &rowCriteria{
			result: []crdt.NamespaceStreamEntry{},
		},
		keys:     []crypto.PublicKey{},
		keyStore: keyStore,
	}
}

func (visitor *NamespaceTreeSelect) RunQuery() api.Response {
	const failMsg = "NamespaceTreeSelect.RunQuery failed"

	fail := api.RESPONSE_FAIL
	fail.Type = api.API_QUERY

	visitErr := visitor.Error()
	if visitErr != nil {
		fail.Err = visitErr
		return fail
	}

	if !visitor.crit.isReady() {
		panic("didn't visit query")
	}

	log.Info("Searching namespaces...")

	searcher := api.SignedTableSearcher{
		Reader: api.SearchResultLambda(visitor.ReadSearchResult),
		Tables: []crdt.TableName{visitor.crit.tableKey},
	}
	searchErr := visitor.Namespace.LoadTraverse(searcher)

	if searchErr != nil {
		fail.Err = errors.Wrap(searchErr, failMsg)
		return fail
	}

	if visitor.indexLoadError {
		fail.Err = errors.New("Index load failure")
		return fail
	}

	log.Info("Search complete")

	response := api.RESPONSE_QUERY

	namespace := visitor.getSelectResults()

	if visitor.namespaceLoadError {
		response.Msg = "ok with load errors"
	}

	response.Namespace = namespace
	return response
}

func (visitor *NamespaceTreeSelect) ReadSearchResult(result api.SearchResult) api.TraversalUpdate {
	if result.NamespaceLoadFailure {
		visitor.namespaceLoadError = true
		return api.TraversalUpdate{More: true}
	}

	if result.IndexLoadFailure {
		visitor.indexLoadError = true
		return api.TraversalUpdate{}
	}

	verified := visitor.filterVerified(result.Namespace)

	return visitor.crit.selectMatching(verified)
}

func (visitor *NamespaceTreeSelect) getSelectResults() crdt.Namespace {
	const failMsg = "NamespaceTreeSelect.resultStream failed"

	stream := visitor.crit.result
	namespace, invalid := crdt.ReadNamespaceStream(stream)
	visitor.logInvalid(invalid)

	return namespace
}

func (visitor *NamespaceTreeSelect) filterVerified(namespace crdt.Namespace) crdt.Namespace {
	if visitor.needsSignature() {
		log.Info("Filtering results by public key...")
		namespace = namespace.FilterVerified(visitor.keys)
		log.Info("Filtering complete")
	}

	namespace, invalid := namespace.Strip()

	visitor.logInvalid(invalid)

	return namespace
}

func (visitor *NamespaceTreeSelect) logInvalid(invalid []crdt.InvalidNamespaceEntry) {
	invalidCount := len(invalid)
	if invalidCount > 0 {
		log.Error("NamespaceTreeSelect found %d invalidEntries", invalidCount)
	}
}

func (visitor *NamespaceTreeSelect) needsSignature() bool {
	return len(visitor.keys) > 0
}

func (visitor *NamespaceTreeSelect) VisitPublicKeyHash(hash crypto.PublicKeyHash) {
	pub, err := visitor.keyStore.GetPublicKey(hash)

	if err != nil {
		log.Warn("Public key lookup failed with: %s", err.Error())
		visitor.BadPublicKey(hash)
		return
	}

	visitor.keys = append(visitor.keys, pub)
}

func (visitor *NamespaceTreeSelect) VisitTableKey(tableKey crdt.TableName) {
	if visitor.Error() != nil {
		return
	}

	visitor.crit.tableKey = tableKey
}

func (visitor *NamespaceTreeSelect) VisitOpCode(opCode query.QueryOpCode) {
	if opCode != query.SELECT {
		visitor.CollectError(errors.New("Expected query OpCode"))
	}
}

func (visitor *NamespaceTreeSelect) VisitSelect(qselect *query.QuerySelect) {
	if visitor.Error() != nil {
		return
	}

	if qselect.Limit < 0 {
		visitor.CollectError(errors.New("Invalid limit"))
		return
	}

	visitor.crit.limit = int(qselect.Limit)

	visitor.crit.rootWhere = &qselect.Where
}

func (visitor *NamespaceTreeSelect) LeaveSelect(*query.QuerySelect) {
}

func (visitor *NamespaceTreeSelect) VisitWhere(position int, where *query.QueryWhere) {
}

func (visitor *NamespaceTreeSelect) LeaveWhere(where *query.QueryWhere) {
}

func (visitor *NamespaceTreeSelect) VisitPredicate(predicate *query.QueryPredicate) {
}

type rowCriteria struct {
	tableKey  crdt.TableName
	count     int
	limit     int
	result    []crdt.NamespaceStreamEntry
	rootWhere *query.QueryWhere
}

func (crit *rowCriteria) selectMatching(namespace crdt.Namespace) api.TraversalUpdate {
	if crit.limit > 0 {
		return crit.selectToLimit(namespace)
	}

	rows := crit.findRows(namespace)
	crit.appendResult(rows)
	return api.TraversalUpdate{More: true}
}

func (crit *rowCriteria) selectToLimit(namespace crdt.Namespace) api.TraversalUpdate {
	remaining := crit.limit - crit.count

	if remaining < 0 {
		panic("Selected over limit")
	}

	if remaining <= 0 {
		return api.TraversalUpdate{}
	}

	rows := crit.findRows(namespace)

	var slurp int
	if len(rows) <= remaining {
		slurp = len(rows)
	} else {
		slurp = remaining
	}

	crit.appendResult(rows[:slurp])

	// logdbg("Found %v more results. Total: %v.  Limit: %v.", slurp, crit.count, crit.limit)
	return api.TraversalUpdate{More: true}
}

func (crit *rowCriteria) appendResult(stream []crdt.NamespaceStreamEntry) {
	crit.result = append(crit.result, stream...)
	crit.count = crit.count + len(stream)
}

func (crit *rowCriteria) findRows(namespace crdt.Namespace) []crdt.NamespaceStreamEntry {
	out := []crdt.NamespaceStreamEntry{}
	invalidEntries := []crdt.InvalidNamespaceEntry{}

	table, err := namespace.GetTable(crit.tableKey)

	if err != nil {
		return out
	}

	if crit.rootWhere.OpCode == query.WHERE_NOOP {
		stream, invalid := crdt.MakeTableStream(crit.tableKey, table)
		crit.logInvalid(invalid)
		return stream
	}

	table.ForeachRow(func(rowKey crdt.RowName, r crdt.Row) {
		eval := makeSelectEvalTree(rowKey, r)
		where := query.MakeWhereStack(crit.rootWhere)

		if eval.evaluate(where) {
			stream, invalid := crdt.MakeRowStream(crit.tableKey, rowKey, r)
			out = append(out, stream...)
			invalidEntries = append(invalidEntries, invalid...)
		}
	})

	crit.logInvalid(invalidEntries)

	return out
}

func (crit *rowCriteria) logInvalid(invalid []crdt.InvalidNamespaceEntry) {
	invalidCount := len(invalid)

	if invalidCount > 0 {
		log.Error("rowCriteria found %d invalid entries", invalidCount)
	}
}

func (crit *rowCriteria) isReady() bool {
	return crit.rootWhere != nil && crit.tableKey != ""
}

type selectEvalTree struct {
	rowKey crdt.RowName
	row    crdt.Row
	root   *expr
	stk    []*expr
}

type exprOpCode uint8

const (
	EXPR_AND = exprOpCode(iota)
	EXPR_OR
	EXPR_TRUE
	EXPR_FALSE
)

type expr struct {
	state    exprOpCode
	children []*expr
	source   *query.QueryWhere
}

func makeSelectEvalTree(rowKey crdt.RowName, row crdt.Row) *selectEvalTree {
	return &selectEvalTree{
		rowKey: rowKey,
		row:    row,
	}
}

func (eval *selectEvalTree) evaluate(stk *query.WhereStack) bool {
	stk.Visit(eval)

	if len(eval.stk) != 0 {
		panic("Eval stack non-empty after visit")
	}

	if eval.root == nil {
		panic("No root for eval stack")
	}

	// logdbg("eval tree: %v", eval.root)

	switch eval.root.state {
	case EXPR_TRUE:
		return true
	case EXPR_FALSE:
		return false
	default:
		panic(fmt.Sprintf("Unevaluated eval root: %v", eval.root))
	}
}

// TODO shortcircuit eval optimisation.
func (eval *selectEvalTree) evalWhere(where *query.QueryWhere) *expr {
	switch where.OpCode {
	case query.AND:
		return &expr{state: EXPR_AND, source: where}
	case query.OR:
		return &expr{state: EXPR_OR, source: where}
	case query.PREDICATE:
		return eval.evalPred(where)
	default:
		panic(fmt.Sprintf("cannot evaluate where with OpCode: %v", where.OpCode))
	}
}

func (eval *selectEvalTree) evalPred(where *query.QueryWhere) *expr {
	pred := where.Predicate

	literals := pred.Literals

	if pred.IncludeRowKey {
		literals = append(literals, string(eval.rowKey))
	}

	entries := []crdt.Entry{}

	for _, key := range pred.Keys {
		more, err := eval.row.GetEntry(key)

		if err == nil {
			// logdbg("entry %v: %v", key, more)
			entries = append(entries, more)
		} else {
			// No key = no match.
			// logdbg("no key = no match for pred %v", pred)
			return &expr{source: where, state: EXPR_FALSE}
		}
	}

	var isMatch bool
	switch pred.OpCode {
	case query.STR_EQ:
		isMatch = StrEq{}.Match(literals, entries)
	default:
		panic(fmt.Sprintf("Unsupported query.QueryPredicate OpCode: %v", pred.OpCode))
	}

	if isMatch {
		return &expr{source: where, state: EXPR_TRUE}
	}

	return &expr{source: where, state: EXPR_FALSE}
}

func (eval *selectEvalTree) VisitWhere(position int, where *query.QueryWhere) {
	e := eval.evalWhere(where)
	if eval.root == nil {
		// Root where
		eval.root = e
		eval.stk = []*expr{eval.root}
	} else {
		head := eval.peek()
		head.children = append(head.children, e)
		eval.stk = append(eval.stk, e)
	}
}

func (eval *selectEvalTree) LeaveWhere(where *query.QueryWhere) {
	head := eval.pop()

	if head.source != where {
		panic(fmt.Sprintf("expr stack corruption, 'head' %v but 'where' %v", head.source, where))
	}

	// TODO dupe code.
	switch head.state {
	case EXPR_AND:
		// TODO should this be invalid?
		if len(head.children) == 0 {
			head.state = EXPR_FALSE
			break
		}

		for _, child := range head.children {
			switch child.state {
			case EXPR_TRUE:
				// Do nothing.
			case EXPR_FALSE:
				head.state = EXPR_FALSE
				return
			default:
				panic(fmt.Sprintf("Unevaluated expr: %v", child))
			}
		}
		head.state = EXPR_TRUE
	case EXPR_OR:
		if len(head.children) == 0 {
			head.state = EXPR_FALSE
			break
		}

		for _, child := range head.children {
			switch child.state {
			case EXPR_TRUE:
				head.state = EXPR_TRUE
				return
			case EXPR_FALSE:
				// Do nothing;
			default:
				panic(fmt.Sprintf("Unevaluated expr: %v", child))
			}
		}
		head.state = EXPR_FALSE
	case EXPR_TRUE:
	case EXPR_FALSE:
		// Do nothing
	default:
		panic(fmt.Sprintf("Unknown state for expr: %v", head))
	}
}

func (eval *selectEvalTree) VisitPredicate(*query.QueryPredicate) {
	// Do nothing.
}

func (eval *selectEvalTree) peek() *expr {
	return eval.stk[len(eval.stk)-1]
}

func (eval *selectEvalTree) pop() *expr {
	head := eval.peek()
	eval.stk = eval.stk[:len(eval.stk)-1]
	return head
}

func (eval *selectEvalTree) push(e *expr) {
	eval.stk = append(eval.stk, e)
}
