package godless

import (
	"github.com/pkg/errors"
)

type QueryOpCode uint8

const (
	SELECT = QueryOpCode(iota)
	JOIN
)

type Query struct {
	OpCode QueryOpCode
	TableKey string
	Update []Row
}

func ParseQuery(source string) (*Query, error) {
	return nil, errors.New("Not implemented")
}

func (query *Query) Run(kvq KvQuery, ns *IpfsNamespace) {

}
