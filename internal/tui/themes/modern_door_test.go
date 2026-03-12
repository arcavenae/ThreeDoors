package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

// TestModernDoorProportions verifies the Modern theme renders door-like
// elements (panel divider, handle, threshold) at proper heights.
func TestModernDoorProportions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int
	}{
		{"min_height_12", 12},
		{"standard_16", 16},
		{"tall_24", 24},
	}

	theme := NewModernTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := theme.Render("Buy groceries", 30, tt.height, false, "", 0.0)
			lines := strings.Split(output, "\n")

			// Should have height rows + 1 threshold line + 1 shadow bottom
			if len(lines) != tt.height+2 {
				t.Errorf("expected %d lines (height + threshold + shadow), got %d", tt.height+2, len(lines))
			}

			anatomy := NewDoorAnatomy(tt.height)

			// Check panel divider contains thin horizontal bar
			if anatomy.PanelDivider < len(lines) {
				dividerLine := lines[anatomy.PanelDivider]
				if !strings.Contains(dividerLine, "─") {
					t.Errorf("panel divider at row %d should contain thin horizontal bars ─, got: %q",
						anatomy.PanelDivider, dividerLine)
				}
			}

			// Check handle row contains ○ (open circle)
			if anatomy.HandleRow < len(lines) {
				handleLine := lines[anatomy.HandleRow]
				if !strings.Contains(handleLine, "○") {
					t.Errorf("handle row at %d should contain ○, got: %q",
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

// TestModernDoorCompactFallback verifies that height < MinHeight uses compact card style.
func TestModernDoorCompactFallback(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()

	// height=0 (compact) should NOT contain door-mode elements
	output := theme.Render("Task text", 30, 0, false, "", 0.0)

	if strings.Contains(output, "○") {
		t.Error("compact mode should not contain minimalist handle ○")
	}
	if strings.Contains(output, "▔") {
		t.Error("compact mode should not contain threshold ▔")
	}

	// Compact mode should still have the filled doorknob ●
	if !strings.Contains(output, "●") {
		t.Error("compact mode should contain filled doorknob ●")
	}
}

// TestModernDoorSelectedVsUnselected verifies that selected doors use heavier frame.
func TestModernDoorSelectedVsUnselected(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()

	unselected := theme.Render("Task", 30, 16, false, "", 0.0)
	selected := theme.Render("Task", 30, 16, true, "", 0.0)

	if unselected == selected {
		t.Error("selected and unselected door-mode output should differ")
	}

	// Unselected uses thin vertical lines
	if !strings.Contains(unselected, "│") {
		t.Error("unselected should use thin vertical line │")
	}

	// Selected uses heavy vertical lines
	if !strings.Contains(selected, "┃") {
		t.Error("selected should use heavy vertical line ┃")
	}

	// Both should have thin panel divider (minimalist)
	anatomy := NewDoorAnatomy(16)
	unselectedLines := strings.Split(unselected, "\n")
	selectedLines := strings.Split(selected, "\n")

	if anatomy.PanelDivider < len(unselectedLines) {
		if !strings.Contains(unselectedLines[anatomy.PanelDivider], "─") {
			t.Error("unselected panel divider should contain thin horizontal ─")
		}
	}
	if anatomy.PanelDivider < len(selectedLines) {
		if !strings.Contains(selectedLines[anatomy.PanelDivider], "─") {
			t.Error("selected panel divider should still contain thin horizontal ─ (minimalist)")
		}
	}
}

// TestModernDoorHingeAsymmetry verifies Modern doors have hinge marks on left,
// standard on right. The right side keeps minimalist no-corner style.
func TestModernDoorHingeAsymmetry(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()

	// Unselected: left hinge ╓/║/╙, right stays minimal (no corners)
	unselected := theme.Render("Task", 30, 16, false, "", 0.0)
	if !strings.Contains(unselected, "╓") {
		t.Error("unselected modern door should have hinge corner ╓ on left")
	}
	if !strings.Contains(unselected, "║") {
		t.Error("unselected modern door should have hinge vertical ║ on left")
	}

	// Right side should NOT have heavy corners
	for _, ch := range []string{"╮", "╯", "┓", "┛"} {
		if strings.Contains(unselected, ch) {
			t.Errorf("modern door right side should not have corner %q", ch)
		}
	}

	// Selected: left hinge ┏/┃/┗, right stays standard │
	selected := theme.Render("Task", 30, 16, true, "", 0.0)
	if !strings.Contains(selected, "┏") {
		t.Error("selected modern door should have heavy hinge corner ┏ on left")
	}
	if !strings.Contains(selected, "┃") {
		t.Error("selected modern door should have heavy hinge vertical ┃ on left")
	}
}

// TestModernDoorVisualWidth verifies all lines have consistent visual width.
func TestModernDoorVisualWidth(t *testing.T) {
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
		{"narrow_unselected", 20, 12, false},
	}

	theme := NewModernTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := theme.Render("Buy groceries for the week", tt.width, tt.height, tt.selected, "", 0.0)
			lines := strings.Split(output, "\n")

			if len(lines) < 2 {
				t.Fatal("expected at least 2 lines")
			}

			// All lines in the door body (not threshold) should have same width
			firstWidth := ansi.StringWidth(lines[0])
			for i := 0; i < len(lines)-1; i++ {
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

// TestModernDoorContentPresent verifies content text appears in door-mode output.
func TestModernDoorContentPresent(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	output := theme.Render("Write unit tests", 30, 16, false, "", 0.0)

	if !strings.Contains(output, "Write unit tests") {
		t.Errorf("door-mode output should contain content text, got:\n%s", output)
	}
}

// TestGolden_ModernDoorHeight tests golden file output at standard and tall heights.
func TestGolden_ModernDoorHeight(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	theme := NewModernTheme()

	tests := []struct {
		name     string
		width    int
		height   int
		selected bool
	}{
		{"h12_unselected_w30", 30, 12, false},
		{"h12_selected_w30", 30, 12, true},
		{"h16_unselected_w30", 30, 16, false},
		{"h16_selected_w30", 30, 16, true},
		{"h24_unselected_w30", 30, 24, false},
		{"h24_selected_w30", 30, 24, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := theme.Render("Buy groceries for the week", tt.width, tt.height, tt.selected, "", 0.0)
			golden.RequireEqual(t, []byte(out))
		})
	}
}
