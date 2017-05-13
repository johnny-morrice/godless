package godless

import "sort"

// FIXME not really a stream, whole is kept in memory.
type NamespaceStreamEntry struct {
	Table  TableName
	Row    RowName
	Entry  EntryName
	Points []Point
}

type byStreamOrder []NamespaceStreamEntry

func (stream byStreamOrder) Len() int {
	return len(stream)
}

func (stream byStreamOrder) Swap(i, j int) {
	stream[i], stream[j] = stream[j], stream[i]
}

func (stream byStreamOrder) Less(i, j int) bool {
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

	minSize := len(a.Points)
	maxSize := len(b.Points)

	if maxSize < minSize {
		minSize, maxSize = maxSize, minSize
	}

	for i := 0; i < minSize; i++ {
		ap := a.Points[i]
		bp := b.Points[j]
		logdbg("%v < %v ? %v", ap, bp, ap < bp)
		if ap < bp {
			return true
		} else if ap > bp {
			return false
		}
	}

	return len(a.Points) < len(b.Points)
}

func MakeStreamEntry(tname TableName, rname RowName, ename EntryName, entry Entry) NamespaceStreamEntry {
	return NamespaceStreamEntry{
		Table:  tname,
		Row:    rname,
		Entry:  ename,
		Points: entry.GetValues(),
	}
}

func MakeNamespaceStream(ns Namespace) []NamespaceStreamEntry {
	stream := []NamespaceStreamEntry{}
	for tableName, table := range ns.Tables {
		for rowName, row := range table.Rows {
			for entryName, entry := range row.Entries {
				streamEntry := MakeStreamEntry(tableName, rowName, entryName, entry)
				stream = append(stream, streamEntry)
			}
		}
	}

	sort.Sort(byStreamOrder(stream))
	return stream
}

func ReadNamespaceStream(stream []NamespaceStreamEntry) Namespace {
	ns := EmptyNamespace()

	for _, streamEntry := range stream {
		ns = ns.JoinStreamEntry(streamEntry)
	}

	return ns
}
