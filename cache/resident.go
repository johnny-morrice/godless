package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

// Memcache style key value store.
// Will drop oldest.
type residentIndexCache struct {
	sync.RWMutex
	buff  []indexCacheItem
	assoc map[crdt.IPFSPath]*indexCacheItem
}

func MakeResidentIndexCache(buffSize int) api.IndexCache {
	if buffSize <= 0 {
		buffSize = __DEFAULT_BUFFER_SIZE
	}

	return &residentIndexCache{
		buff:  make([]indexCacheItem, buffSize),
		assoc: map[crdt.IPFSPath]*indexCacheItem{},
	}
}

type indexCacheItem struct {
	key       crdt.IPFSPath
	index     crdt.Index
	timestamp int
}

func (cache *residentIndexCache) GetIndex(indexAddr crdt.IPFSPath) (crdt.Index, error) {
	cache.RLock()
	defer cache.RUnlock()

	item, present := cache.assoc[indexAddr]

	if !present {
		return crdt.EmptyIndex(), fmt.Errorf("No cached index for: %v", indexAddr)
	}

	return item.index, nil
}

func (cache *residentIndexCache) SetIndex(indexAddr crdt.IPFSPath, index crdt.Index) error {
	cache.Lock()
	defer cache.Unlock()

	item, present := cache.assoc[indexAddr]

	if present {
		item.timestamp = makeTimestamp()
		return nil
	}

	return cache.addNewItem(indexAddr, index)
}

func (cache *residentIndexCache) addNewItem(indexAddr crdt.IPFSPath, index crdt.Index) error {
	newItem := indexCacheItem{
		timestamp: makeTimestamp(),
		key:       indexAddr,
		index:     index,
	}

	bufferedItem := cache.findOldest()
	*bufferedItem = newItem
	cache.assoc[indexAddr] = bufferedItem
	return nil
}

func (cache *residentIndexCache) findOldest() *indexCacheItem {
	var oldest *indexCacheItem

	for _, item := range cache.buff {
		if oldest == nil {
			oldest = &item
			continue
		}

		if item.timestamp < oldest.timestamp {
			oldest = &item
		}
	}

	if oldest == nil {
		panic("Corrupt buffer")
	}

	return oldest
}

func makeTimestamp() int {
	return time.Now().Nanosecond()
}

type residentHeadCache struct {
	sync.RWMutex
	writing  bool
	written  bool
	current  crdt.IPFSPath
	previous crdt.IPFSPath
}

func (cache *residentHeadCache) BeginReadTransaction() error {
	cache.RLock()
	return nil
}

func (cache *residentHeadCache) BeginWriteTransaction() error {
	cache.Lock()
	cache.writing = true
	return nil
}

func (cache *residentHeadCache) SetHead(head crdt.IPFSPath) error {
	cache.previous = cache.current
	cache.current = head
	cache.written = true
	return nil
}

func (cache *residentHeadCache) Commit() error {
	if cache.writing {
		cache.previous = ""
		cache.writing = false
		cache.written = false
		cache.Unlock()
	} else {
		cache.RUnlock()
	}

	return nil
}

func (cache *residentHeadCache) Rollback() error {
	if !cache.writing {
		return errors.New("Cannot rollback without write")
	}

	if cache.written {
		cache.current = cache.previous
	}

	return cache.Commit()
}

func (cache *residentHeadCache) GetHead() (crdt.IPFSPath, error) {
	head := cache.current
	return head, nil
}

func MakeResidentHeadCache() api.HeadCache {
	return &residentHeadCache{}
}

type residentPriorityQueue struct {
	sync.Mutex
	semaphore chan struct{}
	buff      []residentQueueItem
	datach    chan interface{}
}

func MakeResidentBufferQueue(buffSize int) api.RequestPriorityQueue {
	if buffSize <= 0 {
		buffSize = __DEFAULT_BUFFER_SIZE
	}

	queue := &residentPriorityQueue{
		semaphore: make(chan struct{}, buffSize),
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

const __DEFAULT_BUFFER_SIZE = 512
