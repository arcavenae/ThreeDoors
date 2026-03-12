package themes

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestNewClassicTheme(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	if theme.Name != "classic" {
		t.Errorf("got name %q, want %q", theme.Name, "classic")
	}
	if theme.Description == "" {
		t.Error("expected non-empty description")
	}
	if theme.Render == nil {
		t.Fatal("expected non-nil Render function")
		return
	}
	if theme.MinWidth < 1 {
		t.Errorf("expected positive MinWidth, got %d", theme.MinWidth)
	}
}

func TestClassicThemeColors(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()

	if theme.Colors.Frame == nil {
		t.Error("expected non-nil Frame color")
	}
	if theme.Colors.Selected == nil {
		t.Error("expected non-nil Selected color")
	}
}

func TestClassicRenderUnselected(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()
	output := theme.Render("Test task", 30, 0, false, "", 0.0)

	if !strings.Contains(output, "Test task") {
		t.Error("rendered output should contain the content text")
	}
	if output == "" {
		t.Error("rendered output should not be empty")
	}
}

func TestClassicRenderSelected(t *testing.T) {
	t.Parallel()

	theme := NewClassicTheme()
	unselected := theme.Render("Test task", 30, 0, false, "", 0.0)
	selected := theme.Render("Test task", 30, 0, true, "", 0.0)

	if selected == "" {
		t.Error("selected output should not be empty")
	}
	if !strings.Contains(selected, "Test task") {
		t.Error("selected output should contain the content text")
	}
	if selected == unselected {
		t.Error("selected and unselected output should differ")
	}
}

func TestClassicRenderMatchesExistingStyle(t *testing.T) {
	t.Parallel()

	// These styles match the existing doorStyle/selectedDoorStyle from internal/tui/styles.go,
	// now using CompleteColor for adaptive color profile support.
	colorAccent := lipgloss.CompleteColor{TrueColor: "#5f5fff", ANSI256: "63", ANSI: "5"}
	colorDoorBright := lipgloss.CompleteColor{TrueColor: "#eeeeee", ANSI256: "255", ANSI: "15"}

	existingDoorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2)

	existingSelectedStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(colorDoorBright).
		Padding(1, 2)

	theme := NewClassicTheme()
	width := 30
	content := "Test task text"

	// Unselected should match existing doorStyle.
	// Compare with ANSI codes stripped — lipgloss color rendering can vary
	// across environments (CI vs local terminal), but structural content
	// (box-drawing characters, spacing, text) must match.
	themeOutput := ansi.Strip(theme.Render(content, width, 0, false, "", 0.0))
	existingOutput := ansi.Strip(existingDoorStyle.Width(width).Render(content))

	if themeOutput != existingOutput {
		t.Errorf("classic unselected does not match existing doorStyle\ngot:\n%s\nwant:\n%s", themeOutput, existingOutput)
	}

	// Selected should match existing selectedDoorStyle
	themeSelectedOutput := ansi.Strip(theme.Render(content, width, 0, true, "", 0.0))
	existingSelectedOutput := ansi.Strip(existingSelectedStyle.Width(width).Render(content))

	if themeSelectedOutput != existingSelectedOutput {
		t.Errorf("classic selected does not match existing selectedDoorStyle\ngot:\n%s\nwant:\n%s", themeSelectedOutput, existingSelectedOutput)
	}
}

func TestClassicRenderVaryingWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 15},
		{"medium", 30},
		{"wide", 50},
	}

	theme := NewClassicTheme()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := theme.Render("Task", tt.width, 0, false, "", 0.0)
			if output == "" {
				t.Error("output should not be empty")
			}
			if !strings.Contains(output, "Task") {
				t.Error("output should contain the task text")
			}
		})
	}
}
