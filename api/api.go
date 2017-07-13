package api

//go:generate mockgen -package mock_godless -destination ../mock/mock_api.go -imports lib=github.com/johnny-morrice/api -self_package lib github.com/johnny-morrice/godless/api Core,RemoteStore,RemoteNamespace,NamespaceSearcher,DataPeer,PubSubSubscription,PubSubRecord

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/query"
)

type Core interface {
	RunQuery(*query.Query, Command)
	Reflect(ReflectionType, Command)
	Replicate([]crdt.Link, Command)
	WriteMemoryImage() error
	Close()
}

type Service interface {
	CloserService
	RequestService
}

type RequestService interface {
	Call(Request) (<-chan Response, error)
}

type CloserService interface {
	CloseAPI()
}

type Request struct {
	Type       MessageType
	Reflection ReflectionType
	Query      *query.Query
	Replicate  []crdt.Link
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

type Responder interface {
	RunQuery() Response
}

type ResponderLambda func() Response

func (lambda ResponderLambda) RunQuery() Response {
	return lambda()
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

type coreCommand interface {
	Run(Core, Command)
}

type coreReplicator struct {
	links []crdt.Link
}

func (replicator coreReplicator) Run(kvn Core, kvq Command) {
	kvn.Replicate(replicator.links, kvq)
}

type coreQueryRunner struct {
	query *query.Query
}

func (queryRunner coreQueryRunner) Run(core Core, command Command) {
	core.RunQuery(queryRunner.query, command)
}

type coreReflectRunner struct {
	reflection ReflectionType
}

func (reflectRunner coreReflectRunner) Run(core Core, command Command) {
	core.Reflect(reflectRunner.reflection, command)
}

type Command struct {
	runner   coreCommand
	Request  Request
	Response chan Response
}

func makeApiQuery(request Request, runner coreCommand) Command {
	return Command{
		Request:  request,
		runner:   runner,
		Response: make(chan Response),
	}
}

func (command Command) WriteResponse(val Response) {
	command.Response <- val
	close(command.Response)
}

func (command Command) Error(err error) {
	command.WriteResponse(Response{Err: err})
}

func (command Command) Run(core Core) {
	command.runner.Run(core, command)
}

type Response struct {
	Msg       string
	Err       error
	Type      MessageType
	Path      crdt.IPFSPath
	Namespace crdt.Namespace
	Index     crdt.Index
}

func (resp Response) IsEmpty() bool {
	return resp.Equals(Response{})
}

func (resp Response) AsText() (string, error) {
	const failMsg = "AsText failed"

	w := &bytes.Buffer{}
	err := EncodeAPIResponseText(resp, w)

	if err != nil {
		return "", errors.Wrap(err, failMsg)
	}

	return w.String(), nil
}

func (resp Response) Equals(other Response) bool {
	ok := resp.Msg == other.Msg
	ok = ok && resp.Type == other.Type
	ok = ok && resp.Path == other.Path

	if !ok {
		return false
	}

	if resp.Err != nil {
		if other.Err == nil {
			return false
		} else if resp.Err.Error() != other.Err.Error() {
			return false
		}
	}

	if !resp.Namespace.Equals(other.Namespace) {
		return false
	}

	if !resp.Index.Equals(other.Index) {
		return false
	}

	return true
}

var RESPONSE_FAIL_MSG = "error"
var RESPONSE_OK_MSG = "ok"
var RESPONSE_OK Response = Response{Msg: RESPONSE_OK_MSG}
var RESPONSE_FAIL Response = Response{Msg: RESPONSE_FAIL_MSG}
var RESPONSE_QUERY Response = Response{Msg: RESPONSE_OK_MSG, Type: API_QUERY}
var RESPONSE_REPLICATE Response = Response{Msg: RESPONSE_OK_MSG, Type: API_REPLICATE}
var RESPONSE_REFLECT Response = Response{Msg: RESPONSE_OK_MSG, Type: API_REFLECT}
