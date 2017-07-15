package api

import (
	"fmt"
	"io"

	"math/rand"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/proto"
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
	ok := request.Type == other.Type
	ok = ok && request.Reflection == other.Reflection
	ok = ok && len(request.Replicate) == len(other.Replicate)
	ok = ok && (request.Query == nil) == (other.Query == nil)

	if !ok {
		return false
	}

	for i, myLink := range request.Replicate {
		otherLink := other.Replicate[i]

		if !myLink.Equals(otherLink) {
			return false
		}
	}

	if request.Query != nil {
		return request.Query.Equals(other.Query)
	}

	return true
}

func (request Request) Validate() error {
	switch request.Type {
	case API_QUERY:
		return request.validateQuery()
	case API_REFLECT:
		return request.validateReflect()
	case API_REPLICATE:
		return request.validateReplicate()
	default:
		return fmt.Errorf("Invalid MessageType: %v", request.Type)
	}

	return nil
}

func (request Request) validateQuery() error {
	const failMsg = "Request.validateQuery failed"

	if request.Query == nil {
		return fmt.Errorf("Query was nil")
	}

	err := request.Query.Validate()

	return errors.Wrap(err, failMsg)
}

func (request Request) validateReflect() error {
	switch request.Reflection {
	case REFLECT_HEAD_PATH:
	case REFLECT_DUMP_NAMESPACE:
	case REFLECT_INDEX:
	default:
		return fmt.Errorf("Invalid ReflectionType: %v", request.Reflection)
	}

	return nil
}

func (request Request) validateReplicate() error {
	if len(request.Replicate) == 0 {
		return fmt.Errorf("No replication links")
	}

	return nil
}

func EncodeRequest(request Request, w io.Writer) error {
	const failMsg = "EncodeRequest failed"

	message := MakeRequestMessage(request)

	err := util.Encode(message, w)

	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	return nil
}

func DecodeRequest(r io.Reader) (Request, error) {
	const failMsg = "DecodeRequest failed"

	message := &proto.APIRequestMessage{}

	err := util.Decode(message, r)

	if err != nil {
		return Request{}, errors.Wrap(err, failMsg)
	}

	return ReadRequestMessage(message), nil
}

func GenRequest(rand *rand.Rand, size int) Request {
	gen := Request{}

	chooseType := rand.Float32()

	if chooseType < 0.3 {
		generateQueryRequest(rand, size, &gen)
	} else if chooseType < 0.6 {
		generateReflectRequest(rand, size, &gen)
	} else {
		generateReplicateRequest(rand, size, &gen)
	}

	return gen
}

func generateQueryRequest(rand *rand.Rand, size int, gen *Request) {
	gen.Type = API_QUERY
	gen.Query = query.GenQuery(rand, size)
}

func generateReflectRequest(rand *rand.Rand, size int, gen *Request) {
	gen.Type = API_REFLECT

	chooseType := rand.Float32()

	if chooseType < 0.3 {
		gen.Reflection = REFLECT_HEAD_PATH
	} else if chooseType < 0.6 {
		gen.Reflection = REFLECT_INDEX
	} else {
		gen.Reflection = REFLECT_DUMP_NAMESPACE
	}
}

func generateReplicateRequest(rand *rand.Rand, size int, gen *Request) {
	gen.Type = API_REPLICATE

	gen.Replicate = make([]crdt.Link, size)

	for i := 0; i < size; i++ {
		gen.Replicate[i] = crdt.GenLink(rand, size)
	}
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
