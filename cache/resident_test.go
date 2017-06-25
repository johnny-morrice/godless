package cache

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/pkg/errors"
)

func TestResidentMemoryImageConcurrency(t *testing.T) {
	const count = __CONCURRENCY_LEVEL / 30

	memimg := MakeResidentMemoryImage()
	indices := genIndices(count)

	wg := &sync.WaitGroup{}
	for _, idx := range indices {
		index := idx
		wg.Add(3)
		go func() {
			err := memimg.PushIndex(index)
			testutil.AssertNil(t, err)
			wg.Done()
		}()
		go func() {
			_, err := memimg.JoinAllIndices()
			testutil.AssertNil(t, err)
			wg.Done()
		}()
		go func() {
			err := memimg.ForeachIndex(func(index crdt.Index) {
				if index.IsEmpty() {
					t.Error("Unexpected empty index")
				}
			})

			testutil.AssertNil(t, err)
			wg.Done()
		}()
	}

	const timeout = time.Second * 2
	testutil.WaitGroupTimeout(t, wg, timeout)
}

func TestResidentMemoryImagePushIndex(t *testing.T) {
	t.FailNow()
}

func TestResidentMemoryImageForeachIndex(t *testing.T) {
	t.FailNow()
}

func TestResidentMemoryImageJoinAllIndices(t *testing.T) {
	t.FailNow()
}

func TestResidentHeadCacheConcurrency(t *testing.T) {
	const headCount = __CONCURRENCY_LEVEL
	heads := genHeads(headCount)

	cache := MakeResidentHeadCache()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, h := range heads {
			head := h
			err := cache.SetHead(head)
			testutil.AssertNil(t, err)
			wg.Add(1)
			go func() {
				defer wg.Done()
				found, err := cache.GetHead()
				testutil.AssertNil(t, err)
				message := fmt.Sprintf("Head too low, expected at least %v but got %v", found, head)
				testutil.Assert(t, message, headLessEqual(head, found))
			}()
		}
	}()

	const timeout = time.Second * 10
	testutil.WaitGroupTimeout(t, wg, timeout)
}

func TestResidentHeadCacheGetSet(t *testing.T) {
	cache := MakeResidentHeadCache()
	head, err := cache.GetHead()
	testutil.AssertNil(t, err)
	testutil.AssertEquals(t, "Expected empty head", crdt.NIL_PATH, head)
	err = cache.SetHead("Howdy")
	testutil.AssertNil(t, err)
	head, err = cache.GetHead()
	testutil.AssertEquals(t, "Unexpected head", crdt.IPFSPath("Howdy"), head)
}

func TestResidentIndexCacheConcurrency(t *testing.T) {
	const count = __CONCURRENCY_LEVEL / 2

	heads := genHeads(count)
	indices := genIndices(count)

	cache := MakeResidentIndexCache(__CONCURRENCY_LEVEL / 4)

	wg := &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		indexAddr := heads[i]
		index := indices[i]
		go func() {
			defer wg.Done()
			err := cache.SetIndex(indexAddr, index)
			testutil.AssertNil(t, err)

			if err == nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					actual, err := cache.GetIndex(indexAddr)
					if err == nil && !index.Equals(actual) {
						t.Error("Unexpected Index")
					}
				}()
			}
		}()
	}

	const timeout = time.Second * 10
	testutil.WaitGroupTimeout(t, wg, timeout)
}

func TestResidentIndexCacheExpire(t *testing.T) {
	const buffSize = 10
	const dataLength = buffSize * 2

	heads := genHeads(dataLength)
	indices := genIndices(dataLength)

	cache := MakeResidentIndexCache(buffSize)

	for i := 0; i < dataLength; i++ {
		indexAddr := heads[i]
		index := indices[i]
		err := cache.SetIndex(indexAddr, index)
		testutil.AssertNil(t, err)
		// wait()
		actual, getErr := cache.GetIndex(indexAddr)
		testutil.AssertNil(t, getErr)
		if !index.Equals(actual) {
			t.Error("Unexpected Index")
		}
	}

	for i := 0; i < buffSize; i++ {
		indexAddr := heads[i]
		index := indices[i]
		index, err := cache.GetIndex(indexAddr)
		if !index.IsEmpty() {
			t.Error("Expected empty Index")
		}

		testutil.AssertNonNil(t, err)
	}
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
				t.Error("Unexpected drain value: %v", head)
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

func genIndices(count int) []crdt.Index {
	const indexSize = 10

	indices := make([]crdt.Index, count)

	for i := 0; i < count; i++ {
		indices[i] = crdt.GenIndex(testutil.Rand(), indexSize)
	}

	return indices
}

func genHeads(count int) []crdt.IPFSPath {
	heads := make([]crdt.IPFSPath, count)
	for i := 0; i < count; i++ {
		heads[i] = crdt.IPFSPath(strconv.Itoa(i))
	}
	return heads
}

func headLessEqual(x, y crdt.IPFSPath) bool {
	numX, errX := strconv.Atoi(string(x))
	numY, errY := strconv.Atoi(string(y))

	if errX != nil || errY != nil {
		panic("Bad numeric head")
	}

	return numX <= numY
}

const __CONCURRENCY_LEVEL = 1000

func wait() {
	timer := time.NewTimer(time.Millisecond * 10)
	<-timer.C
}

var __REFLECT_REQUEST api.APIRequest = api.APIRequest{Type: api.API_REFLECT}
