package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

// TestClassicDoorProportions verifies the Classic theme renders door-like
// elements (panel divider, doorknob, threshold) at proper heights.
func TestClassicDoorProportions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int
	}{
		{"min_height_10", 10},
		{"standard_16", 16},
		{"tall_24", 24},
	}

	theme := NewClassicTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := theme.Render("Buy groceries", 30, tt.height, false, "")
			lines := strings.Split(output, "\n")

			// Should have height rows + 1 threshold line + 1 shadow bottom
			if len(lines) != tt.height+2 {
				t.Errorf("expected %d lines (height + threshold + shadow), got %d", tt.height+2, len(lines))
			}

			anatomy := NewDoorAnatomy(tt.height)

			// Check panel divider contains horizontal bar characters
			if anatomy.PanelDivider < len(lines) {
				dividerLine := lines[anatomy.PanelDivider]
				if !strings.Contains(dividerLine, "─") && !strings.Contains(dividerLine, "━") {
					t.Errorf("panel divider at row %d should contain horizontal bars, got: %q",
						anatomy.PanelDivider, dividerLine)
				}
				// Hinge junction: ╟ (unselected) or ┣ (selected)
				if !strings.Contains(dividerLine, "╟") && !strings.Contains(dividerLine, "┣") {
					t.Errorf("panel divider should have hinge junction character (╟ or ┣), got: %q", dividerLine)
				}
			}

			// Check doorknob row contains ●
			if anatomy.HandleRow < len(lines) {
				handleLine := lines[anatomy.HandleRow]
				if !strings.Contains(handleLine, "●") {
					t.Errorf("handle row at %d should contain ●, got: %q",
						anatomy.HandleRow, handleLine)
				}
			}

			// Check threshold line (second-to-last; last is shadow bottom) contains ▔
			threshLine := lines[len(lines)-2]
			if !strings.Contains(threshLine, "▔") {
				t.Errorf("threshold should contain ▔, got: %q", threshLine)
			}

			// Check shadow bottom row contains ▄
			shadowLine := lines[len(lines)-1]
			if !strings.Contains(shadowLine, "▄") {
				t.Errorf("shadow bottom should contain ▄, got: %q", shadowLine)
			}
		})
	}
}

// TestClassicDoorCompactFallback verifies that height < MinHeight uses compact card style.
func TestClassicDoorCompactFallback(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	// height=0 (compact) should NOT contain door elements
	output := theme.Render("Task text", 30, 0, false, "")

	if strings.Contains(output, "●") {
		t.Error("compact mode should not contain doorknob ●")
	}
	if strings.Contains(output, "▔") {
		t.Error("compact mode should not contain threshold ▔")
	}
	if strings.Contains(output, "├") {
		t.Error("compact mode should not contain panel divider ├")
	}
}

// TestClassicDoorSelectedVsUnselected verifies that selected doors use different borders.
func TestClassicDoorSelectedVsUnselected(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	unselected := theme.Render("Task", 30, 16, false, "")
	selected := theme.Render("Task", 30, 16, true, "")

	if unselected == selected {
		t.Error("selected and unselected door-mode output should differ")
	}

	// Unselected uses double-line hinge on left (╓)
	if !strings.Contains(unselected, "╓") {
		t.Error("unselected should use double-line hinge corner ╓")
	}

	// Selected uses heavy hinge on left (┏)
	if !strings.Contains(selected, "┏") {
		t.Error("selected should use heavy hinge corner ┏")
	}

	// Opening side (right) uses standard weight
	if !strings.Contains(unselected, "┐") {
		t.Error("unselected should use standard right corner ┐")
	}
	if !strings.Contains(selected, "┐") {
		t.Error("selected should use standard right corner ┐ (lighter than hinge)")
	}

	// Selected panel divider uses heavy hinge junction
	if !strings.Contains(selected, "┣") {
		t.Error("selected panel divider should use heavy hinge junction ┣")
	}
}

// TestClassicDoorVisualWidth verifies all lines have consistent visual width.
func TestClassicDoorVisualWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		width    int
		height   int
		selected bool
	}{
		{"standard_unselected", 30, 16, false},
		{"standard_selected", 30, 16, true},
		{"wide_unselected", 40, 24, false},
		{"narrow_unselected", 20, 10, false},
	}

	theme := NewClassicTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := theme.Render("Buy groceries for the week", tt.width, tt.height, tt.selected, "")
			lines := strings.Split(output, "\n")

			if len(lines) < 2 {
				t.Fatal("expected at least 2 lines")
			}

			// All lines in the door body (not threshold) should have same width
			firstWidth := ansi.StringWidth(lines[0])
			for i := 0; i < len(lines)-1; i++ { // exclude threshold
				lw := ansi.StringWidth(lines[i])
				if lw != firstWidth {
					t.Errorf("line %d visual width %d != first line width %d\nline: %q",
						i, lw, firstWidth, lines[i])
				}
			}

			// Threshold line should also match width
			threshWidth := ansi.StringWidth(lines[len(lines)-1])
			if threshWidth != firstWidth {
				t.Errorf("threshold width %d != door width %d", threshWidth, firstWidth)
			}
		})
	}
}

// TestGolden_ClassicDoorHeight tests golden file output at standard and tall heights.
func TestGolden_ClassicDoorHeight(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	theme := NewClassicTheme()

	tests := []struct {
		name     string
		width    int
		height   int
		selected bool
	}{
		{"h16_unselected_w30", 30, 16, false},
		{"h16_selected_w30", 30, 16, true},
		{"h24_unselected_w30", 30, 24, false},
		{"h24_selected_w30", 30, 24, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := theme.Render("Buy groceries for the week", tt.width, tt.height, tt.selected, "")
			golden.RequireEqual(t, []byte(out))
		})
	}
}
