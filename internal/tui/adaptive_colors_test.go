package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestAdaptiveColors_StatusColorsAreTerminalColor(t *testing.T) {
	t.Parallel()

	// All status color vars must implement lipgloss.TerminalColor.
	colors := []struct {
		name  string
		color lipgloss.TerminalColor
	}{
		{"colorTodo", colorTodo},
		{"colorInProgress", colorInProgress},
		{"colorBlocked", colorBlocked},
		{"colorInReview", colorInReview},
		{"colorComplete", colorComplete},
		{"colorAccent", colorAccent},
		{"colorSelected", colorSelected},
		{"colorGreeting", colorGreeting},
		{"colorDoorBright", colorDoorBright},
	}

	for _, tc := range colors {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.color == nil {
				t.Errorf("%s must not be nil", tc.name)
			}
		})
	}
}

func TestAdaptiveColors_DoorColorsAreTerminalColor(t *testing.T) {
	t.Parallel()

	if len(doorColors) != 3 {
		t.Fatalf("doorColors should have 3 entries, got %d", len(doorColors))
	}

	for i, c := range doorColors {
		if c == nil {
			t.Errorf("doorColors[%d] must not be nil", i)
		}
	}
}

func TestAdaptiveColors_StatusColorReturnsTerminalColor(t *testing.T) {
	t.Parallel()

	statuses := []string{
		"todo", "in-progress", "blocked", "in-review",
		"complete", "deferred", "archived", "unknown",
	}

	for _, s := range statuses {
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			c := StatusColor(s)
			if c == nil {
				t.Errorf("StatusColor(%q) must not return nil", s)
			}
		})
	}
}

func TestAdaptiveColors_StylesRenderWithoutPanic(t *testing.T) {
	t.Parallel()

	// Verify that all styles using adaptive colors render without panic.
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"doorStyle", doorStyle},
		{"selectedDoorStyle", selectedDoorStyle},
		{"headerStyle", headerStyle},
		{"flashStyle", flashStyle},
		{"helpStyle", helpStyle},
		{"searchResultStyle", searchResultStyle},
		{"searchSelectedStyle", searchSelectedStyle},
		{"greetingStyle", greetingStyle},
		{"separatorStyle", separatorStyle},
		{"healthOKStyle", healthOKStyle},
		{"healthFailStyle", healthFailStyle},
		{"healthWarnStyle", healthWarnStyle},
	}

	for _, tc := range styles {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.style.Render("test content")
			if result == "" {
				t.Errorf("%s.Render produced empty output", tc.name)
			}
		})
	}
}

func TestAdaptiveColors_CompleteColorHasAllProfiles(t *testing.T) {
	t.Parallel()

	// Verify high-visibility colors use CompleteColor with all three profiles.
	colors := []struct {
		name  string
		color lipgloss.TerminalColor
	}{
		{"colorTodo", colorTodo},
		{"colorInProgress", colorInProgress},
		{"colorBlocked", colorBlocked},
		{"colorInReview", colorInReview},
		{"colorComplete", colorComplete},
		{"colorAccent", colorAccent},
		{"colorSelected", colorSelected},
		{"colorDoorBright", colorDoorBright},
	}

	for _, tc := range colors {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cc, ok := tc.color.(lipgloss.CompleteColor)
			if !ok {
				t.Errorf("%s should be lipgloss.CompleteColor, got %T", tc.name, tc.color)
				return
			}
			if cc.ANSI256 == "" {
				t.Errorf("%s.ANSI256 must not be empty", tc.name)
			}
			if cc.ANSI == "" {
				t.Errorf("%s.ANSI must not be empty", tc.name)
			}
		})
	}
}
