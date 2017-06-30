package crdt

import (
	"fmt"
	"sort"

	"github.com/johnny-morrice/godless/crypto"
	"github.com/johnny-morrice/godless/log"
	"github.com/pkg/errors"
)

type TableName string
type RowName string
type EntryName string

type byTableName []TableName

func (by byTableName) Len() int {
	return len(by)
}

func (by byTableName) Swap(i, j int) {
	by[i], by[j] = by[j], by[i]
}

func (by byTableName) Less(i, j int) bool {
	return by[i] < by[j]
}

func JoinStreamEntries(stream []NamespaceStreamEntry) ([]NamespaceStreamEntry, []InvalidNamespaceEntry, error) {
	const failMsg = "JoinStreamEntries failed"
	ns, readInvalid, err := ReadNamespaceStream(stream)

	if err != nil {
		return nil, readInvalid, errors.Wrap(err, failMsg)
	}

	stream, makeInvalid := MakeNamespaceStream(ns)

	invalidEntries := append(readInvalid, makeInvalid...)

	return stream, invalidEntries, nil
}

func FilterSignedEntries(stream []NamespaceStreamEntry, keys []crypto.PublicKey) ([]NamespaceStreamEntry, []InvalidNamespaceEntry, error) {
	const failMsg = "FilterSignedEntries failed"

	unsigned, readInvalid, readErr := ReadNamespaceStream(stream)

	if readErr != nil {
		return nil, readInvalid, errors.Wrap(readErr, failMsg)
	}

	signed := unsigned.FilterVerified(keys)
	signedStream, makeInvalid := MakeNamespaceStream(signed)

	invalidEntries := append(readInvalid, makeInvalid...)

	return signedStream, invalidEntries, nil
}

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

func (ns Namespace) FilterVerified(keys []crypto.PublicKey) Namespace {
	verified := EmptyNamespace()

	ns.ForeachEntry(func(t TableName, r RowName, e EntryName, entry Entry) {
		signed := entry.FilterVerified(keys)
		verified.addEntry(t, r, e, signed)
	})

	return verified
}

func (ns Namespace) addEntry(t TableName, r RowName, e EntryName, entry Entry) {
	table := EmptyTable()
	row := EmptyRow()

	row.addEntry(e, entry)
	table.addRow(r, row)
	ns.addTable(t, table)
}

// Strip removes empty tables and rows that would not be saved to the backing store.
func (ns Namespace) Strip() (Namespace, []InvalidNamespaceEntry, error) {
	const failMsg = "Namespace.Strip failed"

	stream, invalid := MakeNamespaceStream(ns)

	invalidCount := len(invalid)
	if invalidCount > 0 {
		log.Warn("Stripped %v invalid points", invalidCount)
	}

	namespace, invalid, err := ReadNamespaceStream(stream)

	if err != nil {
		return EmptyNamespace(), invalid, errors.Wrap(err, failMsg)
	}

	return namespace, invalid, nil
}

func (ns Namespace) GetTableNames() []TableName {
	tableNames := make([]TableName, 0, len(ns.Tables))

	for name := range ns.Tables {
		tableNames = append(tableNames, name)
	}

	sort.Sort(byTableName(tableNames))

	return tableNames
}

// The batch should correspond to a single point.
func (ns Namespace) addPointBatch(stream []NamespaceStreamEntry) ([]InvalidNamespaceEntry, error) {
	const failMsg = "Namespace.addPointBatch failed"

	if len(stream) == 0 {
		log.Warn("Namespace.addPointBatch of length 0")
		return nil, nil
	}

	point, invalid, err := readStreamPoint(stream)

	if err != nil {
		return invalid, errors.Wrap(err, failMsg)
	}

	first := stream[0]
	tableName := first.Table
	rowName := first.Row
	entryName := first.Entry

	ns.addPoint(tableName, rowName, entryName, point)

	return invalid, nil
}

func (ns Namespace) addStreamEntry(entry NamespaceStreamEntry) error {
	const failMsg = "Namespace.addStreamEntry failed"

	var sigs []crypto.Signature

	if !crypto.IsNilSignature(entry.Point.Signature) {
		sig, err := crypto.ParseSignature(entry.Point.Signature)

		if err != nil {
			return errors.Wrap(err, failMsg)
		}

		sigs = []crypto.Signature{sig}
	}

	point := PresignedPoint(entry.Point.Text, sigs)

	ns.addPoint(entry.Table, entry.Row, entry.Entry, point)
	return nil
}

func (ns Namespace) addPoint(tableName TableName, rowName RowName, entryName EntryName, point Point) {
	table := MakeTable(map[RowName]Row{
		rowName: MakeRow(map[EntryName]Entry{
			entryName: MakeEntry([]Point{point}),
		}),
	})

	ns.addTable(tableName, table)
}

