package themes

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

// TestGolden_DoorAllThemes generates golden files across all 4 themes at
// 3 heights × 3 widths × 2 states = 72 golden files (AC1, AC2).
func TestGolden_DoorAllThemes(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	type themeSpec struct {
		theme     *DoorTheme
		minHeight int
	}

	themes := []themeSpec{
		{NewClassicTheme(), 10},
		{NewModernTheme(), 12},
		{NewSciFiTheme(), 14},
		{NewShojiTheme(), 14},
		{NewWinterTheme(), 12},
		{NewSpringTheme(), 12},
		{NewSummerTheme(), 12},
		{NewAutumnTheme(), 12},
	}

	heights := []struct {
		label string
		value func(minH int) int
	}{
		{"min", func(minH int) int { return minH }},
		{"h16", func(_ int) int { return 16 }},
		{"h24", func(_ int) int { return 24 }},
	}

	widths := []struct {
		label string
		value func(theme *DoorTheme) int
	}{
		{"wMin", func(theme *DoorTheme) int { return theme.MinWidth }},
		{"w80", func(_ *DoorTheme) int { return 80 }},
		{"w120", func(_ *DoorTheme) int { return 120 }},
	}

	for _, ts := range themes {
		for _, h := range heights {
			for _, w := range widths {
				for _, sel := range []bool{false, true} {
					state := "unselected"
					if sel {
						state = "selected"
					}
					height := h.value(ts.minHeight)
					width := w.value(ts.theme)
					name := ts.theme.Name + "/" + h.label + "_" + w.label + "_" + state
					t.Run(name, func(t *testing.T) {
						out := ts.theme.Render("Buy groceries for the week", width, height, sel, "", 0.0)
						golden.RequireEqual(t, []byte(out))
					})
				}
			}
		}
	}
}

// TestMonochromeDoorSignifiers verifies that door structural elements are
// distinguishable using structural characters only, without color (AC3).
func TestMonochromeDoorSignifiers(t *testing.T) {
	// Not parallel: sets global lipgloss color profile.
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	type signifierCheck struct {
		theme         *DoorTheme
		dividerChars  []string // characters expected in panel divider row
		handleChars   []string // characters expected in handle row
		thresholdChar string   // character expected in threshold/last row
	}

	checks := []signifierCheck{
		{
			NewClassicTheme(),
			[]string{"─", "╟"},
			[]string{"●"},
			"▔",
		},
		{
			NewModernTheme(),
			[]string{"─"},
			[]string{"○"},
			"▔",
		},
		{
			NewSciFiTheme(),
			[]string{"═", "╞"},
			[]string{"◈"},
			"▓",
		},
		{
			NewShojiTheme(),
			[]string{"─", "┼"},
			[]string{"○"},
			"▔",
		},
		{
			NewWinterTheme(),
			[]string{"─", "╟"},
			[]string{"◆"},
			"▔",
		},
		{
			NewSpringTheme(),
			[]string{"─", "╟"},
			[]string{"○"},
			"▔",
		},
		{
			NewSummerTheme(),
			[]string{"═", "╠"},
			[]string{"■"},
			"▀",
		},
		{
			NewAutumnTheme(),
			[]string{"─", "╟"},
			[]string{"●"},
			"▒",
		},
	}

	for _, c := range checks {
		t.Run(c.theme.Name, func(t *testing.T) {
			output := c.theme.Render("Test task", 40, 16, false, "", 0.0)
			lines := strings.Split(output, "\n")
			anatomy := NewDoorAnatomy(16)

			// Check panel divider row
			if anatomy.PanelDivider < len(lines) {
				dividerLine := lines[anatomy.PanelDivider]
				for _, ch := range c.dividerChars {
					if !strings.Contains(dividerLine, ch) {
						t.Errorf("%s: panel divider at row %d missing %q in: %q",
							c.theme.Name, anatomy.PanelDivider, ch, dividerLine)
					}
				}
			}

			// Check handle row
			if anatomy.HandleRow < len(lines) {
				handleLine := lines[anatomy.HandleRow]
				for _, ch := range c.handleChars {
					if !strings.Contains(handleLine, ch) {
						t.Errorf("%s: handle row at %d missing %q in: %q",
							c.theme.Name, anatomy.HandleRow, ch, handleLine)
					}
				}
			}

			// Check threshold (second-to-last; last is shadow bottom)
			threshLine := lines[len(lines)-2]
			if !strings.Contains(threshLine, c.thresholdChar) {
				t.Errorf("%s: threshold missing %q in: %q",
					c.theme.Name, c.thresholdChar, threshLine)
			}

			// Check shadow bottom row contains ▄
			shadowLine := lines[len(lines)-1]
			if !strings.Contains(shadowLine, "▄") {
				t.Errorf("%s: shadow bottom missing ▄ in: %q",
					c.theme.Name, shadowLine)
			}
		})
	}
}

