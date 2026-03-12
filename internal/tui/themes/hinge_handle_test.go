package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

// TestHingeAsymmetry_AllThemes verifies that all themes render heavier
// left (hinge) borders and lighter right (opening) borders in door mode,
// creating visible left-right asymmetry (Story 48.1 AC2).
func TestHingeAsymmetry_AllThemes(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	tests := []struct {
		theme        *DoorTheme
		hingeChars   []string // characters expected on left (hinge) side
		openingChars []string // characters expected on right (opening) side
	}{
		{
			NewClassicTheme(),
			[]string{"╓", "║", "╟", "╙"},
			[]string{"┐", "│", "┤", "┘"},
		},
		{
			NewModernTheme(),
			[]string{"╓", "║", "╟", "╙"},
			[]string{"│", "┤"},
		},
		{
			NewSciFiTheme(),
			[]string{"╔", "║", "╚"},
			[]string{"╕", "│", "╛"},
		},
		{
			NewShojiTheme(),
			[]string{"╥", "║", "╟", "╨"},
			[]string{"┬", "│", "┤", "┴"},
		},
		{
			NewWinterTheme(),
			[]string{"╓", "║", "╟", "╙"},
			[]string{"┐", "│", "┤", "┘"},
		},
		{
			NewSpringTheme(),
			[]string{"╓", "║", "╟", "╙"},
			[]string{"╮", "│", "┤", "╯"},
		},
		{
			NewSummerTheme(),
			[]string{"╔", "║", "╠", "╚"},
			[]string{"╕", "│", "╡", "╛"},
		},
		{
			NewAutumnTheme(),
			[]string{"╓", "║", "╟", "╙"},
			[]string{"┐", "│", "┤", "┘"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.theme.Name+"_unselected", func(t *testing.T) {
			output := tt.theme.Render("Test task", 40, 16, false, "", 0.0)

			for _, ch := range tt.hingeChars {
				if !strings.Contains(output, ch) {
					t.Errorf("%s: missing hinge character %q", tt.theme.Name, ch)
				}
			}

			for _, ch := range tt.openingChars {
				if !strings.Contains(output, ch) {
					t.Errorf("%s: missing opening-side character %q", tt.theme.Name, ch)
				}
			}
		})
	}
}

// TestHingeAsymmetry_Selected verifies selected doors also have asymmetry
// with heavy left (hinge) and standard right (opening).
func TestHingeAsymmetry_Selected(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	tests := []struct {
		theme       *DoorTheme
		hingeChars  []string // heavy chars on left
		standardRht []string // standard chars on right
	}{
		{
			NewClassicTheme(),
			[]string{"┏", "┃", "┣", "┗"},
			[]string{"┐", "│", "┤", "┘"},
		},
		{
			NewModernTheme(),
			[]string{"┏", "┃", "┣", "┗"},
			[]string{"│", "┤"},
		},
		{
			NewShojiTheme(),
			[]string{"┳", "┃", "┣", "┻"},
			[]string{"┬", "│", "┤", "┴"},
		},
		{
			NewWinterTheme(),
			[]string{"┏", "┃", "┣", "┗"},
			[]string{"┐", "│", "┤", "┘"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.theme.Name+"_selected", func(t *testing.T) {
			output := tt.theme.Render("Test task", 40, 16, true, "", 0.0)

			for _, ch := range tt.hingeChars {
				if !strings.Contains(output, ch) {
					t.Errorf("%s selected: missing heavy hinge character %q", tt.theme.Name, ch)
				}
			}

			for _, ch := range tt.standardRht {
				if !strings.Contains(output, ch) {
					t.Errorf("%s selected: missing standard opening character %q", tt.theme.Name, ch)
				}
			}
		})
	}
}

// TestHandleAtRightEdge verifies the handle character is positioned at the
// rightmost content column, not centered (Story 48.1 AC1).
func TestHandleAtRightEdge(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	tests := []struct {
		theme     *DoorTheme
		handleSym string
	}{
		{NewClassicTheme(), "●"},
		{NewModernTheme(), "○"},
		{NewShojiTheme(), "○"},
		{NewWinterTheme(), "◆"},
		{NewSpringTheme(), "○"},
		{NewSummerTheme(), "■"},
		{NewAutumnTheme(), "●"},
	}

	for _, tt := range tests {
		t.Run(tt.theme.Name, func(t *testing.T) {
			output := tt.theme.Render("Test task", 40, 16, false, "", 0.0)
			lines := strings.Split(output, "\n")
			anatomy := NewDoorAnatomy(16)

			if anatomy.HandleRow >= len(lines) {
				t.Fatalf("handle row %d out of bounds (only %d lines)", anatomy.HandleRow, len(lines))
			}

			handleLine := lines[anatomy.HandleRow]
			if !strings.Contains(handleLine, tt.handleSym) {
				t.Fatalf("handle row should contain %q, got: %q", tt.handleSym, handleLine)
			}

			// The handle should be at or near the rightmost content column.
			// Find the position of the handle symbol within the line.
			idx := strings.LastIndex(handleLine, tt.handleSym)
			handlePos := ansi.StringWidth(handleLine[:idx])
			lineWidth := ansi.StringWidth(handleLine)

			// Handle should be within 2 chars of the right border
			// (1 char for the right border itself, plus possible shadow)
			distFromRight := lineWidth - handlePos - 1
			if distFromRight > 3 {
				t.Errorf("handle %q should be near right edge (within 3 chars), but is %d chars from right in: %q",
					tt.handleSym, distFromRight, handleLine)
			}
		})
	}
}

// TestSciFiHandleAtRightEdge verifies the Sci-Fi ◈──┤ handle is at the right edge.
func TestSciFiHandleAtRightEdge(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	theme := NewSciFiTheme()
	output := theme.Render("Test task", 40, 16, false, "", 0.0)
	lines := strings.Split(output, "\n")
	anatomy := NewDoorAnatomy(16)

	if anatomy.HandleRow >= len(lines) {
		t.Fatalf("handle row out of bounds")
	}

	handleLine := lines[anatomy.HandleRow]
	if !strings.Contains(handleLine, "◈──┤") {
		t.Fatalf("Sci-Fi handle row should contain ◈──┤, got: %q", handleLine)
	}

	// ◈──┤ should be near the right side (within inner right border)
	idx := strings.Index(handleLine, "◈──┤")
	handleEnd := ansi.StringWidth(handleLine[:idx]) + ansi.StringWidth("◈──┤")
	lineWidth := ansi.StringWidth(handleLine)

	distFromRight := lineWidth - handleEnd
	if distFromRight > 4 {
		t.Errorf("Sci-Fi handle ◈──┤ should be near right edge, but %d chars from right in: %q",
			distFromRight, handleLine)
	}
}

// TestHingeCol_AlwaysZero verifies the HingeCol field is always 0.
func TestHingeCol_AlwaysZero(t *testing.T) {
	t.Parallel()

	for _, height := range []int{5, 10, 16, 24, 40} {
		a := NewDoorAnatomy(height)
		if a.HingeCol != 0 {
			t.Errorf("HingeCol should be 0 for height %d, got %d", height, a.HingeCol)
		}
	}
}

// TestMinWidthWithHingeHandle verifies doors render correctly at minimum width
// with side-mounted handle and hinge marks — zero width cost (Story 48.1 AC5).
func TestMinWidthWithHingeHandle(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			output := theme.Render("Task", theme.MinWidth, theme.MinHeight, false, "", 0.0)
			lines := strings.Split(output, "\n")

			if len(lines) < 2 {
				t.Fatal("expected at least 2 lines")
			}

			// All lines should have consistent width
			firstWidth := ansi.StringWidth(lines[0])
			for i, line := range lines {
				w := ansi.StringWidth(line)
				if w != firstWidth {
					t.Errorf("line %d width %d != first line width %d\nline: %q",
						i, w, firstWidth, line)
				}
			}

			// Content should be present
			if !strings.Contains(output, "Task") {
				t.Error("content should be visible at minimum width")
			}
		})
	}
}
