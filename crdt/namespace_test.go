package crdt

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
)

func TestMakeRowStream(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(rowStreamOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func rowStreamOk(expected Row) bool {
	const tableKey = "tableKey"
	const rowKey = "rowKey"
	stream, invalid := MakeRowStream(tableKey, rowKey, expected)

	if len(invalid) > 0 {
		log.Error("Found invalid points")
		return false
	}

	actualNs := readNamespaceStream(stream)
	actualTable := actualNs.Tables[tableKey]
	actual := actualTable.Rows[rowKey]
	same := expected.Equals(actual)

	if !same {
		expectedTable := MakeTable(map[RowName]Row{
			rowKey: expected,
		})
		expectedNs := MakeNamespace(map[TableName]Table{
			tableKey: expectedTable,
		})
		expectedText := namespaceText(expectedNs)
		actualText := namespaceText(actualNs)

		testutil.LogDiff(expectedText, actualText)
	}

	return same
}

func (row Row) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := genRow(rand, size)
	return reflect.ValueOf(gen)
}

func (ns Namespace) Generate(rand *rand.Rand, size int) reflect.Value {
	gen := GenNamespace(rand, size)
	return reflect.ValueOf(gen)
}

func BenchmarkNamespaceEncoding(b *testing.B) {
	seed := time.Now().UTC().UnixNano()
	src := rand.NewSource(seed)
	rand := rand.New(src)
	nsType := reflect.TypeOf(Namespace{})
	for i := 0; i < b.N; i++ {
		nsValue, ok := quick.Value(nsType, rand)

		if !ok {
			panic("Can not generate value")
		}

		ns := nsValue.Interface().(Namespace)
		namespaceSerializationPass(ns)
	}
}

func TestEncodeNamespace(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
		return
	}

	config := &quick.Config{
		MaxCount: testutil.ENCODE_REPEAT_COUNT,
	}

	err := quick.Check(namespaceEncodeOk, config)

	testutil.AssertVerboseErrorIsNil(t, err)
}

func TestEncodeNamespaceStable(t *testing.T) {
	const size = 50
	namespace := GenNamespace(testutil.Rand(), size)

	buff := &bytes.Buffer{}
	err := encodeNamespace(namespace, buff)

	if err != nil {
		panic(err)
	}

	expected := buff.Bytes()
	testutil.AssertEncodingStable(t, expected, func(w io.Writer) {
		err := encodeNamespace(namespace, w)

		if err != nil {
			panic(err)
		}
	})
}

func namespaceEncodeOk(randomNs Namespace) bool {
	expected, invalid, err := randomNs.Strip()

	if len(invalid) > 0 || err != nil {
		panic("namespaceEncodeFailed")
	}

	actual := namespaceSerializationPass(randomNs)
	return expected.Equals(actual)
}

func namespaceSerializationPass(expected Namespace) Namespace {
	buff := &bytes.Buffer{}
	err := encodeNamespace(expected, buff)

	if err != nil {
		panic(err)
	}

	var actual Namespace
	actual, err = decodeNamespace(buff)

	if err != nil {
		panic(err)
	}

	return actual
}

func TestEmptyNamespace(t *testing.T) {
	expected := Namespace{
		Tables: map[TableName]Table{},
	}
	actual := EmptyNamespace()

	assertNamespaceEquals(t, expected, actual)
}

func TestMakeNamespace(t *testing.T) {
	table := EmptyTable()

	expected := Namespace{
		Tables: map[TableName]Table{
			"foo": table,
		},
	}
	actual := MakeNamespace(map[TableName]Table{"foo": table})

	assertNamespaceEquals(t, expected, actual)
}

func TestNamespaceIsEmpty(t *testing.T) {
	table := EmptyTable()

	full := MakeNamespace(map[TableName]Table{"foo": table})

	empty := EmptyNamespace()

	if !empty.IsEmpty() {
		t.Error("Expected IsEmpty():", empty)
	}

	if full.IsEmpty() {
		t.Error("Unexpected IsEmpty():", full)
	}
}

func TestNamespaceCopy(t *testing.T) {
	expected := MakeNamespace(map[TableName]Table{"foo": EmptyTable()})
	actual := expected.Copy()

	assertNamespaceEquals(t, expected, actual)
}

func TestNamespaceJoinNamespace(t *testing.T) {
	table := EmptyTable()
	foo := MakeNamespace(map[TableName]Table{"foo": table})
	bar := MakeNamespace(map[TableName]Table{"bar": table})

	expectedJoin := MakeNamespace(map[TableName]Table{"foo": table, "bar": table})

	actualJoinFooBar := foo.JoinNamespace(bar)
	actualJoinBarFoo := bar.JoinNamespace(foo)

	assertNamespaceEquals(t, expectedJoin, actualJoinFooBar)
	assertNamespaceEquals(t, expectedJoin, actualJoinBarFoo)
}

