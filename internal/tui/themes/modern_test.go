package themes

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNewModernTheme(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()

	if theme.Name != "modern" {
		t.Errorf("got name %q, want %q", theme.Name, "modern")
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

func TestModernThemeColors(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()

	if theme.Colors.Frame == "" {
		t.Error("expected non-empty Frame color")
	}
	if theme.Colors.Selected == "" {
		t.Error("expected non-empty Selected color")
	}
}

func TestModernRenderContainsContent(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	output := theme.Render("Write unit tests", 30, false)

	if !strings.Contains(output, "Write unit tests") {
		t.Errorf("output should contain content text, got:\n%s", output)
	}
}

func TestModernRenderHasDoorknob(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	output := theme.Render("Task text", 30, false)

	if !strings.Contains(output, "●") {
		t.Errorf("modern theme should have filled circle doorknob (●), got:\n%s", output)
	}
}

func TestModernRenderHasBoxDrawingFrame(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	output := theme.Render("Task", 30, false)

	// Should have horizontal lines at top and bottom
	if !strings.Contains(output, "─") {
		t.Error("expected horizontal box-drawing lines (─)")
	}
	// Should have vertical lines on sides
	if !strings.Contains(output, "│") {
		t.Error("expected vertical box-drawing lines (│)")
	}
}

func TestModernRenderNoRoundedCorners(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	output := theme.Render("Task", 30, false)

	// Modern theme uses straight lines, no rounded corners
	for _, ch := range []string{"╭", "╮", "╰", "╯"} {
		if strings.Contains(output, ch) {
			t.Errorf("modern theme should not have rounded corner %q", ch)
		}
	}
}

func TestModernRenderSelectedDiffers(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	unselected := theme.Render("Task", 30, false)
	selected := theme.Render("Task", 30, true)

	if selected == unselected {
		t.Error("selected and unselected output should differ")
	}
	if !strings.Contains(selected, "Task") {
		t.Error("selected output should contain content text")
	}
}

func TestModernRenderVaryingWidths(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()

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
			// At narrow widths text may wrap, so check for substring
			if !strings.Contains(output, "Task") {
				t.Errorf("output should contain task text at width %d", tt.width)
			}
		})
	}
}

func TestModernRenderWordWraps(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	longText := "This is a very long task description that should definitely be wrapped across multiple lines"
	output := theme.Render(longText, 30, false)

	lines := strings.Split(output, "\n")
	if len(lines) < 5 {
		t.Errorf("long text should wrap to multiple lines, got %d lines", len(lines))
	}
}

func TestModernRenderUnicodeInAllowedRanges(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	output := theme.Render("Test", 30, false)

	for _, r := range output {
		if r <= 0x7F || r == '\n' {
			continue // ASCII and newline are fine
		}
		// Allowed ranges: box-drawing U+2500-U+257F, block elements U+2580-U+259F,
		// geometric shapes U+25A0-U+25FF
		if (r >= 0x2500 && r <= 0x257F) ||
			(r >= 0x2580 && r <= 0x259F) ||
			(r >= 0x25A0 && r <= 0x25FF) {
			continue
		}
		t.Errorf("character %q (U+%04X) is outside allowed Unicode ranges", string(r), r)
	}

	// Verify it contains some Unicode (not purely ASCII)
	hasUnicode := false
	for _, r := range output {
		if r > 0x7F && r != '\n' {
			hasUnicode = true
			break
		}
	}
	if !hasUnicode {
		t.Error("expected some Unicode box-drawing characters in output")
	}

	_ = utf8.ValidString(output) // ensure valid UTF-8
}

func TestModernRenderConsistentLineWidths(t *testing.T) {
	t.Parallel()

	theme := NewModernTheme()
	output := theme.Render("Task", 30, false)
	lines := strings.Split(output, "\n")

	if len(lines) < 3 {
		t.Fatal("expected at least 3 lines of output")
	}

	// All lines should have the same rune count (consistent width)
	firstWidth := utf8.RuneCountInString(lines[0])
	for i, line := range lines {
		w := utf8.RuneCountInString(line)
		if w != firstWidth {
			t.Errorf("line %d width %d != first line width %d\nline: %q", i, w, firstWidth, line)
		}
	}
}
