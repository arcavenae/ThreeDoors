package tui

import "github.com/charmbracelet/lipgloss"

// Inline hint ANSI colors per Story 39.9 / 39.10, simplified in 39.13 (fade removed).
var (
	// hintStyleNormal renders key hints in dim gray (ANSI 245).
	hintStyleNormal = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// hintStyleBright renders key hints in bright white (ANSI 255) for selected doors.
	hintStyleBright = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)

	// hintStyleDim renders key hints in dark gray (ANSI 240) for unselected doors when a selection is active.
	hintStyleDim = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// renderInlineHint returns a Lipgloss-styled "[key]" string when enabled,
// or an empty string when disabled.
func renderInlineHint(key string, enabled bool) string {
	if !enabled {
		return ""
	}
	return hintStyleNormal.Render("[" + key + "]")
}

// renderDoorHint returns a selection-state-aware inline hint for a door.
// When selected, the hint brightens (ANSI 255 + bold). When another door is
// selected, unselected doors' hints dim (ANSI 240). When no door is selected,
// all hints use normal brightness.
func renderDoorHint(key string, enabled bool, isSelected bool, hasSelection bool) string {
	if !enabled {
		return ""
	}
	text := "[" + key + "]"
	if hasSelection {
		if isSelected {
			return hintStyleBright.Render(text)
		}
		return hintStyleDim.Render(text)
	}
	return hintStyleNormal.Render(text)
}
