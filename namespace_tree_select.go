package godless

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

type NamespaceTreeSelect struct {
	noJoinVisitor
	noDebugVisitor
	errorCollectVisitor
	Namespace NamespaceTree
	crit      *rowCriteria
}

func MakeNamespaceTreeSelect(namespace NamespaceTree) *NamespaceTreeSelect {
	return &NamespaceTreeSelect{
		Namespace: namespace,
		crit: &rowCriteria{
			result: []NamespaceStreamEntry{},
		},
	}
}

func (visitor *NamespaceTreeSelect) RunQuery() APIResponse {
	fail := RESPONSE_FAIL
	fail.Type = API_QUERY

	err := visitor.Error()
	if err != nil {
		fail.Err = err
		return fail
	}

	if !visitor.crit.isReady() {
		panic("didn't visit query")
	}

	lambda := NamespaceTreeLambda(visitor.crit.selectMatching)
	tables := []TableName{visitor.crit.tableKey}
	tableReader := AddTableHints(tables, lambda)
	err = visitor.Namespace.LoadTraverse(tableReader)

	if err != nil {
		fail.Err = errors.Wrap(err, "NamespaceTreeSelect failed")
		return fail
	}

	response := RESPONSE_QUERY
	stream := visitor.crit.result
	sort.Sort(byNamespaceStreamOrder(stream))
	response.QueryResponse.Entries = stream
	return response
}

func (visitor *NamespaceTreeSelect) VisitTableKey(tableKey TableName) {
	if visitor.Error() != nil {
		return
	}

	visitor.crit.tableKey = tableKey
}

func (visitor *NamespaceTreeSelect) VisitOpCode(opCode QueryOpCode) {
	if opCode != SELECT {
		visitor.collectError(errors.New("Expected SELECT OpCode"))
	}
}

func (visitor *NamespaceTreeSelect) VisitSelect(qselect *QuerySelect) {
	if visitor.Error() != nil {
		return
	}

	if qselect.Limit <= 0 {
		visitor.collectError(errors.New("Invalid limit"))
		return
	}

	visitor.crit.limit = int(qselect.Limit)
	visitor.crit.rootWhere = &qselect.Where
}

func (visitor *NamespaceTreeSelect) LeaveSelect(*QuerySelect) {
}

func (visitor *NamespaceTreeSelect) VisitWhere(position int, where *QueryWhere) {
}

func (visitor *NamespaceTreeSelect) LeaveWhere(where *QueryWhere) {
}

func (visitor *NamespaceTreeSelect) VisitPredicate(predicate *QueryPredicate) {
}

type rowCriteria struct {
	tableKey  TableName
	count     int
	limit     int
	result    []NamespaceStreamEntry
	rootWhere *QueryWhere
}

func (crit *rowCriteria) selectMatching(namespace Namespace) (bool, error) {
	remaining := crit.limit - crit.count

	if remaining < 0 {
		panic("Selected over limit")
	}

	if remaining <= 0 {
		return true, nil
	}

	rows := crit.findRows(namespace)

	var slurp int
	if len(rows) <= remaining {
		slurp = len(rows)
	} else {
		slurp = remaining
	}

	crit.result = append(crit.result, rows[:slurp]...)
	crit.count = crit.count + slurp

	// logdbg("Found %v more results. Total: %v.  Limit: %v.", slurp, crit.count, crit.limit)
	return false, nil
}

func (crit *rowCriteria) findRows(namespace Namespace) []NamespaceStreamEntry {
	out := []NamespaceStreamEntry{}

	table, err := namespace.GetTable(crit.tableKey)

	if err != nil {
		return out
	}

	if crit.rootWhere.OpCode == WHERE_NOOP {
		return MakeTableStream(crit.tableKey, table)
	}

	table.Foreachrow(RowConsumerFunc(func(rowKey RowName, r Row) {
		eval := makeSelectEvalTree(rowKey, r)
		where := makeWhereStack(crit.rootWhere)

		if eval.evaluate(where) {
			stream := MakeRowStream(crit.tableKey, rowKey, r)
			out = append(out, stream...)
		}
	}))

	return out
}

func (crit *rowCriteria) isReady() bool {
	return crit.rootWhere != nil && crit.limit > 0 && crit.tableKey != ""
}

type selectEvalTree struct {
	rowKey RowName
	row    Row
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
	source   *QueryWhere
}

func makeSelectEvalTree(rowKey RowName, row Row) *selectEvalTree {
	return &selectEvalTree{
		rowKey: rowKey,
		row:    row,
	}
}

func (eval *selectEvalTree) evaluate(stk *whereStack) bool {
	stk.visit(eval)

	if len(eval.stk) != 0 {
		panic("Eval stack remains after visit")
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
func (eval *selectEvalTree) evalWhere(where *QueryWhere) *expr {
	switch where.OpCode {
	case AND:
		return &expr{state: EXPR_AND, source: where}
	case OR:
		return &expr{state: EXPR_OR, source: where}
	case PREDICATE:
		return eval.evalPred(where)
	default:
		panic(fmt.Sprintf("cannot evaluate where with OpCode: %v", where.OpCode))
	}
}

func (eval *selectEvalTree) evalPred(where *QueryWhere) *expr {
	pred := where.Predicate

	prefix := []string{}

	prefix = append(prefix, pred.Literals...)

	if pred.IncludeRowKey {
		prefix = append(prefix, string(eval.rowKey))
	}

	entries := []Entry{}

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
	case STR_EQ:
		isMatch = eval.str_eq(prefix, entries)
	case STR_NEQ:
		isMatch = eval.str_neq(prefix, entries)
	default:
		panic(fmt.Sprintf("Unsupported QueryPredicate OpCode: %v", pred.OpCode))
	}

	if isMatch {
		return &expr{source: where, state: EXPR_TRUE}
	}

	return &expr{source: where, state: EXPR_FALSE}
}

func (eval *selectEvalTree) str_eq(prefix []string, entries []Entry) bool {
	m, err := eval.matcher(prefix, entries)

	if err != nil {
		// Don't log this case.
		// logdbg("find matcher error")
		return false
	}

	for _, pfx := range prefix {
		if pfx != m {
			return false
		}
	}

	for _, entry := range entries {
		found := false
		for _, val := range entry.GetValues() {
			if string(val) == m {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (eval *selectEvalTree) str_neq(prefix []string, entries []Entry) bool {
	m, err := eval.matcher(prefix, entries)

	if err != nil {
		return false
	}

	pfxmatch := 0
	for _, pfx := range prefix {
		if pfx == m {
			pfxmatch++
		}
	}

	entrymatch := 0
	for _, entry := range entries {
		for _, val := range entry.GetValues() {
			if string(val) == m {
				entrymatch++
			}
		}
	}

	return !((pfxmatch > 0 && entrymatch > 0) || pfxmatch > 1 || entrymatch > 1)
}

// TODO need user concepts + crypto to narrow row match down.
func (eval *selectEvalTree) matcher(prefix []string, entries []Entry) (string, error) {
	var first string
	var found bool

	if len(prefix) > 0 {
		first = prefix[0]
		found = true
	} else {
		for _, entry := range entries {
			values := entry.GetValues()
			if len(values) > 0 {
				first = string(values[0])
				found = true
				break
			}
		}
	}

	// No values: no match.
	if !found {
		return "", errors.New("no match")
	}

	return first, nil
}

func (eval *selectEvalTree) VisitWhere(position int, where *QueryWhere) {
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

func (eval *selectEvalTree) LeaveWhere(where *QueryWhere) {
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

func (eval *selectEvalTree) VisitPredicate(*QueryPredicate) {
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
