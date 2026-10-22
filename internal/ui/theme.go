package ui

import "charm.land/lipgloss/v2"

type Theme struct {
	PromptSymbol string
	// MarkerSymbol must be one cell wide.
	MarkerSymbol         string
	PlaceholderText      string
	SelectHintText       string
	NoMatchesText        string
	NoSubdirectoriesText string
	// ErrorPrefix goes before the error message in the footer.
	ErrorPrefix string
	// The footer hint texts say what a key does. Each key is shown with the text for its current action.
	HintOpenText   string
	HintSelectText string
	HintBackText   string
	HintClearText  string
	HintTopText    string
	HintCancelText string
	// KeyLabels maps a key name to the symbol shown in the footer. A key without a label shows its name.
	KeyLabels map[string]string
	// SeparatorSymbol must be one cell wide.
	SeparatorSymbol string
	// EllipsisSymbol marks text that is cut to fit the width.
	EllipsisSymbol string

	Separator lipgloss.Style
	Path      lipgloss.Style
	// PathPreview styles the name of the selected entry after the path.
	PathPreview     lipgloss.Style
	Prompt          lipgloss.Style
	PromptWithQuery lipgloss.Style
	Query           lipgloss.Style
	Placeholder     lipgloss.Style
	Entry           lipgloss.Style
	EntryMatch      lipgloss.Style
	EmptyList       lipgloss.Style
	SelectHint      lipgloss.Style
	Selected        lipgloss.Style
	SelectedMatch   lipgloss.Style
	Marker          lipgloss.Style
	Counter         lipgloss.Style
	Hint            lipgloss.Style
	Error           lipgloss.Style
}

// DefaultTheme uses only the 16 ANSI colors, so it follows the colors of the terminal theme.
func DefaultTheme() Theme {
	accent := lipgloss.Magenta
	accentBright := lipgloss.BrightMagenta
	plain := lipgloss.NewStyle()
	dim := plain.Foreground(lipgloss.BrightBlack)
	selected := plain.Background(lipgloss.BrightBlack).Foreground(lipgloss.BrightWhite)

	return Theme{
		PromptSymbol:         ": ",
		MarkerSymbol:         "▌",
		PlaceholderText:      "type to search",
		SelectHintText:       "select directory",
		NoMatchesText:        "no matches",
		NoSubdirectoriesText: "no subdirectories",
		ErrorPrefix:          "error: ",
		HintOpenText:         "open",
		HintSelectText:       "select",
		HintBackText:         "back",
		HintClearText:        "clear",
		HintTopText:          "top",
		HintCancelText:       "cancel",
		KeyLabels: map[string]string{
			"enter":     "↵",
			"up":        "↑",
			"down":      "↓",
			"left":      "←",
			"right":     "→",
			"backspace": "⌫",
			"space":     "␣",
		},
		SeparatorSymbol: "─",
		EllipsisSymbol:  "…",

		Path:            plain,
		PathPreview:     dim,
		Separator:       dim,
		Prompt:          dim,
		PromptWithQuery: plain.Foreground(accent),
		Query:           plain,
		Placeholder:     dim,
		Entry:           plain,
		EntryMatch:      plain.Foreground(accent).Bold(true),
		EmptyList:       dim,
		SelectHint:      selected.Faint(true),
		Selected:        selected,
		SelectedMatch:   selected.Foreground(accentBright).Bold(true),
		Marker:          selected.Foreground(accent),
		Counter:         dim,
		Hint:            dim,
		Error:           plain.Foreground(lipgloss.Red),
	}
}

func (t Theme) confirmHint(action confirmAction) string {
	if action == confirmSelect {
		return t.HintSelectText
	}
	return t.HintOpenText
}

func (t Theme) cancelHint(action cancelAction) string {
	switch action {
	case cancelClear:
		return t.HintClearText
	case cancelTop:
		return t.HintTopText
	default:
		return t.HintCancelText
	}
}
