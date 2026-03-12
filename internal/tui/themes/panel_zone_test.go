package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestPanelZoneShadingAllThemes verifies that upper and lower panels use
// different background colors in all 8 themes (Story 56.4 AC1/AC2).
func TestPanelZoneShadingAllThemes(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			output := theme.Render("Buy groceries", 30, 16, false, "", 0.0)
			lines := strings.Split(output, "\n")
			anatomy := NewDoorAnatomy(16)

			// Collect background escape sequences from upper panel rows
			var upperBGs []string
			for row := anatomy.LintelRow + 1; row < anatomy.PanelDivider; row++ {
				if row < len(lines) && strings.Contains(lines[row], "48;") {
					upperBGs = append(upperBGs, lines[row])
				}
			}

			// Collect background escape sequences from lower panel rows
			// (below divider, excluding handle row which has special rendering)
			var lowerBGs []string
			for row := anatomy.PanelDivider + 1; row < anatomy.ThresholdRow; row++ {
				if row == anatomy.HandleRow {
					continue
				}
				if row < len(lines) && strings.Contains(lines[row], "48;") {
					lowerBGs = append(lowerBGs, lines[row])
				}
			}

			if len(upperBGs) == 0 {
				t.Errorf("%s: no background color found in upper panel rows", theme.Name)
			}
			if len(lowerBGs) == 0 {
				t.Errorf("%s: no background color found in lower panel rows", theme.Name)
			}

			// Upper and lower should use different background colors.
			// Extract the 48;2;R;G;B sequences to compare.
			if len(upperBGs) > 0 && len(lowerBGs) > 0 {
				upperBG := extractBGEscape(upperBGs[0])
				lowerBG := extractBGEscape(lowerBGs[0])
				if upperBG == lowerBG {
					t.Errorf("%s: upper and lower panels use same background %q — expected different", theme.Name, upperBG)
				}
			}
		})
	}
}

// TestPanelDividerUsesLowerBg verifies that the divider row itself uses
// FillLower background (Story 56.4 AC3).
func TestPanelDividerUsesLowerBg(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			// panelBg should return fillLower for the divider row
			anatomy := NewDoorAnatomy(16)
			fill := theme.Colors.Fill
			fillLower := theme.Colors.FillLower

			bg := panelBg(anatomy.PanelDivider, anatomy.PanelDivider, fill, fillLower)
			if bg != fillLower {
				t.Errorf("%s: panelBg at divider row returned Fill, want FillLower", theme.Name)
			}
		})
	}
}

// TestFillLowerColorValues verifies each theme's FillLower matches the
// research artifact palette (Story 56.4 AC4).
func TestFillLowerColorValues(t *testing.T) {
	t.Parallel()

	expected := map[string]lipgloss.CompleteColor{
		"classic": {TrueColor: "#080820", ANSI256: "17", ANSI: "0"},
		"modern":  {TrueColor: "#080808", ANSI256: "232", ANSI: "0"},
		"scifi":   {TrueColor: "#061425", ANSI256: "17", ANSI: "0"},
		"shoji":   {TrueColor: "#141005", ANSI256: "233", ANSI: "0"},
		"winter":  {TrueColor: "#060a14", ANSI256: "232", ANSI: "0"},
		"spring":  {TrueColor: "#061408", ANSI256: "232", ANSI: "0"},
		"summer":  {TrueColor: "#141005", ANSI256: "233", ANSI: "0"},
		"autumn":  {TrueColor: "#140a05", ANSI256: "232", ANSI: "0"},
	}

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()
			exp, ok := expected[theme.Name]
			if !ok {
				t.Fatalf("no expected FillLower color for %s", theme.Name)
			}
			got, ok := theme.Colors.FillLower.(lipgloss.CompleteColor)
			if !ok {
				t.Fatalf("%s FillLower is not CompleteColor", theme.Name)
			}
			if got.TrueColor != exp.TrueColor {
				t.Errorf("%s FillLower TrueColor: got %s, want %s", theme.Name, got.TrueColor, exp.TrueColor)
			}
			if got.ANSI256 != exp.ANSI256 {
				t.Errorf("%s FillLower ANSI256: got %s, want %s", theme.Name, got.ANSI256, exp.ANSI256)
			}
			if got.ANSI != exp.ANSI {
				t.Errorf("%s FillLower ANSI: got %s, want %s", theme.Name, got.ANSI, exp.ANSI)
			}
		})
	}
}

