package tui

import (
	"fmt"
	"strings"
	"time"
)

// BreadcrumbCapacity is the maximum number of entries in the ring buffer.
const BreadcrumbCapacity = 50

// BreadcrumbEntry records a single navigation event.
type BreadcrumbEntry struct {
	ViewMode  string
	Action    string
	Timestamp time.Time
}

// BreadcrumbTrail tracks recent user navigation actions in a fixed-size ring buffer.
// Nothing is persisted, logged, or transmitted — memory only.
// Privacy: tea.KeyRunes (text input) is never recorded by callers.
type BreadcrumbTrail struct {
	entries [BreadcrumbCapacity]BreadcrumbEntry
	head    int // next write position
	count   int // number of valid entries (max BreadcrumbCapacity)
}

// NewBreadcrumbTrail returns an empty breadcrumb trail.
func NewBreadcrumbTrail() BreadcrumbTrail {
	return BreadcrumbTrail{}
}

// Record adds a navigation event to the trail.
// When the buffer is full, the oldest entry is overwritten.
func (b *BreadcrumbTrail) Record(viewMode, action string) {
	b.entries[b.head] = BreadcrumbEntry{
		ViewMode:  viewMode,
		Action:    action,
		Timestamp: time.Now().UTC(),
	}
	b.head = (b.head + 1) % BreadcrumbCapacity
	if b.count < BreadcrumbCapacity {
		b.count++
	}
}

// Entries returns breadcrumb entries in chronological order (oldest first).
func (b *BreadcrumbTrail) Entries() []BreadcrumbEntry {
	if b.count == 0 {
		return nil
	}
	result := make([]BreadcrumbEntry, b.count)
	start := (b.head - b.count + BreadcrumbCapacity) % BreadcrumbCapacity
	for i := range b.count {
		result[i] = b.entries[(start+i)%BreadcrumbCapacity]
	}
	return result
}

// Format returns all entries as human-readable lines in chronological order.
// Each line includes UTC timestamp, view mode name, and action.
// Returns empty string if there are no entries.
func (b *BreadcrumbTrail) Format() string {
	entries := b.Entries()
	if len(entries) == 0 {
		return ""
	}

	var s strings.Builder
	for i, e := range entries {
		if i > 0 {
			s.WriteByte('\n')
		}
		fmt.Fprintf(&s, "%s [%s] %s", e.Timestamp.Format(time.RFC3339), e.ViewMode, e.Action)
	}
	return s.String()
}
