package ui

import (
	"errors"
	"slices"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/vstelmakh/walker/internal/nav"
	"github.com/vstelmakh/walker/internal/search"
)

const (
	testWidth  = 80
	testHeight = 10
)

var testSeparator = " " + strings.Repeat("─", testWidth-1)

var errDenied = errors.New("permission denied")

type fakeLister map[string][]string

func (f fakeLister) List(dir string, showHidden bool) ([]string, error) {
	entries, ok := f[dir]
	if !ok {
		return nil, errDenied
	}
	if showHidden {
		return entries, nil
	}
	return slices.DeleteFunc(slices.Clone(entries), func(name string) bool { return strings.HasPrefix(name, ".") }), nil
}

var tree = fakeLister{
	"/":         {"home"},
	"/home":     {"alice", "bob", "carol", "dave", "erin"},
	"/home/bob": {},
	"/work":     {".git", "app"},
}

var (
	keyUp    = tea.KeyPressMsg{Code: tea.KeyUp}
	keyDown  = tea.KeyPressMsg{Code: tea.KeyDown}
	keyLeft  = tea.KeyPressMsg{Code: tea.KeyLeft}
	keyRight = tea.KeyPressMsg{Code: tea.KeyRight}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyBack  = tea.KeyPressMsg{Code: tea.KeyBackspace}
	keyEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
	keyCtrlC = tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	keyCtrlH = tea.KeyPressMsg{Code: 'h', Mod: tea.ModCtrl}
	keySpace = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
)

func TestDefaultKeyNames(t *testing.T) {
	tests := []struct {
		key  tea.KeyPressMsg
		want string
	}{
		{key: keyUp, want: "up"},
		{key: keyDown, want: "down"},
		{key: keyLeft, want: "left"},
		{key: keyRight, want: "right"},
		{key: keyEnter, want: "enter"},
		{key: keyBack, want: "backspace"},
		{key: keyEsc, want: "esc"},
		{key: keyCtrlC, want: "ctrl+c"},
		{key: keySpace, want: "space"},
	}
	for _, test := range tests {
		if got := test.key.String(); got != test.want {
			t.Errorf("String() = %q, want %q", got, test.want)
		}
	}
}

func TestNavigation(t *testing.T) {
	model := newModel(t, "/home")

	send(model, keyDown, keyDown, keyRight)
	assertDir(t, model, "/home/bob")

	send(model, keyLeft)
	assertDir(t, model, "/home")
	assertView(t, model, "▌bob")

	send(model, keyUp)
	assertView(t, model, "▌alice")

	send(model, keyDown, keyRight, keyBack)
	assertDir(t, model, "/home")
	assertView(t, model, "▌bob")
}

func TestConfirmOpensDir(t *testing.T) {
	model := newModel(t, "/home")

	cmd := send(model, keyDown, keyDown, keyEnter)

	if cmd != nil {
		t.Error("command != nil, want nil")
	}
	assertDir(t, model, "/home/bob")
}

func TestConfirmSelectsCurrent(t *testing.T) {
	model := newModel(t, "/home")

	cmd := send(model, keyDown, keyDown, keyEnter, keyEnter)

	assertQuit(t, cmd)
	path, ok := model.Result()
	if !ok || path != "/home/bob" {
		t.Errorf("Result() = %q, %t, want %q, true", path, ok, "/home/bob")
	}
}

func TestCancelMovesToCurrent(t *testing.T) {
	model := newModel(t, "/home")

	cmd := send(model, keyDown, keyDown, keyEsc)

	if cmd != nil {
		t.Error("command != nil, want nil")
	}
	assertView(t, model, "▌/home")
}

func TestCancelExits(t *testing.T) {
	model := newModel(t, "/home")

	cmd := send(model, keyDown, keyEsc, keyEsc)

	assertQuit(t, cmd)
	assertCancelled(t, model)
}

func TestQuit(t *testing.T) {
	model := newModel(t, "/home")

	cmd := send(model, keyDown, keyCtrlC)

	assertQuit(t, cmd)
	assertCancelled(t, model)
}

