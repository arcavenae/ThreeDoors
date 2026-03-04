package tui

import (
	"strings"
	"testing"

	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	tea "github.com/charmbracelet/bubbletea"
)

func TestThemePicker_InitialState(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	if tp.cursor != 0 {
		t.Errorf("cursor = %d, want 0", tp.cursor)
	}
	if tp.confirmed {
		t.Error("should not be confirmed initially")
	}

	view := tp.View()
	if !strings.Contains(view, "Theme") {
		t.Error("view should contain 'Theme'")
	}
}

func TestThemePicker_NavigateRight(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	initialCursor := tp.cursor
	tp.Update(tea.KeyMsg{Type: tea.KeyRight})

	if tp.cursor != initialCursor+1 {
		t.Errorf("cursor = %d, want %d after right", tp.cursor, initialCursor+1)
	}
}

func TestThemePicker_NavigateLeft(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	// Move right first, then left
	tp.Update(tea.KeyMsg{Type: tea.KeyRight})
	tp.Update(tea.KeyMsg{Type: tea.KeyLeft})

	if tp.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after right+left", tp.cursor)
	}
}

func TestThemePicker_NavigateWrapsAround(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	// Left from 0 should wrap to last
	tp.Update(tea.KeyMsg{Type: tea.KeyLeft})
	expected := len(tp.themeNames) - 1
	if tp.cursor != expected {
		t.Errorf("cursor = %d, want %d after wrapping left", tp.cursor, expected)
	}

	// Right from last should wrap to 0
	tp.Update(tea.KeyMsg{Type: tea.KeyRight})
	if tp.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after wrapping right", tp.cursor)
	}
}

func TestThemePicker_ConfirmWithEnter(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	cmd := tp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on Enter")
	}

	msg := cmd()
	selected, ok := msg.(ThemeSelectedMsg)
	if !ok {
		t.Fatalf("expected ThemeSelectedMsg, got %T", msg)
	}

	if selected.ThemeName == "" {
		t.Error("expected non-empty theme name")
	}
	if !tp.confirmed {
		t.Error("should be confirmed after Enter")
	}
}

func TestThemePicker_SkipWithEscape(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	// Move to a non-default theme first
	tp.Update(tea.KeyMsg{Type: tea.KeyRight})

	cmd := tp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command on Escape")
	}

	msg := cmd()
	selected, ok := msg.(ThemeSelectedMsg)
	if !ok {
		t.Fatalf("expected ThemeSelectedMsg, got %T", msg)
	}

	// Escape should default to modern
	if selected.ThemeName != themes.DefaultThemeName {
		t.Errorf("theme = %q, want %q on escape", selected.ThemeName, themes.DefaultThemeName)
	}
	if !selected.Skipped {
		t.Error("expected Skipped=true on escape")
	}
}

func TestThemePicker_ViewShowsThemeName(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	view := tp.View()
	// Should show the current theme's name
	currentName := tp.themeNames[tp.cursor]
	if !strings.Contains(view, currentName) {
		t.Errorf("view should contain theme name %q", currentName)
	}
}

func TestThemePicker_ViewShowsPreview(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	view := tp.View()
	// The preview should render something (door frame chars)
	if !strings.Contains(view, "Sample Task") {
		t.Error("view should contain sample task preview")
	}
}

func TestThemePicker_NarrowTerminalFallback(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(30) // Very narrow

	view := tp.View()
	// Should still render without crashing
	if view == "" {
		t.Error("view should not be empty even at narrow width")
	}
}

func TestThemePicker_SetWidth(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(100)

	if tp.width != 100 {
		t.Errorf("width = %d, want 100", tp.width)
	}
}

func TestThemePicker_AllThemesAccessible(t *testing.T) {
	t.Parallel()

	reg := themes.NewDefaultRegistry()
	tp := NewThemePicker(reg)
	tp.SetWidth(80)

	names := reg.Names()
	seen := make(map[string]bool)

	for range names {
		seen[tp.themeNames[tp.cursor]] = true
		tp.Update(tea.KeyMsg{Type: tea.KeyRight})
	}

	for _, name := range names {
		if !seen[name] {
			t.Errorf("theme %q was not accessible via navigation", name)
		}
	}
}
