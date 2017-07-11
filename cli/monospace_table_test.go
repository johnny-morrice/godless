package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/johnny-morrice/godless/internal/testutil"
)

const tableText = `
--------------------------------------
| Column A | Column B     | Column C |
--------------------------------------
| Hello    | Mr Gutenberg | Ducks    |
| Wonder   | Soldier      | Zaps     |
--------------------------------------`

func TestMonospaceTable(t *testing.T) {
	expected := strings.Trim(tableText, " \n\t")

	table := &monospaceTable{}
	err := table.addColumn("Column A", "Column B", "Column C")
	testutil.AssertNil(t, err)
	err = table.addRow("Hello", "Mr Gutenberg", "Ducks")
	testutil.AssertNil(t, err)
	err = table.addRow("Wonder", "Soldier", "Zaps")
	testutil.AssertNil(t, err)

	buff := &bytes.Buffer{}
	table.fprint(buff)

	actual := buff.String()
	testutil.AssertEquals(t, "Unexpected table", expected, actual)

	err = table.addRow("Hola")
	testutil.AssertNonNil(t, err)

	err = table.addColumn("More!")
	testutil.AssertNonNil(t, err)
}
