// Package ui renders the navigator in the terminal and maps keys to navigation actions.
package ui

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/vstelmakh/walker/internal/nav"
)

const (
	leftPadding = " "
	// nonListHeight is the number of rows around the list: the path row, the search line, two separators and the footer.
	nonListHeight    = 5
	minFooterGap     = 2
	minSelectHintGap = 2
)

type Config struct {
	Keys  KeyMap
	Theme Theme
	// Home is shown as "~" in the path row. An empty value disables it.
	Home string
}

type Model struct {
	navigator *nav.Navigator
	config    Config
	width     int
	height    int
	offset    int
	err       error
	result    string
	selected  bool
}

func New(navigator *nav.Navigator, config Config) *Model {
	return &Model{navigator: navigator, config: config}
}

// Result returns the selected path. Reports ok as false when the user cancelled.
func (m *Model) Result() (path string, ok bool) {
	return m.result, m.selected
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		cmd = m.handleKey(msg)
	case tea.PasteMsg:
		m.handlePaste(msg)
	}
	m.offset = adjustOffset(m.offset, max(0, m.navigator.Cursor()), m.listHeight(), len(m.navigator.Entries()))
	return m, cmd
}

func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	keys := m.config.Keys
	m.err = nil
	key := msg.String()
	query := m.navigator.Query()
	switch {
	case query != "" && slices.Contains(keys.Erase, key):
		m.navigator.SetQuery(dropLastRune(query))
	case slices.Contains(keys.Up, key):
		m.navigator.MoveUp()
	case slices.Contains(keys.Down, key):
		m.navigator.MoveDown()
	case slices.Contains(keys.Open, key):
		m.err = m.navigator.Open()
	case slices.Contains(keys.Back, key):
		if slices.Contains(keys.Erase, key) && !m.navigator.AtCurrent() {
			m.navigator.MoveToCurrent()
		} else {
			m.err = m.navigator.Back()
		}
	case slices.Contains(keys.Confirm, key):
		switch m.confirmAction() {
		case confirmOpen:
			m.err = m.navigator.Open()
		case confirmSelect:
			m.result = m.navigator.Dir()
			m.selected = true
			return tea.Quit
		}
	case slices.Contains(keys.Cancel, key):
		switch m.cancelAction() {
		case cancelClear:
			m.navigator.SetQuery("")
		case cancelTop:
			m.navigator.MoveToCurrent()
		case cancelQuit:
			return tea.Quit
		}
	case slices.Contains(keys.Quit, key):
		return tea.Quit
	case msg.Text != "":
		m.navigator.SetQuery(query + msg.Text)
	}
	return nil
}

type confirmAction int

const (
	confirmOpen confirmAction = iota
	confirmSelect
)

type cancelAction int

const (
	cancelClear cancelAction = iota
	cancelTop
	cancelQuit
)

// Opens the selected entry. It selects the current directory when the cursor is on it.
func (m *Model) confirmAction() confirmAction {
	if m.navigator.AtCurrent() {
		return confirmSelect
	}
	return confirmOpen
}

// Clears the query first, then moves the cursor to the current directory, then quits.
func (m *Model) cancelAction() cancelAction {
	switch {
	case m.navigator.Query() != "":
		return cancelClear
	case !m.navigator.AtCurrent():
		return cancelTop
	default:
		return cancelQuit
	}
}

// Adds the first line of the pasted text to the query, like typed text.
func (m *Model) handlePaste(msg tea.PasteMsg) {
	m.err = nil
	if text := sanitizePaste(msg.Content); text != "" {
		m.navigator.SetQuery(m.navigator.Query() + text)
	}
}

func dropLastRune(s string) string {
	_, size := utf8.DecodeLastRuneInString(s)
	return s[:len(s)-size]
}

// View is empty until the terminal size is known.
func (m *Model) View() tea.View {
	view := tea.NewView("")
	view.AltScreen = true
	if m.width == 0 || m.height == 0 {
		return view
	}

	lines := []string{m.pathRow(), m.searchLine(), m.separator()}
	lines = append(lines, m.list()...)
	lines = append(lines, m.separator(), m.footer())
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, m.width, "")
	}

	view.SetContent(strings.Join(lines, "\n"))
	view.Cursor = tea.NewCursor(m.cursorX(), 1)
	view.Cursor.Shape = tea.CursorBar
	return view
}

func (m *Model) separator() string {
	theme := m.config.Theme
	return leftPadding + theme.Separator.Render(strings.Repeat(theme.SeparatorSymbol, m.contentWidth()))
}

// pathRow shows the current directory. The row stands for nav.Current, so it is selected when the cursor is on it.
// Otherwise, it ends with the name of the selected entry as a preview of the target path.
func (m *Model) pathRow() string {
	theme := m.config.Theme
	selected := m.navigator.AtCurrent()
	path := abbreviateHome(m.navigator.Dir(), m.config.Home)
	preview := ""
	if entries := m.navigator.Entries(); len(entries) > 0 && !selected {
		preview = entries[m.navigator.Cursor()].Name
		if !strings.HasSuffix(path, string(filepath.Separator)) {
			path += string(filepath.Separator)
		}
	}

	line := path + preview
	line = truncateLeft(line, m.rowWidth(), theme.EllipsisSymbol)
	// When the truncation cuts into the preview, all of the line is the preview.
	split := 0
	if strings.HasSuffix(line, preview) {
		split = len(line) - len(preview)
	}
	if !selected {
		return m.listRow(theme.Path.Render(line[:split])+theme.PathPreview.Render(line[split:]), false)
	}
	return m.listRow(theme.Selected.Render(line)+m.selectHint(ansi.StringWidth(line)), true)
}

