package tui

import (
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/charmbracelet/bubbles/spinner"
)

func TestNewSyncSpinner(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	if s == nil {
		t.Fatal("NewSyncSpinner returned nil")
		return
	}
	if s.Active() {
		t.Error("spinner should not be active initially")
	}
	if s.View() != "" {
		t.Error("inactive spinner should return empty view")
	}
}

func TestSyncSpinnerStartStop(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()

	s.Start("Todoist")
	if !s.Active() {
		t.Error("spinner should be active after Start")
	}
	if s.ProviderName() != "Todoist" {
		t.Errorf("provider name should be 'Todoist', got %q", s.ProviderName())
	}

	s.Stop()
	if s.Active() {
		t.Error("spinner should not be active after Stop")
	}
	if s.View() != "" {
		t.Error("stopped spinner should return empty view")
	}
}

func TestSyncSpinnerViewWhenActiveAndPastThreshold(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	s.Start("Local")
	// Move start time past threshold
	s.startTime = time.Now().UTC().Add(-200 * time.Millisecond)
	view := s.View()
	if view == "" {
		t.Error("active spinner past threshold should return non-empty view")
	}
}

func TestSyncSpinnerUpdate(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	s.Start("Local")

	// Updating with a spinner.TickMsg should produce a command
	msg := spinner.TickMsg{ID: s.model.ID(), Time: time.Now().UTC()}
	cmd := s.Update(msg)
	if cmd == nil {
		t.Error("Update with TickMsg should return a command when active")
	}
}

func TestSyncSpinnerMultipleStartsKeepsLatest(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	s.Start("Alpha")
	s.Start("Beta")
	if s.ProviderName() != "Beta" {
		t.Errorf("should track latest provider, got %q", s.ProviderName())
	}
	if !s.Active() {
		t.Error("spinner should still be active")
	}
}

func TestSyncSpinnerStopIdempotent(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	s.Stop() // stop when not started — should not panic
	s.Start("Local")
	s.Stop()
	s.Stop() // stop again — should not panic
	if s.Active() {
		t.Error("spinner should not be active")
	}
}

func TestSyncSpinnerUsesExpectedStyle(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	// Verify MiniDot spinner style by comparing frame count
	if len(s.model.Spinner.Frames) != len(spinner.MiniDot.Frames) {
		t.Errorf("spinner should use MiniDot style (expected %d frames, got %d)",
			len(spinner.MiniDot.Frames), len(s.model.Spinner.Frames))
	}
}

func TestSyncSpinnerDelayThreshold(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	s.Start("Todoist")
	// Pin startTime to the future so threshold is guaranteed not elapsed
	s.startTime = time.Now().UTC().Add(time.Second)
	if s.ThresholdElapsed() {
		t.Error("threshold should not be elapsed immediately after start")
	}
}

func TestSyncSpinnerDelayThresholdAfterTime(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	s.Start("Todoist")
	// Manually set startTime to 200ms ago
	s.startTime = time.Now().UTC().Add(-200 * time.Millisecond)
	if !s.ThresholdElapsed() {
		t.Error("threshold should be elapsed after 200ms")
	}
}

func TestSyncSpinnerViewRespectsThreshold(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	s.Start("Todoist")
	// Pin startTime to the future so threshold is guaranteed not elapsed
	s.startTime = time.Now().UTC().Add(time.Second)
	if s.View() != "" {
		t.Error("spinner view should be empty before threshold elapsed")
	}

	// Set start time to past threshold
	s.startTime = time.Now().UTC().Add(-200 * time.Millisecond)
	if s.View() == "" {
		t.Error("spinner view should be non-empty after threshold elapsed")
	}
}

func TestRenderSyncStatusBarWithSpinner(t *testing.T) {
	t.Parallel()
	tracker := core.NewSyncStatusTracker()
	tracker.Register("Todoist")
	tracker.SetSyncing("Todoist")

	sp := NewSyncSpinner()
	sp.Start("Todoist")
	sp.startTime = time.Now().UTC().Add(-200 * time.Millisecond) // past threshold

	got := RenderSyncStatusBarWithSpinner(tracker, sp)
	if got == "" {
		t.Error("status bar with syncing provider should not be empty")
	}
}

func TestRenderSyncStatusBarWithSpinnerNilSpinner(t *testing.T) {
	t.Parallel()
	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")

	// Should fall back to regular rendering when spinner is nil.
	// Strip ANSI codes before comparing because lipgloss color profile
	// is global state that parallel tests may mutate.
	got := stripANSI(RenderSyncStatusBarWithSpinner(tracker, nil))
	want := stripANSI(RenderSyncStatusBar(tracker))
	if got != want {
		t.Errorf("nil spinner should fall back to regular render\ngot:  %q\nwant: %q", got, want)
	}
}

func TestRenderSyncStatusBarWithSpinnerInactive(t *testing.T) {
	t.Parallel()
	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")

	sp := NewSyncSpinner()
	// Not started — should fall back to regular rendering
	got := stripANSI(RenderSyncStatusBarWithSpinner(tracker, sp))
	want := stripANSI(RenderSyncStatusBar(tracker))
	if got != want {
		t.Errorf("inactive spinner should fall back to regular render\ngot:  %q\nwant: %q", got, want)
	}
}

func TestSyncSpinnerUpdateWhenInactive(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	// Update when inactive should return nil
	msg := spinner.TickMsg{Time: time.Now().UTC()}
	cmd := s.Update(msg)
	if cmd != nil {
		t.Error("Update when inactive should return nil")
	}
}

func TestSyncSpinnerThresholdWhenInactive(t *testing.T) {
	t.Parallel()
	s := NewSyncSpinner()
	if s.ThresholdElapsed() {
		t.Error("threshold should not be elapsed when inactive")
	}
}

func TestRenderSyncStatusBarWithSpinnerBelowThreshold(t *testing.T) {
	t.Parallel()
	tracker := core.NewSyncStatusTracker()
	tracker.Register("Todoist")
	tracker.SetSyncing("Todoist")

	sp := NewSyncSpinner()
	sp.Start("Todoist")
	// Pin startTime to the future so threshold is guaranteed not elapsed,
	// regardless of CI machine speed or scheduling delays.
	sp.startTime = time.Now().UTC().Add(time.Second)

	got := stripANSI(RenderSyncStatusBarWithSpinner(tracker, sp))
	want := stripANSI(RenderSyncStatusBar(tracker))
	if got != want {
		t.Errorf("below-threshold spinner should fall back to regular render\ngot:  %q\nwant: %q", got, want)
	}
}
