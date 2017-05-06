package godless

//go:generate mockgen -destination mock/mock_namespace.go -imports lib=github.com/johnny-morrice/godless -self_package lib github.com/johnny-morrice/godless RowConsumer

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

// Semi-lattice type that implements our storage
type Namespace struct {
	tables map[string]Table
}

func EmptyNamespace() *Namespace {
	return MakeNamespace(map[string]Table{})
}

func MakeNamespace(tables map[string]Table) *Namespace {
	out := &Namespace{
		tables: map[string]Table{},
	}

	for k, v := range tables {
		out.tables[k] = v
	}

	return out
}

func (ns *Namespace) IsEmpty() bool {
	return len(ns.tables) == 0
}

func (ns *Namespace) Copy() *Namespace {
	out := EmptyNamespace()

	for k, table := range ns.tables {
		out.tables[k] = table
	}

	return out
}

func (ns *Namespace) JoinNamespace(other *Namespace) *Namespace {
	out := ns.Copy()

	for otherk, otherTable := range other.tables {
		out.addTable(otherk, otherTable)
	}

	return out
}

// Destructive
func (ns *Namespace) addTable(key string, table Table) {
	current, present := ns.tables[key]

	if present {
		joined := current.JoinTable(table)

		ns.tables[key] = joined
	} else {
		ns.tables[key] = table
	}
}

func (ns *Namespace) JoinTable(key string, table Table) *Namespace {
	out := ns.Copy()

	out.addTable(key, table)

	return out
}

func (ns *Namespace) GetTable(key string) (Table, error) {
	if table, present := ns.tables[key]; present {
		return table, nil
	} else {
		return Table{}, fmt.Errorf("No such Table in Namespace '%v'", key)
	}
}

func (ns *Namespace) Equals(other *Namespace) bool {
	if len(ns.tables) != len(other.tables) {
		return false
	}

	for k, v := range ns.tables {
		otherv, present := other.tables[k]
		if !present || !v.Equals(otherv) {
			return false
		}
	}

	return true
}

// TODO improved type validation
type Table struct {
	rows map[string]Row
}

func EmptyTable() Table {
	return MakeTable(map[string]Row{})
}

func MakeTable(rows map[string]Row) Table {
	out := Table{
		rows: map[string]Row{},
	}

	for k, v := range rows {
		out.rows[k] = v
	}

	return out
}

type RowConsumer interface {
	Accept(rowKey string, r Row)
}

type RowConsumerFunc func(rowKey string, r Row)

func (rcf RowConsumerFunc) Accept(rowKey string, r Row) {
	rcf(rowKey, r)
}

// TODO easy optimisation: hold slice in Table for fast iteration.
func (t Table) Foreachrow(consumer RowConsumer) {
	for k, r := range t.rows {
		consumer.Accept(k, r)
	}
}

func (t Table) Copy() Table {
	out := Table{rows: map[string]Row{}}

	for k, row := range t.rows {
		out.rows[k] = row
	}

	return out
}

func (t Table) AllRows() []Row {
	out := []Row{}

	for _, v := range t.rows {
		out = append(out, v)
	}

	return out
}

func (t Table) JoinTable(other Table) Table {
	out := t.Copy()

	for otherk, otherRow := range other.rows {
		out.addRow(otherk, otherRow)
	}

	return out
}

func (t Table) JoinRow(rowKey string, row Row) Table {
	out := t.Copy()

	out.addRow(rowKey, row)

	return out
}

func (t Table) GetRow(rowKey string) (Row, error) {
	if row, present := t.rows[rowKey]; present {
		return row, nil
	} else {
		return Row{}, fmt.Errorf("No such Row in Table '%v'", rowKey)
	}
}

// Destructive.
func (t Table) addRow(rowKey string, row Row) {
	if current, present := t.rows[rowKey]; present {
		jrow := current.JoinRow(row)
		t.rows[rowKey] = jrow
	} else {
		t.rows[rowKey] = row
	}
}

func (t Table) Equals(other Table) bool {
	if len(t.rows) != len(other.rows) {
		return false
	}

	for k, v := range t.rows {
		otherv, present := other.rows[k]
		if !present || !v.Equals(otherv) {
			return false
		}
	}

	return true
}

type Row struct {
	entries map[string]Entry
}

func EmptyRow() Row {
	return MakeRow(map[string]Entry{})
}

func MakeRow(entries map[string]Entry) Row {
	out := Row{
		entries: map[string]Entry{},
	}

	for k, v := range entries {
		out.entries[k] = v
	}

	return out
}

func (row Row) Copy() Row {
	out := Row{entries: map[string]Entry{}}

	for k, entry := range row.entries {
		out.entries[k] = entry
	}

	return out
}

func (row Row) JoinRow(other Row) Row {
	out := row.Copy()

	for otherk, otherEntry := range other.entries {
		if oute, present := out.entries[otherk]; present {
			out.entries[otherk] = oute.JoinEntry(otherEntry)
		} else {
			out.entries[otherk] = otherEntry
		}
	}

	return out
}

func (row Row) GetEntry(entryKey string) (Entry, error) {
	if entry, present := row.entries[entryKey]; present {
		return entry, nil
	} else {
		return Entry{}, fmt.Errorf("No such Entry in Row '%v'", entryKey)
	}
}

func (row Row) JoinEntry(entryKey string, entry Entry) Row {
	entryRow := Row{
		entries: map[string]Entry{
			entryKey: entry,
		},
	}

	return row.JoinRow(entryRow)
}

func (row Row) Equals(other Row) bool {
	if len(row.entries) != len(other.entries) {
		return false
	}

	for k, v := range row.entries {
		otherv, present := other.entries[k]
		if !present || !v.Equals(otherv) {
			return false
		}
	}

	return true
}

type Entry struct {
	set []string
}

func EmptyEntry() Entry {
	return MakeEntry([]string{})
}

func MakeEntry(set []string) Entry {
	undupes := uniq256(set)
	sort.Strings(undupes)
	return Entry{set: undupes}
}

func (e Entry) JoinEntry(other Entry) Entry {
	return MakeEntry(append(e.set, other.set...))
}

func (e Entry) Equals(other Entry) bool {
	// Easy because Entry.set is deduplicated and sorted
	if len(e.set) != len(other.set) {
		return false
	}

	for i, v := range e.set {
		if other.set[i] != v {
			return false
		}
	}

	return true
}

func (e Entry) GetValues() []string {
	return e.set
}

// uniq256 deduplicates a slice of strings using sha256.
func uniq256(dups []string) []string {
	dedup := map[[sha256.Size]byte]string{}

	for _, s := range dups {
		bs := []byte(s)
		k := sha256.Sum256(bs)
		if _, present := dedup[k]; !present {
			dedup[k] = s
		}
	}

	out := []string{}

	for _, v := range dedup {
		out = append(out, v)
	}

	return out
}
