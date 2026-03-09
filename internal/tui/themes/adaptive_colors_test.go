package themes

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestThemeColors_FieldsAreTerminalColor(t *testing.T) {
	t.Parallel()

	registry := NewDefaultRegistry()
	for _, name := range registry.Names() {
		theme, _ := registry.Get(name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if theme.Colors.Frame == nil {
				t.Error("Frame color must not be nil")
			}
			if theme.Colors.Fill == nil {
				t.Error("Fill color must not be nil")
			}
			if theme.Colors.Accent == nil {
				t.Error("Accent color must not be nil")
			}
			if theme.Colors.Selected == nil {
				t.Error("Selected color must not be nil")
			}
		})
	}
}

func TestThemeColors_UseCompleteColor(t *testing.T) {
	t.Parallel()

	registry := NewDefaultRegistry()
	for _, name := range registry.Names() {
		theme, _ := registry.Get(name)
		t.Run(name+"/Frame", func(t *testing.T) {
			t.Parallel()
			cc, ok := theme.Colors.Frame.(lipgloss.CompleteColor)
			if !ok {
				t.Errorf("Frame should be CompleteColor, got %T", theme.Colors.Frame)
				return
			}
			if cc.ANSI256 == "" {
				t.Error("Frame.ANSI256 must not be empty")
			}
			if cc.ANSI == "" {
				t.Error("Frame.ANSI must not be empty")
			}
		})
		t.Run(name+"/Selected", func(t *testing.T) {
			t.Parallel()
			cc, ok := theme.Colors.Selected.(lipgloss.CompleteColor)
			if !ok {
				t.Errorf("Selected should be CompleteColor, got %T", theme.Colors.Selected)
				return
			}
			if cc.ANSI256 == "" {
				t.Error("Selected.ANSI256 must not be empty")
			}
			if cc.ANSI == "" {
				t.Error("Selected.ANSI must not be empty")
			}
		})
	}
}

func TestThemeColors_RenderWithoutPanic(t *testing.T) {
	t.Parallel()

	registry := NewDefaultRegistry()
	for _, name := range registry.Names() {
		theme, _ := registry.Get(name)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Rendering with theme colors should not panic
			style := lipgloss.NewStyle().Foreground(theme.Colors.Frame)
			result := style.Render("test")
			if result == "" {
				t.Error("render with Frame color produced empty output")
			}

			style = lipgloss.NewStyle().Foreground(theme.Colors.Selected)
			result = style.Render("test")
			if result == "" {
				t.Error("render with Selected color produced empty output")
			}
		})
	}
}
