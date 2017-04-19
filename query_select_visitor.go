package godless

import (
	"fmt"

	"github.com/pkg/errors"
)

// Specific to IPFS.  We can add more visitors later.
type QuerySelectVisitor struct {
	noJoinVisitor
	noDebugVisitor
	errorCollectVisitor
	Namespace NamespaceTree
	crit *rowCriteria
}

func MakeQuerySelectVisitor(namespace NamespaceTree) *QuerySelectVisitor {
	return &QuerySelectVisitor{
		Namespace: namespace,
		crit: &rowCriteria{},
	}
}

func (visitor *QuerySelectVisitor) RunQuery(kv KvQuery) {
	if visitor.hasError() {
		visitor.reportError(kv)
		return
	}

	if !visitor.crit.isReady() {
		panic("didn't visit query")
	}

	lambda := NamespaceTreeLambda(visitor.crit.selectMatching)
	err := visitor.Namespace.LoadTraverse(lambda)

	if err != nil {
		visitor.collectError(err)
		visitor.reportError(kv)
	}
}

func (visitor *QuerySelectVisitor) VisitTableKey(tableKey string) {
	if visitor.hasError() {
		return
	}

	visitor.crit.tableKey = tableKey
}

func (visitor *QuerySelectVisitor) VisitOpCode(opCode QueryOpCode) {
	if opCode != SELECT {
		panic("Expected SELECT OpCode")
	}
}

func (visitor *QuerySelectVisitor) VisitSelect(qselect *QuerySelect) {
	if visitor.hasError() {
		return
	}

	if qselect.Limit <= 0 {
		visitor.collectError(errors.New("Invalid limit"))
		return
	}

	visitor.crit.limit = int(qselect.Limit)
	visitor.crit.rootWhere = &qselect.Where
}

func (visitor *QuerySelectVisitor) VisitWhere(position int, where *QueryWhere) {
}

func (visitor *QuerySelectVisitor) LeaveWhere(where *QueryWhere) {
}

func (visitor *QuerySelectVisitor) VisitPredicate(predicate *QueryPredicate) {
}

type rowCriteria struct {
	tableKey string
	count int
	limit int
	result []Row
	rootWhere *QueryWhere
}

func (crit *rowCriteria) selectMatching(namespace *Namespace) (bool, error) {
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
	return false, nil
}

func (crit *rowCriteria) findRows(namespace *Namespace) []Row {
	out := []Row{}

	table, present := namespace.Tables[crit.tableKey]

	if !present {
		return out
	}

	table.Foreachrow(func (rowKey string, r Row) {
		eval := makeSelectEvalTree(rowKey, r)
		where := makeWhereStack(crit.rootWhere)

		if eval.evaluate(where) {
			out = append(out, r)
		}
	})

	return out
}

func (crit *rowCriteria) isReady() bool {
	return crit.rootWhere != nil && crit.limit > 0
}

type selectEvalTree struct {
	rowKey string
	row Row
	root *expr
	stk []*expr
}

type exprOpCode uint8

const (
	EXPR_AND = exprOpCode(iota)
	EXPR_OR
	EXPR_TRUE
	EXPR_FALSE
)

type expr struct {
	state exprOpCode
	children []*expr
	source *QueryWhere
}

func makeSelectEvalTree(rowKey string, row Row) *selectEvalTree {
	return &selectEvalTree{
		rowKey: rowKey,
		row: row,
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

	switch eval.root.state {
	case EXPR_TRUE:
		return true
	case EXPR_FALSE:
		return false
	default:
		panic("Unevaluated eval root")
	}
}

// TODO shortcircuit eval optimisation.
func (eval *selectEvalTree) evalWhere(where *QueryWhere) *expr {
	switch where.OpCode {
	case AND:
		return &expr{state: EXPR_AND}
	case OR:
		return &expr{state: EXPR_OR}
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
		prefix = append(prefix, eval.rowKey)
	}

	entries := [][]string{}

	for _, key := range pred.Keys {
		values, present := eval.row.Entries[key]

		if present {
			entries = append(entries, values)
		} else {
			// No key = no match.
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

func (eval *selectEvalTree) str_eq(prefix []string, entries [][]string) bool {
	m, err := eval.matcher(prefix, entries)

	if err != nil {
		// TODO log error?
		return false
	}

	for _, pfx := range prefix {
		if pfx != m {
			return false
		}
	}

	for _, entry := range entries {
		found := false
		for _, val := range entry {
			if val == m {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true;
}

func (eval *selectEvalTree) str_neq(prefix []string, entries [][]string) bool {
	m, err := eval.matcher(prefix, entries)

	if err != nil {
		return false
	}

	neqprefix := false
	for _, pfx := range prefix {
		if pfx != m {
			neqprefix = true
		}
	}

	neqentries := false
	for _, entry := range entries {
		for _, val := range entry {
			if val != m {
				neqentries = true
				break
			}
		}
	}

	return neqprefix || neqentries
}

// TODO need user concepts + crypto to narrow row match down.
func (eval *selectEvalTree) matcher(prefix []string, entries [][]string) (string, error) {
	var first string
	var found bool

	if len(prefix) > 0 {
		first = prefix[0]
		found = true
	} else {
		for _, entry := range entries {
			if len(entry) > 0 {
				first = entry[0]
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
	}
}

func (eval *selectEvalTree) LeaveWhere(where *QueryWhere) {
	head := eval.pop()

	if head.source != where {
		panic(fmt.Sprintf("expr stack corruption, 'head' %v but 'where' %v", head, where))
	}

	switch head.state {
	case EXPR_AND:
		for _, child := range head.children {
			switch child.state {
			case EXPR_TRUE:
				// Do nothing.
			case EXPR_FALSE:
				head.state = EXPR_FALSE
			default:
				panic(fmt.Sprintf("Unevaluated expr: %v", child))
			}
		}
	case EXPR_OR:
		for _, child := range head.children {
			switch child.state {
			case EXPR_TRUE:
				head.state = EXPR_TRUE
				break
			case EXPR_FALSE:
				// Do nothing;
			default:
				panic(fmt.Sprintf("Unevaluated expr: %v", child))
			}
		}
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
	return eval.stk[len(eval.stk) - 1]
}

func (eval *selectEvalTree) pop() *expr {
	head := eval.peek()
	eval.stk = eval.stk[:len(eval.stk) - 1]
	return head
}

func (eval *selectEvalTree) push(e *expr) {
	eval.stk = append(eval.stk, e)
}
