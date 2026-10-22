package nav

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

var errDenied = errors.New("permission denied")

// fakeLister returns errDenied for a dir that is not in the map.
type fakeLister map[string][]string

func (f fakeLister) List(dir string) ([]string, error) {
	entries, ok := f[dir]
	if !ok {
		return nil, errDenied
	}
	return entries, nil
}

type prefixMatcher struct{}

func (prefixMatcher) Match(query string, names []string) []Match {
	var matches []Match
	for i, name := range names {
		if strings.HasPrefix(name, query) {
			matches = append(matches, Match{Index: i, Positions: []int{0}})
		}
	}
	return matches
}

var tree = fakeLister{
	"/":         {"home", "tmp"},
	"/home":     {"alice", "bob", "carol"},
	"/home/bob": {},
	"/tmp":      {"locked"},
}

func TestNew(t *testing.T) {
	navigator := mustNew(t, "/home")

	assertState(t, navigator, "/home", []string{"alice", "bob", "carol"}, -1)
}

func TestNewError(t *testing.T) {
	_, err := New(tree, prefixMatcher{}, "/missing")
	if !errors.Is(err, errDenied) {
		t.Fatalf("New() error = %v, want %v", err, errDenied)
	}
}

func TestUpDown(t *testing.T) {
	navigator := mustNew(t, "/home")

	navigator.MoveUp()
	assertCursor(t, navigator, -1)

	navigator.MoveDown()
	navigator.MoveDown()
	navigator.MoveDown()
	assertCursor(t, navigator, 2)

	navigator.MoveDown()
	assertCursor(t, navigator, 2)

	navigator.MoveUp()
	assertCursor(t, navigator, 1)
}

func TestUpDownEmpty(t *testing.T) {
	navigator := mustNew(t, "/home/bob")

	navigator.MoveDown()
	navigator.MoveUp()
	assertCursor(t, navigator, -1)
}

func TestMoveToCurrent(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.MoveDown()
	navigator.MoveDown()

	navigator.MoveToCurrent()

	assertCursor(t, navigator, -1)
}

func TestOpen(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.MoveDown()
	navigator.MoveDown()

	mustSucceed(t, navigator.Open())

	assertState(t, navigator, "/home/bob", []string{}, -1)
}

func TestOpenCurrent(t *testing.T) {
	navigator := mustNew(t, "/home")

	mustSucceed(t, navigator.Open())

	assertState(t, navigator, "/home", []string{"alice", "bob", "carol"}, -1)
}

func TestOpenError(t *testing.T) {
	navigator := mustNew(t, "/tmp")
	navigator.MoveDown()

	if err := navigator.Open(); !errors.Is(err, errDenied) {
		t.Fatalf("Open() error = %v, want %v", err, errDenied)
	}

	assertState(t, navigator, "/tmp", []string{"locked"}, 0)
}

func TestBack(t *testing.T) {
	navigator := mustNew(t, "/home/bob")

	mustSucceed(t, navigator.Back())
	assertState(t, navigator, "/home", []string{"alice", "bob", "carol"}, 1)

	mustSucceed(t, navigator.Back())
	assertState(t, navigator, "/", []string{"home", "tmp"}, 0)
}

func TestBackRoot(t *testing.T) {
	navigator := mustNew(t, "/")
	navigator.MoveDown()

	mustSucceed(t, navigator.Back())

	assertState(t, navigator, "/", []string{"home", "tmp"}, 0)
}

func TestBackError(t *testing.T) {
	navigator, err := New(fakeLister{"/home": {"alice", "bob", "carol"}}, prefixMatcher{}, "/home")
	mustSucceed(t, err)

	if err := navigator.Back(); !errors.Is(err, errDenied) {
		t.Fatalf("Back() error = %v, want %v", err, errDenied)
	}

	assertState(t, navigator, "/home", []string{"alice", "bob", "carol"}, -1)
}

func TestAtCurrent(t *testing.T) {
	navigator := mustNew(t, "/home")
	if !navigator.AtCurrent() {
		t.Error("AtCurrent() = false, want true")
	}

	navigator.MoveDown()
	if navigator.AtCurrent() {
		t.Error("AtCurrent() = true, want false")
	}
}

func TestSelected(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.MoveDown()
	navigator.MoveDown()

	if got := navigator.Selected(); got != "/home/bob" {
		t.Errorf("Selected() = %q, want %q", got, "/home/bob")
	}
}

