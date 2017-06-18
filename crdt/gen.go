package crdt

import (
	"math/rand"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func GenNamespace(rand *rand.Rand, size int) Namespace {
	const maxStr = 100
	const tableFudge = 0.125
	const rowFudge = 0.25

	gen := EmptyNamespace()

	// FIXME This looks horrific.
	tableCount := testutil.GenCount(rand, size, tableFudge)
	for i := 0; i < tableCount; i++ {
		tableName := TableName(testutil.RandLetters(rand, maxStr))
		table := EmptyTable()
		rowCount := testutil.GenCount(rand, size, rowFudge)
		for j := 0; j < rowCount; j++ {
			rowName := RowName(testutil.RandLetters(rand, maxStr))
			row := genRow(rand, size)
			table.addRow(rowName, row)
		}
		gen.addTable(tableName, table)
	}

	return gen
}

func genRow(rand *rand.Rand, size int) Row {
	const maxStr = 100
	const entryFudge = 0.65
	const pointFudge = 0.85
	row := EmptyRow()
	entryCount := testutil.GenCountRange(rand, 1, size, entryFudge)
	for k := 0; k < entryCount; k++ {
		entryName := EntryName(testutil.RandLetters(rand, maxStr))
		pointCount := testutil.GenCountRange(rand, 1, size, pointFudge)
		points := make([]Point, pointCount)

		for m := 0; m < pointCount; m++ {
			points[m] = genPoint(rand, maxStr)
		}

		entry := MakeEntry(points)
		row.addEntry(entryName, entry)
	}
	return row
}

func genPoint(rand *rand.Rand, size int) Point {
	return Point{Text: PointText(testutil.RandLetters(rand, size))}
}
