package godless

import (
	"reflect"
	"runtime"
	"testing"
)

func TestEmptyNamespace(t *testing.T) {
	expected := &Namespace{
		tables: map[string]Table{},
	}
	actual := EmptyNamespace()

	assertNamespaceEquals(t, expected, actual)
}

func TestMakeNamespace(t *testing.T) {
	table := EmptyTable()

	expected := &Namespace{
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

	if expected == actual {
		t.Error("Unexpected pointer equality")
	}

	assertNamespaceEquals(t, expected, actual)
}

func TestNamespaceJoinNamespace(t *testing.T) {
	table := EmptyTable()
	foo := MakeNamespace(map[string]Table{"foo": table})
	bar := MakeNamespace(map[string]Table{"bar": table})

	expectedJoin := MakeNamespace(map[string]Table{"foo": table, "bar": table})
	expectedFoo := MakeNamespace(map[string]Table{"foo": table})
	expectedBar := MakeNamespace(map[string]Table{"bar": table})

	actualJoinFooBar, fooBarErr := foo.JoinNamespace(bar)
	actualJoinBarFoo, barFooErr := bar.JoinNamespace(foo)

	assertNilError(t, fooBarErr)
	assertNilError(t, barFooErr)

	assertNamespaceEquals(t, expectedJoin, actualJoinFooBar)
	assertNamespaceEquals(t, expectedJoin, actualJoinBarFoo)
	assertNamespaceEquals(t, expectedFoo, foo)
	assertNamespaceEquals(t, expectedBar, bar)
}

func TestNamespaceJoinTable(t *testing.T) {
	table := EmptyTable()
	foo := MakeNamespace(map[string]Table{"foo": table})
	barTable := EmptyTable()

	expectedFoo := MakeNamespace(map[string]Table{"foo": table})
	expectedJoin := MakeNamespace(map[string]Table{"foo": table, "bar": table})

	actual, err := foo.JoinTable("bar", barTable)

	assertNilError(t, err)

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

	foos := []*Namespace{
		nsA, nsB, nsC, nsD,
	}

	bars := make([]*Namespace, len(foos))
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
	t.Fail()
}

func TestRowMakeTable(t *testing.T) {
	t.Fail()
}

func TestTableForeachrow(t *testing.T) {
	t.Fail()
}

func TestTableCopy(t *testing.T) {
	t.Fail()
}

func TestTableAllRows(t *testing.T) {
	t.Fail()
}

func TestTableJoinTable(t *testing.T) {
	t.Fail()
}

func TestTableJoinRow(t *testing.T) {
	t.Fail()
}

func TestTableGetRow(t *testing.T) {
	t.Fail()
}

func TestTableEquals(t *testing.T) {
	t.Fail()
}

func TestEmptyRow(t *testing.T) {
	t.Fail()
}

func TestEntryMakeRow(t *testing.T) {
	t.Fail()
}

func TestRowCopy(t *testing.T) {
	t.Fail()
}

func TestRowJoinRow(t *testing.T) {
	t.Fail()
}

func TestRowGetEntry(t *testing.T) {
	t.Fail()
}

func TestRowJoinEntry(t *testing.T) {
	t.Fail()
}

func TestRowEquals(t *testing.T) {
	t.Fail()
}

func TestEmptyEntry(t *testing.T) {
	t.Fail()
}

func TestMakeEntry(t *testing.T) {
	t.Fail()
}

func TestEntryJoinEntry(t *testing.T) {
	t.Fail()
}

func TestEntryEquals(t *testing.T) {
	t.Fail()
}

func TestEntryGetValues(t *testing.T) {
	t.Fail()
}

func assertNamespaceEquals(t *testing.T, expected, actual *Namespace) {
	if !reflect.DeepEqual(expected, actual) {
		debugLine(t)
		t.Error("Expected Namespace", expected, "but received", actual)
	}
}

func assertNamespaceNotEquals(t *testing.T, other, actual *Namespace) {
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
