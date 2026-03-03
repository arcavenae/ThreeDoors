package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

func TestRenderSyncStatusBarNilTracker(t *testing.T) {
	t.Parallel()
	got := RenderSyncStatusBar(nil)
	if got != "" {
		t.Errorf("expected empty string for nil tracker, got %q", got)
	}
}

func TestRenderSyncStatusBarEmptyTracker(t *testing.T) {
	t.Parallel()
	tracker := core.NewSyncStatusTracker()
	got := RenderSyncStatusBar(tracker)
	if got != "" {
		t.Errorf("expected empty string for empty tracker, got %q", got)
	}
}

func TestRenderSyncStatusBarSingleProvider(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")

	got := RenderSyncStatusBar(tracker)
	if !strings.Contains(got, "✓") {
		t.Errorf("synced provider should show ✓, got %q", got)
	}
	if !strings.Contains(got, "Local") {
		t.Errorf("should contain provider name 'Local', got %q", got)
	}
}

func TestRenderSyncStatusBarMultipleProviders(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.Register("WAL")

	got := RenderSyncStatusBar(tracker)
	if !strings.Contains(got, "Local") {
		t.Errorf("should contain 'Local', got %q", got)
	}
	if !strings.Contains(got, "WAL") {
		t.Errorf("should contain 'WAL', got %q", got)
	}
}

func TestRenderSyncStatusBarPendingState(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("WAL")
	tracker.SetPending("WAL", 3)

	got := RenderSyncStatusBar(tracker)
	if !strings.Contains(got, "⏳") {
		t.Errorf("pending provider should show ⏳, got %q", got)
	}
	if !strings.Contains(got, "(3)") {
		t.Errorf("pending provider should show count '(3)', got %q", got)
	}
}

func TestRenderSyncStatusBarErrorState(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.SetError("Local", "connection refused")

	got := RenderSyncStatusBar(tracker)
	if !strings.Contains(got, "✗") {
		t.Errorf("error provider should show ✗, got %q", got)
	}
}

func TestRenderSyncStatusBarSyncingState(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.SetSyncing("Local")

	got := RenderSyncStatusBar(tracker)
	if !strings.Contains(got, "↻") {
		t.Errorf("syncing provider should show ↻, got %q", got)
	}
}

func TestRenderSyncStatusBarSyncedWithTimestamp(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.SetSynced("Local")

	got := RenderSyncStatusBar(tracker)
	if !strings.Contains(got, "just now") {
		t.Errorf("recently synced should show 'just now', got %q", got)
	}
}

func TestFormatSyncAge(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{"just now", now.Add(-10 * time.Second), "just now"},
		{"1 minute ago", now.Add(-90 * time.Second), "1m ago"},
		{"5 minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"1 hour ago", now.Add(-90 * time.Minute), "1h ago"},
		{"3 hours ago", now.Add(-3 * time.Hour), "3h ago"},
		{"1 day ago", now.Add(-25 * time.Hour), "1d ago"},
		{"3 days ago", now.Add(-72 * time.Hour), "3d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatSyncAge(tt.time)
			if got != tt.want {
				t.Errorf("formatSyncAge() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderSyncStatusBarDeterministicOrder(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Zebra")
	tracker.Register("Alpha")

	got := RenderSyncStatusBar(tracker)
	alphaIdx := strings.Index(got, "Alpha")
	zebraIdx := strings.Index(got, "Zebra")
	if alphaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("expected both provider names in output, got %q", got)
	}
	if alphaIdx >= zebraIdx {
		t.Errorf("providers should be sorted alphabetically: Alpha before Zebra, got Alpha@%d Zebra@%d", alphaIdx, zebraIdx)
	}
}
