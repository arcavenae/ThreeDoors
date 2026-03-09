package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStatusColor_AllStatuses(t *testing.T) {
	tests := []struct {
		status string
		expect lipgloss.TerminalColor
	}{
		{"todo", colorTodo},
		{"in-progress", colorInProgress},
		{"blocked", colorBlocked},
		{"in-review", colorInReview},
		{"complete", colorComplete},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := StatusColor(tt.status)
			if got != tt.expect {
				t.Errorf("StatusColor(%q) = %v, want %v", tt.status, got, tt.expect)
			}
		})
	}
}

func TestStatusColor_UnknownDefaultsToTodo(t *testing.T) {
	got := StatusColor("unknown")
	if got != colorTodo {
		t.Errorf("StatusColor(\"unknown\") should default to colorTodo, got %v", got)
	}
}

// --- Story 1.6: Essential Polish - Message Pool Tests ---

func TestGreetingMessages_NonEmpty(t *testing.T) {
	if len(greetingMessages) == 0 {
		t.Fatal("greetingMessages pool must not be empty")
	}
}

func TestGreetingMessages_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, msg := range greetingMessages {
		if seen[msg] {
			t.Errorf("duplicate greeting message: %q", msg)
		}
		seen[msg] = true
	}
}

func TestGreetingMessages_AllNonEmpty(t *testing.T) {
	for i, msg := range greetingMessages {
		if msg == "" {
			t.Errorf("greetingMessages[%d] is empty", i)
		}
	}
}

func TestCelebrationMessages_NonEmpty(t *testing.T) {
	if len(celebrationMessages) == 0 {
		t.Fatal("celebrationMessages pool must not be empty")
	}
}

func TestCelebrationMessages_NoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, msg := range celebrationMessages {
		if seen[msg] {
			t.Errorf("duplicate celebration message: %q", msg)
		}
		seen[msg] = true
	}
}

func TestCelebrationMessages_AllNonEmpty(t *testing.T) {
	for i, msg := range celebrationMessages {
		if msg == "" {
			t.Errorf("celebrationMessages[%d] is empty", i)
		}
	}
}

func TestDoorColors_HasThreeEntries(t *testing.T) {
	if len(doorColors) != 3 {
		t.Errorf("doorColors should have exactly 3 entries, got %d", len(doorColors))
	}
}

func TestDoorColors_AllDistinct(t *testing.T) {
	if len(doorColors) < 3 {
		t.Skip("doorColors has fewer than 3 entries")
	}
	// Compare RGBA values to verify distinctness since TerminalColor is an interface.
	r0, g0, b0, _ := doorColors[0].RGBA()
	r1, g1, b1, _ := doorColors[1].RGBA()
	r2, g2, b2, _ := doorColors[2].RGBA()
	if r0 == r1 && g0 == g1 && b0 == b1 {
		t.Error("doorColors[0] and doorColors[1] must be distinct")
	}
	if r1 == r2 && g1 == g2 && b1 == b2 {
		t.Error("doorColors[1] and doorColors[2] must be distinct")
	}
	if r0 == r2 && g0 == g2 && b0 == b2 {
		t.Error("doorColors[0] and doorColors[2] must be distinct")
	}
}

func TestGreetingStyle_Exists(t *testing.T) {
	// Verify greetingStyle is defined and renders without panic
	result := greetingStyle.Render("test")
	if result == "" {
		t.Error("greetingStyle.Render should produce output")
	}
}
