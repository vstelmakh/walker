package ui

import (
	"path/filepath"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Shows the home directory and paths inside it with "~".
func abbreviateHome(dir, home string) string {
	if home == "" {
		return dir
	}
	if dir == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(dir, home+string(filepath.Separator)); ok {
		return "~" + string(filepath.Separator) + rest
	}
	return dir
}

// Keeps the end of a string and starts it with the ellipsis when it cuts.
func truncateLeft(s string, width int, ellipsis string) string {
	excess := ansi.StringWidth(s) - width
	if excess <= 0 {
		return s
	}
	return ellipsis + ansi.TruncateLeft(s, excess+ansi.StringWidth(ellipsis), "")
}

// Keeps the first line of pasted text and drops the control characters.
func sanitizePaste(s string) string {
	s, _, _ = strings.Cut(strings.ReplaceAll(s, "\r", "\n"), "\n")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

func highlight(s string, positions []int, base, match lipgloss.Style) string {
	var b strings.Builder
	runStart, runMatch, next := 0, false, 0
	flush := func(end int) {
		if end == runStart {
			return
		}
		style := base
		if runMatch {
			style = match
		}
		b.WriteString(style.Render(s[runStart:end]))
		runStart = end
	}

	for i := range s {
		isMatch := next < len(positions) && positions[next] == i
		if isMatch {
			next++
		}
		if isMatch != runMatch {
			flush(i)
			runMatch = isMatch
		}
	}
	flush(len(s))
	return b.String()
}

func firstKeyLabel(keys []string, labels map[string]string) string {
	if len(keys) == 0 {
		return ""
	}
	if label, ok := labels[keys[0]]; ok {
		return label
	}
	return keys[0]
}
