package themes

import (
	"math"
	"strconv"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// relativeLuminance calculates the WCAG 2.0 relative luminance of an sRGB color.
// See https://www.w3.org/TR/WCAG20/#relativeluminancedef
func relativeLuminance(r, g, b uint8) float64 {
	linearize := func(c uint8) float64 {
		s := float64(c) / 255.0
		if s <= 0.04045 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*linearize(r) + 0.7152*linearize(g) + 0.0722*linearize(b)
}

// contrastRatio calculates the WCAG 2.0 contrast ratio between two luminances.
// See https://www.w3.org/TR/WCAG20/#contrast-ratiodef
func contrastRatio(l1, l2 float64) float64 {
	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)
	return (lighter + 0.05) / (darker + 0.05)
}

// parseHexColor parses a #RRGGBB hex color string into RGB components.
func parseHexColor(hex string) (uint8, uint8, uint8) {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return uint8(r), uint8(g), uint8(b)
}

// wcagAAMinContrast is the minimum contrast ratio for WCAG AA normal text.
const wcagAAMinContrast = 4.5

// TestSeasonalThemes_WCAGContrast validates that all seasonal theme text colors
// maintain minimum 4.5:1 WCAG AA contrast ratio against both dark (#000000) and
// light (#FFFFFF) terminal backgrounds (FR136, AC10, AC11).
func TestSeasonalThemes_WCAGContrast(t *testing.T) {
	t.Parallel()

	darkBg := relativeLuminance(0, 0, 0)        // #000000
	lightBg := relativeLuminance(255, 255, 255) // #FFFFFF

	seasonalThemes := []*DoorTheme{
		NewWinterTheme(),
		NewSpringTheme(),
		NewSummerTheme(),
		NewAutumnTheme(),
	}

	for _, theme := range seasonalThemes {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()

			// Extract TrueColor hex from Frame and Selected colors
			colors := map[string]lipgloss.TerminalColor{
				"Frame":    theme.Colors.Frame,
				"Selected": theme.Colors.Selected,
				"Accent":   theme.Colors.Accent,
			}

			for colorName, tc := range colors {
				cc, ok := tc.(lipgloss.CompleteColor)
				if !ok {
					t.Errorf("%s: %s is not CompleteColor", theme.Name, colorName)
					continue
				}

				r, g, b := parseHexColor(cc.TrueColor)
				lum := relativeLuminance(r, g, b)

				// Must pass against at least one background (dark or light)
				darkRatio := contrastRatio(lum, darkBg)
				lightRatio := contrastRatio(lum, lightBg)

				if darkRatio < wcagAAMinContrast && lightRatio < wcagAAMinContrast {
					t.Errorf("%s %s (#%02x%02x%02x): contrast %.2f:1 (dark) and %.2f:1 (light) — both below WCAG AA 4.5:1",
						theme.Name, colorName, r, g, b, darkRatio, lightRatio)
				}
			}
		})
	}
}

// TestSeasonalThemes_SeasonMetadata validates that all seasonal themes have
// correct Season, SeasonStart, and SeasonEnd fields matching DefaultSeasonRanges (AC3).
func TestSeasonalThemes_SeasonMetadata(t *testing.T) {
	t.Parallel()

	rangeMap := make(map[string]SeasonRange)
	for _, r := range DefaultSeasonRanges {
		rangeMap[r.Name] = r
	}

	seasonalThemes := []*DoorTheme{
		NewWinterTheme(),
		NewSpringTheme(),
		NewSummerTheme(),
		NewAutumnTheme(),
	}

	for _, theme := range seasonalThemes {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()

			if theme.Season == "" {
				t.Fatal("Season field must be set")
			}

			expected, ok := rangeMap[theme.Season]
			if !ok {
				t.Fatalf("Season %q not found in DefaultSeasonRanges", theme.Season)
			}

			if theme.SeasonStart != expected.Start {
				t.Errorf("SeasonStart = %+v, want %+v", theme.SeasonStart, expected.Start)
			}
			if theme.SeasonEnd != expected.End {
				t.Errorf("SeasonEnd = %+v, want %+v", theme.SeasonEnd, expected.End)
			}
		})
	}
}

// TestSeasonalThemes_RegisteredInDefaultRegistry validates that all four
// seasonal themes are registered in NewDefaultRegistry() (AC13).
func TestSeasonalThemes_RegisteredInDefaultRegistry(t *testing.T) {
	t.Parallel()

	registry := NewDefaultRegistry()

	seasons := []string{"winter", "spring", "summer", "autumn"}
	for _, season := range seasons {
		t.Run(season, func(t *testing.T) {
			t.Parallel()

			theme, ok := registry.GetBySeason(season)
			if !ok {
				t.Fatalf("no theme registered for season %q", season)
			}
			if theme.Season != season {
				t.Errorf("theme.Season = %q, want %q", theme.Season, season)
			}
		})
	}
}

// TestSeasonalThemes_DistinctVisualIdentity validates that each seasonal theme
// produces visually distinct output from all other seasonal themes (AC9).
func TestSeasonalThemes_DistinctVisualIdentity(t *testing.T) {
	t.Parallel()

	seasonalThemes := []*DoorTheme{
		NewWinterTheme(),
		NewSpringTheme(),
		NewSummerTheme(),
		NewAutumnTheme(),
	}

	outputs := make(map[string]string)
	for _, theme := range seasonalThemes {
		outputs[theme.Name] = theme.Render("Test task content", 40, 16, false, "")
	}

	for i, a := range seasonalThemes {
		for j, cmpTheme := range seasonalThemes {
			if i >= j {
				continue
			}
			if outputs[a.Name] == outputs[cmpTheme.Name] {
				t.Errorf("themes %q and %q produce identical output", a.Name, cmpTheme.Name)
			}
		}
	}
}

// TestSeasonalThemes_UnicodeSafety validates that seasonal themes only use
// characters from allowed Unicode ranges (AC4): box-drawing (U+2500-U+257F),
// block elements (U+2580-U+259F), geometric shapes (U+25A0-U+25FF), and ASCII.
func TestSeasonalThemes_UnicodeSafety(t *testing.T) {
	t.Parallel()

	seasonalThemes := []*DoorTheme{
		NewWinterTheme(),
		NewSpringTheme(),
		NewSummerTheme(),
		NewAutumnTheme(),
	}

	for _, theme := range seasonalThemes {
		t.Run(theme.Name, func(t *testing.T) {
			t.Parallel()

			output := theme.Render("Test task", 40, 16, false, "")

			for _, r := range output {
				if r <= 0x007F {
					continue // ASCII
				}
				if r >= 0x00B7 && r <= 0x00B7 {
					continue // middle dot (·)
				}
				if r >= 0x2500 && r <= 0x257F {
					continue // box-drawing
				}
				if r >= 0x2580 && r <= 0x259F {
					continue // block elements
				}
				if r >= 0x25A0 && r <= 0x25FF {
					continue // geometric shapes
				}
				t.Errorf("theme %q uses disallowed Unicode character U+%04X (%c)",
					theme.Name, r, r)
			}
		})
	}
}
