package core

import "testing"

func TestMetricsSnapshot_DurationSeconds(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	snap := st.GetMetricsSnapshot()

	dur := snap.DurationSeconds()
	if dur < 0 {
		t.Errorf("expected non-negative duration, got %f", dur)
	}
}

func TestSessionTracker_GetMetricsSnapshot(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	st.RecordTaskCompleted()
	st.RecordTaskCompleted()
	st.RecordTaskCompleted()

	snap := st.GetMetricsSnapshot()
	if snap.TasksCompleted != 3 {
		t.Errorf("expected 3 tasks completed, got %d", snap.TasksCompleted)
	}
}

func TestSessionTracker_GetSessionID(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()
	id := st.GetSessionID()
	if id == "" {
		t.Error("expected non-empty session ID")
	}
	if id != st.metrics.SessionID {
		t.Errorf("GetSessionID() = %q, want %q", id, st.metrics.SessionID)
	}
}

func TestSessionTracker_GetSessionID_Unique(t *testing.T) {
	t.Parallel()
	st1 := NewSessionTracker()
	st2 := NewSessionTracker()
	if st1.GetSessionID() == st2.GetSessionID() {
		t.Error("expected unique session IDs for different trackers")
	}
}
