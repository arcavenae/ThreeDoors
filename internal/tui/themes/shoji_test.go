package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestNewShojiTheme(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()

	if theme.Name != "shoji" {
		t.Errorf("got name %q, want %q", theme.Name, "shoji")
	}
	if theme.Description == "" {
		t.Error("expected non-empty description")
	}
	if theme.Render == nil {
		t.Fatal("expected non-nil Render function")
	}
	if theme.MinWidth < 1 {
		t.Errorf("expected positive MinWidth, got %d", theme.MinWidth)
	}
}

func TestShojiThemeColors(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()

	if theme.Colors.Frame == "" {
		t.Error("expected non-empty Frame color")
	}
	if theme.Colors.Selected == "" {
		t.Error("expected non-empty Selected color")
	}
}

func TestShojiRenderContainsContent(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Clean up backlog", 30, false)

	if !strings.Contains(output, "Clean up") {
		t.Errorf("output should contain content text, got:\n%s", output)
	}
}

func TestShojiRenderHasGridPattern(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, false)

	// Cross junctions for lattice grid
	if !strings.Contains(output, "┼") {
		t.Error("shoji theme should have cross junctions (┼)")
	}
	// Horizontal and vertical lines
	if !strings.Contains(output, "─") {
		t.Error("shoji theme should have horizontal lines (─)")
	}
	if !strings.Contains(output, "│") {
		t.Error("shoji theme should have vertical lines (│)")
	}
}

func TestShojiRenderHasTopBottomRails(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, false)
	lines := strings.Split(output, "\n")

	if len(lines) < 3 {
		t.Fatal("expected at least 3 lines")
	}

	// Top rail should have ┬ characters (unselected)
	if !strings.Contains(lines[0], "┬") {
		t.Errorf("top rail should contain ┬, got: %q", lines[0])
	}
	// Bottom rail should have ┴ characters (unselected)
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "┴") {
		t.Errorf("bottom rail should contain ┴, got: %q", lastLine)
	}
}

func TestShojiRenderHasEdgeConnectors(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, false)

	// Left edge connectors
	if !strings.Contains(output, "├") {
		t.Error("shoji theme should have left edge connectors (├)")
	}
	// Right edge connectors
	if !strings.Contains(output, "┤") {
		t.Error("shoji theme should have right edge connectors (┤)")
	}
}

func TestShojiRenderSelectedDiffers(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	unselected := theme.Render("Task", 30, false)
	selected := theme.Render("Task", 30, true)

	if selected == unselected {
		t.Error("selected and unselected output should differ")
	}
	if !strings.Contains(selected, "Task") {
		t.Error("selected output should contain content text")
	}
}

func TestShojiRenderVaryingWidths(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()

	tests := []struct {
		name  string
		width int
	}{
		{"min_width", theme.MinWidth},
		{"medium", 30},
		{"standard", 40},
		{"wide", 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := theme.Render("Task text", tt.width, false)
			if output == "" {
				t.Error("output should not be empty")
			}
			if !strings.Contains(output, "Task") {
				t.Errorf("output should contain task text at width %d", tt.width)
			}
		})
	}
}

func TestShojiRenderWordWraps(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	longText := "This is a very long task description that should definitely be wrapped across multiple lines"
	output := theme.Render(longText, 30, false)

	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Errorf("long text should wrap to multiple lines, got %d lines", len(lines))
	}
}

func TestShojiRenderUnicodeInAllowedRanges(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Test", 30, false)

	for _, r := range output {
		if r <= 0x7F || r == '\n' {
			continue
		}
		if (r >= 0x2500 && r <= 0x257F) ||
			(r >= 0x2580 && r <= 0x259F) ||
			(r >= 0x25A0 && r <= 0x25FF) {
			continue
		}
		t.Errorf("character %q (U+%04X) is outside allowed Unicode ranges", string(r), r)
	}
}

func TestShojiRenderConsistentLineWidths(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, false)
	lines := strings.Split(output, "\n")

	if len(lines) < 3 {
		t.Fatal("expected at least 3 lines of output")
	}

	firstWidth := ansi.StringWidth(lines[0])
	for i, line := range lines {
		w := ansi.StringWidth(line)
		if w != firstWidth {
			t.Errorf("line %d width %d != first line width %d\nline: %q", i, w, firstWidth, line)
		}
	}
}
