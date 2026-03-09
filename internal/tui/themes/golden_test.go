package themes

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

// setAsciiProfile forces ASCII color profile for deterministic golden output
// and restores the original profile on test cleanup.
func setAsciiProfile(t *testing.T) {
	t.Helper()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })
}

// allThemes returns all built-in themes in a stable order.
func allThemes() []*DoorTheme {
	return []*DoorTheme{
		NewClassicTheme(),
		NewModernTheme(),
		NewSciFiTheme(),
		NewShojiTheme(),
		NewWinterTheme(),
		NewSpringTheme(),
		NewSummerTheme(),
		NewAutumnTheme(),
	}
}

// TestGolden_ThemeRender tests each theme at 28-char and 40-char widths,
// in both selected and unselected states (AC1, AC2).
func TestGolden_ThemeRender(t *testing.T) {
	setAsciiProfile(t)

	tests := []struct {
		width    int
		selected bool
	}{
		{28, false},
		{28, true},
		{40, false},
		{40, true},
	}

	for _, theme := range allThemes() {
		for _, tt := range tests {
			state := "unselected"
			if tt.selected {
				state = "selected"
			}
			name := theme.Name + "/" + state + "_w" + itoa(tt.width)
			t.Run(name, func(t *testing.T) {
				out := theme.Render("Buy groceries for the week", tt.width, 0, tt.selected, "")
				golden.RequireEqual(t, []byte(out))
			})
		}
	}
}

// TestGolden_ThemeBoundaryWidth tests each theme at MinWidth (should render
// correctly) and MinWidth-1 (verifies behavior at below-minimum width) (AC5).
func TestGolden_ThemeBoundaryWidth(t *testing.T) {
	setAsciiProfile(t)

	for _, theme := range allThemes() {
		t.Run(theme.Name+"/at_min_width", func(t *testing.T) {
			out := theme.Render("Task text", theme.MinWidth, 0, false, "")
			golden.RequireEqual(t, []byte(out))
		})
		t.Run(theme.Name+"/below_min_width", func(t *testing.T) {
			out := theme.Render("Task text", theme.MinWidth-1, 0, false, "")
			golden.RequireEqual(t, []byte(out))
		})
	}
}

// TestGolden_ThemeContentLength tests each theme with short, medium, and long
// content to verify wrapping behavior (AC6).
func TestGolden_ThemeContentLength(t *testing.T) {
	setAsciiProfile(t)

	contentCases := []struct {
		label   string
		content string
	}{
		{"short", "Do it"},
		{"medium", "Review the pull request and leave comments on the architecture decisions"},
		{"long", strings.Join([]string{
			"This is a very long task description that should wrap across",
			"multiple lines to verify that each theme handles content",
			"wrapping gracefully without visual artifacts or broken",
			"borders. The text keeps going to ensure at least five",
			"lines of wrapped content appear in the rendered output.",
		}, " ")},
	}

	for _, theme := range allThemes() {
		for _, cc := range contentCases {
			t.Run(theme.Name+"/"+cc.label, func(t *testing.T) {
				out := theme.Render(cc.content, 40, 0, false, "")
				golden.RequireEqual(t, []byte(out))
			})
		}
	}
}

// TestVisualWidthConsistency verifies that every rendered line of every theme
// has the same visual width within a single door (AC6).
func TestVisualWidthConsistency(t *testing.T) {
	t.Parallel()

	widths := []int{28, 40, 60, 80}

	for _, theme := range allThemes() {
		for _, w := range widths {
			for _, sel := range []bool{false, true} {
				state := "unselected"
				if sel {
					state = "selected"
				}
				name := theme.Name + "/" + state + "_w" + itoa(w)
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					out := theme.Render("Buy groceries for the week", w, 0, sel, "")
					lines := strings.Split(out, "\n")
					if len(lines) == 0 {
						t.Fatal("expected at least one line of output")
					}
					firstWidth := ansi.StringWidth(lines[0])
					for i, line := range lines {
						lw := ansi.StringWidth(line)
						if lw != firstWidth {
							t.Errorf("line %d visual width %d != first line width %d\nline: %q",
								i, lw, firstWidth, line)
						}
					}
				})
			}
		}
	}
}

// TestNoANSIEscapeLeak verifies that no raw ANSI escape codes appear as literal
// text (without a preceding ESC byte) in any theme's output.
func TestNoANSIEscapeLeak(t *testing.T) {
	t.Parallel()

	// Matches CSI sequences that are NOT preceded by ESC (0x1b).
	// A properly formed ANSI escape is \x1b[...m — a "leaked" one has
	// the literal bracket-digits-m without the ESC prefix.
	leakPattern := regexp.MustCompile(`[^\x1b]\[[\d;]+m`)

	for _, theme := range allThemes() {
		for _, w := range []int{28, 40, 60, 80} {
			for _, sel := range []bool{false, true} {
				state := "unselected"
				if sel {
					state = "selected"
				}
				name := theme.Name + "/" + state + "_w" + itoa(w)
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					out := theme.Render("Buy groceries for the week", w, 0, sel, "")
					if leakPattern.MatchString(out) {
						matches := leakPattern.FindAllString(out, -1)
						t.Errorf("found leaked ANSI escape sequences: %q", matches)
					}
				})
			}
		}
	}
}

// itoa converts a small int to its string representation without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
