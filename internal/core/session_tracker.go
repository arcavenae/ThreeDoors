package core

import (
	"time"

	"github.com/google/uuid"
)

// UndoCompleteEntry captures when a task completion is reversed.
type UndoCompleteEntry struct {
	Timestamp           time.Time `json:"timestamp"`
	TaskID              string    `json:"task_id"`
	OriginalCompletedAt time.Time `json:"original_completed_at"`
	ElapsedSeconds      float64   `json:"elapsed_seconds"`
}

// DependencyEvent captures a dependency-related action during a session.
type DependencyEvent struct {
	Timestamp    time.Time `json:"timestamp"`
	EventType    string    `json:"event_type"` // dependency_added, dependency_removed, dependency_unblocked, dependency_cycle_rejected
	TaskID       string    `json:"task_id"`
	DependencyID string    `json:"dependency_id"`
}

// DoorFeedbackEntry captures feedback on why a door/task was declined.
type DoorFeedbackEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	TaskID       string    `json:"task_id"`
	FeedbackType string    `json:"feedback_type"` // blocked, not-now, needs-breakdown, other
	Comment      string    `json:"comment,omitempty"`
}

// SnoozeEvent captures when a task is snoozed/deferred.
type SnoozeEvent struct {
	Timestamp  time.Time  `json:"timestamp"`
	TaskID     string     `json:"task_id"`
	DeferUntil *time.Time `json:"defer_until,omitempty"` // nil for "someday"
	Option     string     `json:"option"`                // "tomorrow", "next_week", "pick_date", "someday"
}

// SnoozeReturnEvent captures when a deferred task auto-returns to todo.
type SnoozeReturnEvent struct {
	Timestamp time.Time `json:"timestamp"`
	TaskID    string    `json:"task_id"`
}

// UnsnoozeEvent captures when a user manually un-snoozes a task.
type UnsnoozeEvent struct {
	Timestamp time.Time `json:"timestamp"`
	TaskID    string    `json:"task_id"`
}

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
	SessionID            string                `json:"session_id"`
	StartTime            time.Time             `json:"start_time"`
	EndTime              time.Time             `json:"end_time"`
	DurationSeconds      float64               `json:"duration_seconds"`
	TasksCompleted       int                   `json:"tasks_completed"`
	DoorsViewed          int                   `json:"doors_viewed"`
	RefreshesUsed        int                   `json:"refreshes_used"`
	DetailViews          int                   `json:"detail_views"`
	NotesAdded           int                   `json:"notes_added"`
	StatusChanges        int                   `json:"status_changes"`
	MoodEntryCount       int                   `json:"mood_entries"`
	TimeToFirstDoorSecs  float64               `json:"time_to_first_door_seconds"`
	DoorSelections       []DoorSelectionRecord `json:"door_selections,omitempty"`
	TaskBypasses         [][]string            `json:"task_bypasses,omitempty"`
	MoodEntries          []MoodEntry           `json:"mood_entries_detail,omitempty"`
	DoorFeedback         []DoorFeedbackEntry   `json:"door_feedback,omitempty"`
	DoorFeedbackCount    int                   `json:"door_feedback_count"`
	UndoCompletes        []UndoCompleteEntry   `json:"undo_completes,omitempty"`
	UndoCompleteCount    int                   `json:"undo_complete_count"`
	DependencyEvents     []DependencyEvent     `json:"dependency_events,omitempty"`
	DependencyEventCount int                   `json:"dependency_event_count"`
	SnoozeEvents         []SnoozeEvent         `json:"snooze_events,omitempty"`
	SnoozeCount          int                   `json:"snooze_count"`
	SnoozeReturnEvents   []SnoozeReturnEvent   `json:"snooze_return_events,omitempty"`
	SnoozeReturnCount    int                   `json:"snooze_return_count"`
	UnsnoozeEvents       []UnsnoozeEvent       `json:"unsnooze_events,omitempty"`
	UnsnoozeCount        int                   `json:"unsnooze_count"`
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

// RecordUndoComplete records when a task completion is reversed.
func (st *SessionTracker) RecordUndoComplete(taskID string, originalCompletedAt time.Time) {
	now := time.Now().UTC()
	st.metrics.UndoCompletes = append(st.metrics.UndoCompletes, UndoCompleteEntry{
		Timestamp:           now,
		TaskID:              taskID,
		OriginalCompletedAt: originalCompletedAt,
		ElapsedSeconds:      now.Sub(originalCompletedAt).Seconds(),
	})
	st.metrics.UndoCompleteCount++
}

// RecordDoorFeedback records feedback on a task shown in a door.
func (st *SessionTracker) RecordDoorFeedback(taskID, feedbackType, comment string) {
	st.metrics.DoorFeedback = append(st.metrics.DoorFeedback, DoorFeedbackEntry{
		Timestamp:    time.Now().UTC(),
		TaskID:       taskID,
		FeedbackType: feedbackType,
		Comment:      comment,
	})
	st.metrics.DoorFeedbackCount++
}

