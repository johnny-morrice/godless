package godless

import (
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
)

// Semi-lattice type that implements our storage
type Namespace struct {
	Tables map[string]Table
}

func MakeNamespace() *Namespace {
	return &Namespace{
		Tables: map[string]Table{},
	}
}

func (ns *Namespace) IsEmpty() bool {
	return len(ns.Tables) == 0
}

func (ns *Namespace) Copy() *Namespace {
	out := MakeNamespace()

	for k, table := range ns.Tables {
		out.Tables[k] = table
	}

	return out
}

func (ns *Namespace) JoinNamespace(other *Namespace) (*Namespace, error) {
	out := ns.Copy()

	for otherk, otherTable := range other.Tables {
		out.addTable(otherk, otherTable)
	}

	return out, nil
}

// Destructive
func (ns *Namespace) addTable(key string, table Table) error {
	current, present := ns.Tables[key]

	if present {
		joined, err := current.JoinTable(table)

		if err != nil {
			return errors.Wrap(err, "Namespace join table failed")
		}

		ns.Tables[key] = joined
	} else {
		ns.Tables[key] = table
	}

	return nil
}

func (ns *Namespace) JoinTable(key string, table Table) (*Namespace, error) {
	out := ns.Copy()

	out.addTable(key, table)

	return out, nil
}

func (ns *Namespace) GetTable(key string) (Table, error) {
	if table, present := ns.Tables[key]; present {
		return table, nil
	} else {
		return Table{}, fmt.Errorf("No such table in namespace '%v'", key)
	}
}

// TODO improved type validation
type Table struct {
	Rows map[string]Row
}

func (t Table) Foreachrow(f func (rowKey string, r Row)) {
	for k, r := range t.Rows {
		f(k, r)
	}
}

func (t Table) Copy() Table {
	out := Table{Rows: map[string]Row{}}

	for k, row := range t.Rows {
		out.Rows[k] = row
	}

	return out
}

func (t Table) JoinTable(other Table) (Table, error) {
	out := t.Copy()

	for otherk, otherRow := range other.Rows {
		out.addRow(otherk, otherRow)
	}

	return out, nil
}

func (t Table) JoinRow(rowKey string, row Row) (Table, error) {
	out := t.Copy()

	out.addRow(rowKey, row)

	return out, nil
}

// Destructive.
func (t Table) addRow(rowKey string, row Row) {
	if current, present := t.Rows[rowKey]; present {
		jrow := current.JoinRow(row)
		t.Rows[rowKey] = jrow
	} else {
		t.Rows[rowKey] = row
	}
}

type Row struct {
	Entries map[string][]string
}

func (row Row) Copy() Row {
	out := Row{Entries: map[string][]string{}}

	for k, entry := range row.Entries {
		out.Entries[k] = entry
	}

	return out
}

func (row Row) JoinRow(other Row) Row {
	out := row.Copy()

	for otherk, otherEntry := range other.Entries {
		if entry, present := out.Entries[otherk]; present {
			out.Entries[otherk] = uniq256(append(entry, otherEntry...))
		} else {
			out.Entries[otherk] = otherEntry
		}
	}

	return out
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
