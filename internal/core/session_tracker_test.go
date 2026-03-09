package core

import (
	"testing"
	"time"
)

func TestNewSessionTracker(t *testing.T) {
	st := NewSessionTracker()
	if st.metrics.SessionID == "" {
		t.Error("Expected non-empty session ID")
	}
	if st.metrics.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}
	if st.metrics.TimeToFirstDoorSecs != -1 {
		t.Error("Expected -1 for initial time-to-first-door")
	}
}

func TestSessionTracker_RecordDoorSelection(t *testing.T) {
	st := NewSessionTracker()
	st.RecordDoorSelection(0, "Test task")

	if len(st.metrics.DoorSelections) != 1 {
		t.Fatalf("Expected 1 door selection, got %d", len(st.metrics.DoorSelections))
	}
	if st.metrics.DoorSelections[0].DoorPosition != 0 {
		t.Errorf("Expected door position 0, got %d", st.metrics.DoorSelections[0].DoorPosition)
	}
	if st.metrics.DoorsViewed != 1 {
		t.Errorf("Expected 1 door viewed, got %d", st.metrics.DoorsViewed)
	}
}

func TestSessionTracker_RecordRefresh(t *testing.T) {
	st := NewSessionTracker()
	st.RecordRefresh([]string{"task1", "task2", "task3"})

	if st.metrics.RefreshesUsed != 1 {
		t.Errorf("Expected 1 refresh, got %d", st.metrics.RefreshesUsed)
	}
	if len(st.metrics.TaskBypasses) != 1 {
		t.Fatalf("Expected 1 bypass record, got %d", len(st.metrics.TaskBypasses))
	}
	if len(st.metrics.TaskBypasses[0]) != 3 {
		t.Errorf("Expected 3 bypassed tasks, got %d", len(st.metrics.TaskBypasses[0]))
	}
}

func TestSessionTracker_RecordMood(t *testing.T) {
	st := NewSessionTracker()
	st.RecordMood("Focused", "")
	st.RecordMood("Other", "feeling creative")

	if st.metrics.MoodEntryCount != 2 {
		t.Errorf("Expected 2 mood entries, got %d", st.metrics.MoodEntryCount)
	}
	if st.metrics.MoodEntries[1].CustomText != "feeling creative" {
		t.Errorf("Expected custom text, got %q", st.metrics.MoodEntries[1].CustomText)
	}
}

func TestSessionTracker_Finalize(t *testing.T) {
	st := NewSessionTracker()
	st.RecordDoorViewed()
	st.RecordTaskCompleted()

	metrics := st.Finalize()
	if metrics.EndTime.IsZero() {
		t.Error("Expected non-zero end time")
	}
	if metrics.DurationSeconds <= 0 {
		// Duration might be very small but should be non-negative
		if metrics.DurationSeconds < 0 {
			t.Error("Expected non-negative duration")
		}
	}
	if metrics.TasksCompleted != 1 {
		t.Errorf("Expected 1 task completed, got %d", metrics.TasksCompleted)
	}
}

func TestSessionTracker_RecordDoorViewed_FirstDoorCapturesTime(t *testing.T) {
	st := NewSessionTracker()

	// Initially -1
	if st.metrics.TimeToFirstDoorSecs != -1 {
		t.Errorf("Expected initial TimeToFirstDoorSecs = -1, got %f", st.metrics.TimeToFirstDoorSecs)
	}

	// First call should set time-to-first-door
	st.RecordDoorViewed()
	if st.metrics.TimeToFirstDoorSecs < 0 {
		t.Errorf("Expected TimeToFirstDoorSecs >= 0 after first view, got %f", st.metrics.TimeToFirstDoorSecs)
	}
	firstDoorTime := st.metrics.TimeToFirstDoorSecs

	// Second call should NOT overwrite
	st.RecordDoorViewed()
	if st.metrics.TimeToFirstDoorSecs != firstDoorTime {
		t.Errorf("TimeToFirstDoorSecs changed on second call: %f != %f", st.metrics.TimeToFirstDoorSecs, firstDoorTime)
	}

	// Should increment DoorsViewed for both calls
	if st.metrics.DoorsViewed != 2 {
		t.Errorf("Expected DoorsViewed = 2, got %d", st.metrics.DoorsViewed)
	}
}