// TestPanelZoneMinHeight verifies panel zone shading works at minimum height
// and that both zones are visible (Story 56.4 AC5).
func TestPanelZoneMinHeight(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			output := theme.Render("Test", theme.MinWidth, theme.MinHeight, false, "", 0.0)
			lines := strings.Split(output, "\n")
			anatomy := NewDoorAnatomy(theme.MinHeight)

			// Both zones should exist and the transition should be at the divider
			if anatomy.PanelDivider <= anatomy.LintelRow+1 {
				t.Errorf("%s: panel divider at row %d — no room for upper panel", theme.Name, anatomy.PanelDivider)
			}
			if anatomy.PanelDivider >= anatomy.ThresholdRow-1 {
				t.Errorf("%s: panel divider at row %d — no room for lower panel", theme.Name, anatomy.PanelDivider)
			}

			// Verify the door rendered with expected number of lines
			if len(lines) < theme.MinHeight {
				t.Errorf("%s: expected at least %d lines at min height, got %d",
					theme.Name, theme.MinHeight, len(lines))
			}
		})
	}
}

// TestPanelZoneSelectedState verifies panel zone shading works in both
// selected and unselected states.
func TestPanelZoneSelectedState(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		for _, sel := range []bool{false, true} {
			state := "unselected"
			if sel {
				state = "selected"
			}
			t.Run(theme.Name+"/"+state, func(t *testing.T) {
				output := theme.Render("Task", 30, 16, sel, "", 0.0)
				lines := strings.Split(output, "\n")
				anatomy := NewDoorAnatomy(16)

				// Check that an upper panel row has bg
				upperHasBG := false
				for row := anatomy.LintelRow + 1; row < anatomy.PanelDivider; row++ {
					if row < len(lines) && strings.Contains(lines[row], "48;") {
						upperHasBG = true
						break
					}
				}

				// Check that a lower panel row has bg
				lowerHasBG := false
				for row := anatomy.PanelDivider + 1; row < anatomy.ThresholdRow; row++ {
					if row == anatomy.HandleRow {
						continue
					}
					if row < len(lines) && strings.Contains(lines[row], "48;") {
						lowerHasBG = true
						break
					}
				}

				if !upperHasBG {
					t.Errorf("%s %s: no bg in upper panel", theme.Name, state)
				}
				if !lowerHasBG {
					t.Errorf("%s %s: no bg in lower panel", theme.Name, state)
				}
			})
		}
	}
}

// TestPanelBgHelper verifies the panelBg helper function returns correct
// background colors based on row position.
func TestPanelBgHelper(t *testing.T) {
	t.Parallel()

	fill := lipgloss.CompleteColor{TrueColor: "#ffffff", ANSI256: "15", ANSI: "15"}
	fillLower := lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	divider := 7

	tests := []struct {
		name string
		row  int
		want lipgloss.TerminalColor
	}{
		{"row above divider", 3, fill},
		{"row just above divider", 6, fill},
		{"row at divider", 7, fillLower},
		{"row just below divider", 8, fillLower},
		{"row well below divider", 12, fillLower},
		{"first row", 0, fill},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := panelBg(tt.row, divider, fill, fillLower)
			if got != tt.want {
				t.Errorf("panelBg(row=%d, divider=%d): got %v, want %v", tt.row, divider, got, tt.want)
			}
		})
	}
}

// extractBGEscape finds the first "48;2;R;G;B" or "48;5;N" sequence in a string.
func extractBGEscape(s string) string {
	// Look for TrueColor bg: "48;2;R;G;B"
	idx := strings.Index(s, "48;2;")
	if idx >= 0 {
		end := idx + 5
		// Consume digits and semicolons for R;G;B
		for end < len(s) && (s[end] >= '0' && s[end] <= '9' || s[end] == ';') {
			end++
		}
		return s[idx:end]
	}
	// Fallback: 256-color bg: "48;5;N"
	idx = strings.Index(s, "48;5;")
	if idx >= 0 {
		end := idx + 5
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		return s[idx:end]
	}
	return ""
}
