package godless

import (
	"fmt"
	"strconv"
	"regexp"
	"github.com/pkg/errors"
)

type QueryAST struct {
	Command string
	TableKey string
	Select QuerySelectAST `json:",omitempty"`
	Join QueryJoinAST `json:",omitempty"`

	whereStack []*QueryWhereAST
	lastRowJoinKey string
	lastRowJoin *QueryRowJoinAST
}

func (ast *QueryAST) AddJoin() {
	ast.Command = "join"
}

func (ast *QueryAST) AddJoinRow() {
	row := &QueryRowJoinAST{
		Values: map[string]string{},
	}
	ast.Join.Rows = append(ast.Join.Rows, row)
	ast.lastRowJoin = row
}

func (ast *QueryAST) SetJoinRowKey(key string) {
	ast.lastRowJoin.RowKey = key
}

func (ast *QueryAST) SetJoinKey(key string) {
	ast.lastRowJoinKey = key
}

func (ast *QueryAST) SetJoinValue(value string) {
	ast.lastRowJoin.Values[ast.lastRowJoinKey] = value
}

func (ast *QueryAST) PushWhere() {
	where := &QueryWhereAST{}

	if len(ast.whereStack) == 0 {
		if ast.Select.Where != nil {
			panic("BUG invalid stack")
		}

		ast.Select.Where = where
	} else {
		lastWhere := ast.peekWhere()
		lastWhere.Clauses = append(lastWhere.Clauses, where)
	}

	ast.whereStack = append(ast.whereStack, where)
}

func (ast *QueryAST) PopWhere() {
	ast.whereStack = ast.whereStack[:len(ast.whereStack) - 1]
}

func (ast *QueryAST) SetWhereCommand(command string) {
	lastWhere := ast.peekWhere()
	lastWhere.Command = command
}

func (ast *QueryAST) peekWhere() *QueryWhereAST {
	if (len(ast.whereStack) == 0) {
		panic("BUG where stack empty!")
	}

	return ast.whereStack[len(ast.whereStack) - 1]
}

func (ast *QueryAST) InitPredicate() {
	where := ast.peekWhere()
	where.Command = "predicate"
	where.Predicate = &QueryPredicateAST{}
}

func (ast *QueryAST) UsePredicateRowKey() {
	where := ast.peekWhere()
	where.Predicate.IncludeRowKey = true
}

func (ast *QueryAST) AddPredicateKey(key string) {
	where := ast.peekWhere()
	where.Predicate.Keys = append(where.Predicate.Keys, key)
}

func (ast *QueryAST) AddPredicateLiteral(literal string) {
	where := ast.peekWhere()
	where.Predicate.Literals = append(where.Predicate.Literals, literal)
}

func (ast *QueryAST) SetPredicateCommand(command string) {
	where := ast.peekWhere()
	where.Predicate.Command = command
}

func (ast *QueryAST) AddSelect() {
	ast.Command = "select"
}

func (ast *QueryAST) SetTableName(key string) {
	ast.TableKey = key
}

func (ast *QueryAST) SetLimit(limit string) {
	ast.Select.Limit = limit
}

func (ast *QueryAST) Compile() (*Query, error) {
	query := &Query{}

	switch ast.Command {
	case "select":
		qselect, err := ast.Select.Compile()

		if err != nil {
			return nil, errors.Wrap(err, "BUG select compile failed")
		}

		query.OpCode = SELECT
		query.Select = qselect
	case "join":
		qjoin, err := ast.Join.Compile()

		if err != nil {
			return nil, errors.Wrap(err, "BUG join compile failed")
		}

		query.OpCode = JOIN
		query.Join = qjoin
	default:
		return nil, fmt.Errorf("BUG no command matching '%v'", ast.Command)
	}

	query.AST = ast
	query.TableKey = ast.TableKey

	return query, nil
}

type QueryJoinAST struct {
	Rows []*QueryRowJoinAST `json:",omitempty"`
}

func (ast *QueryJoinAST) Compile() (QueryJoin, error) {
	rows := make([]QueryRowJoin, len(ast.Rows))

	for i, r := range ast.Rows {
		unquoted, err := unquoteMap(r.Values)

		if err != nil {
			return QueryJoin{}, errors.Wrap(err, "Error compiling join")
		}

		rows[i] = QueryRowJoin{RowKey: r.RowKey, Entries: unquoted,}
	}

	qjoin := QueryJoin{
		Rows: rows,
	}

	return qjoin, nil
}