func TestSessionTracker_RecordDoorSelection_IncrementsDoorsViewed(t *testing.T) {
	st := NewSessionTracker()

	// RecordDoorSelection calls RecordDoorViewed internally
	st.RecordDoorSelection(1, "Task A")

	if st.metrics.DoorsViewed != 1 {
		t.Errorf("Expected DoorsViewed = 1, got %d", st.metrics.DoorsViewed)
	}

	// Calling RecordDoorViewed directly adds 1, then RecordDoorSelection adds 1 more internally
	// Total: 1 (first selection) + 1 (direct view) + 1 (second selection's internal view) = 3
	st.RecordDoorViewed()
	st.RecordDoorSelection(2, "Task B")
	if st.metrics.DoorsViewed != 3 {
		t.Errorf("Expected DoorsViewed = 3 (1 selection + 1 direct + 1 selection), got %d", st.metrics.DoorsViewed)
	}
}

func TestSessionTracker_RecordDetailView(t *testing.T) {
	st := NewSessionTracker()

	st.RecordDetailView()
	if st.metrics.DetailViews != 1 {
		t.Errorf("Expected DetailViews = 1, got %d", st.metrics.DetailViews)
	}

	st.RecordDetailView()
	st.RecordDetailView()
	if st.metrics.DetailViews != 3 {
		t.Errorf("Expected DetailViews = 3, got %d", st.metrics.DetailViews)
	}
}

func TestSessionTracker_RecordNoteAdded(t *testing.T) {
	st := NewSessionTracker()

	st.RecordNoteAdded()
	if st.metrics.NotesAdded != 1 {
		t.Errorf("Expected NotesAdded = 1, got %d", st.metrics.NotesAdded)
	}

	st.RecordNoteAdded()
	if st.metrics.NotesAdded != 2 {
		t.Errorf("Expected NotesAdded = 2, got %d", st.metrics.NotesAdded)
	}
}

func TestSessionTracker_RecordStatusChange(t *testing.T) {
	st := NewSessionTracker()

	st.RecordStatusChange()
	if st.metrics.StatusChanges != 1 {
		t.Errorf("Expected StatusChanges = 1, got %d", st.metrics.StatusChanges)
	}

	st.RecordStatusChange()
	st.RecordStatusChange()
	if st.metrics.StatusChanges != 3 {
		t.Errorf("Expected StatusChanges = 3, got %d", st.metrics.StatusChanges)
	}
}

func TestSessionTracker_RecordTaskCompleted(t *testing.T) {
	st := NewSessionTracker()

	st.RecordTaskCompleted()
	if st.metrics.TasksCompleted != 1 {
		t.Errorf("Expected TasksCompleted = 1, got %d", st.metrics.TasksCompleted)
	}

	st.RecordTaskCompleted()
	if st.metrics.TasksCompleted != 2 {
		t.Errorf("Expected TasksCompleted = 2, got %d", st.metrics.TasksCompleted)
	}
}

func TestSessionTracker_RecordRefresh_EmptySlice(t *testing.T) {
	st := NewSessionTracker()

	// Empty slice should still increment refresh count but not add nil bypass entry
	st.RecordRefresh([]string{})

	if st.metrics.RefreshesUsed != 1 {
		t.Errorf("Expected RefreshesUsed = 1, got %d", st.metrics.RefreshesUsed)
	}
	// Empty slice should NOT be appended to TaskBypasses
	if len(st.metrics.TaskBypasses) != 0 {
		t.Errorf("Expected 0 bypass records for empty slice, got %d", len(st.metrics.TaskBypasses))
	}
}

