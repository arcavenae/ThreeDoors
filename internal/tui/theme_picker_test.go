package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestRegistry() *themes.Registry {
	return themes.NewDefaultRegistry()
}

func TestNewThemePicker(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()

	tp := NewThemePicker(reg, "modern")

	if tp == nil {
		t.Fatal("NewThemePicker returned nil")
		return
	}
	if len(tp.themeNames) == 0 {
		t.Fatal("expected theme names to be populated")
	}
	if tp.currentTheme != "modern" {
		t.Errorf("got currentTheme=%q, want %q", tp.currentTheme, "modern")
	}
}

func TestThemePickerCursorStartsOnCurrentTheme(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()

	tests := []struct {
		name         string
		currentTheme string
	}{
		{"starts on modern", "modern"},
		{"starts on classic", "classic"},
		{"starts on scifi", "scifi"},
		{"starts on shoji", "shoji"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tp := NewThemePicker(reg, tt.currentTheme)
			selectedName := tp.themeNames[tp.cursor]
			if selectedName != tt.currentTheme {
				t.Errorf("cursor at %q, want %q", selectedName, tt.currentTheme)
			}
		})
	}
}

func TestThemePickerCursorFallbackOnUnknown(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()

	tp := NewThemePicker(reg, "nonexistent")
	if tp.cursor != 0 {
		t.Errorf("expected cursor=0 for unknown theme, got %d", tp.cursor)
	}
}

func TestThemePickerNavigateLeftRight(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, tp_firstName(reg))

	startCursor := tp.cursor
	// Move right
	tp.Update(tea.KeyMsg{Type: tea.KeyRight})
	if tp.cursor != startCursor+1 {
		t.Errorf("after right: cursor=%d, want %d", tp.cursor, startCursor+1)
	}
	// Move left
	tp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if tp.cursor != startCursor {
		t.Errorf("after left: cursor=%d, want %d", tp.cursor, startCursor)
	}
}

func TestThemePickerNavigateUpDown(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, tp_firstName(reg))

	startCursor := tp.cursor
	tp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if tp.cursor != startCursor+1 {
		t.Errorf("after down: cursor=%d, want %d", tp.cursor, startCursor+1)
	}
	tp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if tp.cursor != startCursor {
		t.Errorf("after up: cursor=%d, want %d", tp.cursor, startCursor)
	}
}

func TestThemePickerWrapAround(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, tp_firstName(reg))

	// Move left past first element — should stay at 0
	tp.cursor = 0
	tp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if tp.cursor != 0 {
		t.Errorf("cursor went below 0: %d", tp.cursor)
	}

	// Move right past last element — should stay at last
	tp.cursor = len(tp.themeNames) - 1
	tp.Update(tea.KeyMsg{Type: tea.KeyRight})
	if tp.cursor != len(tp.themeNames)-1 {
		t.Errorf("cursor went past last: %d", tp.cursor)
	}
}

func TestThemePickerEnterSelectsTheme(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, tp_firstName(reg))

	tp.cursor = 1
	cmd := tp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Enter")
		return
	}

	msg := cmd()
	selected, ok := msg.(ThemeSelectedMsg)
	if !ok {
		t.Fatalf("expected ThemeSelectedMsg, got %T", msg)
	}
	if selected.Name != tp.themeNames[1] {
		t.Errorf("selected %q, want %q", selected.Name, tp.themeNames[1])
	}
}

func TestThemePickerEscapeCancels(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, "modern")

	cmd := tp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Escape")
		return
	}

	msg := cmd()
	if _, ok := msg.(ThemeCancelledMsg); !ok {
		t.Fatalf("expected ThemeCancelledMsg, got %T", msg)
	}
}

func TestThemePickerViewShowsCurrentIndicator(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, "modern")
	tp.SetWidth(80)

	view := tp.View()
	if !strings.Contains(view, "[current]") {
		t.Error("expected [current] indicator in view output")
	}
}

func TestThemePickerViewShowsNonSeasonalThemes(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, "modern")
	tp.SetWidth(80)

	view := tp.View()
	for _, name := range reg.NonSeasonalNames() {
		if !strings.Contains(view, name) {
			t.Errorf("non-seasonal theme %q not found in view output", name)
		}
	}
	for _, name := range reg.SeasonalNames() {
		if strings.Contains(view, name) {
			t.Errorf("seasonal theme %q should not appear in non-seasonal picker", name)
		}
	}
}

