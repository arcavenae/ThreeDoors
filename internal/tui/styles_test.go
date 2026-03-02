package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStatusColor_AllStatuses(t *testing.T) {
	tests := []struct {
		status string
		expect lipgloss.Color
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
