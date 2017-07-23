package function

import (
	"github.com/johnny-morrice/godless/crdt"
)

type StrEq struct{}

func (streq StrEq) FuncName() string {
	return "str_eq"
}

func (StrEq) Match(literals []string, entries []crdt.Entry) bool {
	first, err := firstValue(literals, entries)

	if err != nil {
		return false
	}

	m := eqMatch{text: first}

	return isMatch(m, literals, entries)
}

type StrWildcard struct{}

func (StrWildcard) FuncName() string {
	return "str_wildcard"
}

func (StrWildcard) Match(literals []string, entries []crdt.Entry) bool {
	panic("not implemented")
}

type match interface {
	matchLiteral(literal string) bool
	matchPoint(point crdt.Point) bool
}

type eqMatch struct {
	text string
}

func (eq eqMatch) matchLiteral(literal string) bool {
	return eq.text == literal
}

func (eq eqMatch) matchPoint(point crdt.Point) bool {
	return point.HasText(eq.text)
}

func isMatch(m match, literals []string, entries []crdt.Entry) bool {
	for _, lit := range literals {
		if !m.matchLiteral(lit) {
			return false
		}
	}

	for _, entry := range entries {
		if !matchEntry(m, entry) {
			return false
		}
	}

	return true
}

func matchEntry(m match, entry crdt.Entry) bool {
	for _, point := range entry.GetValues() {
		if m.matchPoint(point) {
			return true
		}
	}

	return false
}