func TestSessionTracker_RecordDoorFeedback(t *testing.T) {
	st := NewSessionTracker()
	st.RecordDoorFeedback("task-123", "blocked", "")
	st.RecordDoorFeedback("task-456", "other", "too vague")

	if st.metrics.DoorFeedbackCount != 2 {
		t.Errorf("Expected 2 door feedback entries, got %d", st.metrics.DoorFeedbackCount)
	}
	if len(st.metrics.DoorFeedback) != 2 {
		t.Fatalf("Expected 2 door feedback records, got %d", len(st.metrics.DoorFeedback))
	}
	if st.metrics.DoorFeedback[0].TaskID != "task-123" {
		t.Errorf("Expected task ID 'task-123', got %q", st.metrics.DoorFeedback[0].TaskID)
	}
	if st.metrics.DoorFeedback[0].FeedbackType != "blocked" {
		t.Errorf("Expected feedback type 'blocked', got %q", st.metrics.DoorFeedback[0].FeedbackType)
	}
	if st.metrics.DoorFeedback[1].Comment != "too vague" {
		t.Errorf("Expected comment 'too vague', got %q", st.metrics.DoorFeedback[1].Comment)
	}
	if st.metrics.DoorFeedback[1].FeedbackType != "other" {
		t.Errorf("Expected feedback type 'other', got %q", st.metrics.DoorFeedback[1].FeedbackType)
	}
}

func TestSessionTracker_RecordDoorFeedback_Timestamp(t *testing.T) {
	st := NewSessionTracker()
	st.RecordDoorFeedback("task-789", "not-now", "")

	if st.metrics.DoorFeedback[0].Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp on door feedback entry")
	}
}

func TestSessionTracker_RecordRefresh_NilSlice(t *testing.T) {
	st := NewSessionTracker()

	st.RecordRefresh(nil)

	if st.metrics.RefreshesUsed != 1 {
		t.Errorf("Expected RefreshesUsed = 1, got %d", st.metrics.RefreshesUsed)
	}
	if len(st.metrics.TaskBypasses) != 0 {
		t.Errorf("Expected 0 bypass records for nil slice, got %d", len(st.metrics.TaskBypasses))
	}
}

// --- LatestMood Tests ---

func TestSessionTracker_LatestMood_NoMoods(t *testing.T) {
	st := NewSessionTracker()
	mood := st.LatestMood()
	if mood != "" {
		t.Errorf("Expected empty string for no moods, got %q", mood)
	}
}

func TestSessionTracker_LatestMood_OneMood(t *testing.T) {
	st := NewSessionTracker()
	st.RecordMood("focused", "")
	mood := st.LatestMood()
	if mood != "focused" {
		t.Errorf("Expected 'focused', got %q", mood)
	}
}

func TestSessionTracker_LatestMood_MultipleMoods(t *testing.T) {
	st := NewSessionTracker()
	st.RecordMood("focused", "")
	st.RecordMood("tired", "")
	st.RecordMood("stressed", "")
	mood := st.LatestMood()
	if mood != "stressed" {
		t.Errorf("Expected 'stressed' (last mood), got %q", mood)
	}
}

// --- RecordUndoComplete Tests ---

func TestSessionTracker_RecordUndoComplete(t *testing.T) {
	st := NewSessionTracker()
	completedAt := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)
	st.RecordUndoComplete("task-123", completedAt)

	if st.metrics.UndoCompleteCount != 1 {
		t.Errorf("Expected UndoCompleteCount = 1, got %d", st.metrics.UndoCompleteCount)
	}
	if len(st.metrics.UndoCompletes) != 1 {
		t.Fatalf("Expected 1 undo entry, got %d", len(st.metrics.UndoCompletes))
	}
	entry := st.metrics.UndoCompletes[0]
	if entry.TaskID != "task-123" {
		t.Errorf("Expected task ID 'task-123', got %q", entry.TaskID)
	}
	if entry.OriginalCompletedAt != completedAt {
		t.Errorf("Expected OriginalCompletedAt = %v, got %v", completedAt, entry.OriginalCompletedAt)
	}
	if entry.ElapsedSeconds <= 0 {
		// elapsed should be positive since completedAt is in the past
		t.Errorf("Expected positive ElapsedSeconds, got %f", entry.ElapsedSeconds)
	}
	if entry.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestSessionTracker_RecordUndoComplete_Multiple(t *testing.T) {
	st := NewSessionTracker()
	completedAt1 := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)
	completedAt2 := time.Date(2025, 3, 1, 11, 0, 0, 0, time.UTC)

	st.RecordUndoComplete("task-1", completedAt1)
	st.RecordUndoComplete("task-2", completedAt2)

	if st.metrics.UndoCompleteCount != 2 {
		t.Errorf("Expected UndoCompleteCount = 2, got %d", st.metrics.UndoCompleteCount)
	}
	if len(st.metrics.UndoCompletes) != 2 {
		t.Fatalf("Expected 2 undo entries, got %d", len(st.metrics.UndoCompletes))
	}
	if st.metrics.UndoCompletes[0].TaskID != "task-1" {
		t.Errorf("Expected first entry task ID 'task-1', got %q", st.metrics.UndoCompletes[0].TaskID)
	}
	if st.metrics.UndoCompletes[1].TaskID != "task-2" {
		t.Errorf("Expected second entry task ID 'task-2', got %q", st.metrics.UndoCompletes[1].TaskID)
	}
}

