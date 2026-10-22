package search

import (
	"slices"
	"testing"

	"github.com/vstelmakh/walker/internal/nav"
)

func TestMatcherMatch(t *testing.T) {
	tests := []struct {
		name  string
		query string
		names []string
		want  []int
	}{
		{name: "empty query matches all", query: "", names: []string{"b", "a"}, want: []int{0, 1}},
		{name: "no match", query: "zz", names: []string{"alpha", "beta"}, want: []int{}},
		{name: "subsequence", query: "aa", names: []string{"alpha", "beta", "banana"}, want: []int{0, 2}},
		{name: "adjacent before scattered", query: "doc", names: []string{"dxxoxxc", "doc"}, want: []int{1, 0}},
		{name: "ties keep order", query: "a", names: []string{"ya", "xa"}, want: []int{0, 1}},
		{name: "lowercase ignores case", query: "doc", names: []string{"docs", "Docs"}, want: []int{0, 1}},
		{name: "uppercase matches case", query: "Doc", names: []string{"docs", "Docs"}, want: []int{1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := indexes(Matcher{}.Match(test.query, test.names))
			if !slices.Equal(got, test.want) {
				t.Errorf("Match(%q, %q) indexes = %v, want %v", test.query, test.names, got, test.want)
			}
		})
	}
}

func TestMatcherMatchPositions(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		matchName string
		want      []int
	}{
		{name: "empty query", query: "", matchName: "rXRc", want: nil},
		{name: "ignore case", query: "rc", matchName: "rXRc", want: []int{0, 3}},
		{name: "exact case", query: "Rc", matchName: "rXRc", want: []int{2, 3}},
		{name: "byte offsets", query: "c", matchName: "érc", want: []int{3}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matches := Matcher{}.Match(test.query, []string{test.matchName})
			if len(matches) != 1 {
				t.Fatalf("Match(%q, %q) = %v, want one match", test.query, test.matchName, matches)
			}
			if got := matches[0].Positions; !slices.Equal(got, test.want) {
				t.Errorf("Match(%q, %q) positions = %v, want %v", test.query, test.matchName, got, test.want)
			}
		})
	}
}

func TestExactSubsequencePositions(t *testing.T) {
	tests := []struct {
		query, name string
		want        []int
		wantOK      bool
	}{
		{query: "", name: "abc", want: []int{}, wantOK: true},
		{query: "ac", name: "abc", want: []int{0, 2}, wantOK: true},
		{query: "c", name: "ébc", want: []int{3}, wantOK: true},
		{query: "ca", name: "abc", wantOK: false},
		{query: "A", name: "abc", wantOK: false},
		{query: "abcd", name: "abc", wantOK: false},
	}
	for _, test := range tests {
		got, ok := exactSubsequencePositions(test.query, test.name)
		if ok != test.wantOK || (ok && !slices.Equal(got, test.want)) {
			t.Errorf("exactSubsequencePositions(%q, %q) = %v, %t, want %v, %t", test.query, test.name, got, ok, test.want, test.wantOK)
		}
	}
}

func indexes(matches []nav.Match) []int {
	result := make([]int, len(matches))
	for i, match := range matches {
		result[i] = match.Index
	}
	return result
}
