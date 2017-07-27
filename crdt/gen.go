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

	tableCount := testutil.GenCountRange(rand, 1, size, tableFudge)
	for i := 0; i < tableCount; i++ {
		tableName := TableName(testutil.RandLettersRange(rand, 1, maxStr))
		table := EmptyTable()
		rowCount := testutil.GenCountRange(rand, 1, size, rowFudge)
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
	if size < 10 {
		size = 10
	}

	const maxStr = 100
	const entryFudge = 0.65
	row := EmptyRow()
	entryCount := testutil.GenCountRange(rand, 1, size, entryFudge)
	for k := 0; k < entryCount; k++ {
		entryName := EntryName(testutil.RandLetters(rand, maxStr))
		entry := GenEntry(rand, size)
		row.addEntry(entryName, entry)
	}
	return row
}

func GenEntry(rand *rand.Rand, size int) Entry {
	if size < 10 {
		size = 10
	}

	const maxStr = 100
	const pointFudge = 0.85
	pointCount := testutil.GenCountRange(rand, 1, size, pointFudge)
	points := make([]Point, pointCount)

	for m := 0; m < pointCount; m++ {
		points[m] = genPoint(rand, maxStr)
	}

	return MakeEntry(points)
}

func genPoint(rand *rand.Rand, size int) Point {
	return UnsignedPoint(PointText(testutil.RandLettersRange(rand, 1, size)))
}

func GenIndex(rand *rand.Rand, size int) Index {
	if size < 20 {
		size = 20
	}

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
	if size < 1 {
		size = 2
	}

	if size > 10 {
		size = 10
	}

	addr := testutil.RandLettersRange(rand, 1, size)
	return UnsignedLink(IPFSPath(addr))
}
