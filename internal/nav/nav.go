// Package nav holds the navigation state: the current directory, its subdirectories, the search query and the cursor.
package nav

import (
	"path/filepath"
	"slices"
)

// Current is the entry that stands for the current directory. The cursor can be on it only when the query is empty.
const Current = "."

// Lister lists the names of the subdirectories in a directory. Hidden ones are included only when showHidden is true.
type Lister interface {
	List(dir string, showHidden bool) ([]string, error)
}

// Matcher filters names against a search query and ranks them, best match first.
type Matcher interface {
	Match(query string, names []string) []Match
}

type Match struct {
	// Index is the index of the name in the names passed to Matcher.Match.
	Index int
	// Positions are the byte offsets of the characters in the name that match the query, in ascending order.
	Positions []int
}

type Entry struct {
	Name string
	// Matched are the byte offsets of the characters in Name that match the query, in ascending order.
	Matched []int
}

type Navigator struct {
	lister     Lister
	matcher    Matcher
	showHidden bool
	dir        string
	names      []string
	query      string
	entries    []Entry
	cursor     int
}

// Creates a navigator that starts in dir, which must be an absolute and clean path.
// showHidden is the start value, ToggleHidden changes it.
func New(lister Lister, matcher Matcher, dir string, showHidden bool) (*Navigator, error) {
	navigator := &Navigator{lister: lister, matcher: matcher, showHidden: showHidden}
	if err := navigator.load(dir, ""); err != nil {
		return nil, err
	}
	return navigator, nil
}

func (n *Navigator) Dir() string {
	return n.dir
}

func (n *Navigator) Query() string {
	return n.query
}

func (n *Navigator) ShowHidden() bool {
	return n.showHidden
}

// Entries are the subdirectory names, without Current.
// When the query is not empty, they are only the names that match it, best match first, and can be empty.
// The caller must not change the slice.
func (n *Navigator) Entries() []Entry {
	if n.hasCurrent() {
		return n.entries[1:]
	}
	return n.entries
}

func (n *Navigator) Count() (matched, total int) {
	return len(n.Entries()), len(n.names)
}

// Cursor is the index of the entry under the cursor in Entries. It is -1 when the cursor is on Current.
func (n *Navigator) Cursor() int {
	if n.hasCurrent() {
		return n.cursor - 1
	}
	return n.cursor
}

func (n *Navigator) AtCurrent() bool {
	return n.focusedName() == Current
}

// Selected returns the full path of the entry under the cursor, or the current directory when there are no entries.
func (n *Navigator) Selected() string {
	return filepath.Join(n.dir, n.focusedName())
}

// SetQuery filters the entries by the query. A non-empty query hides Current and moves the cursor to the best match.
// An empty query shows all entries and keeps the cursor on the entry it was on.
func (n *Navigator) SetQuery(query string) {
	previous := n.focusedName()
	n.query = query
	n.rebuildEntries()
	n.cursor = 0
	if query == "" {
		n.moveCursorTo(previous)
	}
}

func (n *Navigator) MoveUp() {
	if n.cursor > 0 {
		n.cursor--
	}
}

func (n *Navigator) MoveDown() {
	if n.cursor < len(n.entries)-1 {
		n.cursor++
	}
}

// MoveToCurrent moves the cursor to Current. It does nothing when the query is not empty, because Current is hidden.
func (n *Navigator) MoveToCurrent() {
	if n.hasCurrent() {
		n.cursor = 0
	}
}

// Open moves into the entry under the cursor and clears the query.
// It does nothing when the cursor is on Current or there are no entries.
func (n *Navigator) Open() error {
	if len(n.entries) == 0 || n.AtCurrent() {
		return nil
	}
	return n.load(filepath.Join(n.dir, n.focusedName()), "")
}

// Back moves to the parent directory, clears the query and puts the cursor on the directory it came from.
// It does nothing in the root directory.
func (n *Navigator) Back() error {
	parent := filepath.Dir(n.dir)
	if parent == n.dir {
		return nil
	}
	return n.load(parent, filepath.Base(n.dir))
}

// ToggleHidden shows or hides the hidden directories and keeps the query.
// The cursor stays on the entry it was on. When that entry is not in the list anymore, it moves to the first row.
func (n *Navigator) ToggleHidden() error {
	names, err := n.lister.List(n.dir, !n.showHidden)
	if err != nil {
		return err
	}
	previous := n.focusedName()
	n.showHidden = !n.showHidden
	n.names = names
	n.rebuildEntries()
	n.moveCursorTo(previous)
	return nil
}

func (n *Navigator) load(dir, cursorOn string) error {
	names, err := n.lister.List(dir, n.showHidden)
	if err != nil {
		return err
	}
	n.dir = dir
	n.names = names
	n.query = ""
	n.rebuildEntries()
	n.moveCursorTo(cursorOn)
	return nil
}

func (n *Navigator) rebuildEntries() {
	if n.hasCurrent() {
		n.entries = make([]Entry, 0, len(n.names)+1)
		n.entries = append(n.entries, Entry{Name: Current})
		for _, name := range n.names {
			n.entries = append(n.entries, Entry{Name: name})
		}
		return
	}
	matches := n.matcher.Match(n.query, n.names)
	n.entries = make([]Entry, len(matches))
	for i, match := range matches {
		n.entries[i] = Entry{Name: n.names[match.Index], Matched: match.Positions}
	}
}

// Reports whether entries start with Current, which is only when the query is empty.
func (n *Navigator) hasCurrent() bool {
	return n.query == ""
}

func (n *Navigator) focusedName() string {
	if len(n.entries) == 0 {
		return ""
	}
	return n.entries[n.cursor].Name
}

func (n *Navigator) moveCursorTo(name string) {
	n.cursor = max(0, slices.IndexFunc(n.entries, func(entry Entry) bool { return entry.Name == name }))
}
