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

// Non-ACID api.MemoryImage implementation.  For use only in tests.
type residentMemoryImage struct {
	joined crdt.Index
	extra  []crdt.Index
	sync.RWMutex
}

func (memimg *residentMemoryImage) PushIndex(index crdt.Index) error {
	memimg.Lock()
	defer memimg.Unlock()

	memimg.extra = append(memimg.extra, index)
	return nil
}

func (memimg *residentMemoryImage) ForeachIndex(f func(index crdt.Index)) error {
	memimg.RLock()
	defer memimg.RUnlock()

	if !memimg.joined.IsEmpty() {
		f(memimg.joined)
	}

	for _, index := range memimg.extra {
		f(index)
	}

	return nil
}

func (memimg *residentMemoryImage) JoinAllIndices() (crdt.Index, error) {
	defer memimg.Unlock()
	memimg.Lock()

	for _, index := range memimg.extra {
		memimg.joined = memimg.joined.JoinIndex(index)
	}

	memimg.extra = memimg.extra[:0]

	return memimg.joined, nil
}

// MakeResidentMemoryImage makes an non-ACID api.MemoryImage implementation that is only suitable for tests.
func MakeResidentMemoryImage() api.MemoryImage {
	return &residentMemoryImage{
		extra: make([]crdt.Index, 0, __DEFAULT_BUFFER_SIZE),
	}
}

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

	cache := &residentIndexCache{
		buff:  make([]indexCacheItem, buffSize),
		assoc: map[crdt.IPFSPath]*indexCacheItem{},
	}

	cache.initBuff()

	return cache
}

type indexCacheItem struct {
	key           crdt.IPFSPath
	index         crdt.Index
	timestamp     int64
	nanoTimestamp int
}

func (cache *residentIndexCache) initBuff() {
	for i := 0; i < len(cache.buff); i++ {
		item := &cache.buff[i]
		item.timestamp, item.nanoTimestamp = makeTimestamp()
	}
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
		item.timestamp, item.nanoTimestamp = makeTimestamp()
		return nil
	}

	return cache.addNewItem(indexAddr, index)
}

func (cache *residentIndexCache) addNewItem(indexAddr crdt.IPFSPath, index crdt.Index) error {
	newItem := indexCacheItem{
		key:   indexAddr,
		index: index,
	}

	newItem.timestamp, newItem.nanoTimestamp = makeTimestamp()

	bufferedItem := cache.popOldest()
	*bufferedItem = newItem

	cache.assoc[indexAddr] = bufferedItem
	return nil
}

func (cache *residentIndexCache) popOldest() *indexCacheItem {
	var oldest *indexCacheItem

	for i := 0; i < len(cache.buff); i++ {
		item := &cache.buff[i]

		if oldest == nil {
			oldest = item
			continue
		}

		older := item.timestamp < oldest.timestamp
		if !older && item.timestamp == oldest.timestamp {
			older = item.nanoTimestamp < oldest.nanoTimestamp
		}

		if older {
			oldest = item
		}
	}

	if oldest == nil {
		panic("Corrupt buffer")
	}

	delete(cache.assoc, oldest.key)

	return oldest
}

func makeTimestamp() (int64, int) {
	t := time.Now()
	return t.Unix(), t.Nanosecond()
}

type residentHeadCache struct {
	sync.RWMutex
	current crdt.IPFSPath
}

func (cache *residentHeadCache) SetHead(head crdt.IPFSPath) error {
	cache.Lock()
	defer cache.Unlock()
	cache.current = head
	return nil
}

func (cache *residentHeadCache) GetHead() (crdt.IPFSPath, error) {
	cache.RLock()
	defer cache.RUnlock()
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
	stopper   chan struct{}
}

func MakeResidentBufferQueue(buffSize int) api.RequestPriorityQueue {
	if buffSize <= 0 {
		buffSize = __DEFAULT_BUFFER_SIZE
	}

	queue := &residentPriorityQueue{
		semaphore: make(chan struct{}, buffSize),
		buff:      make([]residentQueueItem, buffSize),
		datach:    make(chan interface{}),
		stopper:   make(chan struct{}),
	}

	return queue
}

func (queue *residentPriorityQueue) Len() int {
	queue.Lock()
	defer queue.Unlock()
	count := 0
	for _, item := range queue.buff {
		if item.populated {
			count++
		}
	}

	return count
}

func (queue *residentPriorityQueue) Enqueue(request api.APIRequest, data interface{}) error {
	item, err := makeResidentQueueItem(request, data)

	if err != nil {
		return errors.Wrap(err, "residentPriorityQueue.Enqueue failed")
	}

	queue.Lock()
	defer queue.Unlock()

	for i := 0; i < len(queue.buff); i++ {
		spot := &queue.buff[i]
		if !spot.populated {
			*spot = item
			queue.lockResource()
			return nil
		}
	}
	log.Debug("Queued request.")

	return fullQueue
}

func (queue *residentPriorityQueue) Drain() <-chan interface{} {
	go func() {
	LOOP:
		for {
			popch := queue.waitForPop()

			select {
			case queuePop := <-popch:
				if queuePop.err != nil {
					log.Error("Error draining residentPriorityQueue: %v", queuePop.err.Error())
					close(queue.datach)
					return
				}

				queue.datach <- queuePop.data
				continue LOOP
			case <-queue.stopper:
				close(queue.datach)
				return
			}

		}
	}()

	return queue.datach
}

func (queue *residentPriorityQueue) Close() error {
	close(queue.stopper)
	return nil
}

type queuePop struct {
	data interface{}
	err  error
}

func (queue *residentPriorityQueue) waitForPop() <-chan queuePop {
	popch := make(chan queuePop)

	go func() {
		data, err := queue.popFront()
		popch <- queuePop{data: data, err: err}
	}()

	return popch
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
		log.Error("resitentPriorityQueue buffer is corrupt")
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
var fullQueue error = errors.New("Queue is full")

type residentPriority uint8

const (
	__QUERY_JOIN_PRIORITY = residentPriority(iota)
	__QUERY_REFLECT_PRIORITY
	__QUERY_SELECT_PRIORITY
	__QUERY_REPLICATE_PRIORITY
	__UNKNOWN_PRIORITY
)

const __DEFAULT_BUFFER_SIZE = 1024
