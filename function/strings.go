package function

import (
	"sync"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"

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

type GlobCache struct {
	sync.Mutex
	Globs map[string]glob.Glob
}

func (cache *GlobCache) findOrCreate(pattern string) (glob.Glob, error) {
	const failMsg = "GlobCache.findOrCreate failed"

	cache.Lock()
	defer cache.Unlock()

	if cache.Globs == nil {
		cache.Globs = map[string]glob.Glob{}
	}

	cacheEntry, ok := cache.Globs[pattern]

	if ok {
		return cacheEntry, nil
	}

	cacheEntry, err := glob.Compile(pattern)

	if err != nil {
		return nil, errors.Wrap(err, failMsg)
	}

	cache.Globs[pattern] = cacheEntry
	return cacheEntry, nil
}

type StrGlob struct {
	Cache *GlobCache
}

func (wildcard *StrGlob) FuncName() string {
	return "str_glob"
}

func (wildcard *StrGlob) Match(literals []string, entries []crdt.Entry) bool {
	if wildcard.Cache == nil {
		wildcard.Cache = &GlobCache{}
	}

	first, err := firstValue(literals, entries)

	if err != nil {
		return false
	}

	m := &globMatch{
		cache: wildcard.Cache,
		text:  first,
	}

	return isMatch(m, literals, entries)
}

type match interface {
	matchLiteral(literal string) bool
	matchPoint(point crdt.Point) bool
}

type globMatch struct {
	cache *GlobCache
	glob  glob.Glob
	text  string
}

func (match *globMatch) matchLiteral(literal string) bool {
	if err := match.init(); err != nil {
		return false
	}

	return match.glob.Match(literal)
}

func (match *globMatch) matchPoint(point crdt.Point) bool {
	if err := match.init(); err != nil {
		return false
	}

	pointText := string(point.Text())
	return match.glob.Match(pointText)
}

func (match *globMatch) init() error {
	const failMsg = "globMatch.init failed"

	if match.glob == nil {
		glob, err := match.cache.findOrCreate(match.text)

		if err != nil {
			return errors.Wrap(err, failMsg)
		}

		match.glob = glob
	}

	return nil
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
