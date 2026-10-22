// Package search ranks names against a fuzzy query.
package search

import (
	"slices"
	"unicode"

	"github.com/sahilm/fuzzy"

	"github.com/vstelmakh/walker/internal/nav"
)

// Matcher matches names with smart case: a lowercase query ignores case, a query with an uppercase letter does not.
type Matcher struct{}

// Match returns the names that match the query, best match first.
// Names with the same score keep their order. An empty query matches all names, with no positions.
func (Matcher) Match(query string, names []string) []nav.Match {
	if query == "" {
		matches := make([]nav.Match, len(names))
		for i := range matches {
			matches[i] = nav.Match{Index: i}
		}
		return matches
	}

	caseSensitive := slices.ContainsFunc([]rune(query), unicode.IsUpper)
	found := fuzzy.Find(query, names)
	matches := make([]nav.Match, 0, len(found))
	for _, match := range found {
		positions := match.MatchedIndexes
		if caseSensitive {
			var ok bool
			if positions, ok = exactSubsequencePositions(query, match.Str); !ok {
				continue
			}
		}
		matches = append(matches, nav.Match{Index: match.Index, Positions: positions})
	}
	return matches
}

func exactSubsequencePositions(query, name string) ([]int, bool) {
	runes := []rune(query)
	positions := make([]int, 0, len(runes))
	for i, r := range name {
		if len(positions) == len(runes) {
			break
		}
		if r == runes[len(positions)] {
			positions = append(positions, i)
		}
	}
	return positions, len(positions) == len(runes)
}
