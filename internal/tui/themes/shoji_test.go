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
		return
	}
	if theme.MinWidth < 1 {
		t.Errorf("expected positive MinWidth, got %d", theme.MinWidth)
	}
}

func TestShojiThemeColors(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()

	if theme.Colors.Frame == nil {
		t.Error("expected non-nil Frame color")
	}
	if theme.Colors.Selected == nil {
		t.Error("expected non-nil Selected color")
	}
}

func TestShojiRenderContainsContent(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Clean up backlog", 30, 0, false, "", 0.0)

	if !strings.Contains(output, "Clean up") {
		t.Errorf("output should contain content text, got:\n%s", output)
	}
}

func TestShojiRenderHasLatticePattern(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, 0, false, "", 0.0)

	// Single cross junction on mid-cross bar (AC3)
	if !strings.Contains(output, "┼") {
		t.Error("shoji theme should have a cross junction (┼) on the mid-cross bar")
	}
	if !strings.Contains(output, "─") {
		t.Error("shoji theme should have horizontal lines (─)")
	}
	if !strings.Contains(output, "│") {
		t.Error("shoji theme should have vertical lines (│)")
	}
}

func TestShojiRenderFewLatticeColumns(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, 0, false, "", 0.0)
	lines := strings.Split(output, "\n")

	// AC1: no more than 3-4 columns — with the new design the side frame
	// is a single │ on each side, no internal vertical subdivisions.
	// Cell rows should have exactly 2 vertical bars (left + right frame).
	for i, line := range lines {
		stripped := stripANSI(line)
		vCount := strings.Count(stripped, "│")
		// Lattice bars (├...┤) and top/bottom rails (┬/┴) have 0 │ chars
		// Content/empty rows have exactly 2
		if vCount != 0 && vCount != 2 {
			t.Errorf("line %d has %d vertical bars, expected 0 or 2: %q", i, vCount, stripped)
		}
	}
}

func TestShojiRenderHasTopBottomRails(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, 0, false, "", 0.0)
	lines := strings.Split(output, "\n")

	if len(lines) < 3 {
		t.Fatal("expected at least 3 lines")
	}

	if !strings.Contains(lines[0], "┬") {
		t.Errorf("top rail should contain ┬, got: %q", lines[0])
	}
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "┴") {
		t.Errorf("bottom rail should contain ┴, got: %q", lastLine)
	}
}

func TestShojiRenderHasEdgeConnectors(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 30, 0, false, "", 0.0)

	if !strings.Contains(output, "├") {
		t.Error("shoji theme should have left edge connectors (├)")
	}
	if !strings.Contains(output, "┤") {
		t.Error("shoji theme should have right edge connectors (┤)")
	}
}

func TestShojiRenderSelectedDiffers(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	unselected := theme.Render("Task", 30, 0, false, "", 0.0)
	selected := theme.Render("Task", 30, 0, true, "", 0.0)

	if selected == unselected {
		t.Error("selected and unselected output should differ")
	}
	if !strings.Contains(selected, "Task") {
		t.Error("selected output should contain content text")
	}
}

func TestShojiRenderSelectedHasHeavyChars(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	selected := theme.Render("Task", 30, 0, true, "", 0.0)

	// AC4: selected state uses heavy characters
	if !strings.Contains(selected, "━") {
		t.Error("selected shoji should use heavy horizontal (━)")
	}
	if !strings.Contains(selected, "┃") {
		t.Error("selected shoji should use heavy vertical (┃)")
	}
	if !strings.Contains(selected, "╋") {
		t.Error("selected shoji should use heavy cross (╋)")
	}
	if !strings.Contains(selected, "┳") {
		t.Error("selected shoji should use heavy top T (┳)")
	}
	if !strings.Contains(selected, "┻") {
		t.Error("selected shoji should use heavy bottom T (┻)")
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
			output := theme.Render("Task text", tt.width, 0, false, "", 0.0)
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
	output := theme.Render(longText, 30, 0, false, "", 0.0)

	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Errorf("long text should wrap to multiple lines, got %d lines", len(lines))
	}
}

func TestShojiRenderUnicodeInAllowedRanges(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Test", 30, 0, false, "", 0.0)

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
	output := theme.Render("Task", 30, 0, false, "", 0.0)
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

func TestShojiRenderContentWidth(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()

	// AC5: minimum 15 chars of usable text width at MinWidth
	output := theme.Render("exactly15charss", theme.MinWidth, 0, false, "", 0.0)
	if !strings.Contains(output, "exactly15charss") {
		t.Errorf("at MinWidth %d, should fit 15+ chars of text, got:\n%s", theme.MinWidth, output)
	}
}

func TestShojiRenderContentDominates(t *testing.T) {
	t.Parallel()

	theme := NewShojiTheme()
	output := theme.Render("Task", 28, 0, false, "", 0.0)
	lines := strings.Split(output, "\n")

	// AC2: content-to-decoration ratio favors content
	// Count content rows (│ ... │ with spaces) vs decoration rows (├─┤, ┬─┬, etc.)
	contentRows := 0
	decoRows := 0
	for _, line := range lines {
		stripped := stripANSI(line)
		if strings.Contains(stripped, "├") || strings.Contains(stripped, "┬") || strings.Contains(stripped, "┴") {
			decoRows++
		} else if strings.Contains(stripped, "│") {
			contentRows++
		}
	}
	if decoRows >= contentRows {
		t.Errorf("decoration rows (%d) should be fewer than content rows (%d)", decoRows, contentRows)
	}
}

// stripANSI removes ANSI escape sequences for testing character counts.
func stripANSI(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
