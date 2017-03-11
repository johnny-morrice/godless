package godless

import (
	"github.com/pkg/errors"
)

type QueryOpCode uint16

const (
	QUERY_NOP = QueryOpCode(iota)
	SELECT
	JOIN
)

type Query struct {
	OpCode QueryOpCode
	Update QueryUpdate
	Select QuerySelect
}

type QueryUpdate struct {
	TableKey string
	Rows []Row
}

type QuerySelect struct {
	TableKey string
	RowKeys []string
	Where QueryWhere
	Limit uint
}

type QueryWhereOpCode uint16

const (
	WHERE_NOP = QueryOpCode(iota)
	AND
	OR
	PREDICATE
)

type QueryWhere struct {
	OpCode QueryOpCode
	Clauses []QueryWhere
	Predicate QueryPredicate
}

type QueryPredicateOpCode uint16

const (
	PREDICATE_NOP = QueryPredicateOpCode(iota)
	STR_EQ
	STR_NEQ
	STR_EMPTY
	STR_NEMPTY
	// TODO flesh these out
	// STR_GT
	// STR_LT
	// STR_GTE
	// STR_LTE
	// NUM_EQ
	// NUM_NEQ
	// NUM_GT
	// NUM_LT
	// NUM_GTE
	// NUM_LTE
	// TIME_EQ
	// TIME_NEQ
	// TIME_GT
	// TIME_LT
	// TIME_GTE
	// TIME_LTE
)

type QueryPredicate struct {
	OpCode QueryPredicateOpCode
	Keys []string
	Literals []string
}

func ParseQuery(source string) (*Query, error) {
	return nil, errors.New("Not implemented")
}

func (query *Query) Analyse() string {
	return ""
}

func (query *Query) Run(kvq KvQuery, ns *IpfsNamespace) {

}
