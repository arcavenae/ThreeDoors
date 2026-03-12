package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// TestCrackOfLight_ClassicTheme verifies that the classic theme renders crack-of-light
// characters on the right border when emphasis exceeds the crack threshold.
func TestCrackOfLight_ClassicTheme(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	output := theme.Render("Task text", 30, 16, true, "", 1.0)

	if !strings.Contains(output, "╎") {
		t.Error("selected door with full emphasis should contain crack character ╎")
	}

	if !strings.Contains(output, "╮") {
		t.Error("selected door with full emphasis should contain rounded corner ╮")
	}

	if !strings.Contains(output, "╯") {
		t.Error("selected door with full emphasis should contain rounded corner ╯")
	}
}

// TestCrackOfLight_NotShownWhenUnselected verifies crack does not appear
// when the door is not selected, regardless of emphasis value.
func TestCrackOfLight_NotShownWhenUnselected(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	output := theme.Render("Task text", 30, 16, false, "", 0.0)

	if strings.Contains(output, "╎") {
		t.Error("unselected door should not contain crack character ╎")
	}
}

// TestCrackOfLight_BelowThreshold verifies crack does not appear when
// emphasis is below the crack threshold, even when selected.
func TestCrackOfLight_BelowThreshold(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	output := theme.Render("Task text", 30, 16, true, "", 0.1)

	if strings.Contains(output, "╎") {
		t.Error("selected door below crack threshold should not contain crack character ╎")
	}
}

// TestCrackOfLight_AboveThreshold verifies crack appears when emphasis
// crosses the crack threshold.
func TestCrackOfLight_AboveThreshold(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	output := theme.Render("Task text", 30, 16, true, "", CrackEmphasisThreshold+0.1)

	if !strings.Contains(output, "╎") {
		t.Errorf("selected door above crack threshold (%.1f) should contain crack character ╎",
			CrackEmphasisThreshold+0.1)
	}
}

// TestCrackOfLight_ShadowUsesShade verifies the shadow column uses ░ when cracked.
func TestCrackOfLight_ShadowUsesShade(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	output := theme.Render("Task text", 30, 16, true, "", 1.0)

	if !strings.Contains(output, "░") {
		t.Error("cracked door shadow should use ░ shade character")
	}
}

// TestCrackOfLight_RemovedOnDeselect verifies crack characters are not present
// when a previously selected door is deselected (emphasis returns to 0).
func TestCrackOfLight_RemovedOnDeselect(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	cracked := theme.Render("Task text", 30, 16, true, "", 1.0)
	if !strings.Contains(cracked, "╎") {
		t.Fatal("precondition: cracked render should contain ╎")
	}

	uncracked := theme.Render("Task text", 30, 16, false, "", 0.0)
	if strings.Contains(uncracked, "╎") {
		t.Error("deselected door should not contain crack character ╎")
	}
}

// TestCrackOfLight_MinimumWidth verifies the door still renders correctly
// at minimum width (15 chars) with crack active.
func TestCrackOfLight_MinimumWidth(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	output := theme.Render("Task", 15, 16, true, "", 1.0)
	lines := strings.Split(output, "\n")

	if len(lines) < 3 {
		t.Fatal("expected at least 3 lines")
	}

	firstWidth := ansi.StringWidth(lines[0])
	for i := 0; i < len(lines)-2; i++ {
		lw := ansi.StringWidth(lines[i])
		if lw != firstWidth {
			t.Errorf("line %d visual width %d != first line width %d\nline: %q",
				i, lw, firstWidth, lines[i])
		}
	}
}

// TestCrackOfLight_ContentReduced verifies the content area width is maintained
// when the crack effect is active (crack + shade replace the border character
// while keeping total visual width constant).
func TestCrackOfLight_ContentReduced(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	uncracked := theme.Render("Task", 30, 16, false, "", 0.0)
	cracked := theme.Render("Task", 30, 16, true, "", 1.0)

	uncrackedLines := strings.Split(uncracked, "\n")
	crackedLines := strings.Split(cracked, "\n")

	if len(uncrackedLines) < 2 || len(crackedLines) < 2 {
		t.Fatal("expected at least 2 lines")
	}

	uncrackedWidth := ansi.StringWidth(uncrackedLines[0])
	crackedWidth := ansi.StringWidth(crackedLines[0])

	if uncrackedWidth != crackedWidth {
		t.Errorf("cracked and uncracked door widths should match: uncracked=%d, cracked=%d",
			uncrackedWidth, crackedWidth)
	}
}

// TestCrackOfLight_AllThemes verifies that all door-mode themes support
// the crack-of-light effect when selected with full emphasis.
func TestCrackOfLight_AllThemes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		theme     *DoorTheme
		width     int
		height    int
		crackChar string
	}{
		{"classic", NewClassicTheme(), 30, 16, "╎"},
		{"modern", NewModernTheme(), 30, 16, "╎"},
		{"scifi", NewSciFiTheme(), 30, 16, "╎"},
		{"shoji", NewShojiTheme(), 30, 16, "╎"},
		{"winter", NewWinterTheme(), 30, 16, "╎"},
		{"spring", NewSpringTheme(), 30, 16, "╎"},
		{"summer", NewSummerTheme(), 30, 16, "╎"},
		{"autumn", NewAutumnTheme(), 30, 16, "╎"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := tt.theme.Render("Task text", tt.width, tt.height, true, "", 1.0)
			if !strings.Contains(output, tt.crackChar) {
				t.Errorf("%s theme should contain crack character %q when selected with full emphasis",
					tt.name, tt.crackChar)
			}
		})
	}
}

// TestCrackOfLight_VisualWidthConsistency verifies that all lines of a cracked
// door have consistent visual width across all themes.
func TestCrackOfLight_VisualWidthConsistency(t *testing.T) {
	t.Parallel()

	themeTests := []struct {
		name  string
		theme *DoorTheme
		width int
	}{
		{"classic", NewClassicTheme(), 30},
		{"modern", NewModernTheme(), 30},
		{"winter", NewWinterTheme(), 30},
		{"spring", NewSpringTheme(), 30},
		{"summer", NewSummerTheme(), 30},
		{"autumn", NewAutumnTheme(), 30},
	}

	for _, tt := range themeTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := tt.theme.Render("Task text", tt.width, 16, true, "", 1.0)
			lines := strings.Split(output, "\n")

			if len(lines) < 3 {
				t.Fatal("expected at least 3 lines")
			}

			firstWidth := ansi.StringWidth(lines[0])
			for i := 0; i < len(lines)-2; i++ {
				lw := ansi.StringWidth(lines[i])
				if lw != firstWidth {
					t.Errorf("line %d visual width %d != first line width %d\nline: %q",
						i, lw, firstWidth, lines[i])
				}
			}
		})
	}
}

// TestApplyShadow_CrackMode verifies ApplyShadow uses ░ when cracked=true.
func TestApplyShadow_CrackMode(t *testing.T) {
	t.Parallel()

	input := "line1\nline2\nline3"

	normal := ApplyShadow(input, 20, 15, true)
	if strings.Contains(normal, "░") {
		t.Error("non-cracked shadow should not contain ░")
	}

	cracked := ApplyShadowWithCrack(input, 20, 15, true)
	if !strings.Contains(cracked, "░") {
		t.Error("cracked shadow should contain ░")
	}
}
