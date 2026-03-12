package tui

import (
	"strings"
)

// Breakpoint represents terminal height tiers for graceful degradation (D-119).
type Breakpoint int

const (
	// BreakpointMinimal: height < 10 — doors only, no header/footer/keybinding bar.
	BreakpointMinimal Breakpoint = iota
	// BreakpointCompact: height 10-15 — 1-line header, doors at min 10, 1-line footer.
	BreakpointCompact
	// BreakpointStandard: height 16-24 — full header, proportional doors, footer with bar.
	BreakpointStandard
	// BreakpointComfortable: height 25-40 — breathing room appears.
	BreakpointComfortable
	// BreakpointSpacious: height 40+ — doors capped at 25, generous padding.
	BreakpointSpacious
)

// layoutBreakpoint returns the degradation tier for a given terminal height.
func layoutBreakpoint(height int) Breakpoint {
	switch {
	case height < 10:
		return BreakpointMinimal
	case height <= 15:
		return BreakpointCompact
	case height <= 24:
		return BreakpointStandard
	case height <= 40:
		return BreakpointComfortable
	default:
		return BreakpointSpacious
	}
}

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
