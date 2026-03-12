package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/muesli/termenv"
)

// TestShojiDoorProportions verifies the Shoji theme renders door-like
// elements at proper DoorAnatomy heights.
func TestShojiDoorProportions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int
	}{
		{"min_height_14", 14},
		{"standard_16", 16},
		{"tall_24", 24},
	}

	theme := NewShojiTheme()

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

			// Check panel divider (cross bar) contains cross junction
			if anatomy.PanelDivider < len(lines) {
				dividerLine := lines[anatomy.PanelDivider]
				if !strings.Contains(dividerLine, "┼") && !strings.Contains(dividerLine, "╋") {
					t.Errorf("panel divider at row %d should contain cross junction, got: %q",
						anatomy.PanelDivider, dividerLine)
				}
			}

			// Check handle row contains ○
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

			// Check top rail
			topLine := lines[anatomy.LintelRow]
			if !strings.Contains(topLine, "┬") && !strings.Contains(topLine, "┳") {
				t.Errorf("top rail should contain T-junction, got: %q", topLine)
			}

			// Check bottom rail
			botLine := lines[anatomy.ThresholdRow]
			if !strings.Contains(botLine, "┴") && !strings.Contains(botLine, "┻") {
				t.Errorf("bottom rail should contain T-junction, got: %q", botLine)
			}
		})
	}
}

// TestShojiDoorCompactFallback verifies that height < MinHeight uses compact layout.
func TestShojiDoorCompactFallback(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()

	// height=0 (compact) should NOT contain door elements like threshold ▔
	output := theme.Render("Task text", 30, 0, false, "", 0.0)
	if strings.Contains(output, "▔") {
		t.Error("compact mode should not contain threshold line ▔")
	}
	if strings.Contains(output, "○") {
		t.Error("compact mode should not contain handle ○")
	}

	// Should still contain the shoji lattice elements
	if !strings.Contains(output, "┼") {
		t.Error("compact mode should contain cross junction ┼")
	}
	if !strings.Contains(output, "Task text") {
		t.Error("compact mode should contain content text")
	}
}

// TestShojiDoorSelectedProportions verifies selected state in door mode.
func TestShojiDoorSelectedProportions(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, 16, true, "", 0.0)
	lines := strings.Split(output, "\n")

	anatomy := NewDoorAnatomy(16)

	// Selected cross bar uses heavy cross ╋
	dividerLine := lines[anatomy.PanelDivider]
	if !strings.Contains(dividerLine, "╋") {
		t.Errorf("selected panel divider should contain ╋, got: %q", dividerLine)
	}

	// Selected top rail uses heavy T ┳
	topLine := lines[anatomy.LintelRow]
	if !strings.Contains(topLine, "┳") {
		t.Errorf("selected top rail should contain ┳, got: %q", topLine)
	}

	// Selected bottom rail uses heavy T ┻
	botLine := lines[anatomy.ThresholdRow]
	if !strings.Contains(botLine, "┻") {
		t.Errorf("selected bottom rail should contain ┻, got: %q", botLine)
	}
}

// TestShojiDoorContentPlacement verifies content starts at ContentStart row.
func TestShojiDoorContentPlacement(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Unique marker text", 40, 16, false, "", 0.0)
	lines := strings.Split(output, "\n")

	anatomy := NewDoorAnatomy(16)

	// Content should be at or after ContentStart
	found := false
	for i := anatomy.ContentStart; i < anatomy.PanelDivider && i < len(lines); i++ {
		if strings.Contains(lines[i], "Unique marker") {
			found = true
			break
		}
	}
	if !found {
		t.Error("content text should appear between ContentStart and PanelDivider")
	}
}

// TestShojiDoorConsistentLineWidths verifies all lines have same visual width.
func TestShojiDoorConsistentLineWidths(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()

	tests := []struct {
		name     string
		width    int
		height   int
		selected bool
	}{
		{"h16_unselected", 30, 16, false},
		{"h16_selected", 30, 16, true},
		{"h24_unselected", 30, 24, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := theme.Render("Task", tt.width, tt.height, tt.selected, "", 0.0)
			lines := strings.Split(output, "\n")

			if len(lines) < 3 {
				t.Fatal("expected at least 3 lines")
			}

			// All lines except threshold should have same width as the door
			doorWidth := ansi.StringWidth(lines[0])
			for i := 0; i < len(lines)-1; i++ {
				w := ansi.StringWidth(lines[i])
				if w != doorWidth {
					t.Errorf("line %d width %d != first line width %d\nline: %q",
						i, w, doorWidth, lines[i])
				}
			}
		})
	}
}

// TestShojiDoorGolden tests golden file output for door-proportioned Shoji.
func TestShojiDoorGolden(t *testing.T) {
	t.Parallel()
	lipgloss.SetColorProfile(termenv.Ascii)

	tests := []struct {
		name     string
		content  string
		width    int
		height   int
		selected bool
	}{
		{"h16_w30_unselected", "Buy groceries", 30, 16, false},
		{"h16_w30_selected", "Buy groceries", 30, 16, true},
		{"h24_w30_unselected", "Buy groceries", 30, 24, false},
		{"h24_w30_selected", "Buy groceries", 30, 24, true},
		{"compact_w30", "Buy groceries", 30, 0, false},
	}

	theme := NewShojiTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := theme.Render(tt.content, tt.width, tt.height, tt.selected, "", 0.0)
			golden.RequireEqual(t, []byte(output))
		})
	}
}
