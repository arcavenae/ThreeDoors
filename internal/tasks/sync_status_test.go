package tasks

import (
	"testing"
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