func TestCustomKeyMap(t *testing.T) {
	keys := DefaultKeyMap()
	keys.Down = []string{"j"}
	navigator, err := nav.New(tree, search.Matcher{}, "/home", false)
	if err != nil {
		t.Fatal(err)
	}
	model := New(navigator, Config{Keys: keys, Theme: DefaultTheme()})
	model.Update(tea.WindowSizeMsg{Width: testWidth, Height: testHeight})

	send(model, keyDown)
	assertView(t, model, "▌/home")

	send(model, tea.KeyPressMsg{Code: 'j', Text: "j"})
	assertView(t, model, "▌alice")
	assertQuery(t, model, "")
}

func TestToggleHidden(t *testing.T) {
	model := newModel(t, "/work")
	assertView(t, model, testSeparator+"\n  app\n\n")

	send(model, keyCtrlH)
	assertView(t, model, testSeparator+"\n  .git\n  app\n")

	send(model, keyCtrlH)
	assertView(t, model, testSeparator+"\n  app\n\n")
}

func TestSearchFilters(t *testing.T) {
	model := newModel(t, "/home")

	send(model, typed("ca")...)

	assertQuery(t, model, "ca")
	assertView(t, model, "  /home/carol\n : ca\n"+testSeparator+"\n ▌carol")
	if entries := model.navigator.Entries(); len(entries) != 1 {
		t.Errorf("Entries() = %v, want only carol", entries)
	}
}

func TestPasteFilters(t *testing.T) {
	model := newModel(t, "/home")

	send(model, tea.PasteMsg{Content: "car"})

	assertQuery(t, model, "car")
	assertView(t, model, "  /home/carol\n : car\n"+testSeparator+"\n ▌carol")
}

func TestPasteAppendsToQuery(t *testing.T) {
	model := newModel(t, "/home")
	send(model, typed("c")...)

	send(model, tea.PasteMsg{Content: "ar"})

	assertQuery(t, model, "car")
}

func TestPasteDropsNewlines(t *testing.T) {
	model := newModel(t, "/home")

	send(model, tea.PasteMsg{Content: "car\nbob"})

	assertQuery(t, model, "car")
}

func TestPasteClearsError(t *testing.T) {
	model := newModel(t, "/home")
	send(model, keyDown, keyRight)

	send(model, tea.PasteMsg{Content: "c"})

	if content := plainView(model); strings.Contains(content, "error:") {
		t.Errorf("View() still shows the error:\n%s", content)
	}
}

func TestSearchSpace(t *testing.T) {
	model := newModel(t, "/home")

	send(model, append(typed("a"), keySpace)...)

	assertQuery(t, model, "a ")
}

func TestSearchNoMatches(t *testing.T) {
	model := newModel(t, "/home")

	cmd := send(model, append(typed("zz"), keyEnter)...)

	if cmd != nil {
		t.Error("command != nil, want nil")
	}
	assertDir(t, model, "/home")
	assertView(t, model, "  /home\n : zz\n"+testSeparator+"\n  no matches")
}

func TestSearchConfirmOpensBestMatch(t *testing.T) {
	model := newModel(t, "/home")

	send(model, append(typed("bo"), keyEnter)...)

	assertDir(t, model, "/home/bob")
	assertQuery(t, model, "")
}

func TestSearchErase(t *testing.T) {
	model := newModel(t, "/home")
	send(model, typed("bo")...)

	send(model, keyBack)
	assertQuery(t, model, "b")

	send(model, keyBack)
	assertQuery(t, model, "")
	assertDir(t, model, "/home")
	assertView(t, model, "▌bob")

	send(model, keyBack)
	assertDir(t, model, "/home")
	assertView(t, model, "▌/home")

	send(model, keyBack)
	assertDir(t, model, "/")
}

func TestEraseKeyNotInBack(t *testing.T) {
	keys := DefaultKeyMap()
	keys.Back = []string{"left"}
	model := newModel(t, "/home")
	model.config.Keys = keys

	send(model, typed("b")...)
	send(model, keyBack, keyBack, keyBack)

	assertQuery(t, model, "")
	assertDir(t, model, "/home")
	assertView(t, model, "▌bob")

	send(model, keyLeft)
	assertDir(t, model, "/")
}

