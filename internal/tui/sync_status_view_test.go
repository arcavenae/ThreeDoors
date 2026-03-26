package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
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

func TestRenderProviderStatusCircuitOpen(t *testing.T) {
	t.Parallel()

	s := core.ProviderSyncStatus{
		Name:         "jira",
		Phase:        core.SyncPhaseError,
		CircuitState: core.CircuitOpen,
		RetryIn:      2 * time.Minute,
	}
	got := renderProviderStatus(s)
	if !strings.Contains(got, "✗") {
		t.Errorf("circuit open should show ✗, got %q", got)
	}
	if !strings.Contains(got, "jira") {
		t.Errorf("should contain provider name, got %q", got)
	}
	if !strings.Contains(got, "retry in 2m") {
		t.Errorf("should contain retry info, got %q", got)
	}
}

func TestRenderProviderStatusCircuitHalfOpen(t *testing.T) {
	t.Parallel()

	s := core.ProviderSyncStatus{
		Name:         "reminders",
		CircuitState: core.CircuitHalfOpen,
	}
	got := renderProviderStatus(s)
	if !strings.Contains(got, "↻") {
		t.Errorf("half-open should show ↻, got %q", got)
	}
	if !strings.Contains(got, "probing") {
		t.Errorf("half-open should show probing, got %q", got)
	}
}

func TestRenderProviderStatusStale(t *testing.T) {
	t.Parallel()

	s := core.ProviderSyncStatus{
		Name:         "textfile",
		Phase:        core.SyncPhaseSynced,
		CircuitState: core.CircuitClosed,
		StaleSince:   time.Now().UTC().Add(-10 * time.Minute),
		LastSyncTime: time.Now().UTC().Add(-10 * time.Minute),
	}
	got := renderProviderStatus(s)
	if !strings.Contains(got, "stale") {
		t.Errorf("stale provider should show staleness, got %q", got)
	}
}

func TestRenderWALPendingNoItems(t *testing.T) {
	t.Parallel()

	statuses := []core.ProviderSyncStatus{
		{Name: "Local", PendingCount: 0},
	}
	got := renderWALPending(statuses)
	if got != "" {
		t.Errorf("expected empty string for no pending, got %q", got)
	}
}

func TestRenderWALPendingWithItems(t *testing.T) {
	t.Parallel()

	statuses := []core.ProviderSyncStatus{
		{Name: "Local", PendingCount: 3, OldestPending: time.Now().UTC().Add(-5 * time.Minute)},
		{Name: "Remote", PendingCount: 2},
	}
	got := renderWALPending(statuses)
	if !strings.Contains(got, "WAL pending") {
		t.Errorf("should contain 'WAL pending', got %q", got)
	}
	if !strings.Contains(got, "5 items") {
		t.Errorf("should aggregate pending count to 5, got %q", got)
	}
}

func TestRenderSyncStatusBarWithWALPending(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("WAL")
	tracker.SetPending("WAL", 4)

	got := RenderSyncStatusBar(tracker)
	if !strings.Contains(got, "WAL pending") {
		t.Errorf("should show WAL pending line, got %q", got)
	}
}

func TestRenderOfflineInfoOnline(t *testing.T) {
	t.Parallel()

	info := OfflineInfo{Online: true, UnpushedCount: 0}
	got := renderOfflineInfo(info)
	if !strings.Contains(got, "online") {
		t.Errorf("should contain 'online', got %q", got)
	}
	if !strings.Contains(got, "●") {
		t.Errorf("online should show filled circle ●, got %q", got)
	}
}

func TestRenderOfflineInfoOffline(t *testing.T) {
	t.Parallel()

	info := OfflineInfo{Online: false, UnpushedCount: 3}
	got := renderOfflineInfo(info)
	if !strings.Contains(got, "offline") {
		t.Errorf("should contain 'offline', got %q", got)
	}
	if !strings.Contains(got, "○") {
		t.Errorf("offline should show empty circle ○, got %q", got)
	}
	if !strings.Contains(got, "3 queued") {
		t.Errorf("should show queue depth '3 queued', got %q", got)
	}
}

func TestRenderOfflineInfoProbing(t *testing.T) {
	t.Parallel()

	info := OfflineInfo{Online: true, Probing: true}
	got := renderOfflineInfo(info)
	if !strings.Contains(got, "probing") {
		t.Errorf("should contain 'probing', got %q", got)
	}
	if !strings.Contains(got, "↻") {
		t.Errorf("probing should show ↻, got %q", got)
	}
}

func TestRenderOfflineInfoWithLastSync(t *testing.T) {
	t.Parallel()

	info := OfflineInfo{Online: true, LastSyncTime: time.Now().UTC().Add(-5 * time.Minute)}
	got := renderOfflineInfo(info)
	if !strings.Contains(got, "5m ago") {
		t.Errorf("should show last sync time '5m ago', got %q", got)
	}
}

func TestRenderSyncStatusBarFullWithOfflineInfo(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")
	info := &OfflineInfo{Online: true, UnpushedCount: 2}

	got := RenderSyncStatusBarFull(tracker, info)
	if !strings.Contains(got, "online") {
		t.Errorf("should contain 'online', got %q", got)
	}
	if !strings.Contains(got, "2 queued") {
		t.Errorf("should show queue depth, got %q", got)
	}
	if !strings.Contains(got, "Local") {
		t.Errorf("should still show provider, got %q", got)
	}
}

func TestRenderSyncStatusBarFullNilInfo(t *testing.T) {
	t.Parallel()

	tracker := core.NewSyncStatusTracker()
	tracker.Register("Local")

	got := RenderSyncStatusBarFull(tracker, nil)
	// Should behave like RenderSyncStatusBar
	want := RenderSyncStatusBar(tracker)
	if got != want {
		t.Errorf("RenderSyncStatusBarFull(tracker, nil) should match RenderSyncStatusBar(tracker)")
	}
}

func TestRenderSyncStatusBarFullOnlyOfflineInfo(t *testing.T) {
	t.Parallel()

	info := &OfflineInfo{Online: false, UnpushedCount: 5}
	got := RenderSyncStatusBarFull(nil, info)
	if !strings.Contains(got, "offline") {
		t.Errorf("should contain 'offline', got %q", got)
	}
	if !strings.Contains(got, "5 queued") {
		t.Errorf("should show queue depth, got %q", got)
	}
}

func TestStyleIconCircuitStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status core.ProviderSyncStatus
	}{
		{"circuit open", core.ProviderSyncStatus{CircuitState: core.CircuitOpen}},
		{"circuit half-open", core.ProviderSyncStatus{CircuitState: core.CircuitHalfOpen}},
		{"circuit closed synced", core.ProviderSyncStatus{CircuitState: core.CircuitClosed, Phase: core.SyncPhaseSynced}},
		{"circuit closed error", core.ProviderSyncStatus{CircuitState: core.CircuitClosed, Phase: core.SyncPhaseError}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			icon := tt.status.Icon()
			got := styleIcon(icon, tt.status)
			if got == "" {
				t.Error("styleIcon returned empty string")
			}
		})
	}
}
