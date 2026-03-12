package themes

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// ThemeColors holds the color palette for a door theme.
// All fields use lipgloss.TerminalColor to support adaptive color profiles
// (TrueColor, ANSI256, ANSI 16-color) for graceful degradation on
// constrained terminals.
type ThemeColors struct {
	Frame    lipgloss.TerminalColor
	Fill     lipgloss.TerminalColor
	Accent   lipgloss.TerminalColor
	Selected lipgloss.TerminalColor

	// Stats dashboard colors (Story 40.9). Zero values are safe —
	// InsightsView falls back to the independent palette when these are empty.
	StatsAccent        string // panel borders, hero number (#RRGGBB)
	StatsGradientStart string // sparkline low end (#RRGGBB)
	StatsGradientEnd   string // sparkline high end (#RRGGBB)
}

// MonthDay represents a calendar day within any year (month + day).
type MonthDay struct {
	Month int
	Day   int
}

// DoorTheme defines the visual frame for a door.
type DoorTheme struct {
	Name        string
	Description string
	Render      func(content string, width int, height int, selected bool, hint string, emphasis float64) string
	Colors      ThemeColors
	MinWidth    int
	MinHeight   int

	// Seasonal metadata. Zero-value Season ("") indicates a non-seasonal theme.
	Season      string
	SeasonStart MonthDay
	SeasonEnd   MonthDay
}

// DefaultThemeName is the theme used when no theme is specified.
const DefaultThemeName = "modern"

// HandleFrames defines the 4-frame animation sequence for a door handle
// as it turns during selection. Each theme provides its own character set.
type HandleFrames struct {
	Rest       string // emphasis 0.0 — handle at rest
	Turning    string // emphasis ~0.3 — handle mid-turn (forward)
	Turned     string // emphasis ~0.6+ — handle fully turned
	SpringBack string // emphasis ~0.3–0.6 during deselect — springing back
}

// Standard handle frame sets for themes.
var (
	// RoundKnobFrames uses filled/half/empty circles (classic, autumn).
	RoundKnobFrames = HandleFrames{Rest: "●", Turning: "◐", Turned: "○", SpringBack: "◑"}
	// OpenKnobFrames reverses the rotation for themes with open-circle rest (modern, shoji, spring).
	OpenKnobFrames = HandleFrames{Rest: "○", Turning: "◑", Turned: "●", SpringBack: "◐"}
	// SquareHandleFrames uses geometric squares (summer).
	SquareHandleFrames = HandleFrames{Rest: "■", Turning: "◧", Turned: "□", SpringBack: "◨"}
	// DiamondHandleFrames uses diamond shapes (winter).
	DiamondHandleFrames = HandleFrames{Rest: "◆", Turning: "◇", Turned: "○", SpringBack: "◑"}
	// SciFiHandleFrames uses diamond/circle shapes for the access panel (sci-fi).
	SciFiHandleFrames = HandleFrames{Rest: "◈", Turning: "◇", Turned: "○", SpringBack: "◑"}
)

// HandleCharForEmphasis returns the handle character for a given animation
// state. selected indicates direction: true = forward (selecting),
// false = reverse (deselecting). The emphasis value (0.0–1.0+) maps to
// thresholds at 0.3 and 0.6 to select the appropriate frame.
func HandleCharForEmphasis(emphasis float64, selected bool, frames HandleFrames) string {
	if emphasis < 0.3 {
		return frames.Rest
	}
	if emphasis >= 0.6 {
		return frames.Turned
	}
	// Mid-range 0.3–0.6: direction-dependent
	if selected {
		return frames.Turning
	}
	return frames.SpringBack
}

// renderHandleWithHint builds a handle row line, placing hint text to the left
// of the handle symbol when hint is non-empty. When hint is empty, renders the
// handle in its standard position. innerWidth is the total interior width between
// the vertical border characters. knobPad is the default left padding for the
// handle symbol. handleSym is the handle character (e.g. "●", "○", "◈──┤").
func renderHandleWithHint(innerWidth, knobPad int, handleSym, hint string) string {
	if hint == "" {
		rightPad := innerWidth - knobPad - 1
		if rightPad < 0 {
			rightPad = 0
		}
		return strings.Repeat(" ", knobPad) + handleSym + strings.Repeat(" ", rightPad)
	}
	hintWidth := ansi.StringWidth(hint)
	handleWidth := ansi.StringWidth(handleSym)
	// Layout: [padding] hint [space] handle [rightPad]
	leftPad := innerWidth - hintWidth - 1 - handleWidth - 1
	if leftPad < 1 {
		leftPad = 1
	}
	rightPad := innerWidth - leftPad - hintWidth - 1 - handleWidth
	if rightPad < 0 {
		rightPad = 0
	}
	return strings.Repeat(" ", leftPad) + hint + " " + handleSym + strings.Repeat(" ", rightPad)
}
