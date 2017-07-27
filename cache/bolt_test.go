package cache

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func TestBoltCache(t *testing.T) {
	f := createTempFile()
	defer f.Close()

	options := BoltOptions{
		FilePath: f.Name(),
	}

	boltFactory, err := MakeBoltFactory(options)

	panicOnBadInit(err)

	cache, err := boltFactory.MakeCache()

	panicOnBadInit(err)

	testCacheGetSet(t, cache)
}

func TestBoltCacheConcurrency(t *testing.T) {
	f := createTempFile()
	defer f.Close()

	options := BoltOptions{
		FilePath: f.Name(),
	}

	boltFactory, err := MakeBoltFactory(options)

	panicOnBadInit(err)

	cache, err := boltFactory.MakeCache()

	panicOnBadInit(err)

	count := __CONCURRENCY_LEVEL / 2
	wg := &sync.WaitGroup{}

	wg.Add(3)

	go func() {
		testHeadConcurrency(t, cache, count)
		wg.Done()
	}()

	go func() {
		testIndexConcurrency(t, cache, count)
		wg.Done()
	}()

	go func() {
		testNamespaceConcurrency(t, cache, count)
		wg.Done()
	}()

	wg.Wait()
}

func createTempFile() *os.File {
	file, err := ioutil.TempFile("/tmp", "godless_bolt_test")

	panicOnBadInit(err)

	return file
}

func panicOnBadInit(err error) {
	if err != nil {
		panic(err)
	}
}

func TestBoltMemoryImage(t *testing.T) {
	t.FailNow()
}
