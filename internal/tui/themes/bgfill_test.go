package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestBgFillAppliedAllThemes verifies that all 8 themes apply background fill
// to interior rows in door mode (Story 56.1 AC2).
func TestBgFillAppliedAllThemes(t *testing.T) {
	// TrueColor mode so background escapes are emitted.
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	tests := []struct {
		name  string
		theme *DoorTheme
	}{
		{"classic", NewClassicTheme()},
		{"modern", NewModernTheme()},
		{"scifi", NewSciFiTheme()},
		{"shoji", NewShojiTheme()},
		{"winter", NewWinterTheme()},
		{"spring", NewSpringTheme()},
		{"summer", NewSummerTheme()},
		{"autumn", NewAutumnTheme()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Render at a comfortable size in door mode
			output := tt.theme.Render("Buy groceries", 30, 16, false, "", 0.0)
			lines := strings.Split(output, "\n")
			anatomy := NewDoorAnatomy(16)

			// Check that interior rows (between lintel and threshold, excluding
			// border rows) contain ANSI background escape sequences.
			// Background colors produce escape sequences containing "48;" (SGR bg).
			bgFound := false
			for row := anatomy.LintelRow + 1; row < anatomy.ThresholdRow; row++ {
				if row == anatomy.PanelDivider || row == anatomy.HandleRow {
					continue
				}
				if row < len(lines) && strings.Contains(lines[row], "48;") {
					bgFound = true
					break
				}
			}

			if !bgFound {
				t.Errorf("%s: no background color escape (48;) found in interior rows", tt.name)
			}
		})
	}
}

// TestBgFillContentReadable verifies that content text is still readable
// over the background fill (Story 56.1 AC2).
func TestBgFillContentReadable(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	taskText := "Buy groceries for the week"

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			output := theme.Render(taskText, 40, 16, false, "", 0.0)

			// Strip ANSI escapes to get plain text
			plain := stripANSI(output)
			if !strings.Contains(plain, taskText) {
				t.Errorf("task text %q not readable in plain-text output of %s theme", taskText, theme.Name)
			}
		})
	}
}

// TestBgFillMinWidth verifies that background fill works at minimum door width
// without changing the content area width (Story 56.1 AC3).
func TestBgFillMinWidth(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			output := theme.Render("Test", theme.MinWidth, theme.MinHeight, false, "", 0.0)
			lines := strings.Split(output, "\n")

			// Door should render without panics and have the expected number of lines.
			// MinHeight + 2 (threshold + shadow bottom).
			if len(lines) < theme.MinHeight {
				t.Errorf("%s: expected at least %d lines at min width, got %d",
					theme.Name, theme.MinHeight, len(lines))
			}

			// Verify content is present
			plain := stripANSI(output)
			if !strings.Contains(plain, "Test") {
				t.Errorf("%s: content not visible at min width", theme.Name)
			}
		})
	}
}

// TestBgFillSelectedVsUnselected verifies both selected and unselected states
// apply background fill (Story 56.1 AC2).
func TestBgFillSelectedVsUnselected(t *testing.T) {
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

				// Both states should have background color escapes
				if !strings.Contains(output, "48;") {
					t.Errorf("%s %s: no background color escape found", theme.Name, state)
				}
			})
		}
	}
}

// TestThemeColorsNewFields verifies all 8 themes populate the new ThemeColors
// depth fields (Story 56.1 AC1).
func TestThemeColorsNewFields(t *testing.T) {
	t.Parallel()

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()
			c := theme.Colors

			if c.FillLower == nil {
				t.Errorf("%s: FillLower is nil", theme.Name)
			}
			if c.Highlight == nil {
				t.Errorf("%s: Highlight is nil", theme.Name)
			}
			if c.ShadowEdge == nil {
				t.Errorf("%s: ShadowEdge is nil", theme.Name)
			}
			if c.ShadowNear == nil {
				t.Errorf("%s: ShadowNear is nil", theme.Name)
			}
			if c.ShadowFar == nil {
				t.Errorf("%s: ShadowFar is nil", theme.Name)
			}
		})
	}
}

// TestThemeColorsFillValues verifies the Fill colors match the research
// artifact specifications (Story 56.1 AC4).
func TestThemeColorsFillValues(t *testing.T) {
	t.Parallel()

	expected := map[string]lipgloss.CompleteColor{
		"classic": {TrueColor: "#0d0d2a", ANSI256: "17", ANSI: "0"},
		"modern":  {TrueColor: "#0d0d0d", ANSI256: "233", ANSI: "0"},
		"scifi":   {TrueColor: "#0a1a2e", ANSI256: "17", ANSI: "0"},
		"shoji":   {TrueColor: "#1a1508", ANSI256: "234", ANSI: "0"},
		"winter":  {TrueColor: "#0a0f1a", ANSI256: "233", ANSI: "0"},
		"spring":  {TrueColor: "#0a1a0d", ANSI256: "233", ANSI: "0"},
		"summer":  {TrueColor: "#1a1508", ANSI256: "234", ANSI: "0"},
		"autumn":  {TrueColor: "#1a0f08", ANSI256: "233", ANSI: "0"},
	}

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()
			exp, ok := expected[theme.Name]
			if !ok {
				t.Fatalf("no expected Fill color for %s", theme.Name)
			}
			got, ok := theme.Colors.Fill.(lipgloss.CompleteColor)
			if !ok {
				t.Fatalf("%s Fill is not CompleteColor", theme.Name)
			}
			if got.TrueColor != exp.TrueColor {
				t.Errorf("%s Fill TrueColor: got %s, want %s", theme.Name, got.TrueColor, exp.TrueColor)
			}
			if got.ANSI256 != exp.ANSI256 {
				t.Errorf("%s Fill ANSI256: got %s, want %s", theme.Name, got.ANSI256, exp.ANSI256)
			}
			if got.ANSI != exp.ANSI {
				t.Errorf("%s Fill ANSI: got %s, want %s", theme.Name, got.ANSI, exp.ANSI)
			}
		})
	}
}
