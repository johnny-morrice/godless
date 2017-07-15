package crdt

import (
	"math"
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
		tableName := TableName(testutil.RandLettersRange(rand, 1, maxStr))
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
	return UnsignedPoint(PointText(testutil.RandLettersRange(rand, 1, size)))
}

func GenIndex(rand *rand.Rand, size int) Index {
	index := EmptyIndex()
	const ADDR_SCALE = 1
	const KEY_SCALE = 0.5
	const PATH_SCALE = 0.5

	for i := 0; i < size; i++ {
		keyCount := testutil.GenCountRange(rand, 2, size, KEY_SCALE)
		indexKey := TableName(testutil.RandLettersRange(rand, 1, keyCount))
		addrCount := testutil.GenCountRange(rand, 1, size, ADDR_SCALE)
		addrs := make([]Link, addrCount)
		for j := 0; j < addrCount; j++ {
			pathCount := testutil.GenCountRange(rand, 2, size, PATH_SCALE)
			a := testutil.RandLettersRange(rand, 1, pathCount)
			addrs[j] = UnsignedLink(IPFSPath(a))
		}

		index.addTable(indexKey, addrs...)
	}

	return index
}

func GenLink(rand *rand.Rand, size int) Link {
	const PATH_SCALE = 0.5
	maxLen := int(math.Floor(float64(size) / PATH_SCALE))
	addr := testutil.RandLettersRange(rand, 1, maxLen)
	return UnsignedLink(IPFSPath(addr))
}
