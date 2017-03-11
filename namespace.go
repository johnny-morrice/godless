package godless

import (
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
)

type SemiLattice interface {
	Join(other SemiLattice) (SemiLattice, error)
}

// Semi-lattice type that implements our storage
type Namespace struct {
	Tables map[string]Table
}

func MakeNamespace() *Namespace {
	return &Namespace{
		Tables: map[string]Table{},
	}
}

func (ns *Namespace) JoinNamespace(other *Namespace) (*Namespace, error) {
	build := map[string]Table{}

	for k, v := range ns.Tables {
		build[k] = v
	}

	for k, otherv := range other.Tables {
		if v, present := ns.Tables[k]; present {
			joined, err := v.JoinTable(otherv)

			if err != nil {
				return nil, errors.Wrap(err, "Error in Namespace join")
			}

			build[k] = joined
		} else {
			build[k] = otherv
		}
	}

	return &Namespace{Tables: build}, nil
}

func (ns *Namespace) Join(other SemiLattice) (SemiLattice, error) {
	if ons, ok := other.(*Namespace); ok {
		return ns.JoinNamespace(ons)
	}

	return nil, errors.New("Expected *Namespace in Join")
}

func (ns *Namespace) JoinTable(key string, table Table) error {
	joined, err := table.JoinTable(ns.Tables[key])

	if err != nil {
		return errors.Wrap(err, "Namespace JoinTable failed")
	}

	ns.Tables[key] = joined

	return nil
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

func (t Table) JoinTable(other Table) (Table, error) {
	out := Table{
		Rows: map[string]Row{},
	}

	for k, v := range t.Rows {
		if othv, present := other.Rows[k]; present {
			jrow := v.JoinRow(othv)
			out.Rows[k] = jrow
		} else {
			out.Rows[k] = v
		}
	}

	return out, nil
}

func (t Table) Join(other SemiLattice) (SemiLattice, error) {
	otherobj, ok := other.(Table)

	if !ok {
		return nil, errors.New("Expected Table in Join")
	}

	return t.JoinTable(otherobj)
}

type Row struct {
	Entries map[string][]string
}

func (row Row) JoinRow(other Row) Row {
	out := Row{
		Entries: map[string][]string{},
	}

	for k, v := range row.Entries {
		if othv, present := other.Entries[k]; present {
			out.Entries[k] = append(v, othv...)
		} else {
			out.Entries[k] = v
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
