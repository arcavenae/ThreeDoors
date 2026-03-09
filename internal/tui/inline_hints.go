package tui

import "github.com/charmbracelet/lipgloss"

// Inline hint ANSI colors per Story 39.9.
var (
	// hintStyleNormal renders key hints in dim gray (ANSI 245).
	hintStyleNormal = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// hintStyleFade renders key hints in extra-dim gray (ANSI 240) for fade mode.
	hintStyleFade = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// renderInlineHint returns a Lipgloss-styled "[key]" string when enabled,
// or an empty string when disabled. When fade is true, uses the extra-dim
// ANSI 240 style instead of the normal ANSI 245.
func renderInlineHint(key string, enabled bool, fade bool) string {
	if !enabled {
		return ""
	}
	style := hintStyleNormal
	if fade {
		style = hintStyleFade
	}
	return style.Render("[" + key + "]")
}
