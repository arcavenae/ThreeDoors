package tasks

import "testing"

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
