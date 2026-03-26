package todoist

import (
	"testing"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestMapPriorityToEffort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priority int
		want     core.TaskEffort
	}{
		{"priority 0 (none) maps to quick-win", 0, core.EffortQuickWin},
		{"priority 1 (normal) maps to quick-win", 1, core.EffortQuickWin},
		{"priority 2 (high) maps to medium", 2, core.EffortMedium},
		{"priority 3 (urgent) maps to deep-work", 3, core.EffortDeepWork},
		{"priority 4 (critical) maps to deep-work", 4, core.EffortDeepWork},
		{"negative priority defaults to quick-win", -1, core.EffortQuickWin},
		{"out-of-range priority defaults to quick-win", 99, core.EffortQuickWin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapPriorityToEffort(tt.priority)
			if got != tt.want {
				t.Errorf("MapPriorityToEffort(%d) = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

func TestMapStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		isCompleted bool
		want        core.TaskStatus
	}{
		{"completed maps to complete", true, core.StatusComplete},
		{"not completed maps to todo", false, core.StatusTodo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapStatus(tt.isCompleted)
			if got != tt.want {
				t.Errorf("MapStatus(%v) = %q, want %q", tt.isCompleted, got, tt.want)
			}
		})
	}
}
