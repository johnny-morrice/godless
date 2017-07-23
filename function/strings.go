package function

import (
	"github.com/johnny-morrice/godless/crdt"
)

type StrEq struct{}

func (streq StrEq) Match(literals []string, entries []crdt.Entry) bool {
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

func (streq StrEq) FuncName() string {
	return "str_eq"
}
