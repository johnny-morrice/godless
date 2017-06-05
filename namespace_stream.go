package godless

import "sort"

// FIXME not really a stream, whole is kept in memory.
type NamespaceStreamEntry struct {
	Table  TableName
	Row    RowName
	Entry  EntryName
	Points []Point
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

func ReadNamespaceEntryMessage(message *NamespaceEntryMessage) NamespaceStreamEntry {
	entry := NamespaceStreamEntry{
		Table:  TableName(message.Table),
		Row:    RowName(message.Row),
		Entry:  EntryName(message.Entry),
		Points: make([]Point, len(message.Points)),
	}

	for i, p := range message.Points {
		entry.Points[i] = Point(p)
	}

	return entry
}

func MakeNamespaceEntryMessage(entry NamespaceStreamEntry) *NamespaceEntryMessage {
	pb := &NamespaceEntryMessage{
		Table:  string(entry.Table),
		Row:    string(entry.Row),
		Entry:  string(entry.Entry),
		Points: make([]string, len(entry.Points)),
	}

	for i, p := range entry.Points {
		pb.Points[i] = string(p)
	}

	return pb
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

func (entries byEntryOrder) Less(i, j int) bool {
	a := entries[i]
	b := entries[j]
	return pointLess(a.Set, b.Set)
}

func pointLess(a, b []Point) bool {
	minSize := imin(len(a), len(b))

	for i := 0; i < minSize; i++ {
		ap := a[i]
		bp := b[i]
		if ap < bp {
			return true
		} else if ap > bp {
			return false
		}
	}

	return len(a) < len(b)
}

func MakeStreamEntry(tname TableName, rname RowName, ename EntryName, entry Entry) NamespaceStreamEntry {
	return NamespaceStreamEntry{
		Table:  tname,
		Row:    rname,
		Entry:  ename,
		Points: entry.GetValues(),
	}
}

func MakeTableStream(tableKey TableName, table Table) []NamespaceStreamEntry {
	subNamespace := MakeNamespace(map[TableName]Table{
		tableKey: table,
	})
	return MakeNamespaceStream(subNamespace)
}

func MakeRowStream(tableKey TableName, rowKey RowName, row Row) []NamespaceStreamEntry {
	count := len(row.Entries)
	entryKeys := make([]string, count)
	i := 0
	for ek, _ := range row.Entries {
		entryKeys[i] = string(ek)
		i++
	}
	sort.Strings(entryKeys)

	entries := make([]Entry, count)
	for i, ek := range entryKeys {
		entry := row.Entries[EntryName(ek)]
		entries[i] = entry
	}

	sort.Sort(byEntryOrder(entries))

	stream := make([]NamespaceStreamEntry, count)
	for i, e := range entries {
		entryKey := EntryName(entryKeys[i])
		stream[i] = NamespaceStreamEntry{
			Points: e.Set,
			Table:  tableKey,
			Row:    rowKey,
			Entry:  entryKey,
		}
	}

	return stream
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

func MakeNamespaceStreamMessage(stream []NamespaceStreamEntry) *NamespaceMessage {
	message := &NamespaceMessage{Entries: make([]*NamespaceEntryMessage, len(stream))}

	for i, entry := range stream {
		message.Entries[i] = MakeNamespaceEntryMessage(entry)
	}

	assertLenEquals(stream, message.Entries)

	return message
}

func ReadNamespaceStreamMessage(message *NamespaceMessage) []NamespaceStreamEntry {
	stream := make([]NamespaceStreamEntry, len(message.Entries))

	for i, emsg := range message.Entries {
		stream[i] = ReadNamespaceEntryMessage(emsg)
	}

	assertLenEquals(message.Entries, stream)

	return stream
}

func ReadNamespaceStream(stream []NamespaceStreamEntry) Namespace {
	ns := EmptyNamespace()

	for _, streamEntry := range stream {
		ns.addStreamEntry(streamEntry)
	}

	return ns
}
