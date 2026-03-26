package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/tui/themes"
)

// --- AC1: Seasonal resolution at construction time ---

func TestDoorsView_ResolveSeasonalTheme_OverridesBase(t *testing.T) {
	t.Parallel()

	dv := newTestDoorsView("t1", "t2", "t3")
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:   "summer-beach",
		Season: "summer",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return "SUMMER\n" + content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetBaseThemeName("modern")
	dv.SetSeasonalEnabled(true)

	// July 15 is summer
	dv.ResolveSeasonalTheme(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC))

	if dv.Theme() == nil {
		t.Fatal("theme should not be nil after seasonal resolution")
	}
	if dv.Theme().Name != "summer-beach" {
		t.Errorf("expected summer-beach, got %s", dv.Theme().Name)
	}
}

// --- AC2: No seasonal theme registered → base theme unchanged ---

func TestDoorsView_ResolveSeasonalTheme_NoMatch_KeepsBase(t *testing.T) {
	t.Parallel()

	dv := newTestDoorsView("t1", "t2", "t3")
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name: "classic",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetBaseThemeName("classic")
	dv.SetSeasonalEnabled(true)

	// No seasonal themes registered
	dv.ResolveSeasonalTheme(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC))

	if dv.Theme() == nil {
		t.Fatal("theme should not be nil")
	}
	if dv.Theme().Name != "classic" {
		t.Errorf("expected classic (base), got %s", dv.Theme().Name)
	}
}

// --- AC3: Resolved theme stored at construction time, no per-render rechecks ---

func TestDoorsView_SeasonalTheme_StoredOnce(t *testing.T) {
	t.Parallel()

	dv := newTestDoorsView("t1", "t2", "t3")
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:   "spring-flowers",
		Season: "spring",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return "SPRING\n" + content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetBaseThemeName("modern")
	dv.SetSeasonalEnabled(true)
	dv.ResolveSeasonalTheme(time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC))

	// Theme should be spring
	if dv.Theme().Name != "spring-flowers" {
		t.Fatalf("expected spring-flowers, got %s", dv.Theme().Name)
	}

	// Calling View() multiple times should not change the theme
	dv.SetWidth(120)
	dv.View()
	dv.View()
	if dv.Theme().Name != "spring-flowers" {
		t.Errorf("theme changed after View(), got %s", dv.Theme().Name)
	}
}

// --- AC4: seasonal_themes: false disables resolution ---

func TestDoorsView_SeasonalDisabled_UsesBaseTheme(t *testing.T) {
	t.Parallel()

	dv := newTestDoorsView("t1", "t2", "t3")
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name: "modern",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return content
		},
		MinWidth: 10,
	})
	registry.Register(&themes.DoorTheme{
		Name:   "summer-beach",
		Season: "summer",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return "SUMMER\n" + content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetBaseThemeName("modern")
	dv.SetSeasonalEnabled(false)

	dv.ResolveSeasonalTheme(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC))

	if dv.Theme() == nil {
		t.Fatal("theme should not be nil")
	}
	if dv.Theme().Name != "modern" {
		t.Errorf("expected modern (base), got %s", dv.Theme().Name)
	}
}

// --- AC6: MinWidth fallback uses base theme for seasonal themes ---

func TestDoorsView_SeasonalTheme_WidthFallback_UsesBaseTheme(t *testing.T) {
	t.Parallel()

	dv := newTestDoorsView("t1", "t2", "t3")
	registry := themes.NewRegistry()

	baseMarker := "BASE_THEME"
	registry.Register(&themes.DoorTheme{
		Name: "my-base",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return baseMarker + "\n" + content
		},
		MinWidth: 10,
	})

	seasonalMarker := "SEASONAL_THEME"
	registry.Register(&themes.DoorTheme{
		Name:   "wide-seasonal",
		Season: "summer",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return seasonalMarker + "\n" + content
		},
		MinWidth: 100, // requires very wide terminal
	})
	dv.SetThemeRegistry(registry)
	dv.SetBaseThemeName("my-base")
	dv.SetSeasonalEnabled(true)
	dv.ResolveSeasonalTheme(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC))

	if dv.Theme().Name != "wide-seasonal" {
		t.Fatalf("expected wide-seasonal, got %s", dv.Theme().Name)
	}

	// Narrow terminal — should fall back to base theme, not classic
	dv.SetWidth(50)
	view := dv.View()
	if strings.Contains(view, seasonalMarker) {
		t.Error("narrow terminal should not use seasonal theme")
	}
	if !strings.Contains(view, baseMarker) {
		t.Error("narrow terminal should fall back to base theme for seasonal themes")
	}
}

// --- AC7: Planning session re-checks seasonal resolution ---

func TestDoorsView_ResolveSeasonalTheme_PlanningRecheck(t *testing.T) {
	t.Parallel()

	dv := newTestDoorsView("t1", "t2", "t3")
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:   "autumn-leaves",
		Season: "autumn",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return "AUTUMN\n" + content
		},
		MinWidth: 10,
	})
	registry.Register(&themes.DoorTheme{
		Name:   "winter-frost",
		Season: "winter",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return "WINTER\n" + content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetBaseThemeName("modern")
	dv.SetSeasonalEnabled(true)

	// Initial resolution in autumn
	dv.ResolveSeasonalTheme(time.Date(2026, 11, 30, 23, 0, 0, 0, time.UTC))
	if dv.Theme().Name != "autumn-leaves" {
		t.Fatalf("expected autumn-leaves, got %s", dv.Theme().Name)
	}

	// Overnight session: planning starts in winter
	dv.ResolveSeasonalTheme(time.Date(2026, 12, 1, 8, 0, 0, 0, time.UTC))
	if dv.Theme().Name != "winter-frost" {
		t.Errorf("expected winter-frost after recheck, got %s", dv.Theme().Name)
	}
}

// --- AC9: Integration with mocked time ---

func TestDoorsView_SeasonalTheme_AllSeasons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		date       time.Time
		wantSeason string
	}{
		{"spring", time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC), "spring"},
		{"summer", time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), "summer"},
		{"autumn", time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), "autumn"},
		{"winter dec", time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC), "winter"},
		{"winter jan", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), "winter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dv := newTestDoorsView("t1", "t2", "t3")
			registry := themes.NewRegistry()
			for _, season := range []string{"spring", "summer", "autumn", "winter"} {
				s := season // capture
				registry.Register(&themes.DoorTheme{
					Name:   s + "-theme",
					Season: s,
					Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
						return content
					},
					MinWidth: 10,
				})
			}
			dv.SetThemeRegistry(registry)
			dv.SetBaseThemeName("modern")
			dv.SetSeasonalEnabled(true)

			dv.ResolveSeasonalTheme(tt.date)

			want := tt.wantSeason + "-theme"
			if dv.Theme().Name != want {
				t.Errorf("for date %v: expected %s, got %s", tt.date, want, dv.Theme().Name)
			}
		})
	}
}

// --- AC10: Existing behavior unchanged when seasonal_themes: false ---

func TestDoorsView_SeasonalDisabled_ThemeSelectionUnchanged(t *testing.T) {
	t.Parallel()

	dv := newTestDoorsView("t1", "t2", "t3")
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name: "scifi",
		Render: func(content string, width, height int, selected bool, hint string, emphasis float64) string {
			return content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetBaseThemeName("scifi")
	dv.SetSeasonalEnabled(false)
	dv.ResolveSeasonalTheme(time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))

	if dv.Theme().Name != "scifi" {
		t.Errorf("with seasonal disabled, expected scifi, got %s", dv.Theme().Name)
	}
}