// TestCompactFallbackAllThemes verifies that all themes fall back to card
// style when height < MinHeight, and that content remains readable (AC4, AC5).
func TestCompactFallbackAllThemes(t *testing.T) {
	t.Parallel()

	// Door-only structural elements per theme that should NOT appear in compact mode.
	// Note: some themes reuse characters like ● in compact card layout (e.g. Modern),
	// so we check only elements exclusive to door mode.
	themeChecks := map[string][]string{
		"classic": {"▔", "╟"},      // threshold and hinge divider junction
		"modern":  {"▔", "○"},      // threshold and open circle handle (door mode)
		"scifi":   {"▓", "◈", "╞"}, // floor grating, access panel, bulkhead junction
		"shoji":   {"▔", "○"},      // threshold and handle (lattice chars shared with compact)
		"winter":  {"▔", "╟"},      // threshold and hinge divider junction
		"spring":  {"▔", "╟"},      // threshold and hinge divider junction
		"summer":  {"▀", "╠"},      // bold threshold and hinge divider
		"autumn":  {"▒", "╟"},      // block threshold and hinge divider junction
	}

	content := "Important task to complete"

	for _, theme := range allThemes() {
		t.Run(theme.Name+"/height_zero", func(t *testing.T) {
			t.Parallel()
			output := theme.Render(content, 40, 0, false, "", 0.0)

			for _, elem := range themeChecks[theme.Name] {
				if strings.Contains(output, elem) {
					t.Errorf("compact mode (height=0) should not contain door element %q", elem)
				}
			}

			if !strings.Contains(output, content) {
				t.Errorf("compact mode should contain task content %q", content)
			}
		})

		t.Run(theme.Name+"/below_min_height", func(t *testing.T) {
			t.Parallel()
			output := theme.Render(content, 40, theme.MinHeight-1, false, "", 0.0)

			// Content must be present
			if !strings.Contains(output, content) {
				t.Errorf("below-MinHeight mode should contain task content %q", content)
			}
		})
	}
}

// TestScreenReaderTextExtraction verifies that task text is accessible in
// reading order when decorative characters are stripped (AC6).
func TestScreenReaderTextExtraction(t *testing.T) {
	// Not parallel: sets global lipgloss color profile.
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	// Regex to strip non-alphanumeric characters (keeping spaces)
	stripDecorative := regexp.MustCompile(`[^\p{L}\p{N}\s]`)

	taskText := "Buy groceries for the week"

	for _, theme := range allThemes() {
		for _, height := range []int{0, 16, 24} {
			label := "compact"
			if height > 0 {
				label = "h" + itoa(height)
			}
			t.Run(theme.Name+"/"+label, func(t *testing.T) {
				output := theme.Render(taskText, 40, height, false, "", 0.0)
				plainText := stripDecorative.ReplaceAllString(output, "")

				// Normalize whitespace
				words := strings.Fields(plainText)
				joined := strings.Join(words, " ")

				if !strings.Contains(joined, taskText) {
					t.Errorf("task text %q not found in plain-text extraction: %q",
						taskText, joined)
				}
			})
		}
	}
}
