package ui

// KeyMap binds actions to key names. A key name is the value of tea.KeyPressMsg.String, for example "up" or "ctrl+c".
// A key with text that is not bound to an action is added to the search query.
type KeyMap struct {
	Up   []string
	Down []string
	Open []string
	Back []string
	// Confirm opens the directory under the cursor, or selects the current directory when the cursor is on nav.Current.
	Confirm []string
	// Erase deletes the last character of the query. With an empty query, a key that is also in Back moves the cursor
	// to nav.Current first, and then goes back. A key that is not in Back does nothing.
	Erase []string
	// Cancel clears the query, then moves the cursor to nav.Current, then exits without a result.
	Cancel []string
	Quit   []string
}

// DefaultKeyMap leaves letters free for the search.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:      []string{"up"},
		Down:    []string{"down"},
		Open:    []string{"right"},
		Back:    []string{"left", "backspace"},
		Confirm: []string{"enter"},
		Erase:   []string{"backspace"},
		Cancel:  []string{"esc"},
		Quit:    []string{"ctrl+c"},
	}
}
