package core

import (
	"fmt"
	"sync"
	"time"
)

// SyncPhase represents the current synchronization state of a provider.
type SyncPhase string

const (
	SyncPhaseSynced  SyncPhase = "synced"
	SyncPhaseSyncing SyncPhase = "syncing"
	SyncPhasePending SyncPhase = "pending"
	SyncPhaseError   SyncPhase = "error"
)

// ProviderSyncStatus holds the sync state for a single provider.
type ProviderSyncStatus struct {
	Name          string
	Phase         SyncPhase
	LastSyncTime  time.Time
	PendingCount  int
	ErrorMsg      string
	CircuitState  CircuitState
	RetryIn       time.Duration
	StaleSince    time.Time
	SyncCount24h  int
	ErrorCount24h int
	OldestPending time.Time
}

// Icon returns the unicode icon based on circuit state.
// Circuit state takes priority: ✓ (closed/healthy), ✗ (open/error), ↻ (half-open/probing).
func (s ProviderSyncStatus) Icon() string {
	switch s.CircuitState {
	case CircuitOpen:
		return "✗"
	case CircuitHalfOpen:
		return "↻"
	default:
		// CircuitClosed — use phase-based icon
		switch s.Phase {
		case SyncPhaseSynced:
			return "✓"
		case SyncPhaseSyncing:
			return "↻"
		case SyncPhasePending:
			return "⏳"
		case SyncPhaseError:
			return "✗"
		default:
			return "?"
		}
	}
}

// StatusText returns a compact display string for the provider status.
// Format: `✓ textfile 5s ago | ✗ jira error (retry in 2m) | ↻ reminders probing...`
func (s ProviderSyncStatus) StatusText() string {
	icon := s.Icon()

	// Circuit open: show error with retry info
	if s.CircuitState == CircuitOpen {
		if s.RetryIn > 0 {
			return fmt.Sprintf("%s %s error (retry in %s)", icon, s.Name, FormatDuration(s.RetryIn))
		}
		return fmt.Sprintf("%s %s error", icon, s.Name)
	}

	// Circuit half-open: probing
	if s.CircuitState == CircuitHalfOpen {
		return fmt.Sprintf("%s %s probing...", icon, s.Name)
	}

	// Circuit closed — phase-based display
	switch s.Phase {
	case SyncPhaseSynced:
		if !s.LastSyncTime.IsZero() {
			return fmt.Sprintf("%s %s %s ago", icon, s.Name, FormatDuration(time.Since(s.LastSyncTime)))
		}
		return fmt.Sprintf("%s %s synced", icon, s.Name)
	case SyncPhaseSyncing:
		return fmt.Sprintf("%s %s syncing", icon, s.Name)
	case SyncPhasePending:
		return fmt.Sprintf("%s %s pending (%d items)", icon, s.Name, s.PendingCount)
	case SyncPhaseError:
		return fmt.Sprintf("%s %s error", icon, s.Name)
	default:
		return fmt.Sprintf("? %s unknown", s.Name)
	}
}

// FormatDuration returns a human-friendly duration string: "5s", "2m", "1h 30m".
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		s := int(d.Seconds())
		if s <= 0 {
			s = 1
		}
		return fmt.Sprintf("%ds", s)
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

// IsStaleSince returns the staleness duration if the provider exceeds its threshold.
// Returns zero if the provider is not stale.
func (s ProviderSyncStatus) IsStaleSince(fileThreshold, networkThreshold time.Duration) time.Duration {
	if s.StaleSince.IsZero() || s.LastSyncTime.IsZero() {
		return 0
	}
	return time.Since(s.StaleSince)
}

// WALPendingText returns a display string for WAL pending items.
// Returns empty string if no items are pending.
func (s ProviderSyncStatus) WALPendingText() string {
	if s.PendingCount == 0 {
		return ""
	}
	if !s.OldestPending.IsZero() {
		age := time.Since(s.OldestPending)
		return fmt.Sprintf("WAL pending (%d items, oldest %s)", s.PendingCount, FormatDuration(age))
	}
	return fmt.Sprintf("WAL pending (%d items)", s.PendingCount)
}

// SyncStatusTracker manages sync status for multiple providers.
type SyncStatusTracker struct {
	mu       sync.RWMutex
	statuses map[string]*ProviderSyncStatus
}

// NewSyncStatusTracker creates a new tracker with no providers registered.
func NewSyncStatusTracker() *SyncStatusTracker {
	return &SyncStatusTracker{
		statuses: make(map[string]*ProviderSyncStatus),
	}
}

// Register adds a provider to the tracker with initial "synced" state.
func (t *SyncStatusTracker) Register(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.statuses[name] = &ProviderSyncStatus{
		Name:  name,
		Phase: SyncPhaseSynced,
	}
}

// SetSyncing marks a provider as currently syncing.
func (t *SyncStatusTracker) SetSyncing(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.statuses[name]; ok {
		s.Phase = SyncPhaseSyncing
		s.ErrorMsg = ""
	}
}

// SetSynced marks a provider as synced with the current timestamp.
func (t *SyncStatusTracker) SetSynced(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.statuses[name]; ok {
		s.Phase = SyncPhaseSynced
		s.LastSyncTime = time.Now().UTC()
		s.PendingCount = 0
		s.ErrorMsg = ""
	}
}

// SetPending marks a provider as having pending items.
func (t *SyncStatusTracker) SetPending(name string, count int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.statuses[name]; ok {
		s.Phase = SyncPhasePending
		s.PendingCount = count
		s.ErrorMsg = ""
	}
}

// SetError marks a provider as having an error.
func (t *SyncStatusTracker) SetError(name string, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.statuses[name]; ok {
		s.Phase = SyncPhaseError
		s.ErrorMsg = errMsg
	}
}

// Get returns the status for a specific provider.
// Returns nil if the provider is not registered.
func (t *SyncStatusTracker) Get(name string) *ProviderSyncStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.statuses[name]
	if !ok {
		return nil
	}
	// Return a copy to avoid data races
	cp := *s
	return &cp
}

// All returns a copy of all provider statuses.
func (t *SyncStatusTracker) All() []ProviderSyncStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]ProviderSyncStatus, 0, len(t.statuses))
	for _, s := range t.statuses {
		result = append(result, *s)
	}
	return result
}

// Count returns the number of registered providers.
func (t *SyncStatusTracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.statuses)
}
