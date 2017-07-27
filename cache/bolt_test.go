package cache

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
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

func TestBoltCacheConcurrency(t *testing.T) {

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