func TestThemePickerViewShowsCursorIndicator(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, tp_firstName(reg))
	tp.SetWidth(80)

	view := tp.View()
	if !strings.Contains(view, "▸") {
		t.Error("expected cursor indicator ▸ in view output")
	}
}

// --- Seasonal theme picker tests ---

func TestNewSeasonalThemePicker(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()

	tp := NewSeasonalThemePicker(reg, "winter")

	if tp == nil {
		t.Fatal("NewSeasonalThemePicker returned nil")
	}
	if !tp.IsSeasonal() {
		t.Error("expected IsSeasonal() to be true")
	}
	if tp.title != "Select Seasonal Theme" {
		t.Errorf("got title=%q, want %q", tp.title, "Select Seasonal Theme")
	}

	// Should only contain seasonal themes
	expected := []string{"autumn", "spring", "summer", "winter"}
	if len(tp.themeNames) != len(expected) {
		t.Fatalf("got %d themes, want %d: %v", len(tp.themeNames), len(expected), tp.themeNames)
	}
	for i, name := range expected {
		if tp.themeNames[i] != name {
			t.Errorf("themeNames[%d] = %q, want %q", i, tp.themeNames[i], name)
		}
	}
}

func TestNewThemePickerExcludesSeasonal(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()

	tp := NewThemePicker(reg, "modern")

	// Should only contain non-seasonal themes
	for _, name := range tp.themeNames {
		theme, ok := reg.Get(name)
		if !ok {
			t.Errorf("theme %q not in registry", name)
			continue
		}
		if theme.Season != "" {
			t.Errorf("non-seasonal picker contains seasonal theme %q (season=%q)", name, theme.Season)
		}
	}
}

func TestSeasonalThemePickerCursorOnCurrent(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()

	tests := []struct {
		name    string
		current string
	}{
		{"winter", "winter"},
		{"spring", "spring"},
		{"summer", "summer"},
		{"autumn", "autumn"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tp := NewSeasonalThemePicker(reg, tt.current)
			selectedName := tp.themeNames[tp.cursor]
			if selectedName != tt.current {
				t.Errorf("cursor at %q, want %q", selectedName, tt.current)
			}
		})
	}
}

func TestSeasonalThemePickerEnterSelects(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewSeasonalThemePicker(reg, "winter")

	// Move to second seasonal theme
	tp.cursor = 1
	cmd := tp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Enter")
	}

	msg := cmd()
	selected, ok := msg.(ThemeSelectedMsg)
	if !ok {
		t.Fatalf("expected ThemeSelectedMsg, got %T", msg)
	}
	if selected.Name != tp.themeNames[1] {
		t.Errorf("selected %q, want %q", selected.Name, tp.themeNames[1])
	}
}

func TestSeasonalThemePickerEscapeCancels(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewSeasonalThemePicker(reg, "winter")

	cmd := tp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Escape")
	}

	msg := cmd()
	if _, ok := msg.(ThemeCancelledMsg); !ok {
		t.Fatalf("expected ThemeCancelledMsg, got %T", msg)
	}
}

func TestSeasonalThemePickerViewShowsTitle(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewSeasonalThemePicker(reg, "winter")
	tp.SetWidth(80)

	view := tp.View()
	if !strings.Contains(view, "Select Seasonal Theme") {
		t.Error("expected 'Select Seasonal Theme' title in view output")
	}
}

func TestSeasonalThemePickerViewShowsOnlySeasonalThemes(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewSeasonalThemePicker(reg, "winter")
	tp.SetWidth(80)

	view := tp.View()
	seasonal := []string{"autumn", "spring", "summer", "winter"}
	for _, name := range seasonal {
		if !strings.Contains(view, name) {
			t.Errorf("seasonal theme %q not found in view output", name)
		}
	}
	nonSeasonal := []string{"classic", "modern", "scifi", "shoji"}
	for _, name := range nonSeasonal {
		if strings.Contains(view, name) {
			t.Errorf("non-seasonal theme %q should not appear in seasonal picker view", name)
		}
	}
}

func TestNonSeasonalPickerIsNotSeasonal(t *testing.T) {
	t.Parallel()
	reg := newTestRegistry()
	tp := NewThemePicker(reg, "modern")

	if tp.IsSeasonal() {
		t.Error("expected IsSeasonal() to be false for non-seasonal picker")
	}
}

// tp_firstName returns the first theme name from the registry (sorted).
func tp_firstName(reg *themes.Registry) string {
	names := reg.Names()
	if len(names) == 0 {
		return ""
	}
	return names[0]
}