func TestBackKeyFromEntry(t *testing.T) {
	model := newModel(t, "/home")

	send(model, keyDown, keyLeft)

	assertDir(t, model, "/")
}

func TestCancelClearsQuery(t *testing.T) {
	model := newModel(t, "/home")
	send(model, typed("bo")...)

	cmd := send(model, keyEsc)
	if cmd != nil {
		t.Error("command != nil, want nil")
	}
	assertQuery(t, model, "")
	assertView(t, model, "▌bob")

	send(model, keyEsc)
	assertView(t, model, "▌/home")

	assertQuit(t, send(model, keyEsc))
	assertCancelled(t, model)
}

func TestErrorIsShownAndCleared(t *testing.T) {
	model := newModel(t, "/home")

	send(model, keyDown, keyRight)
	assertView(t, model, "error: permission denied")
	assertDir(t, model, "/home")

	send(model, keyUp)
	if content := plainView(model); strings.Contains(content, "error:") {
		t.Errorf("View() still shows the error:\n%s", content)
	}
}

func TestViewEmpty(t *testing.T) {
	model := newModel(t, "/home/bob")

	assertView(t, model, " ▌/home/bob")
	assertView(t, model, " : type to search\n"+testSeparator+"\n  no subdirectories")
	assertView(t, model, testSeparator+"\n 0/0")
}

func TestViewScrolls(t *testing.T) {
	model := newModel(t, "/home")
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 7})
	separator := " " + strings.Repeat("─", 79)
	pathRow := " ▌/home" + strings.Repeat(" ", 57) + "select directory"

	assertView(t, model, pathRow+"\n : type to search\n"+separator+"\n  alice\n  bob\n"+separator+"\n 5/5")

	send(model, keyDown, keyDown, keyDown)
	assertView(t, model, "  /home/carol\n : type to search\n"+separator+"\n  bob\n ▌carol\n"+separator)
}

func TestViewFitsWidth(t *testing.T) {
	model := newModel(t, "/home")
	model.config.Home = ""
	model.Update(tea.WindowSizeMsg{Width: 5, Height: 10})

	lines := strings.Split(model.View().Content, "\n")
	for i, line := range lines {
		if width := ansi.StringWidth(line); width > 5 {
			t.Errorf("line %d width = %d, want at most 5: %q", i, width, ansi.Strip(line))
		}
	}
	if got := ansi.Strip(lines[2]); got != " ────" {
		t.Errorf("separator = %q, want %q", got, " ────")
	}
	if got := ansi.Strip(lines[0]); got != " ▌…me" {
		t.Errorf("path row = %q, want %q", got, " ▌…me")
	}
	if got := ansi.StringWidth(lines[0]); got != 5 {
		t.Errorf("path row width = %d, want 5", got)
	}
}

func TestViewEmptyUntilSizeKnown(t *testing.T) {
	navigator, err := nav.New(tree, search.Matcher{}, "/home", false)
	if err != nil {
		t.Fatal(err)
	}
	model := New(navigator, Config{Keys: DefaultKeyMap(), Theme: DefaultTheme()})

	if view := model.View(); view.Content != "" || view.Cursor != nil {
		t.Errorf("View() = %q, cursor %+v, want empty", view.Content, view.Cursor)
	}
}

func TestViewFitsNarrowWidth(t *testing.T) {
	states := map[string][]tea.Msg{
		"current selected": nil,
		"entry selected":   {keyDown},
		"query":            typed("a"),
		"no matches":       typed("zz"),
		"error":            {keyDown, keyDown, keyRight},
	}
	for name, msgs := range states {
		for width := 1; width <= 10; width++ {
			model := newModel(t, "/home")
			model.Update(tea.WindowSizeMsg{Width: width, Height: testHeight})
			send(model, msgs...)

			for i, line := range strings.Split(model.View().Content, "\n") {
				if got := ansi.StringWidth(line); got > width {
					t.Errorf("%s, width %d, line %d width = %d: %q", name, width, i, got, ansi.Strip(line))
				}
			}
		}
	}
}

