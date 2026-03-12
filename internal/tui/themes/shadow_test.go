package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestApplyShadow_AddsRightAndBottomShadow(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	input := "AAA\nBBB\nCCC"
	result := ApplyShadow(input, 10, 5, false)
	lines := strings.Split(result, "\n")

	// Should have original 3 lines + 1 shadow bottom
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	// First line: no right shadow (offset effect), trailing space
	if strings.Contains(lines[0], shadowRight) {
		t.Error("first line should not have right shadow")
	}

	// Lines 1-2: should have right shadow ▐
	for i := 1; i < 3; i++ {
		if !strings.Contains(lines[i], shadowRight) {
			t.Errorf("line %d should contain right shadow %q", i, shadowRight)
		}
	}

	// Last line: bottom shadow with ▄
	if !strings.Contains(lines[3], shadowBottom) {
		t.Errorf("bottom shadow should contain %q, got: %q", shadowBottom, lines[3])
	}

	// Bottom shadow should be offset by 1 space
	if !strings.HasPrefix(lines[3], " ") {
		t.Error("bottom shadow should start with space (offset)")
	}
}

func TestApplyShadow_EnhancedForSelected(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	input := "AAA\nBBB"
	unselected := ApplyShadow(input, 10, 5, false)
	selected := ApplyShadow(input, 10, 5, true)

	if unselected == selected {
		t.Error("selected and unselected shadow should differ")
	}

	// Selected uses full block █ instead of half block ▐
	if !strings.Contains(selected, shadowRightSelected) {
		t.Errorf("selected shadow should contain %q", shadowRightSelected)
	}
}

func TestApplyShadow_OmittedWhenTooNarrow(t *testing.T) {
	input := "AAA\nBBB"

	// Width exactly at MinWidth — too narrow for shadow
	result := ApplyShadow(input, 5, 5, false)
	if result != input {
		t.Error("shadow should be omitted when width < minWidth + 2")
	}

	// Width = MinWidth + 1 — still too narrow
	result = ApplyShadow(input, 6, 5, false)
	if result != input {
		t.Error("shadow should be omitted when width < minWidth + 2")
	}

	// Width = MinWidth + 2 — shadow applied
	result = ApplyShadow(input, 7, 5, false)
	if result == input {
		t.Error("shadow should be applied when width >= minWidth + 2")
	}
}

func TestApplyShadow_ConsistentVisualWidth(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	// Build a simple 3-line input where each line is exactly 20 chars wide
	line := strings.Repeat("X", 20)
	input := line + "\n" + line + "\n" + line
	result := ApplyShadow(input, 20, 5, false)
	lines := strings.Split(result, "\n")

	// All lines should have the same visual width
	firstWidth := ansi.StringWidth(lines[0])
	for i, l := range lines {
		w := ansi.StringWidth(l)
		if w != firstWidth {
			t.Errorf("line %d width %d != first line width %d\nline: %q",
				i, w, firstWidth, l)
		}
	}
}

func TestApplyShadow_AllThemesIntegration(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	themes := []*DoorTheme{
		NewClassicTheme(),
		NewModernTheme(),
		NewSciFiTheme(),
		NewShojiTheme(),
	}

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			// Render at sufficient width for shadow
			output := theme.Render("Test task", 40, 16, false, "", 0.0)
			lines := strings.Split(output, "\n")

			// Last line should be shadow bottom row with ▄
			lastLine := lines[len(lines)-1]
			if !strings.Contains(lastLine, shadowBottom) {
				t.Errorf("last line should contain bottom shadow %q, got: %q",
					shadowBottom, lastLine)
			}

			// Interior lines (not first, not last two) should have right shadow
			for i := 1; i < len(lines)-2; i++ {
				if !strings.Contains(lines[i], shadowRight) {
					t.Errorf("line %d should contain right shadow, got: %q", i, lines[i])
				}
			}
		})
	}
}

func TestApplyShadow_CompactModeNoShadow(t *testing.T) {
	themes := []*DoorTheme{
		NewClassicTheme(),
		NewModernTheme(),
		NewSciFiTheme(),
		NewShojiTheme(),
	}

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			// Compact mode (height=0) should not have shadow elements
			output := theme.Render("Test task", 40, 0, false, "", 0.0)

			if strings.Contains(output, shadowBottom) {
				t.Error("compact mode should not contain bottom shadow ▄")
			}
		})
	}
}