func TestNamespaceJoinTable(t *testing.T) {
	table := EmptyTable()
	foo := MakeNamespace(map[TableName]Table{"foo": table})
	barTable := EmptyTable()

	expectedFoo := MakeNamespace(map[TableName]Table{"foo": table})
	expectedJoin := MakeNamespace(map[TableName]Table{"foo": table, "bar": table})

	actual := foo.JoinTable("bar", barTable)

	assertNamespaceEquals(t, expectedFoo, foo)
	assertNamespaceEquals(t, expectedJoin, actual)
}

func TestNamespaceGetTable(t *testing.T) {
	expectedTable := MakeTable(map[RowName]Row{"foo": EmptyRow()})
	expectedEmptyTable := Table{}

	hasTable := MakeNamespace(map[TableName]Table{"bar": expectedTable})
	hasNoTable := EmptyNamespace()

	var actualTable Table
	var err error
	actualTable, err = hasTable.GetTable("bar")

	assertTableEquals(t, expectedTable, actualTable)

	testutil.AssertNil(t, err)

	actualTable, err = hasNoTable.GetTable("bar")

	assertTableEquals(t, expectedEmptyTable, actualTable)

	testutil.AssertNonNil(t, err)
}

func TestNamespaceEquals(t *testing.T) {

	table := EmptyTable()
	nsA := MakeNamespace(map[TableName]Table{"foo": table})
	nsB := MakeNamespace(map[TableName]Table{"bar": table})
	nsC := EmptyNamespace()
	nsD := MakeNamespace(map[TableName]Table{
		"foo": MakeTable(map[RowName]Row{
			"howdy": EmptyRow(),
		}),
	})

	foos := []Namespace{
		nsA, nsB, nsC, nsD,
	}

	bars := make([]Namespace, len(foos))
	for i, f := range foos {
		bars[i] = f.Copy()
	}

	for i, f := range foos {
		assertNamespaceEquals(t, f, f)
		for j, b := range bars {
			if i == j {
				assertNamespaceEquals(t, f, b)
			} else {
				assertNamespaceNotEquals(t, f, b)
			}

		}
	}
}

func TestEmptyTable(t *testing.T) {
	expected := Table{Rows: map[RowName]Row{}}
	actual := EmptyTable()

	assertTableEquals(t, expected, actual)
}

func TestMakeTable(t *testing.T) {
	expected := Table{
		Rows: map[RowName]Row{
			"foo": EmptyRow(),
		},
	}
	actual := MakeTable(map[RowName]Row{"foo": EmptyRow()})

	assertTableEquals(t, expected, actual)
}

func TestTableCopy(t *testing.T) {
	expected := MakeTable(map[RowName]Row{"foo": EmptyRow()})
	actual := expected.Copy()

	assertTableEquals(t, expected, actual)
}

func TestTableAllRows(t *testing.T) {
	emptyRow := EmptyRow()
	fullRow := MakeRow(map[EntryName]Entry{
		"baz": EmptyEntry(),
	})

	expected := []Row{
		emptyRow, fullRow,
	}

	table := MakeTable(map[RowName]Row{
		"foo": emptyRow,
		"bar": fullRow,
	})

	actual := table.AllRows()

	if len(expected) != len(actual) {
		t.Error("Length mismatch in expected/actual", actual)
	}

	for _, a := range actual {
		found := false
		for _, e := range expected {
			found = found || reflect.DeepEqual(e, a)
		}

		if !found {
			t.Error("Unexpected row:", a)
		}
	}
}

func TestTableJoinTable(t *testing.T) {
	emptyRow := EmptyRow()
	barRow := MakeRow(map[EntryName]Entry{
		"Baz": EmptyEntry(),
	})

	foo := MakeTable(map[RowName]Row{
		"foo": emptyRow,
	})

	bar := MakeTable(map[RowName]Row{
		"bar": barRow,
	})

	expected := MakeTable(map[RowName]Row{
		"foo": emptyRow,
		"bar": barRow,
	})

	actual := foo.JoinTable(bar)

	assertTableEquals(t, expected, actual)
}

func TestTableJoinRow(t *testing.T) {
	emptyTable := EmptyTable()
	row := MakeRow(map[EntryName]Entry{
		"bar": MakeEntry([]Point{UnsignedPoint("hello")}),
	})

	expected := MakeTable(map[RowName]Row{
		"foo": row,
	})

	actual := emptyTable.JoinRow("foo", row)

	assertTableEquals(t, expected, actual)
}