func TestViewPathRow(t *testing.T) {
	theme := DefaultTheme()
	marker := leftPadding + theme.Marker.Render(theme.MarkerSymbol)
	tests := []struct {
		name  string
		dir   string
		width int
		down  int
		want  string
	}{
		{name: "selected", dir: "/home", width: 60, down: 0, want: marker + theme.Selected.Render("/home")},
		{name: "entry", dir: "/home", width: 60, down: 1, want: "  " + theme.Path.Render("/home/") + theme.PathPreview.Render("alice")},
		{name: "root", dir: "/", width: 60, down: 1, want: "  " + theme.Path.Render("/") + theme.PathPreview.Render("home")},
		{name: "cut path", dir: "/home", width: 9, down: 1, want: "  " + theme.Path.Render("…/") + theme.PathPreview.Render("alice")},
		{name: "cut preview", dir: "/home", width: 5, down: 1, want: "  " + theme.Path.Render("") + theme.PathPreview.Render("…ce")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := newModel(t, test.dir)
			model.Update(tea.WindowSizeMsg{Width: test.width, Height: 10})
			for range test.down {
				send(model, keyDown)
			}

			row := strings.Split(model.View().Content, "\n")[0]
			if !strings.HasPrefix(row, test.want) {
				t.Errorf("path row = %q, want prefix %q", row, test.want)
			}
		})
	}
}

func TestViewSelectHint(t *testing.T) {
	tests := []struct {
		name  string
		width int
		down  int
		want  string
	}{
		{name: "fits", width: 25, want: " ▌/home  select directory"},
		{name: "too narrow", width: 24, want: " ▌/home"},
		{name: "entry selected", width: 25, down: 1, want: "  /home/alice"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := newModel(t, "/home")
			model.Update(tea.WindowSizeMsg{Width: test.width, Height: 10})
			for range test.down {
				send(model, keyDown)
			}

			row := strings.Split(model.View().Content, "\n")[0]
			if got := strings.TrimRight(ansi.Strip(row), " "); got != test.want {
				t.Errorf("path row = %q, want %q", got, test.want)
			}
			if got := ansi.StringWidth(row); test.down == 0 && got != test.width {
				t.Errorf("path row width = %d, want %d", got, test.width)
			}
		})
	}
}

func TestViewPromptStyle(t *testing.T) {
	model := newModel(t, "/home")
	theme := model.config.Theme
	idle := leftPadding + theme.Prompt.Render(theme.PromptSymbol)
	active := leftPadding + theme.PromptWithQuery.Render(theme.PromptSymbol)
	searchLine := func() string { return strings.Split(model.View().Content, "\n")[1] }

	if line := searchLine(); !strings.HasPrefix(line, idle) {
		t.Errorf("search line = %q, want prefix %q", line, idle)
	}

	send(model, typed("a")...)
	if line := searchLine(); !strings.HasPrefix(line, active) {
		t.Errorf("search line = %q, want prefix %q", line, active)
	}

	send(model, keyBack)
	if line := searchLine(); !strings.HasPrefix(line, idle) {
		t.Errorf("search line = %q, want prefix %q", line, idle)
	}
}

func TestViewHome(t *testing.T) {
	model := newModel(t, "/home")
	model.config.Home = "/home"

	assertView(t, model, " ▌~")
}

func TestViewHighlightsMatches(t *testing.T) {
	model := newModel(t, "/home")

	send(model, typed("bo")...)

	theme := model.config.Theme
	want := theme.SelectedMatch.Render("bo") + theme.Selected.Render("b")
	if content := model.View().Content; !strings.Contains(content, want) {
		t.Errorf("View() does not contain the highlighted match %q:\n%q", want, content)
	}
}

