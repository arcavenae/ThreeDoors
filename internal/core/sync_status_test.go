package core

import (
	"testing"
	"time"
)

func TestProviderSyncStatusIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		phase SyncPhase
		want  string
	}{
		{"synced icon", SyncPhaseSynced, "✓"},
		{"syncing icon", SyncPhaseSyncing, "↻"},
		{"pending icon", SyncPhasePending, "⏳"},
		{"error icon", SyncPhaseError, "✗"},
		{"unknown icon", SyncPhase("unknown"), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := ProviderSyncStatus{Phase: tt.phase}
			got := s.Icon()
			if got != tt.want {
				t.Errorf("Icon() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProviderSyncStatusText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		status       ProviderSyncStatus
		wantContains string
	}{
		{
			"synced text",
			ProviderSyncStatus{Name: "Local", Phase: SyncPhaseSynced},
			"✓ Local synced",
		},
		{
			"syncing text",
			ProviderSyncStatus{Name: "WAL", Phase: SyncPhaseSyncing},
			"↻ WAL syncing",
		},
		{
			"pending text with count",
			ProviderSyncStatus{Name: "WAL", Phase: SyncPhasePending, PendingCount: 3},
			"⏳ WAL pending (3 items)",
		},
		{
			"error text",
			ProviderSyncStatus{Name: "Local", Phase: SyncPhaseError},
			"✗ Local error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.status.StatusText()
			if got != tt.wantContains {
				t.Errorf("StatusText() = %q, want %q", got, tt.wantContains)
			}
		})
	}
}

func TestSyncStatusTrackerRegisterAndGet(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("Local")

	got := tracker.Get("Local")
	if got == nil {
		t.Fatal("expected status for 'Local', got nil")
	}
	if got.Name != "Local" {
		t.Errorf("Name = %q, want %q", got.Name, "Local")
	}
	if got.Phase != SyncPhaseSynced {
		t.Errorf("Phase = %q, want %q", got.Phase, SyncPhaseSynced)
	}
}

func TestSyncStatusTrackerGetUnregistered(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	got := tracker.Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil for unregistered provider, got %+v", got)
	}
}

func TestSyncStatusTrackerSetSyncing(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.SetSyncing("Local")

	got := tracker.Get("Local")
	if got.Phase != SyncPhaseSyncing {
		t.Errorf("Phase = %q, want %q", got.Phase, SyncPhaseSyncing)
	}
}

func TestSyncStatusTrackerSetSynced(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("WAL")
	tracker.SetPending("WAL", 5) // set pending first
	tracker.SetSynced("WAL")

	got := tracker.Get("WAL")
	if got.Phase != SyncPhaseSynced {
		t.Errorf("Phase = %q, want %q", got.Phase, SyncPhaseSynced)
	}
	if got.PendingCount != 0 {
		t.Errorf("PendingCount = %d, want 0", got.PendingCount)
	}
	if got.LastSyncTime.IsZero() {
		t.Error("LastSyncTime should be set after SetSynced")
	}
}

func TestSyncStatusTrackerSetPending(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("WAL")
	tracker.SetPending("WAL", 7)

	got := tracker.Get("WAL")
	if got.Phase != SyncPhasePending {
		t.Errorf("Phase = %q, want %q", got.Phase, SyncPhasePending)
	}
	if got.PendingCount != 7 {
		t.Errorf("PendingCount = %d, want 7", got.PendingCount)
	}
}

func TestSyncStatusTrackerSetError(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.SetError("Local", "connection refused")

	got := tracker.Get("Local")
	if got.Phase != SyncPhaseError {
		t.Errorf("Phase = %q, want %q", got.Phase, SyncPhaseError)
	}
	if got.ErrorMsg != "connection refused" {
		t.Errorf("ErrorMsg = %q, want %q", got.ErrorMsg, "connection refused")
	}
}

func TestSyncStatusTrackerAll(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.Register("WAL")

	all := tracker.All()
	if len(all) != 2 {
		t.Fatalf("All() returned %d statuses, want 2", len(all))
	}
}

func TestSyncStatusTrackerCount(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	if tracker.Count() != 0 {
		t.Errorf("Count() = %d, want 0", tracker.Count())
	}

	tracker.Register("Local")
	if tracker.Count() != 1 {
		t.Errorf("Count() = %d, want 1", tracker.Count())
	}

	tracker.Register("WAL")
	if tracker.Count() != 2 {
		t.Errorf("Count() = %d, want 2", tracker.Count())
	}
}

func TestSyncStatusTrackerSetOnUnregistered(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	// These should not panic on unregistered providers
	tracker.SetSyncing("nonexistent")
	tracker.SetSynced("nonexistent")
	tracker.SetPending("nonexistent", 5)
	tracker.SetError("nonexistent", "err")
}