func TestTableGetRow(t *testing.T) {
	expected := MakeRow(map[EntryName]Entry{
		"bar": EmptyEntry(),
	})

	table := MakeTable(map[RowName]Row{
		"foo": expected,
	})

	actual, err := table.GetRow("foo")

	if err != nil {
		t.Error("Unexpected error", err)
	}

	assertRowEquals(t, expected, actual)
}

func TestTableEquals(t *testing.T) {
	tables := []Table{
		EmptyTable(),
		MakeTable(map[RowName]Row{
			"foo": EmptyRow(),
		}),
		MakeTable(map[RowName]Row{
			"bar": EmptyRow(),
		}),
		MakeTable(map[RowName]Row{
			"foo": MakeRow(map[EntryName]Entry{
				"baz": EmptyEntry(),
			}),
		}),
	}

	for i := 0; i < len(tables); i++ {
		tab := tables[i]
		if !tab.Equals(tab) {
			t.Error("Expected Table equality at ", i)
		}

		for j := 0; j < len(tables); j++ {
			if i == j {
				continue
			}
			other := tables[j]

			if tab.Equals(other) {
				t.Error("Unexpected Table equality at ", i, j)
			}
		}
	}
}

func TestEmptyRow(t *testing.T) {
	expected := Row{
		Entries: map[EntryName]Entry{},
	}

	actual := EmptyRow()

	assertRowEquals(t, expected, actual)
}

func TestMakeRow(t *testing.T) {
	entry := EmptyEntry()

	expected := Row{
		Entries: map[EntryName]Entry{
			"foo": entry,
		},
	}

	actual := MakeRow(map[EntryName]Entry{
		"foo": entry,
	})

	assertRowEquals(t, expected, actual)
}

func TestRowCopy(t *testing.T) {
	expected := MakeRow(map[EntryName]Entry{"foo": EmptyEntry()})
	actual := expected.Copy()
	assertRowEquals(t, expected, actual)
}

func TestRowJoinRow(t *testing.T) {
	emptyEntry := EmptyEntry()
	fullEntry := MakeEntry([]Point{UnsignedPoint("hi")})

	expected := MakeRow(map[EntryName]Entry{
		"foo": emptyEntry,
		"bar": fullEntry,
	})

	foo := MakeRow(map[EntryName]Entry{
		"foo": emptyEntry,
	})

	bar := MakeRow(map[EntryName]Entry{
		"bar": fullEntry,
	})

	actual := foo.JoinRow(bar)

	assertRowEquals(t, expected, actual)
}

func TestRowGetEntry(t *testing.T) {
	expected := MakeEntry([]Point{UnsignedPoint("hi")})

	row := MakeRow(map[EntryName]Entry{
		"foo": expected,
	})

	actual, err := row.GetEntry("foo")

	if err != nil {
		t.Error("Unexpected error", err)
	}

	assertEntryEquals(t, expected, actual)
}

func TestRowJoinEntry(t *testing.T) {
	emptyEntry := EmptyEntry()
	fullEntry := MakeEntry([]Point{UnsignedPoint("hi")})

	expected := MakeRow(map[EntryName]Entry{
		"foo": emptyEntry,
		"bar": fullEntry,
	})

	foo := MakeRow(map[EntryName]Entry{
		"foo": emptyEntry,
	})

	actual := foo.JoinEntry("bar", fullEntry)

	assertRowEquals(t, expected, actual)
}

func TestRowEquals(t *testing.T) {
	emptyEntry := EmptyEntry()
	fullEntry := MakeEntry([]Point{UnsignedPoint("hi")})

	rows := []Row{
		EmptyRow(),
		MakeRow(map[EntryName]Entry{
			"foo": emptyEntry,
		}),
		MakeRow(map[EntryName]Entry{
			"bar": emptyEntry,
		}),
		MakeRow(map[EntryName]Entry{
			"foo": fullEntry,
		}),
	}

	for i := 0; i < len(rows); i++ {
		r := rows[i]

		if !r.Equals(r) {
			t.Error("Expected equality at ", i)
		}

		for j := 0; j < len(rows); j++ {
			if i == j {
				continue
			}

			other := rows[j]
			if r.Equals(other) {
				t.Error("Unexpected equality at ", j)
			}
		}
	}
}

func TestEmptyEntry(t *testing.T) {
	expected := Entry{
		Set: []Point{},
	}

	actual := EmptyEntry()

	assertEntryEquals(t, expected, actual)
}

