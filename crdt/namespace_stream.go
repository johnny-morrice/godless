package crdt

import (
	"sort"

	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/pkg/errors"
)

type StreamPoint struct {
	Text      string
	Signature crypto.SignatureText
}

// FIXME not really a stream, whole is kept in memory.
type NamespaceStreamEntry struct {
	Table  TableName
	Row    RowName
	Entry  EntryName
	Points []StreamPoint
}

func (entry NamespaceStreamEntry) Equals(other NamespaceStreamEntry) bool {
	ok := entry.Table == other.Table
	ok = ok && entry.Row == other.Row
	ok = ok && entry.Entry == other.Entry

	if !ok {
		return false
	}

	for i, myPoint := range entry.Points {
		theirPoint := other.Points[i]
		if myPoint != theirPoint {
			return false
		}
	}

	return true
}

func StreamEquals(a, b []NamespaceStreamEntry) bool {
	for i, ar := range a {
		br := b[i]
		if !ar.Equals(br) {
			return false
		}
	}

	return true
}

func SortNamespaceStream(stream []NamespaceStreamEntry) {
	sort.Sort(byNamespaceStreamOrder(stream))
}

type byNamespaceStreamOrder []NamespaceStreamEntry

func (stream byNamespaceStreamOrder) Len() int {
	return len(stream)
}

func (stream byNamespaceStreamOrder) Swap(i, j int) {
	stream[i], stream[j] = stream[j], stream[i]
}

func (stream byNamespaceStreamOrder) Less(i, j int) bool {
	a, b := stream[i], stream[j]

	if a.Table < b.Table {
		return true
	} else if a.Table > b.Table {
		return false
	}

	if a.Row < b.Row {
		return true
	} else if a.Row > b.Row {
		return false
	}

	if a.Entry < b.Entry {
		return true
	} else if a.Entry > b.Entry {
		return false
	}

	return pointLess(a.Points, b.Points)
}

type byEntryOrder []Entry

func (entries byEntryOrder) Len() int {
	return len(entries)
}

func (entries byEntryOrder) Swap(i, j int) {
	entries[i], entries[j] = entries[j], entries[i]
}

// FIXME actually do comparsion
func (entries byEntryOrder) Less(i, j int) bool {
	panic("not implemented")
	return false
}

func pointLess(a, b []StreamPoint) bool {
	minSize := util.Imin(len(a), len(b))

	for i := 0; i < minSize; i++ {
		ap := a[i]
		bp := b[i]

		if ap.Text < bp.Text {
			return true
		} else if ap.Text > bp.Text {
			return false
		}

		if ap.Signature < bp.Signature {
			return true
		} else if ap.Signature > bp.Signature {
			return false
		}
	}

	return len(a) < len(b)
}

// TODO does not support unsigned links.
func MakeStreamPoints(point Point) []StreamPoint {
	count := 1
	sigCount := len(point.Signatures)
	if sigCount > 0 {
		count = sigCount
	}
	stream := make([]StreamPoint, count)

	for _, sig := range point.Signatures {
		stream = append(stream, MakeStreamPoint(point.Text, sig))
	}

	return stream
}

func MakeStreamPoint(text string, sig crypto.Signature) StreamPoint {
	return StreamPoint{Text: text, Signature: crypto.PrintSignature(sig)}
}

// TODO should take slice of StreamPoint
func ReadStreamPoint(stream StreamPoint) (Point, error) {
	const failMsg = "ReadStreamPoint failed"

	_, err := crypto.ParseSignature(stream.Signature)

	if err != nil {
		return Point{}, errors.Wrap(err, failMsg)
	}

	point := Point{
		Text: stream.Text,
	}

	panic("not implemented")

	return point, nil
}

func MakeStreamEntry(tname TableName, rname RowName, ename EntryName, entry Entry) NamespaceStreamEntry {
	panic("not implemented")

	count := 0

	points := entry.GetValues()
	streamPoints := make([]StreamPoint, count)

	for _, p := range points {
		streamPoints = append(streamPoints, MakeStreamPoints(p)...)
	}

	return NamespaceStreamEntry{
		Table:  tname,
		Row:    rname,
		Entry:  ename,
		Points: streamPoints,
	}
}

func MakeTableStream(tableKey TableName, table Table) []NamespaceStreamEntry {
	subNamespace := MakeNamespace(map[TableName]Table{
		tableKey: table,
	})
	return MakeNamespaceStream(subNamespace)
}

func MakeRowStream(tableKey TableName, rowKey RowName, row Row) []NamespaceStreamEntry {
	table := MakeTable(map[RowName]Row{
		rowKey: row,
	})
	return MakeTableStream(tableKey, table)
}

func MakeNamespaceStream(ns Namespace) []NamespaceStreamEntry {
	stream := []NamespaceStreamEntry{}
	for tableName, table := range ns.Tables {
		for rowName, row := range table.Rows {
			for entryName, entry := range row.Entries {
				if len(entry.Set) > 0 {
					streamEntry := MakeStreamEntry(tableName, rowName, entryName, entry)
					stream = append(stream, streamEntry)
				}
			}
		}
	}

	sort.Sort(byNamespaceStreamOrder(stream))
	return stream
}

func ReadNamespaceStream(stream []NamespaceStreamEntry) Namespace {
	ns := EmptyNamespace()

	for _, streamEntry := range stream {
		ns.addStreamEntry(streamEntry)
	}

	return ns
}
