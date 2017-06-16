package testutil

import (
	"math/rand"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/testutil"
)

func GenIndex(rand *rand.Rand, size int) crdt.Index {
	index := crdt.EmptyIndex()
	const ADDR_SCALE = 1
	const KEY_SCALE = 0.5
	const PATH_SCALE = 0.5

	for i := 0; i < size; i++ {
		keyCount := testutil.GenCountRange(rand, 1, size, KEY_SCALE)
		indexKey := crdt.TableName(testutil.RandPoint(rand, keyCount))
		addrCount := testutil.GenCountRange(rand, 1, size, ADDR_SCALE)
		addrs := make([]crdt.IPFSPath, addrCount)
		for j := 0; j < addrCount; j++ {
			pathCount := testutil.GenCountRange(rand, 1, size, PATH_SCALE)
			a := testutil.RandPoint(rand, pathCount)
			addrs[j] = crdt.IPFSPath(a)
		}

		index.Index[indexKey] = addrs
	}

	return index
}

func GenNamespace(rand *rand.Rand, size int) crdt.Namespace {
	const maxStr = 100
	const tableFudge = 0.125
	const rowFudge = 0.25

	gen := crdt.EmptyNamespace()

	// FIXME This looks horrific.
	tableCount := testutil.GenCount(rand, size, tableFudge)
	for i := 0; i < tableCount; i++ {
		tableName := crdt.TableName(testutil.RandLetters(rand, maxStr))
		table := crdt.EmptyTable()
		rowCount := testutil.GenCount(rand, size, rowFudge)
		for j := 0; j < rowCount; j++ {
			rowName := crdt.RowName(testutil.RandLetters(rand, maxStr))
			row := GenRow(rand, size)
			table = table.JoinRow(rowName, row)
		}
		gen = gen.JoinTable(tableName, table)
	}

	return gen
}

func GenRow(rand *rand.Rand, size int) crdt.Row {
	const maxStr = 100
	const entryFudge = 0.65
	const pointFudge = 0.85
	row := crdt.EmptyRow()
	entryCount := testutil.GenCountRange(rand, 1, size, entryFudge)
	for k := 0; k < entryCount; k++ {
		entryName := crdt.EntryName(testutil.RandLetters(rand, maxStr))
		pointCount := testutil.GenCountRange(rand, 1, size, pointFudge)
		points := make([]crdt.Point, pointCount)

		for m := 0; m < pointCount; m++ {
			points[m] = crdt.Point(testutil.RandLetters(rand, maxStr))
		}

		entry := crdt.MakeEntry(points)
		row = row.JoinEntry(entryName, entry)
	}
	return row
}