func TestMakeEntry(t *testing.T) {
	expected := MakeEntry([]Point{UnsignedPoint("hello"), UnsignedPoint("world")})

	actuals := []Entry{
		MakeEntry([]Point{UnsignedPoint("hello"), UnsignedPoint("world")}),
		MakeEntry([]Point{UnsignedPoint("world"), UnsignedPoint("hello")}),
		MakeEntry([]Point{UnsignedPoint("hello"), UnsignedPoint("hello"), UnsignedPoint("world"), UnsignedPoint("world")}),
	}

	for _, a := range actuals {
		assertEntryEquals(t, expected, a)
	}
}

func TestEntryJoinEntry(t *testing.T) {
	expected := MakeEntry([]Point{UnsignedPoint("hello"), UnsignedPoint("world")})

	hello := MakeEntry([]Point{UnsignedPoint("hello")})
	world := MakeEntry([]Point{UnsignedPoint("world")})

	actualFront := hello.JoinEntry(world)
	actualBack := world.JoinEntry(hello)
	actualHello := hello.JoinEntry(hello)

	assertEntryEquals(t, expected, actualFront)
	assertEntryEquals(t, expected, actualBack)
	assertEntryEquals(t, hello, actualHello)
}

func TestEntryEquals(t *testing.T) {
	entries := []Entry{
		EmptyEntry(),
		MakeEntry([]Point{UnsignedPoint("hi")}),
		MakeEntry([]Point{UnsignedPoint("hello"), UnsignedPoint("world")}),
	}

	for i := 0; i < len(entries); i++ {
		e := entries[i]

		if !e.Equals(e) {
			t.Error("Expected equality at ", i)
		}

		for j := 0; j < len(entries); j++ {
			if i == j {
				continue
			}

			other := entries[j]
			if e.Equals(other) {
				t.Error("Unexpected equality at ", j)
			}
		}
	}
}

func TestEntryGetValues(t *testing.T) {
	expected := []Point{UnsignedPoint("hello")}
	entry := MakeEntry([]Point{UnsignedPoint("hello")})
	actual := entry.GetValues()
	if !reflect.DeepEqual(expected, actual) {
		t.Error("Expected", expected, "but was", actual)
	}
}

func assertEntryEquals(t *testing.T, expected, actual Entry) {
	if !reflect.DeepEqual(expected, actual) {
		testutil.DebugLine(t)
		t.Error("Expected Entry", expected, "but received", actual)
	}
}

func assertNamespaceEquals(t *testing.T, expected, actual Namespace) {
	if !reflect.DeepEqual(expected, actual) {
		testutil.DebugLine(t)
		t.Error("Expected Namespace", expected, "but received", actual)
	}
}

func assertNamespaceNotEquals(t *testing.T, other, actual Namespace) {
	if reflect.DeepEqual(other, actual) {
		testutil.DebugLine(t)
		t.Error("Unexpected Namespace", other, "was equal to", actual)
	}
}

func assertTableEquals(t *testing.T, expected, actual Table) {
	if !reflect.DeepEqual(expected, actual) {
		testutil.DebugLine(t)
		t.Error("Expected Table", expected, "but received", actual)
	}
}

func assertRowEquals(t *testing.T, expected, actual Row) {
	if !reflect.DeepEqual(expected, actual) {
		testutil.DebugLine(t)
		t.Error("Expected Row", expected, "but received", actual)
	}
}

func readNamespaceStream(stream []NamespaceStreamEntry) Namespace {
	ns, invalid, err := ReadNamespaceStream(stream)

	panicInvalidNamespace(invalid)

	if err != nil {
		panic(err)
	}

	return ns
}

func encodeNamespace(ns Namespace, w io.Writer) error {
	invalid, err := EncodeNamespace(ns, w)

	panicInvalidNamespace(invalid)

	return err
}

func decodeNamespace(r io.Reader) (Namespace, error) {
	namespace, invalid, err := DecodeNamespace(r)

	panicInvalidNamespace(invalid)

	return namespace, err
}

func namespaceText(namespace Namespace) string {
	message, invalid := MakeNamespaceMessage(namespace)

	panicInvalidNamespace(invalid)

	buff := &bytes.Buffer{}

	err := proto.MarshalText(buff, message)

	if err != nil {
		panic(err)
	}

	return buff.String()
}

func panicInvalidNamespace(invalid []InvalidNamespaceEntry) {
	invalidCount := len(invalid)
	if invalidCount > 0 {
		panic(fmt.Sprintf("%v invalid entries"))
	}
}
