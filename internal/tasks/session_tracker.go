package tasks

import (
	"time"

	"github.com/google/uuid"
)

// SessionMetrics captures behavioral data for a single app session
// Used for validation analysis to determine if Three Doors UX reduces friction
type SessionMetrics struct {
	SessionID           string    `json:"session_id"`
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	DurationSeconds     float64   `json:"duration_seconds"`
	TasksCompleted      int       `json:"tasks_completed"`
	DoorsViewed         int       `json:"doors_viewed"`
	RefreshesUsed       int       `json:"refreshes_used"`
	DetailViews         int       `json:"detail_views"`
	NotesAdded          int       `json:"notes_added"`
	StatusChanges       int       `json:"status_changes"`
	TimeToFirstDoorSecs float64   `json:"time_to_first_door_seconds"` // -1 if no door selected
}

// SessionTracker provides in-memory tracking of user behavior during an app session
// Metrics are persisted to sessions.jsonl on app exit for validation analysis
type SessionTracker struct {
	metrics       *SessionMetrics
	firstDoorTime *time.Time
}

// NewSessionTracker creates a new session tracker with initialized metrics
// Call this once at app startup
func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		metrics: &SessionMetrics{
			SessionID:           uuid.New().String(),
			StartTime:           time.Now().UTC(),
			TimeToFirstDoorSecs: -1, // Sentinel: not yet recorded
		},
	}
}

// RecordDoorViewed increments the door selection counter
// On first call, also captures time-to-first-door metric
func (st *SessionTracker) RecordDoorViewed() {
	st.metrics.DoorsViewed++

	// Capture time to first door if not yet recorded
	if st.firstDoorTime == nil {
		now := time.Now().UTC()
		st.firstDoorTime = &now
		st.metrics.TimeToFirstDoorSecs = now.Sub(st.metrics.StartTime).Seconds()
	}
}

// RecordRefresh increments the refresh counter (R keypress)
func (st *SessionTracker) RecordRefresh() {
	st.metrics.RefreshesUsed++
}

// RecordDetailView increments the task detail view counter
func (st *SessionTracker) RecordDetailView() {
	st.metrics.DetailViews++
}

// RecordTaskCompleted increments the completion counter
func (st *SessionTracker) RecordTaskCompleted() {
	st.metrics.TasksCompleted++
}

// RecordNoteAdded increments the notes counter
func (st *SessionTracker) RecordNoteAdded() {
	st.metrics.NotesAdded++
}

// RecordStatusChange increments the status change counter
func (st *SessionTracker) RecordStatusChange() {
	st.metrics.StatusChanges++
}

// Finalize calculates session duration and returns metrics for persistence
// Call this on app exit before persisting to file
func (st *SessionTracker) Finalize() *SessionMetrics {
	st.metrics.EndTime = time.Now().UTC()
	st.metrics.DurationSeconds = st.metrics.EndTime.Sub(st.metrics.StartTime).Seconds()
	return st.metrics
}