// RecordSnooze records when a task is snoozed/deferred.
func (st *SessionTracker) RecordSnooze(taskID string, deferUntil *time.Time, option string) {
	st.metrics.SnoozeEvents = append(st.metrics.SnoozeEvents, SnoozeEvent{
		Timestamp:  time.Now().UTC(),
		TaskID:     taskID,
		DeferUntil: deferUntil,
		Option:     option,
	})
	st.metrics.SnoozeCount++
}

// RecordSnoozeReturn records when a deferred task auto-returns to todo.
func (st *SessionTracker) RecordSnoozeReturn(taskID string) {
	st.metrics.SnoozeReturnEvents = append(st.metrics.SnoozeReturnEvents, SnoozeReturnEvent{
		Timestamp: time.Now().UTC(),
		TaskID:    taskID,
	})
	st.metrics.SnoozeReturnCount++
}

// RecordUnsnooze records when a user manually un-snoozes a task.
func (st *SessionTracker) RecordUnsnooze(taskID string) {
	st.metrics.UnsnoozeEvents = append(st.metrics.UnsnoozeEvents, UnsnoozeEvent{
		Timestamp: time.Now().UTC(),
		TaskID:    taskID,
	})
	st.metrics.UnsnoozeCount++
}

// MetricsSnapshot provides a read-only view of current session state without finalizing.
type MetricsSnapshot struct {
	TasksCompleted int
	startTime      time.Time
}

// DurationSeconds returns the elapsed session duration in seconds.
func (ms *MetricsSnapshot) DurationSeconds() float64 {
	return time.Since(ms.startTime).Seconds()
}

// GetMetricsSnapshot returns a snapshot of current session metrics without finalizing.
func (st *SessionTracker) GetMetricsSnapshot() *MetricsSnapshot {
	return &MetricsSnapshot{
		TasksCompleted: st.metrics.TasksCompleted,
		startTime:      st.metrics.StartTime,
	}
}

// GetSessionID returns the current session's unique identifier.
func (st *SessionTracker) GetSessionID() string {
	return st.metrics.SessionID
}

// LatestMood returns the most recently recorded mood string, or empty if none.
func (st *SessionTracker) LatestMood() string {
	if len(st.metrics.MoodEntries) == 0 {
		return ""
	}
	return st.metrics.MoodEntries[len(st.metrics.MoodEntries)-1].Mood
}

// RecordDependencyAdded records when a user adds a dependency to a task.
func (st *SessionTracker) RecordDependencyAdded(taskID, dependencyID string) {
	st.metrics.DependencyEvents = append(st.metrics.DependencyEvents, DependencyEvent{
		Timestamp:    time.Now().UTC(),
		EventType:    "dependency_added",
		TaskID:       taskID,
		DependencyID: dependencyID,
	})
	st.metrics.DependencyEventCount++
}

// RecordDependencyRemoved records when a user removes a dependency from a task.
func (st *SessionTracker) RecordDependencyRemoved(taskID, dependencyID string) {
	st.metrics.DependencyEvents = append(st.metrics.DependencyEvents, DependencyEvent{
		Timestamp:    time.Now().UTC(),
		EventType:    "dependency_removed",
		TaskID:       taskID,
		DependencyID: dependencyID,
	})
	st.metrics.DependencyEventCount++
}

// RecordDependencyUnblocked records when a task becomes unblocked because
// its dependencies have all been completed.
func (st *SessionTracker) RecordDependencyUnblocked(taskID, completedDependencyID string) {
	st.metrics.DependencyEvents = append(st.metrics.DependencyEvents, DependencyEvent{
		Timestamp:    time.Now().UTC(),
		EventType:    "dependency_unblocked",
		TaskID:       taskID,
		DependencyID: completedDependencyID,
	})
	st.metrics.DependencyEventCount++
}

// RecordDependencyCycleRejected records when a dependency addition is rejected
// because it would create a circular dependency chain.
func (st *SessionTracker) RecordDependencyCycleRejected(taskID, attemptedDependencyID string) {
	st.metrics.DependencyEvents = append(st.metrics.DependencyEvents, DependencyEvent{
		Timestamp:    time.Now().UTC(),
		EventType:    "dependency_cycle_rejected",
		TaskID:       taskID,
		DependencyID: attemptedDependencyID,
	})
	st.metrics.DependencyEventCount++
}

// Finalize calculates session duration and returns metrics for persistence.
func (st *SessionTracker) Finalize() *SessionMetrics {
	st.metrics.EndTime = time.Now().UTC()
	st.metrics.DurationSeconds = st.metrics.EndTime.Sub(st.metrics.StartTime).Seconds()
	return st.metrics
}
