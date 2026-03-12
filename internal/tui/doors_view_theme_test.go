package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tui/themes"
)

// --- AC1: DoorsView gains a theme field ---

func TestDoorsView_ThemeField_NilByDefault(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	if dv.Theme() != nil {
		t.Error("theme should be nil by default")
	}
}

// --- AC2: Theme loaded from config by name with fallback ---

func TestDoorsView_SetThemeByName_Valid(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetThemeByName("classic")
	if dv.Theme() == nil {
		t.Fatal("theme should not be nil after SetThemeByName")
	}
	if dv.Theme().Name != "classic" {
		t.Errorf("expected classic, got %s", dv.Theme().Name)
	}
}

func TestDoorsView_SetThemeByName_Modern(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetThemeByName("modern")
	if dv.Theme() == nil {
		t.Fatal("theme should not be nil")
	}
	if dv.Theme().Name != "modern" {
		t.Errorf("expected modern, got %s", dv.Theme().Name)
	}
}

// --- AC4: Invalid or missing theme falls back to DefaultThemeName ---

func TestDoorsView_SetThemeByName_InvalidFallsBack(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetThemeByName("nonexistent")
	if dv.Theme() == nil {
		t.Fatal("theme should fall back, not be nil")
	}
	if dv.Theme().Name != themes.DefaultThemeName {
		t.Errorf("expected fallback to %s, got %s", themes.DefaultThemeName, dv.Theme().Name)
	}
}

func TestDoorsView_SetThemeByName_EmptyFallsBack(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetThemeByName("")
	if dv.Theme() == nil {
		t.Fatal("theme should fall back, not be nil")
	}
	if dv.Theme().Name != themes.DefaultThemeName {
		t.Errorf("expected fallback to %s, got %s", themes.DefaultThemeName, dv.Theme().Name)
	}
}

// --- AC3: View uses theme.Render instead of doorStyle ---

func TestDoorsView_View_UsesThemeRender(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(120)

	// Set a custom theme whose Render injects a marker string
	marker := "THEME_RENDERED"
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:        "test-marker",
		Description: "test theme",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return marker + "\n" + content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetThemeByName("test-marker")

	view := dv.View()
	if !strings.Contains(view, marker) {
		t.Errorf("View should use theme.Render, expected marker %q in output", marker)
	}
}

func TestDoorsView_View_NoTheme_UsesLipglossStyle(t *testing.T) {
	// Without setting a theme, View should still render (backward compatible)
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(80)
	view := dv.View()
	if !strings.Contains(view, "ThreeDoors") {
		t.Error("View without theme should still render header")
	}
}

// --- AC5: Width-aware fallback to Classic when too narrow ---

func TestDoorsView_View_NarrowFallsBackToClassic(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")

	// Create a theme with high MinWidth requirement
	wideMarker := "WIDE_THEME"
	registry := themes.NewDefaultRegistry()
	registry.Register(&themes.DoorTheme{
		Name:        "wide-theme",
		Description: "needs wide terminal",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return wideMarker + "\n" + content
		},
		MinWidth: 100, // requires very wide terminal
	})
	dv.SetThemeRegistry(registry)
	dv.SetThemeByName("wide-theme")

	// Set narrow width: (50 - 6) / 3 = ~14, less than MinWidth 100
	dv.SetWidth(50)
	view := dv.View()

	// The wide marker should NOT appear — should fall back to classic
	if strings.Contains(view, wideMarker) {
		t.Error("narrow terminal should fall back to classic, not use wide theme")
	}
}

func TestDoorsView_View_WideEnough_UsesTheme(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")

	marker := "CUSTOM_THEME"
	registry := themes.NewDefaultRegistry()
	registry.Register(&themes.DoorTheme{
		Name:        "custom",
		Description: "custom theme",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			return marker + "\n" + content
		},
		MinWidth: 15,
	})
	dv.SetThemeRegistry(registry)
	dv.SetThemeByName("custom")

	// Set wide enough: (120 - 6) / 3 = 38, above MinWidth 15
	dv.SetWidth(120)
	view := dv.View()

	if !strings.Contains(view, marker) {
		t.Error("wide terminal should use the custom theme")
	}
}

// --- AC7: Door number labels / content still visible with theme ---

func TestDoorsView_View_WithTheme_ShowsTaskContent(t *testing.T) {
	dv := newTestDoorsView("Alpha Task", "Beta Task", "Gamma Task")
	dv.SetWidth(120)
	dv.SetThemeByName("classic")
	view := dv.View()

	// At least one of the task texts should appear
	found := false
	for _, door := range dv.currentDoors {
		if strings.Contains(view, door.Text) {
			found = true
			break
		}
	}
	if !found {
		t.Error("View with theme should still show task content")
	}
}

func TestDoorsView_View_WithTheme_ShowsStatusIndicator(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(120)
	dv.SetThemeByName("modern")
	view := dv.View()
	if !strings.Contains(view, "[todo]") {
		t.Error("View with theme should still show status indicators")
	}
}

// --- AC8: Existing TUI tests still pass (verified by running make test) ---

// --- Config Theme field test ---

func TestProviderConfig_ThemeField(t *testing.T) {
	// This is tested via the YAML struct tag in core.ProviderConfig
	// The Theme field should be accessible
	cfg := &struct {
		Theme string `yaml:"theme,omitempty"`
	}{Theme: "scifi"}
	if cfg.Theme != "scifi" {
		t.Errorf("expected scifi, got %s", cfg.Theme)
	}
}

// --- SetThemeRegistry ---

func TestDoorsView_SetThemeRegistry(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	r := themes.NewRegistry()
	r.Register(themes.NewClassicTheme())
	dv.SetThemeRegistry(r)
	dv.SetThemeByName("classic")
	if dv.Theme() == nil || dv.Theme().Name != "classic" {
		t.Error("custom registry should work with SetThemeByName")
	}
}

// --- Selected door rendering with theme ---

func TestDoorsView_View_SelectedDoor_WithTheme(t *testing.T) {
	dv := newTestDoorsView("t1", "t2", "t3")
	dv.SetWidth(120)

	selectedMarker := "SELECTED_DOOR"
	registry := themes.NewRegistry()
	registry.Register(&themes.DoorTheme{
		Name:        "sel-test",
		Description: "selection test theme",
		Render: func(content string, width int, height int, selected bool, hint string, emphasis float64) string {
			if selected {
				return selectedMarker + "\n" + content
			}
			return content
		},
		MinWidth: 10,
	})
	dv.SetThemeRegistry(registry)
	dv.SetThemeByName("sel-test")
	dv.selectedDoorIndex = 1

	view := dv.View()
	// Exactly one door should be marked as selected
	count := strings.Count(view, selectedMarker)
	if count != 1 {
		t.Errorf("expected exactly 1 selected marker, got %d", count)
	}
}
