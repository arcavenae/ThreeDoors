package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Keybinding bar styles — dim/recessive to avoid competing with primary content.
var (
	// barKeyStyle renders key names slightly brighter for visual scanning.
	barKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	// barDescStyle renders descriptions in dim foreground.
	barDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	// barSeparatorStyle renders the thin separator line above the bar.
	barSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))
)

// Width breakpoints for bar content adaptation.
const (
	barWidthMinimal = 40 // Below this: show only "? Help"
	barWidthNarrow  = 60 // 40-60: show 3 keys + ?
	barWidthMedium  = 80 // 60-80: show 5 keys + ?
)

// Height breakpoints for bar visibility.
const (
	barHeightHidden  = 10 // Below this: bar hidden entirely
	barHeightCompact = 15 // 10-15: compact mode (keys only)
)

// RenderKeybindingBar renders a concise keybinding bar for the given view mode.
// It returns a two-line string (separator + bar) or empty string if disabled or
// terminal is too small. The function is stateless — all inputs are parameters.
func RenderKeybindingBar(mode ViewMode, width, height int, enabled bool, doorSelected bool) string {
	if !enabled {
		return ""
	}
	if height < barHeightHidden {
		return ""
	}

	bindings := barBindings(mode, doorSelected)
	if len(bindings) == 0 {
		return ""
	}

	compact := height <= barHeightCompact

	// Separate help binding from the rest — it's always last.
	var helpBinding KeyBinding
	var otherBindings []KeyBinding
	for _, b := range bindings {
		if b.Key == "?" {
			helpBinding = b
		} else {
			otherBindings = append(otherBindings, b)
		}
	}

	// Determine how many bindings to show based on width.
	maxBindings := len(otherBindings)
	switch {
	case width < barWidthMinimal:
		maxBindings = 0 // Only show help
	case width < barWidthNarrow:
		if maxBindings > 3 {
			maxBindings = 3
		}
	case width < barWidthMedium:
		if maxBindings > 5 {
			maxBindings = 5
		}
	}

	// Build the visible binding list: selected bindings + help at end.
	visible := otherBindings
	if maxBindings < len(visible) {
		visible = visible[:maxBindings]
	}
	if helpBinding.Key != "" {
		visible = append(visible, helpBinding)
	}

	// Format the bar content.
	bar := formatBar(visible, compact, width)

	// Build separator + bar.
	separator := barSeparatorStyle.Render(strings.Repeat("─", width))
	return separator + "\n" + bar
}

// formatBar renders the binding pairs as a single line, truncating if needed.
func formatBar(bindings []KeyBinding, compact bool, maxWidth int) string {
	if len(bindings) == 0 {
		return ""
	}

	// Build individual segments.
	segments := make([]string, 0, len(bindings))
	for _, b := range bindings {
		if compact {
			segments = append(segments, barKeyStyle.Render(b.Key))
		} else {
			segments = append(segments, fmt.Sprintf("%s %s",
				barKeyStyle.Render(b.Key),
				barDescStyle.Render(b.Description)))
		}
	}

	// Join with double space separator and check fit.
	joined := strings.Join(segments, "  ")
	if lipgloss.Width(joined) <= maxWidth {
		return joined
	}

	// Truncate from right, always keeping the last segment (help).
	if len(segments) <= 1 {
		return segments[0]
	}

	help := segments[len(segments)-1]
	for i := len(segments) - 2; i >= 0; i-- {
		candidate := strings.Join(segments[:i+1], "  ") + "  " + help
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}

	// Fall back to help only.
	return help
}