func (ns Namespace) IsEmpty() bool {
	return len(ns.Tables) == 0
}

func (ns Namespace) Copy() Namespace {
	cpy := EmptyNamespace()

	for k, table := range ns.Tables {
		cpy.Tables[k] = table.Copy()
	}

	return cpy
}

func (ns Namespace) JoinNamespace(other Namespace) Namespace {
	joined := ns.Copy()

	for otherk, otherTable := range other.Tables {
		joined.addTable(otherk, otherTable)
	}

	return joined
}

func (ns Namespace) addTable(key TableName, table Table) {
	current, present := ns.Tables[key]

	if present {
		current.addTable(table)
	} else {
		ns.Tables[key] = table
	}
}

func (ns Namespace) JoinTable(key TableName, table Table) Namespace {
	joined := ns.Copy()
	joined.addTable(key, table)
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

		if !present {
			return false
		}

		if !v.Equals(otherv) {
			return false
		}
	}

	return true
}

// Iterate through all entries
func (ns Namespace) ForeachEntry(f func(t TableName, r RowName, e EntryName, entry Entry)) {
	for tableName, table := range ns.Tables {
		for rowName, row := range table.Rows {
			for entryName, entry := range row.Entries {
				f(tableName, rowName, entryName, entry)
			}
		}
	}
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

func (t Table) ForeachRow(f func(rowName RowName, r Row)) {
	for k, r := range t.Rows {
		f(k, r)
	}
}

func (t Table) ForeachEntry(f func(rowName RowName, entryName EntryName, entry Entry)) {
	for rowName, row := range t.Rows {
		for entryName, entry := range row.Entries {
			f(rowName, entryName, entry)
		}
	}
}

func (t Table) Copy() Table {
	cpy := EmptyTable()

	for k, row := range t.Rows {
		cpy.Rows[k] = row.Copy()
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
	joined.addTable(other)
	return joined
}

func (t Table) addTable(other Table) {
	for otherk, otherRow := range other.Rows {
		t.addRow(otherk, otherRow)
	}
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
		current.addRow(row)
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
		if !present {
			return false
		}

		if !v.Equals(otherv) {
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

func (row Row) ForeachEntry(f func(entryName EntryName, entry Entry)) {
	for name, entry := range row.Entries {
		f(name, entry)
	}
}

func (row Row) Copy() Row {
	cpy := Row{Entries: map[EntryName]Entry{}}

	for k, entry := range row.Entries {
		cpy.Entries[k] = entry.Copy()
	}

	return cpy
}

func (row Row) addRow(other Row) {
	for otherk, otherEntry := range other.Entries {
		if joinEntry, present := row.Entries[otherk]; present {
			row.Entries[otherk] = joinEntry.JoinEntry(otherEntry)
		} else {
			row.Entries[otherk] = otherEntry
		}
	}
}

func (row Row) JoinRow(other Row) Row {
	joined := row.Copy()
	joined.addRow(other)
	return joined
}

func (row Row) GetEntry(entryKey EntryName) (Entry, error) {
	if entry, present := row.Entries[entryKey]; present {
		return entry, nil
	} else {
		return Entry{}, fmt.Errorf("No such Entry in Row '%v'", entryKey)
	}
}

func (row Row) addEntry(entryKey EntryName, other Entry) {
	if entry, present := row.Entries[entryKey]; present {
		row.Entries[entryKey] = entry.JoinEntry(other)
	} else {
		row.Entries[entryKey] = other
	}
}

func (row Row) JoinEntry(entryKey EntryName, entry Entry) Row {
	cpy := row.Copy()
	cpy.addEntry(entryKey, entry)
	return cpy
}

func (row Row) Equals(other Row) bool {
	if len(row.Entries) != len(other.Entries) {
		return false
	}

	for k, v := range row.Entries {
		otherv, present := other.Entries[k]

		if !present {
			return false
		}

		if !v.Equals(otherv) {
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
	sort.Sort(byPointValue(set))
	undupes := uniqPointSorted(set)
	return Entry{Set: undupes}
}

func (e Entry) FilterVerified(keys []crypto.PublicKey) Entry {
	verified := make([]Point, 0, len(e.Set))

	for _, p := range e.Set {
		if p.IsVerifiedByAny(keys) {
			verified = append(verified, p)
		}
	}

	verifiedEntry := Entry{Set: verified}

	return verifiedEntry
}

func (e Entry) Copy() Entry {
	return MakeEntry(e.Set)
}

func (e Entry) JoinEntry(other Entry) Entry {
	return MakeEntry(append(e.Set, other.Set...))
}

func (e Entry) Equals(other Entry) bool {
	// Easy because Entry.set is deduplicated and sorted
	if len(e.Set) != len(other.Set) {
		return false
	}

	for i, myPoint := range e.Set {
		theirPoint := other.Set[i]

		if !myPoint.Equals(theirPoint) {
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
