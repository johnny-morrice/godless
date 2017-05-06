package godless

import (
	"reflect"
	"runtime"
	"testing"
)

func TestEmptyNamespace(t *testing.T) {
	expected := Namespace{
		tables: map[string]Table{},
	}
	actual := EmptyNamespace()

	assertNamespaceEquals(t, expected, actual)
}

func TestMakeNamespace(t *testing.T) {
	table := EmptyTable()

	expected := Namespace{
		tables: map[string]Table{
			"foo": table,
		},
	}
	actual := MakeNamespace(map[string]Table{"foo": table})

	assertNamespaceEquals(t, expected, actual)
}

func TestNamespaceIsEmpty(t *testing.T) {
	table := EmptyTable()

	full := MakeNamespace(map[string]Table{"foo": table})

	empty := EmptyNamespace()

	if !empty.IsEmpty() {
		t.Error("Expected IsEmpty():", empty)
	}

	if full.IsEmpty() {
		t.Error("Unexpected IsEmpty():", full)
	}
}

func TestNamespaceCopy(t *testing.T) {
	expected := MakeNamespace(map[string]Table{"foo": EmptyTable()})
	actual := expected.Copy()

	assertNamespaceEquals(t, expected, actual)
}

func TestNamespaceJoinNamespace(t *testing.T) {
	table := EmptyTable()
	foo := MakeNamespace(map[string]Table{"foo": table})
	bar := MakeNamespace(map[string]Table{"bar": table})

	expectedJoin := MakeNamespace(map[string]Table{"foo": table, "bar": table})

	actualJoinFooBar := foo.JoinNamespace(bar)
	actualJoinBarFoo := bar.JoinNamespace(foo)

	assertNamespaceEquals(t, expectedJoin, actualJoinFooBar)
	assertNamespaceEquals(t, expectedJoin, actualJoinBarFoo)
}

func TestNamespaceJoinTable(t *testing.T) {
	table := EmptyTable()
	foo := MakeNamespace(map[string]Table{"foo": table})
	barTable := EmptyTable()

	expectedFoo := MakeNamespace(map[string]Table{"foo": table})
	expectedJoin := MakeNamespace(map[string]Table{"foo": table, "bar": table})

	actual := foo.JoinTable("bar", barTable)

	assertNamespaceEquals(t, expectedFoo, foo)
	assertNamespaceEquals(t, expectedJoin, actual)
}

func TestNamespaceGetTable(t *testing.T) {
	expectedTable := MakeTable(map[string]Row{"foo": EmptyRow()})
	expectedEmptyTable := Table{}

	hasTable := MakeNamespace(map[string]Table{"bar": expectedTable})
	hasNoTable := EmptyNamespace()

	var actualTable Table
	var err error
	actualTable, err = hasTable.GetTable("bar")

	assertTableEquals(t, expectedTable, actualTable)

	assertNilError(t, err)

	actualTable, err = hasNoTable.GetTable("bar")

	assertTableEquals(t, expectedEmptyTable, actualTable)

	assertNonNilError(t, err)
}

