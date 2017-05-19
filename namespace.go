package godless

//go:generate mockgen -destination mock/mock_namespace.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless RowConsumer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

type TableName string
type RowName string
type EntryName string
type Point string

// Semi-lattice type that implements our storage
type Namespace struct {
	Tables map[TableName]Table
}

func EmptyNamespace() Namespace {
	return MakeNamespace(map[TableName]Table{})
}

func MakeNamespace(tables map[TableName]Table) Namespace {
	out := Namespace{
		Tables: map[TableName]Table{},
	}

	for k, v := range tables {
		out.Tables[k] = v
	}

	return out
}

func EncodeNamespace(ns Namespace, w io.Writer) error {
	const failMsg = "EncodeNamespace failed"

	pb := MakeNamespaceMessage(ns)

	bs, err := proto.Marshal(pb)
	if err != nil {
		return errors.Wrap(err, failMsg)
	}

	var written int
	written, err = w.Write(bs)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("%v after %v bytes", failMsg, written))
	}

	return nil
}

func DecodeNamespace(r io.Reader) (Namespace, error) {
	const failMsg = "DecodeNamespace failed"
	bs, err := ioutil.ReadAll(r)

	if err != nil {
		return EmptyNamespace(), errors.Wrap(err, failMsg)
	}

	pb := &NamespaceMessage{}
	err = proto.Unmarshal(bs, pb)

	if err != nil {
		return EmptyNamespace(), errors.Wrap(err, failMsg)
	}

	return ReadNamespaceMessage(pb), nil
}

func ReadNamespaceMessage(message *NamespaceMessage) Namespace {
	stream := ReadNamespaceStreamMessage(message)
	return ReadNamespaceStream(stream)
}

func MakeNamespaceMessage(ns Namespace) *NamespaceMessage {
	stream := MakeNamespaceStream(ns)
	pb := &NamespaceMessage{Entries: make([]*NamespaceEntryMessage, len(stream))}

	for i, entry := range stream {
		pb.Entries[i] = MakeNamespaceEntryMessage(entry)
	}

	return pb
}

func (ns Namespace) GetTableNames() []TableName {
	tableNames := make([]TableName, len(ns.Tables))

	i := 0
	for name := range ns.Tables {
		tableNames[i] = name
		i++
	}

	return tableNames
}

func (ns Namespace) JoinStreamEntry(streamEntry NamespaceStreamEntry) Namespace {
	entryNamespace := MakeNamespace(map[TableName]Table{
		streamEntry.Table: MakeTable(map[RowName]Row{
			streamEntry.Row: MakeRow(map[EntryName]Entry{
				streamEntry.Entry: MakeEntry(streamEntry.Points),
			}),
		}),
	})

	return ns.JoinNamespace(entryNamespace)
}

func (ns Namespace) IsEmpty() bool {
	return len(ns.Tables) == 0
}

func (ns Namespace) Copy() Namespace {
	cpy := EmptyNamespace()

	for k, table := range ns.Tables {
		cpy.Tables[k] = table
	}

	return cpy
}

func (ns Namespace) JoinNamespace(other Namespace) Namespace {
	joined := ns.Copy()

	for otherk, otherTable := range other.Tables {
		joined = joined.addTable(otherk, otherTable)
	}

	return joined
}

// Destructive
func (ns Namespace) addTable(key TableName, table Table) Namespace {
	current, present := ns.Tables[key]

	if present {
		joined := current.JoinTable(table)

		ns.Tables[key] = joined
	} else {
		ns.Tables[key] = table
	}

	return ns
}

func (ns Namespace) JoinTable(key TableName, table Table) Namespace {
	joined := ns.Copy()
	joined = joined.addTable(key, table)
	return joined
}

func (ns Namespace) GetTable(key TableName) (Table, error) {
	if table, present := ns.Tables[key]; present {
		return table, nil
	} else {
		return Table{}, fmt.Errorf("No such Table in Namespace '%v'", key)
	}
}

func (ns Namespace) Equals(other Namespace) bool {
	if len(ns.Tables) != len(other.Tables) {
		return false
	}

	for k, v := range ns.Tables {
		otherv, present := other.Tables[k]
		if !present || !v.Equals(otherv) {
			return false
		}
	}

	return true
}

// TODO improved type validation
type Table struct {
	Rows map[RowName]Row
}

func EmptyTable() Table {
	return MakeTable(map[RowName]Row{})
}

func MakeTable(rows map[RowName]Row) Table {
	out := Table{
		Rows: map[RowName]Row{},
	}

	for k, v := range rows {
		out.Rows[k] = v
	}

	return out
}

type RowConsumer interface {
	Accept(rowKey RowName, r Row)
}

type RowConsumerFunc func(rowKey RowName, r Row)

func (rcf RowConsumerFunc) Accept(rowKey RowName, r Row) {
	rcf(rowKey, r)
}

