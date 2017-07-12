package service

import (
	"fmt"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/log"
)

type QueuedApiServiceOptions struct {
	Core       api.Core
	Queue      api.RequestPriorityQueue
	QueryLimit int
}

type queuedApiService struct {
	QueuedApiServiceOptions
	Debug     bool
	semaphore chan struct{}
	stopch    chan struct{}
}

func LaunchQueuedApiService(options QueuedApiServiceOptions) (api.Service, <-chan error) {
	errch := make(chan error, 1)

	service := &queuedApiService{
		QueuedApiServiceOptions: options,
		stopch:                  make(chan struct{}),
	}

	if service.QueryLimit > 0 {
		service.semaphore = make(chan struct{}, service.QueryLimit)
	}

	go service.executeLoop(errch)

	return service, errch
}

func (service *queuedApiService) executeLoop(errch chan<- error) {
	defer close(errch)
	drainch := service.Queue.Drain()
	for {
		select {
		case anything, ok := <-drainch:
			if ok {
				thing := anything
				service.fork(func() { service.runQueueItem(errch, thing) })
			}
		case <-service.stopch:
			return
		}
	}
}

func (service *queuedApiService) fork(f func()) {
	if service.Debug {
		f()
		return
	}

	go f()
}

func (service *queuedApiService) runQueueItem(errch chan<- error, thing interface{}) {
	service.lockResource()
	defer service.unlockResource()

	log.Info("API executing request, %d remain in queue", service.Queue.Len())
	kvq, ok := thing.(api.Command)

	if !ok {
		log.Error("Corrupt queue")
		errch <- fmt.Errorf("Corrupt queue")
	}

	service.run(kvq)
}

func (service *queuedApiService) run(kvq api.Command) {
	go kvq.Run(service.Core)
}

func (service *queuedApiService) lockResource() {
	if service.semaphore == nil {
		return
	}

	log.Debug("API waiting for resource...")
	service.semaphore <- struct{}{}
	log.Debug("API found resource")
}

func (service *queuedApiService) unlockResource() {
	if service.semaphore == nil {
		return
	}

	log.Debug("API releasing resource...")
	<-service.semaphore
	log.Debug("API released resource")
}

func (service *queuedApiService) Call(request api.Request) (<-chan api.Response, error) {
	switch request.Type {
	case api.API_QUERY:
		return service.runQuery(request)
	case api.API_REPLICATE:
		return service.replicate(request)
	case api.API_REFLECT:
		return service.reflect(request)
	default:
		return nil, fmt.Errorf("Unknown request.Type: %v", request.Type)
	}
}

func (service *queuedApiService) writeResponse(respch chan<- api.Response, resp api.Response) {
	select {
	case <-service.stopch:
		return
	case respch <- resp:
		return
	}
}

func (service *queuedApiService) enqueue(kvq api.Command) {
	log.Info("Enqueing request...")
	service.fork(func() {
		err := service.Queue.Enqueue(kvq.Request, kvq)
		if err != nil {
			log.Error("Failed to enqueue request: %s", err.Error())
			fail := api.RESPONSE_FAIL
			fail.Err = err
			go service.writeResponse(kvq.Response, fail)
		}
	})
}

func (service *queuedApiService) replicate(request api.Request) (<-chan api.Response, error) {
	log.Info("api.APIService Replicating...")
	command, err := request.MakeCommand()

	if err != nil {
		return nil, err
	}

	service.enqueue(command)

	return command.Response, nil
}

func (service *queuedApiService) runQuery(request api.Request) (<-chan api.Response, error) {
	query := request.Query
	if log.CanLog(log.LOG_INFO) {
		text, err := query.PrettyText()
		if err == nil {
			log.Info("api.APIService running query.Query:\n%s", text)
		} else {
			log.Debug("Failed to pretty print query: %s", err.Error())
		}

	}

	if err := query.Validate(); err != nil {
		log.Warn("Invalid query.Query")
		return nil, err
	}

	command, err := request.MakeCommand()

	if err != nil {
		return nil, err
	}

	service.enqueue(command)

	return command.Response, nil
}

func (service *queuedApiService) reflect(request api.Request) (<-chan api.Response, error) {
	log.Info("api.APIService running reflect request...")
	command, err := request.MakeCommand()

	if err != nil {
		return nil, err
	}

	service.enqueue(command)

	return command.Response, nil
}

func (service *queuedApiService) CloseAPI() {
	close(service.stopch)
	service.Core.Close()
	service.Queue.Close()
	log.Info("API closed")
}
