package api

import (
	"fmt"
	"io"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/query"
)

type Request struct {
	Type       MessageType
	Reflection ReflectionType
	Query      *query.Query
	Replicate  []crdt.Link
}

func MakeQueryRequest(query *query.Query) Request {
	return Request{
		Type:  API_QUERY,
		Query: query,
	}
}

func MakeReflectRequest(reflection ReflectionType) Request {
	return Request{
		Type:       API_REFLECT,
		Reflection: reflection,
	}
}

func MakeReplicateRequest(replicate []crdt.Link) Request {
	return Request{
		Type:      API_REPLICATE,
		Replicate: replicate,
	}
}

func (request Request) MakeCommand() (Command, error) {
	switch request.Type {
	case API_QUERY:
		return makeApiQuery(request, coreQueryRunner{query: request.Query}), nil
	case API_REFLECT:
		return makeApiQuery(request, coreReflectRunner{reflection: request.Reflection}), nil
	case API_REPLICATE:
		return makeApiQuery(request, coreReplicator{links: request.Replicate}), nil
	default:
		return Command{}, fmt.Errorf("Invalid request.Type: %d", request.Type)
	}
}

func (request Request) Equals(other Request) bool {
	panic("not implemented")
}

func (request Request) Validate() error {
	panic("not implemented")
}

func EncodeRequest(request Request, w io.Writer) error {
	panic("not implemented")
}

func DecodeRequest(r io.Reader) (Request, error) {
	panic("not implemented")
}

type ReflectionType uint16

const (
	REFLECT_NOOP = ReflectionType(iota)
	REFLECT_HEAD_PATH
	REFLECT_DUMP_NAMESPACE
	REFLECT_INDEX
)

type MessageType uint8

const (
	API_MESSAGE_NOOP = MessageType(iota)
	API_QUERY
	API_REFLECT
	API_REPLICATE
)
