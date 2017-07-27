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
)

func testMemoryImageConcurrency(t *testing.T, memimg api.MemoryImage, count int) {
	indices := genIndices(count)

	wg := &sync.WaitGroup{}

	initialIndex := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		"hi": crdt.UnsignedLink("world"),
	})
	memimg.JoinIndex(initialIndex)

	for _, idx := range indices {
		index := idx
		wg.Add(2)
		go func() {
			err := memimg.JoinIndex(index)
			testutil.AssertNil(t, err)
			wg.Done()
		}()

		go func() {
			index, err := memimg.GetIndex()
			if index.IsEmpty() {
				t.Error("Unexpected empty index")
			}

			testutil.AssertNil(t, err)
			wg.Done()
		}()
	}

	const timeout = time.Second * 2
	testutil.WaitGroupTimeout(t, wg, timeout)
}

func testMemoryImage(t *testing.T, memimg api.MemoryImage) {
	indexA := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		"hi": crdt.UnsignedLink("world"),
	})
	indexB := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		"dude": crdt.UnsignedLink("yes"),
	})

	expected := indexA.JoinIndex(indexB)

	memimg.JoinIndex(indexA)
	memimg.JoinIndex(indexB)
	actual, err := memimg.GetIndex()

	testutil.AssertNil(t, err)
	testutil.Assert(t, "Unexpected index", expected.Equals(actual))
}

func testIndexExpire(t *testing.T, cache api.IndexCache, count, buffsize int) {
	if count <= buffsize {
		panic("count must exceed buffsize")
	}

	heads := genHeads(count)
	indices := genIndices(count)

	for i := 0; i < count; i++ {
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

	for i := 0; i < buffsize; i++ {
		indexAddr := heads[i]
		index := indices[i]
		index, err := cache.GetIndex(indexAddr)
		if !index.IsEmpty() {
			t.Error("Expected empty Index")
		}

		testutil.AssertNonNil(t, err)
	}
}

func testNamespaceExpire(t *testing.T, cache api.NamespaceCache, count, buffsize int) {
	if count <= buffsize {
		panic("count must exceed buffsize")
	}

	heads := genHeads(count)
	indices := genNamespaces(count)

	for i := 0; i < count; i++ {
		namespaceAddr := heads[i]
		namespace := indices[i]
		err := cache.SetNamespace(namespaceAddr, namespace)
		testutil.AssertNil(t, err)
		// wait()
		actual, getErr := cache.GetNamespace(namespaceAddr)
		testutil.AssertNil(t, getErr)
		if !namespace.Equals(actual) {
			t.Error("Unexpected Namespace")
		}
	}

	for i := 0; i < buffsize; i++ {
		namespaceAddr := heads[i]
		namespace := indices[i]
		namespace, err := cache.GetNamespace(namespaceAddr)
		if !namespace.IsEmpty() {
			t.Error("Expected empty Namespace")
		}

		testutil.AssertNonNil(t, err)
	}
}

func testHeadConcurrency(t *testing.T, cache api.HeadCache, count int) {
	heads := genHeads(count)

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
				message := fmt.Sprintf("Head too low, expected at least %s but got %s", found, head)
				testutil.Assert(t, message, headLessEqual(head, found))
			}()
		}
	}()

	const timeout = time.Second * 10
	testutil.WaitGroupTimeout(t, wg, timeout)
}

func testIndexConcurrency(t *testing.T, cache api.IndexCache, count int) {
	heads := genHeads(count)
	indices := genIndices(count)

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

func testNamespaceConcurrency(t *testing.T, cache api.NamespaceCache, count int) {
	heads := genHeads(count)
	namespaces := genNamespaces(count)

	wg := &sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		namespaceAddr := heads[i]
		namespace := namespaces[i]
		go func() {
			defer wg.Done()
			err := cache.SetNamespace(namespaceAddr, namespace)
			testutil.AssertNil(t, err)

			if err == nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					actual, err := cache.GetNamespace(namespaceAddr)
					if err == nil && !namespace.Equals(actual) {
						t.Error("Unexpected Index")
					}
				}()
			}
		}()
	}

	const timeout = time.Second * 10
	testutil.WaitGroupTimeout(t, wg, timeout)
}

func testCacheGetSet(t *testing.T, cache api.Cache) {
	const head = crdt.IPFSPath("Hello")

	actualHead, err := cache.GetHead()
	testutil.AssertNil(t, err)
	testutil.AssertEquals(t, "Expected empty head", crdt.NIL_PATH, actualHead)

	err = cache.SetHead(head)
	testutil.AssertNil(t, err)

	actualHead, err = cache.GetHead()
	testutil.AssertNil(t, err)
	testutil.AssertEquals(t, "Unexpected head", head, actualHead)

	namespace := crdt.MakeNamespace(map[crdt.TableName]crdt.Table{
		"Hi": crdt.MakeTable(map[crdt.RowName]crdt.Row{
			"Dude": crdt.MakeRow(map[crdt.EntryName]crdt.Entry{
				"Welcome": crdt.MakeEntry([]crdt.Point{
					crdt.UnsignedPoint("Wowzer"),
				}),
			}),
		}),
	})
	const namespaceAddr = crdt.IPFSPath("Namespace Addr")

	err = cache.SetNamespace(namespaceAddr, namespace)
	testutil.AssertNil(t, err)

	actualNamespace, err := cache.GetNamespace(namespaceAddr)
	testutil.AssertNil(t, err)
	testutil.Assert(t, "Unexpected namespace", namespace.Equals(actualNamespace))

	actualNamespace, err = cache.GetNamespace("Bad namespace addr")
	testutil.AssertNonNil(t, err)
	testutil.Assert(t, "Expected empty namespace", crdt.EmptyNamespace().Equals(actualNamespace))

	const indexAddr = "Index Addr"
	index := crdt.MakeIndex(map[crdt.TableName]crdt.Link{
		"hi": crdt.UnsignedLink("world"),
	})

	err = cache.SetIndex(indexAddr, index)
	testutil.AssertNil(t, err)

	actualIndex, err := cache.GetIndex(indexAddr)
	testutil.AssertNil(t, err)
	testutil.Assert(t, "Unexpected index", index.Equals(actualIndex))

	actualIndex, err = cache.GetIndex("Bad index addr")
	testutil.AssertNonNil(t, err)
	testutil.Assert(t, "Expected empty index", crdt.EmptyIndex().Equals(actualIndex))
}

func genIndices(count int) []crdt.Index {
	const size = 10

	indices := make([]crdt.Index, count)

	for i := 0; i < count; i++ {
		indices[i] = crdt.GenIndex(testutil.Rand(), size)
	}

	return indices
}

func genNamespaces(count int) []crdt.Namespace {
	const size = 10

	namespaces := make([]crdt.Namespace, count)

	for i := 0; i < count; i++ {
		namespaces[i] = crdt.GenNamespace(testutil.Rand(), size)
	}

	return namespaces
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
