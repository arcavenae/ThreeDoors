package tui

import (
	"strings"
)

// layoutFull pads the combined header + content + footer output to exactly
// totalHeight lines, filling the terminal vertically. Remaining vertical
// space is inserted between the content and footer regions.
//
// If the combined output already meets or exceeds totalHeight, lines are
// returned as-is (no truncation — Bubbletea handles overflow via AltScreen
// scrolling).
func layoutFull(header, content, footer string, totalHeight int) string {
	if totalHeight <= 0 {
		return joinNonEmpty(header, content, footer)
	}

	combined := joinNonEmpty(header, content, footer)
	currentLines := strings.Count(combined, "\n") + 1

	if currentLines >= totalHeight {
		return combined
	}

	// Insert padding between content and footer to fill the terminal.
	padding := totalHeight - currentLines
	padStr := strings.Repeat("\n", padding)

	if footer == "" {
		return joinNonEmpty(header, content) + padStr
	}

	return joinNonEmpty(header, content) + padStr + "\n" + footer
}

// joinNonEmpty joins non-empty strings with newlines.
func joinNonEmpty(parts ...string) string {
	var nonEmpty []string
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, "\n")
}