func TestSelectedCurrent(t *testing.T) {
	navigator := mustNew(t, "/home")

	if got := navigator.Selected(); got != "/home" {
		t.Errorf("Selected() = %q, want %q", got, "/home")
	}
}

func TestSetQuery(t *testing.T) {
	navigator := mustNew(t, "/home")

	navigator.SetQuery("b")

	if got := navigator.Entries()[0].Matched; !slices.Equal(got, []int{0}) {
		t.Errorf("Entries()[0].Matched = %v, want [0]", got)
	}

	if got := navigator.Query(); got != "b" {
		t.Errorf("Query() = %q, want %q", got, "b")
	}
	assertState(t, navigator, "/home", []string{"bob"}, 0)
	if navigator.AtCurrent() {
		t.Error("AtCurrent() = true, want false")
	}
	if got := navigator.Selected(); got != "/home/bob" {
		t.Errorf("Selected() = %q, want %q", got, "/home/bob")
	}
}

func TestSetQueryMovesCursorToBestMatch(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.MoveDown()
	navigator.MoveDown()
	navigator.MoveDown()

	navigator.SetQuery("c")

	assertState(t, navigator, "/home", []string{"carol"}, 0)
}

func TestSetQueryNoMatches(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.SetQuery("z")

	navigator.MoveDown()
	navigator.MoveUp()
	mustSucceed(t, navigator.Open())

	assertState(t, navigator, "/home", []string{}, 0)
	if navigator.AtCurrent() {
		t.Error("AtCurrent() = true, want false")
	}
	if got := navigator.Selected(); got != "/home" {
		t.Errorf("Selected() = %q, want %q", got, "/home")
	}
}

func TestSetQueryEmptyKeepsCursorEntry(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.SetQuery("b")

	navigator.SetQuery("")

	assertState(t, navigator, "/home", []string{"alice", "bob", "carol"}, 1)
}

func TestMoveToCurrentWithQuery(t *testing.T) {
	navigator := mustNew(t, "/")
	navigator.SetQuery("t")

	navigator.MoveToCurrent()

	assertState(t, navigator, "/", []string{"tmp"}, 0)
}

func TestOpenClearsQuery(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.SetQuery("b")

	mustSucceed(t, navigator.Open())

	assertState(t, navigator, "/home/bob", []string{}, -1)
	if got := navigator.Query(); got != "" {
		t.Errorf("Query() = %q, want empty", got)
	}
}

func TestBackClearsQuery(t *testing.T) {
	navigator := mustNew(t, "/home")
	navigator.SetQuery("b")

	mustSucceed(t, navigator.Back())

	assertState(t, navigator, "/", []string{"home", "tmp"}, 0)
	if got := navigator.Query(); got != "" {
		t.Errorf("Query() = %q, want empty", got)
	}
}

func TestCount(t *testing.T) {
	navigator := mustNew(t, "/home")
	assertCount(t, navigator, 3, 3)

	navigator.SetQuery("b")
	assertCount(t, navigator, 1, 3)

	navigator.SetQuery("z")
	assertCount(t, navigator, 0, 3)
}

func mustNew(t *testing.T, dir string) *Navigator {
	t.Helper()
	navigator, err := New(tree, prefixMatcher{}, dir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return navigator
}

func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertState(t *testing.T, navigator *Navigator, dir string, entries []string, cursor int) {
	t.Helper()
	if got := navigator.Dir(); got != dir {
		t.Errorf("Dir() = %q, want %q", got, dir)
	}
	var names []string
	for _, entry := range navigator.Entries() {
		names = append(names, entry.Name)
	}
	if !slices.Equal(names, entries) {
		t.Errorf("Entries() names = %q, want %q", names, entries)
	}
	assertCursor(t, navigator, cursor)
}

func assertCount(t *testing.T, navigator *Navigator, matched, total int) {
	t.Helper()
	if gotMatched, gotTotal := navigator.Count(); gotMatched != matched || gotTotal != total {
		t.Errorf("Count() = %d, %d, want %d, %d", gotMatched, gotTotal, matched, total)
	}
}

func assertCursor(t *testing.T, navigator *Navigator, cursor int) {
	t.Helper()
	if got := navigator.Cursor(); got != cursor {
		t.Errorf("Cursor() = %d, want %d", got, cursor)
	}
}