type QueryRowJoinAST struct {
	RowKey string
	Values map[string]string `json:",omitempty"`
}



type QuerySelectAST struct {
	Where *QueryWhereAST `json:",omitempty"`
	Limit string
}

func (ast *QuerySelectAST) Compile() (QuerySelect, error) {
	qselect := QuerySelect{}

	if (ast.Limit != "") {
		limit, converr := strconv.Atoi(ast.Limit)

		if converr != nil {
			return QuerySelect{}, errors.Wrap(converr, "BUG convert limit failed")
		}

		qselect.Limit = uint(limit)
	}

	if (ast.Where != nil) {
		where, err := ast.Where.Compile()

		if err != nil {
			return QuerySelect{}, errors.Wrap(err, "BUG where clause compile failed")
		}

		qselect.Where = where
	}

	return qselect, nil
}

type QueryWhereAST struct {
	Command string
	Clauses []*QueryWhereAST `json:",omitempty"`
	Predicate *QueryPredicateAST `json:",omitempty"`
}

func (ast *QueryWhereAST) Compile() (QueryWhere, error) {
	where := QueryWhere{}

	if ast.Command == "and" {
		clauses, err := ast.CompileClauses()

		if err != nil {
			return QueryWhere{}, errors.Wrap(err, "BUG and clause compile failed")
		}

		where.Clauses = clauses
		where.OpCode = AND
	} else if ast.Command == "or" {
		clauses, err := ast.CompileClauses()

		if err != nil {
			return QueryWhere{}, errors.Wrap(err, "BUG or clause compile failed")
		}

		where.Clauses = clauses
		where.OpCode = OR
	} else if ast.Command == "predicate" && ast.Predicate != nil {
		predicate, err := ast.Predicate.Compile()

		if err != nil {
			return QueryWhere{}, errors.Wrap(err, "BUG predicate compile failed")
		}

		where.OpCode = PREDICATE
		where.Predicate = predicate
	} else {
		return QueryWhere{}, fmt.Errorf("BUG unsupported where OpCode: '%v'", ast.Command)
	}

	return where, nil
}

func (ast *QueryWhereAST) CompileClauses() ([]QueryWhere, error) {
	clauses := make([]QueryWhere, len(ast.Clauses))

	for i, child := range ast.Clauses {
		clause, err := child.Compile()

		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("BUG failed compiling child clause '%v'", i))
		}

		clauses[i] = clause
	}

	return clauses, nil
}

type QueryPredicateAST struct {
	Command string
	Keys []string
	Literals []string
	IncludeRowKey bool
}

func (ast *QueryPredicateAST) Compile() (QueryPredicate, error) {
	predicate := QueryPredicate{}

	// TODO flesh out
	switch (ast.Command) {
	case "str_eq":
		predicate.OpCode = STR_EQ
	case "str_neq":
		predicate.OpCode = STR_NEQ
	default:
		return QueryPredicate{}, fmt.Errorf("BUG unsupported predicate '%v'", ast.Command)
	}

	literals, err := unquoteAll(ast.Literals)

	if err != nil {
		return QueryPredicate{}, errors.Wrap(err, "Error compiling predicate")
	}

	predicate.Keys = ast.Keys
	predicate.Literals = literals
	predicate.IncludeRowKey = ast.IncludeRowKey

	return predicate, nil
}

func unquoteAll(values []string) ([]string, error) {
	quoted := make([]string, len(values))

	for i, v := range values {
		q, err := unquote(v)

		if err != nil {
			return nil, err
		}

		quoted[i] = q
	}

	return quoted, nil
}

func unquoteMap(values map[string]string) (map[string]string, error) {
	quoted := map[string]string{}

	for k, v := range values {
		q, err := unquote(v)

		if err != nil {
			return nil, err
		}

		quoted[k] = q
	}

	return quoted, nil
}

func unquote(value string) (string, error) {
	regex, regerr := regexp.Compile("\\\\'")

	if regerr != nil {
		return "", errors.Wrap(regerr, "BUG regex compile failed in string unquote")
	}

	desingled := string(regex.ReplaceAll([]byte(value), []byte("'")))

	dquote := fmt.Sprintf("\"%v\"", desingled)
	quoted, err := strconv.Unquote(dquote)

	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Invalid string escape: '%v'", desingled))
	}

	return quoted, nil
}
