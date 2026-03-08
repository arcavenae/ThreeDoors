package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

// TestSciFiDoorProportions verifies the Sci-Fi theme renders door-like
// elements (bulkhead divider, handle, floor grating) at proper heights.
func TestSciFiDoorProportions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int
	}{
		{"min_height_14", 14},
		{"standard_16", 16},
		{"tall_24", 24},
	}

	theme := NewSciFiTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := theme.Render("Deploy to staging", 30, tt.height, false)
			lines := strings.Split(output, "\n")

			// Should have height rows + 1 floor grating line
			if len(lines) != tt.height+1 {
				t.Errorf("expected %d lines (height + grating), got %d", tt.height+1, len(lines))
			}

			anatomy := NewDoorAnatomy(tt.height)

			// Check bulkhead divider contains ╞ and ╡
			if anatomy.PanelDivider < len(lines) {
				dividerLine := lines[anatomy.PanelDivider]
				if !strings.Contains(dividerLine, "╞") {
					t.Errorf("bulkhead divider at row %d should contain ╞, got: %q",
						anatomy.PanelDivider, dividerLine)
				}
				if !strings.Contains(dividerLine, "╡") {
					t.Errorf("bulkhead divider at row %d should contain ╡, got: %q",
						anatomy.PanelDivider, dividerLine)
				}
			}

			// Check handle row contains ◈
			if anatomy.HandleRow < len(lines) {
				handleLine := lines[anatomy.HandleRow]
				if !strings.Contains(handleLine, "◈") {
					t.Errorf("handle row at %d should contain ◈, got: %q",
						anatomy.HandleRow, handleLine)
				}
			}

			// Check ACCESS label exists in lower panel area
			foundAccess := false
			for i := anatomy.HandleRow; i < anatomy.ThresholdRow; i++ {
				if i < len(lines) && strings.Contains(lines[i], "ACCESS") {
					foundAccess = true
					break
				}
			}
			if !foundAccess {
				t.Error("expected [ACCESS] label in lower panel area")
			}

			// Check floor grating (last line) contains ▓
			lastLine := lines[len(lines)-1]
			if !strings.Contains(lastLine, "▓") {
				t.Errorf("floor grating (last line) should contain ▓, got: %q", lastLine)
			}

			// Check shade rails on sides
			for i := 1; i < anatomy.ThresholdRow; i++ {
				if i == anatomy.PanelDivider {
					continue // divider row has different structure
				}
				if i < len(lines) && !strings.Contains(lines[i], "░") {
					t.Errorf("row %d should contain shade rail ░, got: %q", i, lines[i])
				}
			}
		})
	}
}

// TestSciFiDoorCompactFallback verifies that height < MinHeight uses compact card style.
func TestSciFiDoorCompactFallback(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()

	// height=0 (compact) should NOT contain door-like elements
	output := theme.Render("Task text", 30, 0, false)

	if strings.Contains(output, "◈") {
		t.Error("compact mode should not contain access handle ◈")
	}
	if strings.Contains(output, "╞") {
		t.Error("compact mode should not contain bulkhead divider ╞")
	}

	// But should still contain basic Sci-Fi elements
	if !strings.Contains(output, "╔") {
		t.Error("compact mode should still have double-line frame ╔")
	}
	if !strings.Contains(output, "ACCESS") {
		t.Error("compact mode should still have ACCESS label")
	}
}

// TestSciFiDoorSelectedVsUnselected verifies selected doors use bright shade.
func TestSciFiDoorSelectedVsUnselected(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()

	unselected := theme.Render("Task", 30, 16, false)
	selected := theme.Render("Task", 30, 16, true)

	if unselected == selected {
		t.Error("selected and unselected door-mode output should differ")
	}

	// Unselected uses light shade
	if !strings.Contains(unselected, "░") {
		t.Error("unselected should use light shade ░")
	}

	// Selected uses bright shade
	unselectedLines := strings.Split(unselected, "\n")
	selectedLines := strings.Split(selected, "\n")

	// Check that interior rows of selected use ▓ for rails
	for i := 1; i < len(selectedLines)-2; i++ {
		line := selectedLines[i]
		if strings.Contains(line, "╞") || strings.Contains(line, "╡") {
			continue // divider row
		}
		if strings.Contains(line, "║") && !strings.Contains(line, "▓") {
			t.Errorf("selected row %d should use bright shade ▓, got: %q", i, line)
		}
	}

	// Both should have same number of lines
	if len(unselectedLines) != len(selectedLines) {
		t.Errorf("selected (%d lines) and unselected (%d lines) should have same line count",
			len(selectedLines), len(unselectedLines))
	}
}

// TestSciFiDoorVisualWidth verifies all lines have consistent visual width.
func TestSciFiDoorVisualWidth(t *testing.T) {
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
		{"narrow_unselected", 20, 14, false},
	}

	theme := NewSciFiTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := theme.Render("Buy groceries for the week", tt.width, tt.height, tt.selected)
			lines := strings.Split(output, "\n")

			if len(lines) < 3 {
				t.Fatal("expected at least 3 lines")
			}

			firstWidth := ansi.StringWidth(lines[0])
			for i, line := range lines {
				w := ansi.StringWidth(line)
				if w != firstWidth {
					t.Errorf("line %d width %d != first line width %d\nline: %q",
						i, w, firstWidth, line)
				}
			}
		})
	}
}

// TestSciFiDoorContentPlacement verifies content is placed in upper panel.
func TestSciFiDoorContentPlacement(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Deploy to staging", 30, 16, false)
	lines := strings.Split(output, "\n")

	anatomy := NewDoorAnatomy(16)

	// Content should appear between ContentStart and PanelDivider
	foundContent := false
	for i := anatomy.ContentStart; i < anatomy.PanelDivider; i++ {
		if i < len(lines) && strings.Contains(lines[i], "Deploy") {
			foundContent = true
			break
		}
	}
	if !foundContent {
		t.Error("content should be placed in upper panel (between ContentStart and PanelDivider)")
	}

	// Content should NOT appear below the divider
	for i := anatomy.PanelDivider; i < len(lines); i++ {
		if strings.Contains(lines[i], "Deploy") {
			t.Errorf("content should not appear below panel divider, found at row %d", i)
		}
	}
}

// TestGolden_SciFiDoorRender tests golden file output for door-like rendering.
func TestGolden_SciFiDoorRender(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	tests := []struct {
		name     string
		width    int
		height   int
		selected bool
	}{
		{"h14_w30_unselected", 30, 14, false},
		{"h14_w30_selected", 30, 14, true},
		{"h16_w30_unselected", 30, 16, false},
		{"h16_w30_selected", 30, 16, true},
		{"h24_w40_unselected", 40, 24, false},
		{"h24_w40_selected", 40, 24, true},
	}

	theme := NewSciFiTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := theme.Render("Buy groceries for the week", tt.width, tt.height, tt.selected)
			golden.RequireEqual(t, []byte(out))
		})
	}
}
