package eval

import (
	"errors"

	"github.com/johnny-morrice/godless/crdt"
)

type MatchFunction interface {
	Match(literals []string, entries []crdt.Entry) bool
}

type MatchFunctionLambda func(literals []string, entries []crdt.Entry) bool

func (lambda MatchFunctionLambda) Match(literals []string, entries []crdt.Entry) bool {
	return lambda(literals, entries)
}

var _ MatchFunction = MatchFunctionLambda(StrEq)

func StrEq(literals []string, entries []crdt.Entry) bool {
	m, err := firstValue(literals, entries)

	if err != nil {
		return false
	}

	for _, pfx := range literals {
		if pfx != m {
			return false
		}
	}

	for _, entry := range entries {
		found := false
		for _, val := range entry.GetValues() {
			if val.HasText(m) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func firstValue(literals []string, entries []crdt.Entry) (string, error) {
	var first string
	var found bool

	if len(literals) > 0 {
		first = literals[0]
		found = true
	} else {
		for _, entry := range entries {
			values := entry.GetValues()
			if len(values) > 0 {
				point := values[0]
				first = string(point.Text())
				found = true
				break
			}
		}
	}

	// No values: no firstValue.
	if !found {
		return "", errors.New("no firstValue")
	}

	return first, nil
}
