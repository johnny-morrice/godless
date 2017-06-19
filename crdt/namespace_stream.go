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

type InvalidNamespaceEntry NamespaceStreamEntry

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
func makeStreamPoints(proto NamespaceStreamEntry, point Point, index int, stream []NamespaceStreamEntry) (int, []InvalidNamespaceEntry) {
	if len(point.Signatures) == 0 {
		entry := proto
		entry.Point = StreamPoint{Text: point.Text}
		stream[index] = entry
		return index + 1, nil
	}

	invalid := []InvalidNamespaceEntry{}

	for _, sig := range point.Signatures {
		entry := proto
		streamPoint, err := MakeStreamPoint(point.Text, sig)

		if err != nil {
			log.Warn("Invalid stream point: %v", err)
			entry.Point.Text = point.Text
			invalid = append(invalid, InvalidNamespaceEntry(entry))
			continue
		}

		entry.Point = streamPoint
		stream[index] = entry
		index++
	}

	return index, invalid
}

func MakeStreamPoint(text PointText, sig crypto.Signature) (StreamPoint, error) {
	sigText, err := crypto.PrintSignature(sig)

	if err != nil {
		return StreamPoint{}, err
	}

	return StreamPoint{Text: text, Signature: sigText}, nil
}

func readStreamPoint(stream []NamespaceStreamEntry) (Point, []InvalidNamespaceEntry, error) {
	const failMsg = "readStreamPoint failed"

	if len(stream) == 0 {
		return Point{}, nil, nil
	}

	first := stream[0]
	point := Point{
		Text:       first.Point.Text,
		Signatures: make([]crypto.Signature, 0, len(stream)),
	}

	invalid := make([]InvalidNamespaceEntry, 0, len(stream))

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
			invalid = append(invalid, InvalidNamespaceEntry(entry))
			continue
		}

		point.Signatures = append(point.Signatures, sig)
	}

	return point, invalid, nil
}

func MakeTableStream(tableKey TableName, table Table) ([]NamespaceStreamEntry, []InvalidNamespaceEntry) {
	subNamespace := MakeNamespace(map[TableName]Table{
		tableKey: table,
	})
	return MakeNamespaceStream(subNamespace)
}

func MakeRowStream(tableKey TableName, rowKey RowName, row Row) ([]NamespaceStreamEntry, []InvalidNamespaceEntry) {
	table := MakeTable(map[RowName]Row{
		rowKey: row,
	})
	return MakeTableStream(tableKey, table)
}

func MakeNamespaceStream(ns Namespace) ([]NamespaceStreamEntry, []InvalidNamespaceEntry) {
	count := streamLength(ns)

	index := 0
	stream := make([]NamespaceStreamEntry, 0, count)
	invalid := []InvalidNamespaceEntry{}

	ns.ForeachEntry(func(t TableName, r RowName, e EntryName, entry Entry) {
		proto := NamespaceStreamEntry{
			Table: t,
			Row:   r,
			Entry: e,
		}

		for _, point := range entry.GetValues() {
			var moreInvalid []InvalidNamespaceEntry
			index, moreInvalid = makeStreamPoints(proto, point, index, stream)

			if len(moreInvalid) > 0 {
				invalid = append(invalid, moreInvalid...)
			}
		}
	})

	sort.Sort(byNamespaceStreamOrder(stream))
	return stream, invalid
}

func streamLength(ns Namespace) int {
	count := 0

	ns.ForeachEntry(func(t TableName, r RowName, e EntryName, entry Entry) {
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

func ReadNamespaceStream(stream []NamespaceStreamEntry) (Namespace, []InvalidNamespaceEntry, error) {
	const failMsg = "ReadNamespaceStream failed"

	ns := EmptyNamespace()
	invalidEntries := []InvalidNamespaceEntry{}

	batchStart := 0
	for i, entry := range stream {
		startEntry := stream[batchStart]

		if !entry.SamePoint(startEntry) {
			batchEnd := i + 1
			invalid, err := ns.addPointBatch(stream[batchStart:batchEnd])

			invalidEntries = append(invalidEntries, invalid...)

			if err != nil {
				return EmptyNamespace(), invalidEntries, errors.Wrap(err, failMsg)
			}

			batchStart = batchEnd
		}

	}

	return ns, invalidEntries, nil
}
