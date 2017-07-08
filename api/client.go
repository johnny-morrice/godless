package api

import (
	"github.com/johnny-morrice/godless/query"
)

type Client interface {
	SendReflection(command ReflectionType) (Response, error)
	SendQuery(q *query.Query) (Response, error)
}