func TestViewHints(t *testing.T) {
	model := newModel(t, "/home")
	assertView(t, model, "↵ select  ← back  esc cancel")

	send(model, keyDown)
	assertView(t, model, "↵ open  ← back  esc top")

	send(model, typed("b")...)
	assertView(t, model, "↵ open  ← back  esc clear")
}

func TestViewCustomTexts(t *testing.T) {
	model := newModel(t, "/home")
	model.config.Theme.HintSelectText = "pick"
	model.config.Theme.HintBackText = "up"
	model.config.Theme.HintCancelText = "quit"
	model.config.Theme.ErrorPrefix = "oops: "
	model.config.Theme.KeyLabels = map[string]string{"enter": "RET"}

	assertView(t, model, "RET pick  left up  esc quit")

	send(model, keyDown, keyRight)
	assertView(t, model, "oops: permission denied")
}

func TestViewCounter(t *testing.T) {
	model := newModel(t, "/home")
	assertView(t, model, " 5/5")

	send(model, typed("ca")...)
	assertView(t, model, " 1/5")
}

func TestViewCursor(t *testing.T) {
	model := newModel(t, "/home")

	send(model, typed("bo")...)

	cursor := model.View().Cursor
	if cursor == nil || cursor.X != 5 || cursor.Y != 1 {
		t.Errorf("Cursor = %+v, want X 5, Y 1", cursor)
	}
}

func TestAdjustOffset(t *testing.T) {
	tests := []struct {
		name                          string
		offset, cursor, height, total int
		want                          int
	}{
		{name: "no room", offset: 3, cursor: 5, height: 0, total: 10, want: 0},
		{name: "cursor visible", offset: 2, cursor: 3, height: 3, total: 10, want: 2},
		{name: "cursor above", offset: 4, cursor: 1, height: 3, total: 10, want: 1},
		{name: "cursor below", offset: 0, cursor: 5, height: 3, total: 10, want: 3},
		{name: "list shrank", offset: 8, cursor: 2, height: 3, total: 4, want: 1},
		{name: "list fits", offset: 2, cursor: 1, height: 5, total: 3, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := adjustOffset(test.offset, test.cursor, test.height, test.total); got != test.want {
				t.Errorf("adjustOffset() = %d, want %d", got, test.want)
			}
		})
	}
}

func newModel(t *testing.T, dir string) *Model {
	t.Helper()
	navigator, err := nav.New(tree, search.Matcher{}, dir, false)
	if err != nil {
		t.Fatalf("nav.New() error: %v", err)
	}
	model := New(navigator, Config{Keys: DefaultKeyMap(), Theme: DefaultTheme()})
	model.Update(tea.WindowSizeMsg{Width: testWidth, Height: testHeight})
	return model
}

// send returns the command of the last message.
func send(model *Model, msgs ...tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	for _, msg := range msgs {
		_, cmd = model.Update(msg)
	}
	return cmd
}

func typed(text string) []tea.Msg {
	var msgs []tea.Msg
	for _, r := range text {
		msgs = append(msgs, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return msgs
}

func assertQuery(t *testing.T, model *Model, query string) {
	t.Helper()
	if got := model.navigator.Query(); got != query {
		t.Errorf("Query() = %q, want %q", got, query)
	}
}

func assertDir(t *testing.T, model *Model, dir string) {
	t.Helper()
	if got := model.navigator.Dir(); got != dir {
		t.Errorf("Dir() = %q, want %q", got, dir)
	}
}

func assertView(t *testing.T, model *Model, want string) {
	t.Helper()
	if content := plainView(model); !strings.Contains(content, want) {
		t.Errorf("View() does not contain %q:\n%s", want, content)
	}
}

func assertCancelled(t *testing.T, model *Model) {
	t.Helper()
	if path, ok := model.Result(); ok {
		t.Errorf("Result() = %q, true, want cancelled", path)
	}
}

// plainView strips ANSI codes and trailing spaces.
func plainView(model *Model) string {
	lines := strings.Split(ansi.Strip(model.View().Content), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	return strings.Join(lines, "\n")
}

func assertQuit(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("command = nil, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("command does not return tea.QuitMsg")
	}
}
