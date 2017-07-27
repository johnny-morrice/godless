package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/pkg/errors"
)

func TestResidentMemoryImageConcurrency(t *testing.T) {
	const count = __CONCURRENCY_LEVEL / 10
	memimg := MakeResidentMemoryImage()
	testMemoryImageConcurrency(t, memimg, count)
}

func TestResidentMemoryImageGetAndJoinIndex(t *testing.T) {
	memimg := MakeResidentMemoryImage()
	testMemoryImage(t, memimg)
}

func TestResidentHeadCacheConcurrency(t *testing.T) {
	const count = __CONCURRENCY_LEVEL
	cache := MakeResidentHeadCache()
	testHeadConcurrency(t, cache, count)
}

func TestResidentCache(t *testing.T) {
	cache := MakeResidentMemoryCache(0, 0)
	testCacheGetSet(t, cache)
}

func TestResidentIndexCacheConcurrency(t *testing.T) {
	const count = __CONCURRENCY_LEVEL / 2
	cache := MakeResidentIndexCache(__CONCURRENCY_LEVEL / 4)
	testIndexConcurrency(t, cache, count)
}

func TestResidentIndexCacheExpire(t *testing.T) {
	const buffsize = 10
	const count = buffsize * 2
	cache := MakeResidentIndexCache(buffsize)
	testIndexExpire(t, cache, count, buffsize)
}

func TestResidentNamespaceCacheConcurrency(t *testing.T) {
	const count = __CONCURRENCY_LEVEL / 2
	cache := MakeResidentNamespaceCache(__CONCURRENCY_LEVEL / 4)
	testNamespaceConcurrency(t, cache, count)
}

func TestResidentNamespaceCacheExpire(t *testing.T) {
	const buffsize = 10
	const count = buffsize * 2
	cache := MakeResidentNamespaceCache(buffsize)
	testNamespaceExpire(t, cache, count, buffsize)
}

func TestResidentPriorityQueueDrain(t *testing.T) {
	const buffSize = 10
	const dataLength = 20

	heads := genHeads(dataLength)

	queue := MakeResidentBufferQueue(buffSize)

	wg := &sync.WaitGroup{}

	for i := 0; i < dataLength; i++ {
		head := heads[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			fullErr := errors.New("try again")
			for fullErr != nil {
				fullErr = queue.Enqueue(__REFLECT_REQUEST, head)
			}

		}()
	}

	// Need to work on the closing logic for this.
	wg.Add(1)
	go func() {
		defer wg.Done()
		drainCount := 0
		for head := range queue.Drain() {
			drainCount++
			found := false
			for _, inputHead := range heads {
				if head == inputHead {
					found = true
				}
			}
			if !found {
				t.Error("Unexpected drain value:", head)
			}

			if drainCount >= dataLength {
				queue.Close()
			}
		}
	}()

	const timeout = time.Second * 2
	testutil.WaitGroupTimeout(t, wg, timeout)

}
func TestResidentPriorityQueueLen(t *testing.T) {
	const buffSize = 10
	queue := MakeResidentBufferQueue(buffSize)

	queue.Enqueue(__REFLECT_REQUEST, "howdy")

	const expected = 1
	testutil.AssertEquals(t, "Unexpected length", expected, queue.Len())
}

var __REFLECT_REQUEST api.Request = api.Request{Type: api.API_REFLECT}
