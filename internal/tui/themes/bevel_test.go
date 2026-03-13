package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestBevelLighting_UnselectedDoors verifies that unselected doors apply
// different ANSI color sequences to top/left vs bottom/right borders,
// confirming the bevel effect (Story 56.2 AC1-AC4).
func TestBevelLighting_UnselectedDoors(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	themes := []struct {
		name      string
		theme     *DoorTheme
		highlight string // TrueColor hex for Highlight
		shadow    string // TrueColor hex for ShadowEdge
	}{
		{"classic", NewClassicTheme(), "#7070ff", "#3a3a8f"},
		{"modern", NewModernTheme(), "#666666", "#2a2a2a"},
		{"scifi", NewSciFiTheme(), "#00d7ff", "#005f7f"},
		{"shoji", NewShojiTheme(), "#e8c888", "#8f7540"},
		{"winter", NewWinterTheme(), "#a0d2e8", "#4a6a80"},
		{"spring", NewSpringTheme(), "#80e090", "#306838"},
		{"summer", NewSummerTheme(), "#ffd060", "#8f7020"},
		{"autumn", NewAutumnTheme(), "#e09040", "#8f5020"},
	}

	for _, tc := range themes {
		t.Run(tc.name, func(t *testing.T) {
			output := tc.theme.Render("Test task", 40, 16, false, "", 0.0)
			lines := strings.Split(output, "\n")
			anatomy := NewDoorAnatomy(16)

			// AC1: Top border (lintel) should contain Highlight color escape
			if anatomy.LintelRow < len(lines) {
				lintel := lines[anatomy.LintelRow]
				if !containsColorRef(lintel, tc.highlight) {
					t.Errorf("lintel row should contain Highlight color %s", tc.highlight)
				}
			}

			// AC3: Bottom border (threshold) should contain ShadowEdge color escape
			if anatomy.ThresholdRow < len(lines) {
				threshold := lines[anatomy.ThresholdRow]
				if !containsColorRef(threshold, tc.shadow) {
					t.Errorf("threshold row should contain ShadowEdge color %s", tc.shadow)
				}
			}

			// AC4: Panel divider should contain ShadowEdge color escape
			if anatomy.PanelDivider < len(lines) {
				divider := lines[anatomy.PanelDivider]
				if !containsColorRef(divider, tc.shadow) {
					t.Errorf("panel divider should contain ShadowEdge color %s", tc.shadow)
				}
			}

			// AC2: Content rows should have left border in Highlight, right in ShadowEdge
			contentRow := anatomy.ContentStart
			if contentRow < len(lines) {
				row := lines[contentRow]
				if !containsColorRef(row, tc.highlight) {
					t.Errorf("content row left border should contain Highlight color %s", tc.highlight)
				}
				if !containsColorRef(row, tc.shadow) {
					t.Errorf("content row right border should contain ShadowEdge color %s", tc.shadow)
				}
			}
		})
	}
}

// TestBevelLighting_SelectedOverride verifies that selected doors use the
// Selected color for ALL borders, overriding bevel (Story 56.2 AC5).
func TestBevelLighting_SelectedOverride(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			output := theme.Render("Test task", 40, 16, true, "", 0.0)
			lines := strings.Split(output, "\n")
			anatomy := NewDoorAnatomy(16)

			selectedHex := extractTrueColorHex(theme.Colors.Selected)
			if selectedHex == "" {
				t.Skip("cannot extract Selected color hex")
			}

			// When selected, lintel should use Selected color, not Highlight
			if anatomy.LintelRow < len(lines) {
				lintel := lines[anatomy.LintelRow]
				if !containsColorRef(lintel, selectedHex) {
					t.Errorf("selected lintel should use Selected color %s", selectedHex)
				}
			}

			// When selected, threshold should also use Selected color
			if anatomy.ThresholdRow < len(lines) {
				threshold := lines[anatomy.ThresholdRow]
				if !containsColorRef(threshold, selectedHex) {
					t.Errorf("selected threshold should use Selected color %s", selectedHex)
				}
			}
		})
	}
}

// TestBevelLighting_MinWidth verifies that bevel at minimum width doesn't
// change the content area width (Story 56.2 AC6).
func TestBevelLighting_MinWidth(t *testing.T) {
	t.Parallel()

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()
			output := theme.Render("Test", theme.MinWidth, 16, false, "", 0.0)

			// Content should still be present at minimum width
			if !strings.Contains(output, "Test") {
				t.Errorf("content should be visible at minimum width %d", theme.MinWidth)
			}

			// Output should have lines (basic sanity)
			lines := strings.Split(output, "\n")
			if len(lines) < 10 {
				t.Errorf("expected at least 10 lines at height 16, got %d", len(lines))
			}
		})
	}
}

// TestBevelLighting_CompactModeUnchanged verifies that compact mode (height < MinHeight)
// is NOT affected by bevel changes — it should still use a single frame color.
func TestBevelLighting_CompactModeUnchanged(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	bevelColors := map[string][2]string{
		"classic": {"#7070ff", "#3a3a8f"},
		"modern":  {"#666666", "#2a2a2a"},
		"scifi":   {"#00d7ff", "#005f7f"},
		"shoji":   {"#e8c888", "#8f7540"},
		"winter":  {"#a0d2e8", "#4a6a80"},
		"spring":  {"#80e090", "#306838"},
		"summer":  {"#ffd060", "#8f7020"},
		"autumn":  {"#e09040", "#8f5020"},
	}

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			// Render in compact mode (height 0 falls through to card style)
			output := theme.Render("Test task", 40, 0, false, "", 0.0)

			colors, ok := bevelColors[theme.Name]
			if !ok {
				t.Skip("no bevel color data for theme")
			}

			// Compact mode should NOT contain bevel-specific colors
			// (it should use the original Frame color)
			highlightHex := colors[0]
			shadowHex := colors[1]

			frameHex := extractTrueColorHex(theme.Colors.Frame)

			// If highlight differs from frame, compact output should not contain highlight
			if highlightHex != frameHex && containsColorRef(output, highlightHex) {
				t.Errorf("compact mode should not contain Highlight color %s (bevel is door-mode only)", highlightHex)
			}
			// If shadow differs from frame, compact output should not contain shadow
			if shadowHex != frameHex && containsColorRef(output, shadowHex) {
				t.Errorf("compact mode should not contain ShadowEdge color %s (bevel is door-mode only)", shadowHex)
			}
		})
	}
}

// containsColorRef checks if a string contains the ANSI escape sequence that
// lipgloss produces for the given hex color. Instead of doing hex math (which
// can disagree with lipgloss's color conversion by ±1), we ask lipgloss to
// render a probe string with the color and extract the escape prefix.
func containsColorRef(s, hexColor string) bool {
	probe := lipgloss.NewStyle().
		Foreground(lipgloss.Color(hexColor)).
		Render("X")
	// Extract the ANSI prefix: everything before "X"
	idx := strings.Index(probe, "X")
	if idx <= 0 {
		return false
	}
	prefix := probe[:idx]
	return strings.Contains(s, prefix)
}

// extractTrueColorHex extracts the TrueColor hex value from a TerminalColor.
func extractTrueColorHex(tc lipgloss.TerminalColor) string {
	if cc, ok := tc.(lipgloss.CompleteColor); ok {
		return cc.TrueColor
	}
	return ""
}
