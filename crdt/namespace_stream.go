package crdt

import (
	"sort"

	"github.com/ethereum/go-ethereum/log"
	"github.com/johnny-morrice/godless/internal/crypto"
	"github.com/pkg/errors"
)

type StreamPoint struct {
	Text      PointText
	Signature crypto.SignatureText
}

func (point StreamPoint) Equals(other StreamPoint) bool {
	return point.Text == other.Text && point.Signature == other.Signature
}

func (point StreamPoint) Less(other StreamPoint) bool {
	if point.Text < other.Text {
		return true
	} else if point.Text > other.Text {
		return false
	}

	return point.Signature < other.Signature
}

// FIXME not really a stream, whole is kept in memory.
type NamespaceStreamEntry struct {
	Table TableName
	Row   RowName
	Entry EntryName
	Point StreamPoint
}

func (entry NamespaceStreamEntry) SamePoint(other NamespaceStreamEntry) bool {
	ok := entry.Table == other.Table
	ok = ok && entry.Row == other.Row
	ok = ok && entry.Entry == other.Entry
	ok = ok && entry.Point.Text == other.Point.Text
	return ok
}

func (entry NamespaceStreamEntry) Equals(other NamespaceStreamEntry) bool {
	ok := entry.Table == other.Table
	ok = ok && entry.Row == other.Row
	ok = ok && entry.Entry == other.Entry

	if !ok {
		return false
	}

	if !entry.Point.Equals(other.Point) {
		return false
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

	return a.Point.Less(b.Point)
}

// TODO does not support unsigned links.
func makeStreamPoints(proto NamespaceStreamEntry, point Point, index int, stream []NamespaceStreamEntry) int {
	if len(point.Signatures) == 0 {
		entry := proto
		entry.Point = MakeStreamPoint(point.Text, crypto.Signature{})
		stream[index] = entry
		return index + 1
	}

	for _, sig := range point.Signatures {
		entry := proto
		proto.Point = MakeStreamPoint(point.Text, sig)
		stream[index] = entry
		index++
	}

	return index
}

func MakeStreamPoint(text PointText, sig crypto.Signature) StreamPoint {
	return StreamPoint{Text: text, Signature: crypto.PrintSignature(sig)}
}

func readStreamPoint(stream []NamespaceStreamEntry) (Point, []InvalidStreamEntry, error) {
	const failMsg = "readStreamPoint failed"

	if len(stream) == 0 {
		return Point{}, nil, nil
	}

	first := stream[0]
	point := Point{
		Text:       first.Point.Text,
		Signatures: make([]crypto.Signature, 0, len(stream)),
	}

	invalid := make([]InvalidStreamEntry, 0, len(stream))

	for _, entry := range stream {
		if !entry.SamePoint(first) {
			notSame := errors.New("Corrupt stream")
			return Point{}, nil, errors.Wrap(notSame, failMsg)
		}

		if entry.Point.Signature == "" {
			continue
		}

		sig, err := crypto.ParseSignature(entry.Point.Signature)

		if err != nil {
			log.Warn("Failed to parse signature")
			invalid = append(invalid, InvalidStreamEntry(entry))
			continue
		}

		point.Signatures = append(point.Signatures, sig)
	}

	return point, invalid, nil
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
	count := streamLength(ns)

	index := 0
	stream := make([]NamespaceStreamEntry, 0, count)

	foreachEntry(ns, func(t TableName, r RowName, e EntryName, entry Entry) {
		proto := NamespaceStreamEntry{
			Table: t,
			Row:   r,
			Entry: e,
		}

		for _, point := range entry.GetValues() {
			index = makeStreamPoints(proto, point, index, stream)
		}
	})

	sort.Sort(byNamespaceStreamOrder(stream))
	return stream
}

func streamLength(ns Namespace) int {
	count := 0

	foreachEntry(ns, func(t TableName, r RowName, e EntryName, entry Entry) {
		for _, point := range entry.GetValues() {
			sigCount := len(point.Signatures)
			if sigCount > 0 {
				count += sigCount
			} else {
				count++
			}
		}
	})

	return count
}

// Iterate through all entries
func foreachEntry(ns Namespace, f func(t TableName, r RowName, e EntryName, entry Entry)) {
	for tableName, table := range ns.Tables {
		for rowName, row := range table.Rows {
			for entryName, entry := range row.Entries {
				f(tableName, rowName, entryName, entry)
			}
		}
	}
}

func ReadNamespaceStream(stream []NamespaceStreamEntry) Namespace {
	ns := EmptyNamespace()

	batchStart := 0
	for i, entry := range stream {
		startEntry := stream[batchStart]

		if !entry.SamePoint(startEntry) {
			batchEnd := i + 1
			ns.addPointBatch(stream[batchStart:batchEnd])
			batchStart = batchEnd
		}

	}

	return ns
}
