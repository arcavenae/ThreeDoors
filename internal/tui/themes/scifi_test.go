package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestNewSciFiTheme(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()

	if theme.Name != "scifi" {
		t.Errorf("got name %q, want %q", theme.Name, "scifi")
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

func TestSciFiThemeColors(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()

	if theme.Colors.Frame == nil {
		t.Error("expected non-nil Frame color")
	}
	if theme.Colors.Selected == nil {
		t.Error("expected non-nil Selected color")
	}
}

func TestSciFiRenderContainsContent(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Deploy to staging", 30, 0, false, "")

	if !strings.Contains(output, "Deploy") {
		t.Errorf("output should contain content text, got:\n%s", output)
	}
}

func TestSciFiRenderHasDoubleLineFrame(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Task", 30, 0, false, "")

	for _, ch := range []string{"╔", "╗", "╚", "╝", "═", "║"} {
		if !strings.Contains(output, ch) {
			t.Errorf("sci-fi theme should have double-line character %q", ch)
		}
	}
}

func TestSciFiRenderHasShadeBlocks(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Task", 30, 0, false, "")

	if !strings.Contains(output, "░") {
		t.Error("sci-fi theme should have light shade blocks (░)")
	}
}

func TestSciFiRenderNoMidBar(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Task", 30, 0, false, "")

	// Mid-bar separator was removed — single content panel
	if strings.Contains(output, "╠") || strings.Contains(output, "╣") {
		t.Error("sci-fi theme should not have mid-bar separator (╠ and ╣)")
	}
}

func TestSciFiRenderHasAccessLabel(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Task", 30, 0, false, "")

	if !strings.Contains(output, "ACCESS") {
		t.Errorf("sci-fi theme should have ACCESS label, got:\n%s", output)
	}
}

func TestSciFiRenderSelectedDiffers(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	unselected := theme.Render("Task", 30, 0, false, "")
	selected := theme.Render("Task", 30, 0, true, "")

	if selected == unselected {
		t.Error("selected and unselected output should differ")
	}
	if !strings.Contains(selected, "Task") {
		t.Error("selected output should contain content text")
	}
}

func TestSciFiRenderVaryingWidths(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()

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
			output := theme.Render("Task text", tt.width, 0, false, "")
			if output == "" {
				t.Error("output should not be empty")
			}
			if !strings.Contains(output, "Task") {
				t.Errorf("output should contain task text at width %d", tt.width)
			}
		})
	}
}

func TestSciFiRenderWordWraps(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	longText := "This is a very long task description that should definitely be wrapped across multiple lines"
	output := theme.Render(longText, 30, 0, false, "")

	lines := strings.Split(output, "\n")
	if len(lines) < 7 {
		t.Errorf("long text should produce many lines (with upper/lower panels), got %d lines", len(lines))
	}
}

func TestSciFiRenderUnicodeInAllowedRanges(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Test", 30, 0, false, "")

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

func TestSciFiRenderConsistentLineWidths(t *testing.T) {
	t.Parallel()

	theme := NewSciFiTheme()
	output := theme.Render("Task", 30, 0, false, "")
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
