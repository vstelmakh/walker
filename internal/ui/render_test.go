package ui

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func TestAbbreviateHome(t *testing.T) {
	tests := []struct {
		dir, home, want string
	}{
		{dir: "/home/alice", home: "/home/alice", want: "~"},
		{dir: "/home/alice/projects", home: "/home/alice", want: "~/projects"},
		{dir: "/home/alicia", home: "/home/alice", want: "/home/alicia"},
		{dir: "/tmp", home: "/home/alice", want: "/tmp"},
		{dir: "/home/alice", home: "", want: "/home/alice"},
	}
	for _, test := range tests {
		if got := abbreviateHome(test.dir, test.home); got != test.want {
			t.Errorf("abbreviateHome(%q, %q) = %q, want %q", test.dir, test.home, got, test.want)
		}
	}
}

func TestTruncateLeft(t *testing.T) {
	tests := []struct {
		s     string
		width int
		want  string
	}{
		{s: "/home/alice", width: 20, want: "/home/alice"},
		{s: "/home/alice", width: 11, want: "/home/alice"},
		{s: "/home/alice", width: 8, want: "…e/alice"},
		{s: "~/проекты", width: 5, want: "…екты"},
	}
	for _, test := range tests {
		if got := truncateLeft(test.s, test.width, "…"); got != test.want {
			t.Errorf("truncateLeft(%q, %d) = %q, want %q", test.s, test.width, got, test.want)
		}
	}
}

func TestHighlight(t *testing.T) {
	base := lipgloss.NewStyle().Italic(true)
	match := lipgloss.NewStyle().Bold(true)

	tests := []struct {
		name      string
		s         string
		positions []int
		want      string
	}{
		{name: "no positions", s: "abc", positions: nil, want: base.Render("abc")},
		{name: "runs", s: "abcd", positions: []int{0, 1, 3}, want: match.Render("ab") + base.Render("c") + match.Render("d")},
		{name: "byte offsets", s: "ébc", positions: []int{0, 3}, want: match.Render("é") + base.Render("b") + match.Render("c")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := highlight(test.s, test.positions, base, match); got != test.want {
				t.Errorf("highlight(%q, %v) = %q, want %q", test.s, test.positions, got, test.want)
			}
		})
	}
}

func TestFirstKeyLabel(t *testing.T) {
	tests := []struct {
		keys []string
		want string
	}{
		{keys: []string{"enter", "l"}, want: "↵"},
		{keys: []string{"ctrl+o"}, want: "ctrl+o"},
		{keys: nil, want: ""},
	}
	for _, test := range tests {
		if got := firstKeyLabel(test.keys, DefaultTheme().KeyLabels); got != test.want {
			t.Errorf("firstKeyLabel(%q) = %q, want %q", test.keys, got, test.want)
		}
	}
}

func TestSanitizePaste(t *testing.T) {
	tests := []struct {
		name, in, want string
	}{
		{name: "plain", in: "car", want: "car"},
		{name: "spaces stay", in: "my dir", want: "my dir"},
		{name: "first line", in: "car\nbob", want: "car"},
		{name: "windows newline", in: "car\r\nbob", want: "car"},
		{name: "carriage return", in: "car\rbob", want: "car"},
		{name: "leading newline", in: "\ncar", want: ""},
		{name: "control characters", in: "c\ta\x1br\x00", want: "car"},
		{name: "unicode", in: "тека", want: "тека"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := sanitizePaste(test.in); got != test.want {
				t.Errorf("sanitizePaste(%q) = %q, want %q", test.in, got, test.want)
			}
		})
	}
}
