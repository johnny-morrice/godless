package crdt

import (
	"math/rand"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func GenNamespace(rand *rand.Rand, size int) Namespace {
	if size > 30 {
		size = 30
	}

	const tableMax = 10
	const maxStr = 20

	gen := EmptyNamespace()

	tableCount := testutil.GenCountRange(rand, 1, tableMax)
	for i := 0; i < tableCount; i++ {
		tableName := TableName(testutil.RandLettersRange(rand, 1, maxStr))
		table := EmptyTable()
		rowCount := testutil.GenCountRange(rand, 1, size)
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
	if size > 10 {
		size = 10
	}

	if size < 5 {
		size = 5
	}

	const maxStr = 20
	row := EmptyRow()
	entryCount := testutil.GenCountRange(rand, 1, size)
	for k := 0; k < entryCount; k++ {
		entryName := EntryName(testutil.RandLetters(rand, maxStr))
		entry := GenEntry(rand, size)
		row.addEntry(entryName, entry)
	}
	return row
}

func GenEntry(rand *rand.Rand, size int) Entry {
	if size > 5 {
		size = 5
	}

	if size < 2 {
		size = 2
	}

	const maxStr = 20
	pointCount := testutil.GenCountRange(rand, 1, size)
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
	if size > 20 {
		size = 20
	}

	if size < 10 {
		size = 10
	}

	index := EmptyIndex()

	for i := 0; i < size; i++ {
		keyCount := testutil.GenCountRange(rand, 2, size)
		indexKey := TableName(testutil.RandLettersRange(rand, 1, keyCount))
		addrCount := testutil.GenCountRange(rand, 1, size)
		addrs := make([]Link, addrCount)
		for j := 0; j < addrCount; j++ {
			pathCount := testutil.GenCountRange(rand, 2, size)
			a := testutil.RandLettersRange(rand, 1, pathCount)
			addrs[j] = UnsignedLink(IPFSPath(a))
		}

		index.addTable(indexKey, addrs...)
	}

	return index
}

func GenLink(rand *rand.Rand, size int) Link {
	if size > 10 {
		size = 10
	}

	if size < 2 {
		size = 2
	}

	addr := testutil.RandLettersRange(rand, 1, size)
	return UnsignedLink(IPFSPath(addr))
}
