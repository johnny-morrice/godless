package eval

import (
	"errors"

	"github.com/johnny-morrice/godless/crdt"
	"github.com/johnny-morrice/godless/internal/crypto"
)

type MatchFunction func(first string, prefix []string, entries []crdt.Entry) bool

var _ MatchFunction = MatchFunction(StrEq)
var _ MatchFunction = MatchFunction(StrNeq)

type SignatureCheck struct {
	PublicKeys []crypto.PublicKey
}

func (check SignatureCheck) IsSigned(first string, prefix []string, entries []crdt.Entry) bool {
	for _, entry := range entries {
		for _, point := range entry.GetValues() {
			for _, key := range check.PublicKeys {
				if point.Verify(key) {
					return true
				}
			}
		}
	}

	return false
}

func StrEq(first string, prefix []string, entries []crdt.Entry) bool {
	prefix = append(prefix, first)
	m, err := match(prefix, entries)

	if err != nil {
		return false
	}

	for _, pfx := range prefix {
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

func StrNeq(first string, prefix []string, entries []crdt.Entry) bool {
	prefix = append(prefix, first)
	m, err := match(prefix, entries)

	if err != nil {
		return false
	}

	pfxmatch := 0
	for _, pfx := range prefix {
		if pfx == m {
			pfxmatch++
		}
	}

	entrymatch := 0
	for _, entry := range entries {
		for _, val := range entry.GetValues() {
			if val.HasText(m) {
				entrymatch++
			}
		}
	}

	return !((pfxmatch > 0 && entrymatch > 0) || pfxmatch > 1 || entrymatch > 1)
}

// TODO need user concepts + crypto to narrow row match down.
func match(prefix []string, entries []crdt.Entry) (string, error) {
	var first string
	var found bool

	if len(prefix) > 0 {
		first = prefix[0]
		found = true
	} else {
		for _, entry := range entries {
			values := entry.GetValues()
			if len(values) > 0 {
				point := values[0]
				first = string(point.Text)
				found = true
				break
			}
		}
	}

	// No values: no match.
	if !found {
		return "", errors.New("no match")
	}

	return first, nil
}
