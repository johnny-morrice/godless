package cache

import (
	"fmt"
	"sync"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

type residentHeadCache struct {
	sync.RWMutex
	current  crdt.IPFSPath
	previous crdt.IPFSPath
}

func (cache *residentHeadCache) SetHead(head crdt.IPFSPath) error {
	cache.Lock()
	cache.previous = cache.current
	cache.current = head
	return nil
}

func (cache *residentHeadCache) Commit() error {
	cache.previous = ""
	cache.Unlock()
	return nil
}

func (cache *residentHeadCache) Rollback() error {
	cache.current = cache.previous
	return cache.Commit()
}

func (cache *residentHeadCache) GetHead() (crdt.IPFSPath, error) {
	cache.RLock()
	head := cache.current
	cache.RUnlock()
	return head, nil
}

func MakeResidentHeadCache() api.HeadCache {
	return &residentHeadCache{}
}

type residentPriorityQueue struct {
	sync.Mutex
	semaphore chan interface{}
	buff      []residentQueueItem
	datach    chan interface{}
}

func MakeResidentBufferQueue(buffSize int) api.RequestPriorityQueue {
	if buffSize <= 0 {
		panic("Expect positive buffSize")
	}

	queue := &residentPriorityQueue{
		semaphore: make(chan interface{}, buffSize),
		buff:      make([]residentQueueItem, buffSize),
		datach:    make(chan interface{}),
	}

	return queue
}

func (queue *residentPriorityQueue) Enqueue(request api.APIRequest, data interface{}) error {
	item, err := makeResidentQueueItem(request, data)

	if err != nil {
		return errors.Wrap(err, "residentPriorityQueue.Enqueue failed")
	}

	queue.lockResource()
	queue.Lock()
	defer queue.Unlock()

	for i := 0; i < len(queue.buff); i++ {
		spot := &queue.buff[i]
		if !spot.populated {
			*spot = item
			return nil
		}
	}

	return corruptBuffer
}

func (queue *residentPriorityQueue) Drain() <-chan interface{} {
	go func() {
		for {
			data, err := queue.popFront()

			if err != nil {
				log.Error("Error draining residentPriorityQueue: %v", err)
				return
			}

			queue.datach <- data
		}
	}()

	return queue.datach
}

func (queue *residentPriorityQueue) Close() error {
	close(queue.datach)
	return nil
}

func (queue *residentPriorityQueue) popFront() (interface{}, error) {
	queue.unlockResource()
	queue.Lock()
	defer queue.Unlock()

	var best *residentQueueItem
	for i := 0; i < len(queue.buff); i++ {
		spot := &queue.buff[i]
		if !spot.populated {
			continue
		}

		if best == nil {
			best = spot
			continue
		}

		if spot.priority < best.priority {
			best = spot
		}
	}

	if best == nil {
		log.Error("Buffer is corrupt")
		return nil, corruptBuffer
	}

	best.populated = false
	return best.data, nil
}

func (queue *residentPriorityQueue) lockResource() {
	queue.semaphore <- struct{}{}
}

func (queue *residentPriorityQueue) unlockResource() {
	<-queue.semaphore
}

type residentQueueItem struct {
	populated bool
	data      interface{}
	priority  residentPriority
}

func makeResidentQueueItem(request api.APIRequest, data interface{}) (residentQueueItem, error) {
	priority, err := findRequestPriority(request)

	if err != nil {
		return residentQueueItem{}, err
	}

	item := residentQueueItem{
		data:      data,
		priority:  priority,
		populated: true,
	}

	return item, nil
}

func findRequestPriority(request api.APIRequest) (residentPriority, error) {
	switch request.Type {
	case api.API_QUERY:
		if request.Query.OpCode == query.JOIN {
			return __QUERY_JOIN_PRIORITY, nil
		} else {
			return __QUERY_SELECT_PRIORITY, nil
		}
	case api.API_REFLECT:
		return __QUERY_REFLECT_PRIORITY, nil
	case api.API_REPLICATE:
		return __QUERY_REPLICATE_PRIORITY, nil
	default:
		return __UNKNOWN_PRIORITY, fmt.Errorf("Unknown request.Type: %v", request.Type)
	}
}

var corruptBuffer error = errors.New("Corrupt residentPriorityQueue buffer")

type residentPriority uint8

const (
	__QUERY_JOIN_PRIORITY = residentPriority(iota)
	__QUERY_REFLECT_PRIORITY
	__QUERY_SELECT_PRIORITY
	__QUERY_REPLICATE_PRIORITY
	__UNKNOWN_PRIORITY
)

const DEFAULT_BUFFER_SIZE = 512
