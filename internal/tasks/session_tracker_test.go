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
