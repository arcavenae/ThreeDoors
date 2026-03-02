package tasks

import (
	"time"

	"github.com/google/uuid"
)

// MoodEntry captures a timestamped mood record.
type MoodEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Mood       string    `json:"mood"`
	CustomText string    `json:"custom_text,omitempty"`
}

// DoorSelectionRecord captures which door position was selected and what task.
type DoorSelectionRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	DoorPosition int       `json:"door_position"` // 0=left, 1=center, 2=right
	TaskText     string    `json:"task_text"`
}

// SessionMetrics captures behavioral data for a single app session.
type SessionMetrics struct {
	SessionID           string                `json:"session_id"`
	StartTime           time.Time             `json:"start_time"`
	EndTime             time.Time             `json:"end_time"`
	DurationSeconds     float64               `json:"duration_seconds"`
	TasksCompleted      int                   `json:"tasks_completed"`
	DoorsViewed         int                   `json:"doors_viewed"`
	RefreshesUsed       int                   `json:"refreshes_used"`
	DetailViews         int                   `json:"detail_views"`
	NotesAdded          int                   `json:"notes_added"`
	StatusChanges       int                   `json:"status_changes"`
	MoodEntryCount      int                   `json:"mood_entries"`
	TimeToFirstDoorSecs float64               `json:"time_to_first_door_seconds"`
	DoorSelections      []DoorSelectionRecord `json:"door_selections,omitempty"`
	TaskBypasses        [][]string            `json:"task_bypasses,omitempty"`
	MoodEntries         []MoodEntry           `json:"mood_entries_detail,omitempty"`
}

// SessionTracker provides in-memory tracking of user behavior during an app session.
type SessionTracker struct {
	metrics       *SessionMetrics
	firstDoorTime *time.Time
}

// NewSessionTracker creates a new session tracker.
func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		metrics: &SessionMetrics{
			SessionID:           uuid.New().String(),
			StartTime:           time.Now().UTC(),
			TimeToFirstDoorSecs: -1,
		},
	}
}

// RecordDoorViewed increments the door view counter and captures time-to-first-door.
func (st *SessionTracker) RecordDoorViewed() {
	st.metrics.DoorsViewed++
	if st.firstDoorTime == nil {
		now := time.Now().UTC()
		st.firstDoorTime = &now
		st.metrics.TimeToFirstDoorSecs = now.Sub(st.metrics.StartTime).Seconds()
	}
}

// RecordDoorSelection records which door position was selected and the task text.
func (st *SessionTracker) RecordDoorSelection(position int, taskText string) {
	st.metrics.DoorSelections = append(st.metrics.DoorSelections, DoorSelectionRecord{
		Timestamp:    time.Now().UTC(),
		DoorPosition: position,
		TaskText:     taskText,
	})
	st.RecordDoorViewed()
}

// RecordRefresh increments the refresh counter and records bypassed tasks.
func (st *SessionTracker) RecordRefresh(doorTasks []string) {
	st.metrics.RefreshesUsed++
	if len(doorTasks) > 0 {
		st.metrics.TaskBypasses = append(st.metrics.TaskBypasses, doorTasks)
	}
}

// RecordDetailView increments the detail view counter.
func (st *SessionTracker) RecordDetailView() {
	st.metrics.DetailViews++
}

// RecordTaskCompleted increments the completion counter.
func (st *SessionTracker) RecordTaskCompleted() {
	st.metrics.TasksCompleted++
}

// RecordNoteAdded increments the notes counter.
func (st *SessionTracker) RecordNoteAdded() {
	st.metrics.NotesAdded++
}

// RecordStatusChange increments the status change counter.
func (st *SessionTracker) RecordStatusChange() {
	st.metrics.StatusChanges++
}

// RecordMood records a mood entry with timestamp.
func (st *SessionTracker) RecordMood(mood string, customText string) {
	st.metrics.MoodEntries = append(st.metrics.MoodEntries, MoodEntry{
		Timestamp:  time.Now().UTC(),
		Mood:       mood,
		CustomText: customText,
	})
	st.metrics.MoodEntryCount++
}

// Finalize calculates session duration and returns metrics for persistence.
func (st *SessionTracker) Finalize() *SessionMetrics {
	st.metrics.EndTime = time.Now().UTC()
	st.metrics.DurationSeconds = st.metrics.EndTime.Sub(st.metrics.StartTime).Seconds()
	return st.metrics
}
