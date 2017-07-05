package api

import (
	"github.com/johnny-morrice/godless/query"
)

type Client interface {
	SendReflection(command APIReflectionType) (APIResponse, error)
	SendQuery(q *query.Query) (APIResponse, error)
}
