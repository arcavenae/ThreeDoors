package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestShadowColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		width    int
		minWidth int
		selected bool
		want     int
	}{
		{"below threshold", 5, 5, false, 0},
		{"at threshold minus one", 6, 5, false, 0},
		{"at threshold", 7, 5, false, 1},
		{"at threshold selected", 7, 5, true, 2},
		{"two col threshold", 9, 5, false, 2},
		{"two col selected", 9, 5, true, 3},
		{"three col threshold", 11, 5, false, 3},
		{"three col selected", 11, 5, true, 3},
		{"wide unselected", 40, 15, false, 3},
		{"wide selected", 40, 15, true, 3},
		{"medium unselected", 18, 15, false, 1},
		{"medium selected", 18, 15, true, 2},
		{"shoji narrow", 20, 19, false, 0},
		{"shoji medium", 23, 19, false, 2},
		{"shoji wide", 26, 19, true, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ShadowColumns(tt.width, tt.minWidth, tt.selected)
			if got != tt.want {
				t.Errorf("ShadowColumns(%d, %d, %v) = %d, want %d",
					tt.width, tt.minWidth, tt.selected, got, tt.want)
			}
		})
	}
}

func TestApplyShadow_GradientColumns(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	input := "AAA\nBBB\nCCC"

	// Width 10, minWidth 5: extra=5 → 2 cols (unselected)
	result := ApplyShadow(input, 10, 5, false, nil, nil)
	lines := strings.Split(result, "\n")

	// Should have original 3 lines + 1 shadow bottom
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	// First line: no right shadow (padding only)
	if strings.Contains(lines[0], shadowGradientNear) {
		t.Error("first line should not have shadow gradient")
	}

	// Lines 1-2: should have gradient chars ▓ and ░ (2 columns)
	for i := 1; i < 3; i++ {
		if !strings.Contains(lines[i], shadowGradientNear) {
			t.Errorf("line %d should contain near gradient %q", i, shadowGradientNear)
		}
		if !strings.Contains(lines[i], shadowGradientFar) {
			t.Errorf("line %d should contain far gradient %q", i, shadowGradientFar)
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

func TestApplyShadow_SelectedWider(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	input := "AAA\nBBB"

	// Width 10, minWidth 5: extra=5
	// Unselected: 2 cols, Selected: 3 cols
	unselected := ApplyShadow(input, 10, 5, false, nil, nil)
	selected := ApplyShadow(input, 10, 5, true, nil, nil)

	if unselected == selected {
		t.Error("selected and unselected shadow should differ")
	}

	// Selected uses full block █ for near column
	if !strings.Contains(selected, shadowRightSelected) {
		t.Errorf("selected shadow should contain %q", shadowRightSelected)
	}

	// Selected should have 3 gradient columns, unselected 2
	unselLines := strings.Split(unselected, "\n")
	selLines := strings.Split(selected, "\n")

	// Check line 1 visual widths — selected should be 1 wider
	unselW := ansi.StringWidth(unselLines[1])
	selW := ansi.StringWidth(selLines[1])
	if selW != unselW+1 {
		t.Errorf("selected line width %d should be unselected %d + 1", selW, unselW)
	}
}

func TestApplyShadow_OmittedWhenTooNarrow(t *testing.T) {
	input := "AAA\nBBB"

	// Width exactly at MinWidth — too narrow for shadow
	result := ApplyShadow(input, 5, 5, false, nil, nil)
	if result != input {
		t.Error("shadow should be omitted when width < minWidth + 2")
	}

	// Width = MinWidth + 1 — still too narrow
	result = ApplyShadow(input, 6, 5, false, nil, nil)
	if result != input {
		t.Error("shadow should be omitted when width < minWidth + 2")
	}

	// Width = MinWidth + 2 — shadow applied
	result = ApplyShadow(input, 7, 5, false, nil, nil)
	if result == input {
		t.Error("shadow should be applied when width >= minWidth + 2")
	}
}

func TestApplyShadow_ConsistentVisualWidth(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	line := strings.Repeat("X", 20)
	input := line + "\n" + line + "\n" + line
	result := ApplyShadow(input, 20, 5, false, nil, nil)
	lines := strings.Split(result, "\n")

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

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			// Render at sufficient width for shadow (door mode)
			output := theme.Render("Test task", 40, 16, false, "", 0.0)
			lines := strings.Split(output, "\n")

			// Last line should be shadow bottom row with ▄
			lastLine := lines[len(lines)-1]
			if !strings.Contains(lastLine, shadowBottom) {
				t.Errorf("last line should contain bottom shadow %q, got: %q",
					shadowBottom, lastLine)
			}

			// Interior lines (not first, not last two) should have gradient chars
			hasGradient := false
			for i := 1; i < len(lines)-2; i++ {
				if strings.Contains(lines[i], shadowGradientNear) ||
					strings.Contains(lines[i], shadowRight) {
					hasGradient = true
					break
				}
			}
			if !hasGradient {
				t.Error("interior lines should contain shadow gradient characters")
			}
		})
	}
}

func TestApplyShadow_CompactModeNoShadow(t *testing.T) {
	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			// Compact mode (height=0) should not have shadow elements
			output := theme.Render("Test task", 40, 0, false, "", 0.0)

			if strings.Contains(output, shadowBottom) {
				t.Error("compact mode should not contain bottom shadow ▄")
			}
		})
	}
}

func TestApplyShadow_SingleColumn(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	input := "AAA\nBBB"

	// Width 7, minWidth 5: extra=2 → 1 col (unselected)
	result := ApplyShadow(input, 7, 5, false, nil, nil)
	lines := strings.Split(result, "\n")

	// Line 1 should have ▐ (single column unselected)
	if !strings.Contains(lines[1], shadowRight) {
		t.Errorf("single-column shadow should use %q, got: %q", shadowRight, lines[1])
	}

	// Should NOT have gradient chars
	if strings.Contains(lines[1], shadowGradientNear) {
		t.Error("single-column shadow should not have gradient near char")
	}
}

func TestApplyShadow_ThreeColumns(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	input := "AAA\nBBB"

	// Width 12, minWidth 5: extra=7 → 3 cols (unselected)
	result := ApplyShadow(input, 12, 5, false, nil, nil)
	lines := strings.Split(result, "\n")

	// Line 1 should have full gradient: ▓▒░
	if !strings.Contains(lines[1], shadowGradientNear) {
		t.Errorf("3-col shadow should contain %q", shadowGradientNear)
	}
	if !strings.Contains(lines[1], shadowGradientMid) {
		t.Errorf("3-col shadow should contain %q", shadowGradientMid)
	}
	if !strings.Contains(lines[1], shadowGradientFar) {
		t.Errorf("3-col shadow should contain %q", shadowGradientFar)
	}
}

// TestShadowNearContrastRatio validates that each theme's ShadowNear color
// has sufficient contrast against dark terminal backgrounds (≥4:1).
func TestShadowNearContrastRatio(t *testing.T) {
	t.Parallel()

	darkBgLum := relativeLuminance(0, 0, 0) // #000000

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()

			cc, ok := theme.Colors.ShadowNear.(lipgloss.CompleteColor)
			if !ok {
				t.Skip("ShadowNear is not a CompleteColor")
			}

			r, g, b := parseHexColor(cc.TrueColor)
			lum := relativeLuminance(r, g, b)
			ratio := contrastRatio(lum, darkBgLum)
			if ratio < 4.0 {
				t.Errorf("ShadowNear %s contrast ratio %.2f:1 against #000000 is below 4:1 minimum",
					cc.TrueColor, ratio)
			}
		})
	}
}