// selectHint is right-aligned and includes the space before it. It is empty when it does not fit next to the path.
func (m *Model) selectHint(pathWidth int) string {
	theme := m.config.Theme
	hint := theme.SelectHintText
	space := m.rowWidth() - pathWidth - ansi.StringWidth(hint)
	if hint == "" || space < minSelectHintGap {
		return ""
	}
	return theme.Selected.Render(strings.Repeat(" ", space)) + theme.SelectHint.Render(hint)
}

func (m *Model) searchLine() string {
	theme := m.config.Theme
	query := m.navigator.Query()
	prompt := theme.Prompt
	if query != "" {
		prompt = theme.PromptWithQuery
	}
	line := leftPadding + prompt.Render(theme.PromptSymbol)
	width := m.contentWidth() - ansi.StringWidth(theme.PromptSymbol)
	if query == "" {
		return line + theme.Placeholder.Render(ansi.Truncate(theme.PlaceholderText, max(0, width), theme.EllipsisSymbol))
	}
	// One cell stays free for the terminal cursor after the query.
	return line + theme.Query.Render(truncateLeft(query, max(0, width-1), theme.EllipsisSymbol))
}

func (m *Model) cursorX() int {
	x := ansi.StringWidth(leftPadding + m.config.Theme.PromptSymbol + m.navigator.Query())
	return min(x, m.width-1)
}

func (m *Model) list() []string {
	theme := m.config.Theme
	entries, cursor := m.navigator.Entries(), m.navigator.Cursor()
	height := m.listHeight()

	var lines []string
	if len(entries) == 0 {
		text := theme.NoSubdirectoriesText
		if m.navigator.Query() != "" {
			text = theme.NoMatchesText
		}
		text = ansi.Truncate(text, m.rowWidth(), theme.EllipsisSymbol)
		lines = append(lines, m.listRow(theme.EmptyList.Render(text), false))
	}
	end := min(len(entries), m.offset+height)
	for i := m.offset; i < end; i++ {
		lines = append(lines, m.entry(entries[i], i == cursor))
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return lines
}

func (m *Model) entry(entry nav.Entry, selected bool) string {
	theme := m.config.Theme
	base, match := theme.Entry, theme.EntryMatch
	if selected {
		base, match = theme.Selected, theme.SelectedMatch
	}

	text := highlight(entry.Name, entry.Matched, base, match)
	if width := m.rowWidth(); ansi.StringWidth(text) > width {
		text = ansi.Truncate(text, width, base.Render(theme.EllipsisSymbol))
	}
	return m.listRow(text, selected)
}

// listRow starts a selected row with the marker and fills it to the full width.
// Other rows start with spaces of the marker width, so the text lines up.
func (m *Model) listRow(text string, selected bool) string {
	theme := m.config.Theme
	if !selected {
		return leftPadding + strings.Repeat(" ", ansi.StringWidth(theme.MarkerSymbol)) + text
	}
	line := leftPadding + theme.Marker.Render(theme.MarkerSymbol) + text
	if fill := m.width - ansi.StringWidth(line); fill > 0 {
		line += theme.Selected.Render(strings.Repeat(" ", fill))
	}
	return line
}

func (m *Model) footer() string {
	theme := m.config.Theme
	matched, total := m.navigator.Count()
	left := theme.Counter.Render(fmt.Sprintf("%d/%d", matched, total))

	right := theme.Hint.Render(m.hints())
	if m.err != nil {
		right = theme.Error.Render(theme.ErrorPrefix + m.err.Error())
	}

	width := m.contentWidth()
	space := width - ansi.StringWidth(left) - minFooterGap
	if ansi.StringWidth(right) > space {
		right = ansi.Truncate(right, max(0, space), theme.EllipsisSymbol)
	}
	gap := width - ansi.StringWidth(left) - ansi.StringWidth(right)
	return leftPadding + left + strings.Repeat(" ", max(0, gap)) + right
}

func (m *Model) hints() string {
	keys, theme := m.config.Keys, m.config.Theme
	var hints []string
	for _, hint := range []struct {
		keys   []string
		action string
	}{
		{keys: keys.Confirm, action: theme.confirmHint(m.confirmAction())},
		{keys: keys.Back, action: theme.HintBackText},
		{keys: keys.Cancel, action: theme.cancelHint(m.cancelAction())},
	} {
		if label := firstKeyLabel(hint.keys, theme.KeyLabels); label != "" {
			hints = append(hints, label+" "+hint.action)
		}
	}
	return strings.Join(hints, "  ")
}

// Returns the width for the text of a list row.
func (m *Model) rowWidth() int {
	return max(0, m.contentWidth()-ansi.StringWidth(m.config.Theme.MarkerSymbol))
}

// Returns the width without the padding.
func (m *Model) contentWidth() int {
	return max(0, m.width-ansi.StringWidth(leftPadding))
}

// Returns the number of rows for the list. It is at least one.
func (m *Model) listHeight() int {
	return max(1, m.height-nonListHeight)
}

// Returns the first visible index that keeps the cursor inside the window. Returns zero for a height of zero.
func adjustOffset(offset, cursor, height, total int) int {
	if height <= 0 {
		return 0
	}
	offset = min(offset, max(0, total-height))
	if cursor < offset {
		return cursor
	}
	if cursor >= offset+height {
		return cursor - height + 1
	}
	return offset
}