// TODO easy optimisation: hold slice in Table for fast iteration.
func (t Table) Foreachrow(consumer RowConsumer) {
	for k, r := range t.Rows {
		consumer.Accept(k, r)
	}
}

func (t Table) Copy() Table {
	cpy := EmptyTable()

	for k, row := range t.Rows {
		cpy.Rows[k] = row
	}

	return cpy
}

func (t Table) AllRows() []Row {
	rows := []Row{}

	for _, v := range t.Rows {
		rows = append(rows, v)
	}

	return rows
}

func (t Table) JoinTable(other Table) Table {
	joined := t.Copy()

	for otherk, otherRow := range other.Rows {
		joined.addRow(otherk, otherRow)
	}

	return joined
}

func (t Table) JoinRow(rowKey RowName, row Row) Table {
	joined := t.Copy()
	joined.addRow(rowKey, row)
	return joined
}

func (t Table) GetRow(rowKey RowName) (Row, error) {
	if row, present := t.Rows[rowKey]; present {
		return row, nil
	} else {
		return Row{}, fmt.Errorf("No such Row in Table '%v'", rowKey)
	}
}

// Destructive.
func (t Table) addRow(rowKey RowName, row Row) {
	if current, present := t.Rows[rowKey]; present {
		jrow := current.JoinRow(row)
		t.Rows[rowKey] = jrow
	} else {
		t.Rows[rowKey] = row
	}
}

func (t Table) Equals(other Table) bool {
	if len(t.Rows) != len(other.Rows) {
		return false
	}

	for k, v := range t.Rows {
		otherv, present := other.Rows[k]
		if !present || !v.Equals(otherv) {
			return false
		}
	}

	return true
}

type Row struct {
	Entries map[EntryName]Entry
}

func EmptyRow() Row {
	return MakeRow(map[EntryName]Entry{})
}

func MakeRow(entries map[EntryName]Entry) Row {
	out := Row{
		Entries: map[EntryName]Entry{},
	}

	for k, v := range entries {
		out.Entries[k] = v
	}

	return out
}

func (row Row) Copy() Row {
	cpy := Row{Entries: map[EntryName]Entry{}}

	for k, entry := range row.Entries {
		cpy.Entries[k] = entry
	}

	return cpy
}

func (row Row) JoinRow(other Row) Row {
	joined := row.Copy()

	for otherk, otherEntry := range other.Entries {
		if joinEntry, present := joined.Entries[otherk]; present {
			joined.Entries[otherk] = joinEntry.JoinEntry(otherEntry)
		} else {
			joined.Entries[otherk] = otherEntry
		}
	}

	return joined
}

func (row Row) GetEntry(entryKey EntryName) (Entry, error) {
	if entry, present := row.Entries[entryKey]; present {
		return entry, nil
	} else {
		return Entry{}, fmt.Errorf("No such Entry in Row '%v'", entryKey)
	}
}

func (row Row) JoinEntry(entryKey EntryName, entry Entry) Row {
	entryRow := Row{
		Entries: map[EntryName]Entry{
			entryKey: entry,
		},
	}

	return row.JoinRow(entryRow)
}

func (row Row) Equals(other Row) bool {
	if len(row.Entries) != len(other.Entries) {
		return false
	}

	for k, v := range row.Entries {
		otherv, present := other.Entries[k]
		if !present || !v.Equals(otherv) {
			return false
		}
	}

	return true
}

type Entry struct {
	Set []Point
}

func EmptyEntry() Entry {
	return MakeEntry([]Point{})
}

func MakeEntry(set []Point) Entry {
	undupes := uniq256(set)
	sort.Sort(byPointValue(undupes))
	return Entry{Set: undupes}
}

func (e Entry) JoinEntry(other Entry) Entry {
	return MakeEntry(append(e.Set, other.Set...))
}

func (e Entry) Equals(other Entry) bool {
	// Easy because Entry.set is deduplicated and sorted
	if len(e.Set) != len(other.Set) {
		return false
	}

	for i, v := range e.Set {
		if other.Set[i] != v {
			return false
		}
	}

	return true
}

func (e Entry) GetValues() []Point {
	cpy := make([]Point, len(e.Set))

	for i, p := range e.Set {
		cpy[i] = p
	}

	return cpy
}

type byPointValue []Point

func (p byPointValue) Len() int {
	return len(p)
}

func (p byPointValue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p byPointValue) Less(i, j int) bool {
	return p[i] < p[j]
}

// uniq256 deduplicates a slice of Values using sha256.
func uniq256(dupes []Point) []Point {
	dedup := map[[sha256.Size]byte]Point{}

	for _, point := range dupes {
		bs := []byte(string(point))
		k := sha256.Sum256(bs)
		if _, present := dedup[k]; !present {
			dedup[k] = point
		}
	}

	out := make([]Point, len(dedup))

	i := 0
	for _, point := range dedup {
		out[i] = point
		i++
	}

	return out
}
