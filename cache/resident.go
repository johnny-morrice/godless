package cache

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/log"
	"github.com/johnny-morrice/godless/query"
	"github.com/pkg/errors"
)

// Non-ACID api.MemoryImage implementation.  For use only in tests.
type residentMemoryImage struct {
	joined crdt.Index
	sync.Mutex
}

// MakeResidentMemoryImage makes an non-ACID api.MemoryImage implementation that is only suitable for tests.
func MakeResidentMemoryImage() api.MemoryImage {
	return &residentMemoryImage{joined: crdt.EmptyIndex()}
}

func (memimg *residentMemoryImage) JoinIndex(index crdt.Index) error {
	memimg.Lock()
	defer memimg.Unlock()

	memimg.joined = memimg.joined.JoinIndex(index)
	return nil
}

func (memimg *residentMemoryImage) GetIndex() (crdt.Index, error) {
	memimg.Lock()
	defer memimg.Unlock()
	return memimg.joined, nil
}

func (memimg *residentMemoryImage) CloseMemoryImage() error {
	log.Info("Closed residentMemoryImage")
	return nil
}

func MakeResidentMemoryCache(indexBufferSize, namespaceBufferSize int) api.Cache {
	return Union{
		HeadCache:      MakeResidentHeadCache(),
		IndexCache:     MakeResidentIndexCache(indexBufferSize),
		NamespaceCache: MakeResidentNamespaceCache(namespaceBufferSize),
	}
}

type residentNamespaceCache struct {
	*kvCache
}

func MakeResidentNamespaceCache(buffSize int) api.NamespaceCache {
	return residentNamespaceCache{kvCache: makeKvCache(buffSize)}
}

func (cache residentNamespaceCache) SetNamespace(addr crdt.IPFSPath, namespace crdt.Namespace) error {
	ptr := unsafe.Pointer(&namespace)
	return cache.set(addr, ptr)
}

func (cache residentNamespaceCache) GetNamespace(addr crdt.IPFSPath) (crdt.Namespace, error) {
	ptr, err := cache.get(addr)

	if err != nil {
		return crdt.EmptyNamespace(), fmt.Errorf("Cache miss for Namespace at: %s", addr)
	}

	namespacePtr := (*crdt.Namespace)(ptr)
	return *namespacePtr, nil
}

type residentIndexCache struct {
	*kvCache
}

func MakeResidentIndexCache(buffSize int) api.IndexCache {
	return residentIndexCache{kvCache: makeKvCache(buffSize)}
}

func (cache residentIndexCache) SetIndex(addr crdt.IPFSPath, index crdt.Index) error {
	ptr := unsafe.Pointer(&index)
	return cache.set(addr, ptr)
}

func (cache residentIndexCache) GetIndex(addr crdt.IPFSPath) (crdt.Index, error) {
	ptr, err := cache.get(addr)

	if err != nil {
		return crdt.EmptyIndex(), fmt.Errorf("Cache miss for Index at: %s", addr)
	}

	indexPtr := (*crdt.Index)(ptr)
	return *indexPtr, nil
}

// Memcache style key value store.
// Will drop oldest.
type kvCache struct {
	sync.RWMutex
	buff  []cacheItem
	assoc map[crdt.IPFSPath]*cacheItem
}

type cacheItem struct {
	timestamp
	key crdt.IPFSPath
	obj unsafe.Pointer
}

func makeKvCache(buffSize int) *kvCache {
	if buffSize <= 0 {
		buffSize = __DEFAULT_BUFFER_SIZE
	}

	cache := &kvCache{
		buff:  make([]cacheItem, buffSize),
		assoc: map[crdt.IPFSPath]*cacheItem{},
	}

	cache.initBuff()

	return cache
}

func (cache *kvCache) initBuff() {
	timestamp := makeTimestamp()
	for i := 0; i < len(cache.buff); i++ {
		item := &cache.buff[i]
		item.timestamp = timestamp
	}
}

func (cache *kvCache) get(addr crdt.IPFSPath) (unsafe.Pointer, error) {
	cache.RLock()
	defer cache.RUnlock()

	item, present := cache.assoc[addr]

	if !present {
		return nil, fmt.Errorf("No cached item for: %s", addr)
	}

	return item.obj, nil
}

func (cache *kvCache) set(addr crdt.IPFSPath, pointer unsafe.Pointer) error {
	cache.Lock()
	defer cache.Unlock()

	item, present := cache.assoc[addr]

	if present {
		item.timestamp = makeTimestamp()
		return nil
	}

	return cache.addNewItem(addr, pointer)
}

func (cache *kvCache) addNewItem(addr crdt.IPFSPath, pointer unsafe.Pointer) error {
	newItem := cacheItem{
		key: addr,
		obj: pointer,
	}

	newItem.timestamp = makeTimestamp()

	bufferedItem := cache.popOldest()
	*bufferedItem = newItem

	cache.assoc[addr] = bufferedItem
	return nil
}

func (cache *kvCache) popOldest() *cacheItem {
	var oldest *cacheItem

	for i := 0; i < len(cache.buff); i++ {
		item := &cache.buff[i]

		if oldest == nil {
			oldest = item
			continue
		}

		if item.older(oldest.timestamp) {
			oldest = item
		}
	}

	if oldest == nil {
		panic("Corrupt buffer")
	}

	delete(cache.assoc, oldest.key)

	return oldest
}

type timestamp struct {
	seconds     int64
	nanoseconds int64
}

func makeTimestamp() timestamp {
	t := time.Now()
	return timestamp{
		seconds:     t.Unix(),
		nanoseconds: int64(t.Nanosecond()),
	}
}

func (stamp timestamp) older(other timestamp) bool {
	if stamp.seconds < other.seconds {
		return true
	}

	if stamp.seconds == other.seconds && stamp.nanoseconds < other.nanoseconds {
		return true
	}

	return false
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

func (queue *residentPriorityQueue) Enqueue(request api.Request, data interface{}) error {
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
					log.Error("Error draining residentPriorityQueue: %s", queuePop.err.Error())
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
	log.Info("Closed residentPriorityQueue")
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

func makeResidentQueueItem(request api.Request, data interface{}) (residentQueueItem, error) {
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

func findRequestPriority(request api.Request) (residentPriority, error) {
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
