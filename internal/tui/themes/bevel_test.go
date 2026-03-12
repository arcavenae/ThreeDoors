package themes

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ansiColorRe matches SGR color sequences (foreground).
var ansiColorRe = regexp.MustCompile(`\x1b\[[\d;]*m`)

// extractColors returns all unique ANSI SGR sequences found in s.
func extractColors(s string) []string {
	matches := ansiColorRe.FindAllString(s, -1)
	seen := map[string]bool{}
	var unique []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}
	return unique
}

func TestBevelStyles_Unit(t *testing.T) {
	// Not parallel: sets global lipgloss color profile.
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	hi := lipgloss.CompleteColor{TrueColor: "#ff0000"}
	sh := lipgloss.CompleteColor{TrueColor: "#0000ff"}
	sel := lipgloss.CompleteColor{TrueColor: "#ffffff"}

	t.Run("unselected_styles_differ", func(t *testing.T) {
		hiS, shS := bevelStyles(hi, sh, sel, false)
		hiOut := hiS.Render("X")
		shOut := shS.Render("X")
		if hiOut == shOut {
			t.Error("expected different output for highlight vs shadow styles when unselected")
		}
	})

	t.Run("selected_styles_match", func(t *testing.T) {
		hiS, shS := bevelStyles(hi, sh, sel, true)
		hiOut := hiS.Render("X")
		shOut := shS.Render("X")
		if hiOut != shOut {
			t.Errorf("expected identical output when selected, got %q vs %q", hiOut, shOut)
		}
	})
}

func TestBevelLighting_AllThemes(t *testing.T) {
	// Not parallel: sets global lipgloss color profile.
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name+"/unselected_bevel_differentiation", func(t *testing.T) {
			out := theme.Render("Test content", 30, 16, false, "", 0.0)
			lines := strings.Split(out, "\n")
			anatomy := NewDoorAnatomy(16)

			if anatomy.LintelRow >= len(lines) || anatomy.ThresholdRow >= len(lines) {
				t.Fatalf("not enough lines: got %d, need lintel=%d threshold=%d",
					len(lines), anatomy.LintelRow, anatomy.ThresholdRow)
			}

			topColors := extractColors(lines[anatomy.LintelRow])
			botColors := extractColors(lines[anatomy.ThresholdRow])

			if len(topColors) == 0 {
				t.Error("top border has no ANSI color codes — bevel not applied")
			}
			if len(botColors) == 0 {
				t.Error("bottom border has no ANSI color codes — bevel not applied")
			}

			// Top (Highlight) and bottom (ShadowEdge) should use different colors
			topJoined := strings.Join(topColors, "|")
			botJoined := strings.Join(botColors, "|")
			if topJoined == botJoined {
				t.Errorf("expected different bevel colors for top (Highlight) vs bottom (ShadowEdge), both have: %s", topJoined)
			}
		})

		t.Run(theme.Name+"/selected_override", func(t *testing.T) {
			out := theme.Render("Test content", 30, 16, true, "", 0.0)
			lines := strings.Split(out, "\n")
			anatomy := NewDoorAnatomy(16)

			if anatomy.LintelRow >= len(lines) || anatomy.ThresholdRow >= len(lines) {
				t.Fatalf("not enough lines: got %d", len(lines))
			}

			topColors := extractColors(lines[anatomy.LintelRow])
			botColors := extractColors(lines[anatomy.ThresholdRow])

			// When selected, the first color in both rows should be the same
			// Selected color. Bottom may have extra shadow colors appended by ApplyShadow.
			if len(topColors) == 0 || len(botColors) == 0 {
				t.Fatal("missing color codes in border rows")
			}
			if topColors[0] != botColors[0] {
				t.Errorf("expected same Selected color for top and bottom border, got %s vs %s", topColors[0], botColors[0])
			}
		})
	}
}

func TestBevelLighting_LeftRight(t *testing.T) {
	// Not parallel: sets global lipgloss color profile.
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			out := theme.Render("Test content", 30, 16, false, "", 0.0)
			lines := strings.Split(out, "\n")
			anatomy := NewDoorAnatomy(16)

			// Check a content row: left border should use Highlight, right should use ShadowEdge
			contentRow := anatomy.ContentStart
			if contentRow >= len(lines) {
				t.Fatalf("content row %d out of range (only %d lines)", contentRow, len(lines))
			}

			line := lines[contentRow]
			colors := extractColors(line)
			if len(colors) < 2 {
				t.Errorf("content row should have at least 2 different color sequences (left+right bevel), got %d", len(colors))
			}
		})
	}
}

func TestBevelLighting_MinWidth(t *testing.T) {
	// Not parallel: sets global lipgloss color profile.
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			// Render at minimum width with bevel — content should still be present
			out := theme.Render("Test", theme.MinWidth, 16, false, "", 0.0)
			if !strings.Contains(out, "Test") {
				t.Error("content missing at minimum width with bevel")
			}
		})
	}
}

func TestBevelLighting_Divider(t *testing.T) {
	// Not parallel: sets global lipgloss color profile.
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(termenv.TrueColor) })

	for _, theme := range allThemes() {
		t.Run(theme.Name, func(t *testing.T) {
			out := theme.Render("Test", 30, 16, false, "", 0.0)
			lines := strings.Split(out, "\n")
			anatomy := NewDoorAnatomy(16)

			if anatomy.PanelDivider >= len(lines) {
				t.Fatalf("panel divider row %d out of range", anatomy.PanelDivider)
			}

			divColors := extractColors(lines[anatomy.PanelDivider])
			botColors := extractColors(lines[anatomy.ThresholdRow])

			if len(divColors) == 0 {
				t.Error("divider has no color codes — bevel not applied")
				return
			}

			// For most themes, the first color in both divider and bottom should be
			// the same ShadowEdge. Bottom may have extra shadow colors from ApplyShadow.
			// For scifi, the divider has mixed styles (outer left=Highlight, interior=ShadowEdge).
			if theme.Name != "scifi" {
				if len(divColors) == 0 || len(botColors) == 0 {
					t.Fatal("missing color codes")
				}
				if divColors[0] != botColors[0] {
					t.Errorf("divider and bottom should use same ShadowEdge color, first color: %s vs %s", divColors[0], botColors[0])
				}
			}
		})
	}
}