func TestSessionTracker_RecordUndoComplete_ElapsedTime(t *testing.T) {
	st := NewSessionTracker()
	// Use a time 1 hour in the past — elapsed should be roughly 3600s
	completedAt := time.Now().UTC().Add(-1 * time.Hour)
	st.RecordUndoComplete("task-abc", completedAt)

	entry := st.metrics.UndoCompletes[0]
	// Allow some tolerance for test execution time
	if entry.ElapsedSeconds < 3590 || entry.ElapsedSeconds > 3610 {
		t.Errorf("Expected ElapsedSeconds ~3600, got %f", entry.ElapsedSeconds)
	}
}

func TestSessionTracker_RecordUndoComplete_IncludedInFinalize(t *testing.T) {
	st := NewSessionTracker()
	completedAt := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)
	st.RecordUndoComplete("task-fin", completedAt)

	metrics := st.Finalize()
	if metrics.UndoCompleteCount != 1 {
		t.Errorf("Expected finalized UndoCompleteCount = 1, got %d", metrics.UndoCompleteCount)
	}
	if len(metrics.UndoCompletes) != 1 {
		t.Errorf("Expected finalized 1 undo entry, got %d", len(metrics.UndoCompletes))
	}
}

// --- RecordSnooze Tests ---

func TestSessionTracker_RecordSnooze(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	deferUntil := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	st.RecordSnooze("task-abc", &deferUntil, "tomorrow")

	if st.metrics.SnoozeCount != 1 {
		t.Errorf("Expected SnoozeCount = 1, got %d", st.metrics.SnoozeCount)
	}
	if len(st.metrics.SnoozeEvents) != 1 {
		t.Fatalf("Expected 1 snooze event, got %d", len(st.metrics.SnoozeEvents))
	}
	e := st.metrics.SnoozeEvents[0]
	if e.TaskID != "task-abc" {
		t.Errorf("Expected task ID 'task-abc', got %q", e.TaskID)
	}
	if e.DeferUntil == nil || !e.DeferUntil.Equal(deferUntil) {
		t.Errorf("Expected DeferUntil = %v, got %v", deferUntil, e.DeferUntil)
	}
	if e.Option != "tomorrow" {
		t.Errorf("Expected option 'tomorrow', got %q", e.Option)
	}
	if e.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestSessionTracker_RecordSnooze_Someday(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	st.RecordSnooze("task-xyz", nil, "someday")

	if st.metrics.SnoozeCount != 1 {
		t.Errorf("Expected SnoozeCount = 1, got %d", st.metrics.SnoozeCount)
	}
	e := st.metrics.SnoozeEvents[0]
	if e.DeferUntil != nil {
		t.Errorf("Expected nil DeferUntil for someday, got %v", e.DeferUntil)
	}
	if e.Option != "someday" {
		t.Errorf("Expected option 'someday', got %q", e.Option)
	}
}

func TestSessionTracker_RecordSnooze_Multiple(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	d1 := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 3, 14, 9, 0, 0, 0, time.UTC)

	st.RecordSnooze("task-1", &d1, "tomorrow")
	st.RecordSnooze("task-2", &d2, "next_week")
	st.RecordSnooze("task-3", nil, "someday")

	if st.metrics.SnoozeCount != 3 {
		t.Errorf("Expected SnoozeCount = 3, got %d", st.metrics.SnoozeCount)
	}
	if len(st.metrics.SnoozeEvents) != 3 {
		t.Errorf("Expected 3 snooze events, got %d", len(st.metrics.SnoozeEvents))
	}
}