func TestSyncStatusTrackerGetReturnsCopy(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("Local")

	got := tracker.Get("Local")
	got.Phase = SyncPhaseError // mutate the copy

	original := tracker.Get("Local")
	if original.Phase != SyncPhaseSynced {
		t.Errorf("mutating Get() result should not affect tracker, got Phase=%q", original.Phase)
	}
}

func TestSyncStatusTrackerClearErrorOnSync(t *testing.T) {
	t.Parallel()

	tracker := NewSyncStatusTracker()
	tracker.Register("Local")
	tracker.SetError("Local", "file not found")
	tracker.SetSynced("Local")

	got := tracker.Get("Local")
	if got.ErrorMsg != "" {
		t.Errorf("ErrorMsg should be cleared after SetSynced, got %q", got.ErrorMsg)
	}
}

func TestIconCircuitStateOverridesPhase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		circuitState CircuitState
		phase        SyncPhase
		wantIcon     string
	}{
		{"circuit open overrides synced", CircuitOpen, SyncPhaseSynced, "✗"},
		{"circuit open overrides syncing", CircuitOpen, SyncPhaseSyncing, "✗"},
		{"circuit half-open overrides synced", CircuitHalfOpen, SyncPhaseSynced, "↻"},
		{"circuit half-open overrides error", CircuitHalfOpen, SyncPhaseError, "↻"},
		{"circuit closed uses synced", CircuitClosed, SyncPhaseSynced, "✓"},
		{"circuit closed uses syncing", CircuitClosed, SyncPhaseSyncing, "↻"},
		{"circuit closed uses pending", CircuitClosed, SyncPhasePending, "⏳"},
		{"circuit closed uses error", CircuitClosed, SyncPhaseError, "✗"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := ProviderSyncStatus{
				CircuitState: tt.circuitState,
				Phase:        tt.phase,
			}
			got := s.Icon()
			if got != tt.wantIcon {
				t.Errorf("Icon() = %q, want %q", got, tt.wantIcon)
			}
		})
	}
}

func TestStatusTextCircuitStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		status       ProviderSyncStatus
		wantContains string
	}{
		{
			"circuit open with retry",
			ProviderSyncStatus{Name: "jira", CircuitState: CircuitOpen, RetryIn: 2 * time.Minute},
			"✗ jira error (retry in 2m)",
		},
		{
			"circuit open no retry",
			ProviderSyncStatus{Name: "jira", CircuitState: CircuitOpen},
			"✗ jira error",
		},
		{
			"circuit half-open probing",
			ProviderSyncStatus{Name: "reminders", CircuitState: CircuitHalfOpen},
			"↻ reminders probing...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.status.StatusText()
			if got != tt.wantContains {
				t.Errorf("StatusText() = %q, want %q", got, tt.wantContains)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"5 seconds", 5 * time.Second, "5s"},
		{"sub-second rounds to 1s", 500 * time.Millisecond, "1s"},
		{"30 seconds", 30 * time.Second, "30s"},
		{"2 minutes", 2 * time.Minute, "2m"},
		{"59 minutes", 59 * time.Minute, "59m"},
		{"1 hour", time.Hour, "1h"},
		{"1 hour 30 min", 90 * time.Minute, "1h 30m"},
		{"2 hours", 2 * time.Hour, "2h"},
		{"2h 15m", 2*time.Hour + 15*time.Minute, "2h 15m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatDuration(tt.d)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestWALPendingText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    ProviderSyncStatus
		want string
	}{
		{
			"no pending",
			ProviderSyncStatus{PendingCount: 0},
			"",
		},
		{
			"pending without oldest",
			ProviderSyncStatus{PendingCount: 5},
			"WAL pending (5 items)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.s.WALPendingText()
			if got != tt.want {
				t.Errorf("WALPendingText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWALPendingTextWithOldest(t *testing.T) {
	t.Parallel()

	s := ProviderSyncStatus{
		PendingCount:  3,
		OldestPending: time.Now().UTC().Add(-5 * time.Minute),
	}
	got := s.WALPendingText()
	if got == "" {
		t.Fatal("expected non-empty WALPendingText")
	}
	if got == "WAL pending (3 items)" {
		t.Error("expected oldest timestamp in output")
	}
}

func TestProviderSyncStatusNewFields(t *testing.T) {
	t.Parallel()

	s := ProviderSyncStatus{
		Name:          "textfile",
		Phase:         SyncPhaseSynced,
		CircuitState:  CircuitClosed,
		RetryIn:       0,
		StaleSince:    time.Time{},
		SyncCount24h:  42,
		ErrorCount24h: 3,
	}

	if s.SyncCount24h != 42 {
		t.Errorf("SyncCount24h = %d, want 42", s.SyncCount24h)
	}
	if s.ErrorCount24h != 3 {
		t.Errorf("ErrorCount24h = %d, want 3", s.ErrorCount24h)
	}
}