func TestNamespaceEquals(t *testing.T) {

	table := EmptyTable()
	nsA := MakeNamespace(map[string]Table{"foo": table})
	nsB := MakeNamespace(map[string]Table{"bar": table})
	nsC := EmptyNamespace()
	nsD := MakeNamespace(map[string]Table{
		"foo": MakeTable(map[string]Row{
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
	expected := Table{rows: map[string]Row{}}
	actual := EmptyTable()

	assertTableEquals(t, expected, actual)
}

func TestMakeTable(t *testing.T) {
	expected := Table{
		rows: map[string]Row{
			"foo": EmptyRow(),
		},
	}
	actual := MakeTable(map[string]Row{"foo": EmptyRow()})

	assertTableEquals(t, expected, actual)
}

func TestTableCopy(t *testing.T) {
	expected := MakeTable(map[string]Row{"foo": EmptyRow()})
	actual := expected.Copy()

	assertTableEquals(t, expected, actual)
}

func TestTableAllRows(t *testing.T) {
	emptyRow := EmptyRow()
	fullRow := MakeRow(map[string]Entry{
		"baz": EmptyEntry(),
	})

	expected := []Row{
		emptyRow, fullRow,
	}

	table := MakeTable(map[string]Row{
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
	barRow := MakeRow(map[string]Entry{
		"Baz": EmptyEntry(),
	})

	foo := MakeTable(map[string]Row{
		"foo": emptyRow,
	})

	bar := MakeTable(map[string]Row{
		"bar": barRow,
	})

	expected := MakeTable(map[string]Row{
		"foo": emptyRow,
		"bar": barRow,
	})

	actual := foo.JoinTable(bar)

	assertTableEquals(t, expected, actual)
}

func TestTableJoinRow(t *testing.T) {
	emptyTable := EmptyTable()
	row := MakeRow(map[string]Entry{
		"bar": MakeEntry([]string{"hello"}),
	})

	expected := MakeTable(map[string]Row{
		"foo": row,
	})

	actual := emptyTable.JoinRow("foo", row)

	assertTableEquals(t, expected, actual)
}

func TestTableGetRow(t *testing.T) {
	expected := MakeRow(map[string]Entry{
		"bar": EmptyEntry(),
	})

	table := MakeTable(map[string]Row{
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
		MakeTable(map[string]Row{
			"foo": EmptyRow(),
		}),
		MakeTable(map[string]Row{
			"bar": EmptyRow(),
		}),
		MakeTable(map[string]Row{
			"foo": MakeRow(map[string]Entry{
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
		entries: map[string]Entry{},
	}

	actual := EmptyRow()

	assertRowEquals(t, expected, actual)
}

func TestMakeRow(t *testing.T) {
	entry := EmptyEntry()

	expected := Row{
		entries: map[string]Entry{
			"foo": entry,
		},
	}

	actual := MakeRow(map[string]Entry{
		"foo": entry,
	})

	assertRowEquals(t, expected, actual)
}

func TestRowCopy(t *testing.T) {
	expected := MakeRow(map[string]Entry{"foo": EmptyEntry()})
	actual := expected.Copy()
	assertRowEquals(t, expected, actual)
}

func TestRowJoinRow(t *testing.T) {
	emptyEntry := EmptyEntry()
	fullEntry := MakeEntry([]string{"hi"})

	expected := MakeRow(map[string]Entry{
		"foo": emptyEntry,
		"bar": fullEntry,
	})

	foo := MakeRow(map[string]Entry{
		"foo": emptyEntry,
	})

	bar := MakeRow(map[string]Entry{
		"bar": fullEntry,
	})

	actual := foo.JoinRow(bar)

	assertRowEquals(t, expected, actual)
}

func TestRowGetEntry(t *testing.T) {
	expected := MakeEntry([]string{"hi"})

	row := MakeRow(map[string]Entry{
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
	fullEntry := MakeEntry([]string{"hi"})

	expected := MakeRow(map[string]Entry{
		"foo": emptyEntry,
		"bar": fullEntry,
	})

	foo := MakeRow(map[string]Entry{
		"foo": emptyEntry,
	})

	actual := foo.JoinEntry("bar", fullEntry)

	assertRowEquals(t, expected, actual)
}

func TestRowEquals(t *testing.T) {
	emptyEntry := EmptyEntry()
	fullEntry := MakeEntry([]string{"hi"})

	rows := []Row{
		EmptyRow(),
		MakeRow(map[string]Entry{
			"foo": emptyEntry,
		}),
		MakeRow(map[string]Entry{
			"bar": emptyEntry,
		}),
		MakeRow(map[string]Entry{
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
		set: []string{},
	}

	actual := EmptyEntry()

	assertEntryEquals(t, expected, actual)
}

func TestMakeEntry(t *testing.T) {
	expected := Entry{
		set: []string{"hello", "world"},
	}

	actuals := []Entry{
		MakeEntry([]string{"hello", "world"}),
		MakeEntry([]string{"world", "hello"}),
		MakeEntry([]string{"hello", "hello", "world", "world"}),
	}

	for _, a := range actuals {
		assertEntryEquals(t, expected, a)
	}
}

func TestEntryJoinEntry(t *testing.T) {
	expected := Entry{
		set: []string{"hello", "world"},
	}

	hello := MakeEntry([]string{"hello"})
	world := MakeEntry([]string{"world"})

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
		MakeEntry([]string{"hi"}),
		MakeEntry([]string{"hello", "world"}),
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
	expected := []string{"hello"}
	entry := MakeEntry([]string{"hello"})
	actual := entry.GetValues()
	if !reflect.DeepEqual(expected, actual) {
		t.Error("Expected", expected, "but was", actual)
	}
}

func assertEntryEquals(t *testing.T, expected, actual Entry) {
	if !reflect.DeepEqual(expected, actual) {
		debugLine(t)
		t.Error("Expected Entry", expected, "but received", actual)
	}
}

func assertNamespaceEquals(t *testing.T, expected, actual Namespace) {
	if !reflect.DeepEqual(expected, actual) {
		debugLine(t)
		t.Error("Expected Namespace", expected, "but received", actual)
	}
}

func assertNamespaceNotEquals(t *testing.T, other, actual Namespace) {
	if reflect.DeepEqual(other, actual) {
		debugLine(t)
		t.Error("Unexpected Namespace", other, "was equal to", actual)
	}
}

func assertTableEquals(t *testing.T, expected, actual Table) {
	if !reflect.DeepEqual(expected, actual) {
		debugLine(t)
		t.Error("Expected Table", expected, "but received", actual)
	}
}

func assertRowEquals(t *testing.T, expected, actual Row) {
	if !reflect.DeepEqual(expected, actual) {
		debugLine(t)
		t.Error("Expected Row", expected, "but received", actual)
	}
}

func assertNilError(t *testing.T, err error) {
	if err != nil {
		debugLine(t)
		t.Error("Unexpected error:", err)
	}
}

func assertNonNilError(t *testing.T, err error) {
	if err == nil {
		debugLine(t)
		t.Error("Expected error")
	}
}

func debugLine(t *testing.T) {
	_, _, line, ok := runtime.Caller(CALLER_DEPTH)

	if !ok {
		panic("debugLine failed")
	}

	t.Log("Test failed at line", line)
}

const CALLER_DEPTH = 2