func TestSessionTracker_RecordSnooze_AllOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		option string
	}{
		{"tomorrow", "tomorrow"},
		{"next_week", "next_week"},
		{"pick_date", "pick_date"},
		{"someday", "someday"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			st := NewSessionTracker()
			st.RecordSnooze("task-1", nil, tt.option)
			if st.metrics.SnoozeEvents[0].Option != tt.option {
				t.Errorf("Expected option %q, got %q", tt.option, st.metrics.SnoozeEvents[0].Option)
			}
		})
	}
}

// --- RecordSnoozeReturn Tests ---

func TestSessionTracker_RecordSnoozeReturn(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	st.RecordSnoozeReturn("task-ret-1")

	if st.metrics.SnoozeReturnCount != 1 {
		t.Errorf("Expected SnoozeReturnCount = 1, got %d", st.metrics.SnoozeReturnCount)
	}
	if len(st.metrics.SnoozeReturnEvents) != 1 {
		t.Fatalf("Expected 1 snooze return event, got %d", len(st.metrics.SnoozeReturnEvents))
	}
	e := st.metrics.SnoozeReturnEvents[0]
	if e.TaskID != "task-ret-1" {
		t.Errorf("Expected task ID 'task-ret-1', got %q", e.TaskID)
	}
	if e.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestSessionTracker_RecordSnoozeReturn_Multiple(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	st.RecordSnoozeReturn("task-1")
	st.RecordSnoozeReturn("task-2")

	if st.metrics.SnoozeReturnCount != 2 {
		t.Errorf("Expected SnoozeReturnCount = 2, got %d", st.metrics.SnoozeReturnCount)
	}
	if len(st.metrics.SnoozeReturnEvents) != 2 {
		t.Errorf("Expected 2 snooze return events, got %d", len(st.metrics.SnoozeReturnEvents))
	}
}

// --- RecordUnsnooze Tests ---

func TestSessionTracker_RecordUnsnooze(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	st.RecordUnsnooze("task-unsn-1")

	if st.metrics.UnsnoozeCount != 1 {
		t.Errorf("Expected UnsnoozeCount = 1, got %d", st.metrics.UnsnoozeCount)
	}
	if len(st.metrics.UnsnoozeEvents) != 1 {
		t.Fatalf("Expected 1 unsnooze event, got %d", len(st.metrics.UnsnoozeEvents))
	}
	e := st.metrics.UnsnoozeEvents[0]
	if e.TaskID != "task-unsn-1" {
		t.Errorf("Expected task ID 'task-unsn-1', got %q", e.TaskID)
	}
	if e.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestSessionTracker_RecordUnsnooze_Multiple(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	st.RecordUnsnooze("task-1")
	st.RecordUnsnooze("task-2")
	st.RecordUnsnooze("task-3")

	if st.metrics.UnsnoozeCount != 3 {
		t.Errorf("Expected UnsnoozeCount = 3, got %d", st.metrics.UnsnoozeCount)
	}
	if len(st.metrics.UnsnoozeEvents) != 3 {
		t.Errorf("Expected 3 unsnooze events, got %d", len(st.metrics.UnsnoozeEvents))
	}
}

// --- Snooze events in Finalize ---

func TestSessionTracker_SnoozeEventsIncludedInFinalize(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	d := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	st.RecordSnooze("task-s", &d, "tomorrow")
	st.RecordSnoozeReturn("task-r")
	st.RecordUnsnooze("task-u")

	metrics := st.Finalize()

	if metrics.SnoozeCount != 1 {
		t.Errorf("Expected finalized SnoozeCount = 1, got %d", metrics.SnoozeCount)
	}
	if len(metrics.SnoozeEvents) != 1 {
		t.Errorf("Expected finalized 1 snooze event, got %d", len(metrics.SnoozeEvents))
	}
	if metrics.SnoozeReturnCount != 1 {
		t.Errorf("Expected finalized SnoozeReturnCount = 1, got %d", metrics.SnoozeReturnCount)
	}
	if len(metrics.SnoozeReturnEvents) != 1 {
		t.Errorf("Expected finalized 1 snooze return event, got %d", len(metrics.SnoozeReturnEvents))
	}
	if metrics.UnsnoozeCount != 1 {
		t.Errorf("Expected finalized UnsnoozeCount = 1, got %d", metrics.UnsnoozeCount)
	}
	if len(metrics.UnsnoozeEvents) != 1 {
		t.Errorf("Expected finalized 1 unsnooze event, got %d", len(metrics.UnsnoozeEvents))
	}
}

// --- CheckDeferredReturnsWithTracker Tests ---

func TestCheckDeferredReturnsWithTracker_LogsSnoozeReturn(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("deferred task")
	task.Status = StatusDeferred
	past := time.Now().UTC().Add(-1 * time.Hour)
	task.DeferUntil = &past
	pool.AddTask(task)

	tracker := NewSessionTracker()
	returned := CheckDeferredReturnsWithTracker(pool, tracker)

	if returned != 1 {
		t.Errorf("returned = %d, want 1", returned)
	}
	if tracker.metrics.SnoozeReturnCount != 1 {
		t.Errorf("SnoozeReturnCount = %d, want 1", tracker.metrics.SnoozeReturnCount)
	}
	if len(tracker.metrics.SnoozeReturnEvents) != 1 {
		t.Fatalf("Expected 1 snooze return event, got %d", len(tracker.metrics.SnoozeReturnEvents))
	}
	if tracker.metrics.SnoozeReturnEvents[0].TaskID != task.ID {
		t.Errorf("TaskID = %q, want %q", tracker.metrics.SnoozeReturnEvents[0].TaskID, task.ID)
	}
}

func TestCheckDeferredReturnsWithTracker_NilTracker(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	task := NewTask("deferred task")
	task.Status = StatusDeferred
	past := time.Now().UTC().Add(-1 * time.Hour)
	task.DeferUntil = &past
	pool.AddTask(task)

	// Should not panic with nil tracker
	returned := CheckDeferredReturnsWithTracker(pool, nil)

	if returned != 1 {
		t.Errorf("returned = %d, want 1", returned)
	}
	if task.Status != StatusTodo {
		t.Errorf("status = %q, want %q", task.Status, StatusTodo)
	}
}

func TestCheckDeferredReturnsWithTracker_MultipleReturns(t *testing.T) {
	t.Parallel()

	pool := NewTaskPool()
	past := time.Now().UTC().Add(-2 * time.Hour)

	task1 := NewTask("expired 1")
	task1.Status = StatusDeferred
	task1.DeferUntil = &past
	pool.AddTask(task1)

	pastCopy := past
	task2 := NewTask("expired 2")
	task2.Status = StatusDeferred
	task2.DeferUntil = &pastCopy
	pool.AddTask(task2)

	future := time.Now().UTC().Add(24 * time.Hour)
	task3 := NewTask("future")
	task3.Status = StatusDeferred
	task3.DeferUntil = &future
	pool.AddTask(task3)

	tracker := NewSessionTracker()
	returned := CheckDeferredReturnsWithTracker(pool, tracker)

	if returned != 2 {
		t.Errorf("returned = %d, want 2", returned)
	}
	if tracker.metrics.SnoozeReturnCount != 2 {
		t.Errorf("SnoozeReturnCount = %d, want 2", tracker.metrics.SnoozeReturnCount)
	}
	if len(tracker.metrics.SnoozeReturnEvents) != 2 {
		t.Errorf("Expected 2 snooze return events, got %d", len(tracker.metrics.SnoozeReturnEvents))
	}
}
