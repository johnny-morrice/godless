package api

//go:generate mockgen -package mock_godless -destination ../mock/mock_api.go -imports lib=github.com/johnny-morrice/api -self_package lib github.com/johnny-morrice/godless/api Core,RemoteStore,RemoteNamespace,NamespaceSearcher,DataPeer,PubSubSubscription,PubSubRecord,Service

import (
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

type Responder interface {
	RunQuery() Response
}

type ResponderLambda func() Response

func (lambda ResponderLambda) RunQuery() Response {
	return lambda()
}

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
		Response: make(chan Response, 1),
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
