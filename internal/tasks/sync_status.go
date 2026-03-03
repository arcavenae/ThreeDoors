package tasks

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
	Name         string
	Phase        SyncPhase
	LastSyncTime time.Time
	PendingCount int
	ErrorMsg     string
}

// Icon returns the unicode icon for the current sync phase.
func (s ProviderSyncStatus) Icon() string {
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

// StatusText returns a compact display string for the provider status.
func (s ProviderSyncStatus) StatusText() string {
	switch s.Phase {
	case SyncPhaseSynced:
		return fmt.Sprintf("%s %s synced", s.Icon(), s.Name)
	case SyncPhaseSyncing:
		return fmt.Sprintf("%s %s syncing", s.Icon(), s.Name)
	case SyncPhasePending:
		return fmt.Sprintf("%s %s pending (%d items)", s.Icon(), s.Name, s.PendingCount)
	case SyncPhaseError:
		return fmt.Sprintf("%s %s error", s.Icon(), s.Name)
	default:
		return fmt.Sprintf("? %s unknown", s.Name)
	}
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
