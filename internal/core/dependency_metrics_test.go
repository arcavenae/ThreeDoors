package core

import (
	"encoding/json"
	"testing"
)

func TestRecordDependencyAdded(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	st.RecordDependencyAdded("task-1", "dep-1")

	if st.metrics.DependencyEventCount != 1 {
		t.Fatalf("DependencyEventCount = %d, want 1", st.metrics.DependencyEventCount)
	}
	if len(st.metrics.DependencyEvents) != 1 {
		t.Fatalf("len(DependencyEvents) = %d, want 1", len(st.metrics.DependencyEvents))
	}

	evt := st.metrics.DependencyEvents[0]
	if evt.EventType != "dependency_added" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "dependency_added")
	}
	if evt.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want %q", evt.TaskID, "task-1")
	}
	if evt.DependencyID != "dep-1" {
		t.Errorf("DependencyID = %q, want %q", evt.DependencyID, "dep-1")
	}
	if evt.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestRecordDependencyRemoved(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	st.RecordDependencyRemoved("task-2", "dep-2")

	if st.metrics.DependencyEventCount != 1 {
		t.Fatalf("DependencyEventCount = %d, want 1", st.metrics.DependencyEventCount)
	}

	evt := st.metrics.DependencyEvents[0]
	if evt.EventType != "dependency_removed" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "dependency_removed")
	}
	if evt.TaskID != "task-2" {
		t.Errorf("TaskID = %q, want %q", evt.TaskID, "task-2")
	}
	if evt.DependencyID != "dep-2" {
		t.Errorf("DependencyID = %q, want %q", evt.DependencyID, "dep-2")
	}
}

func TestRecordDependencyUnblocked(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	st.RecordDependencyUnblocked("task-3", "completed-dep")

	if st.metrics.DependencyEventCount != 1 {
		t.Fatalf("DependencyEventCount = %d, want 1", st.metrics.DependencyEventCount)
	}

	evt := st.metrics.DependencyEvents[0]
	if evt.EventType != "dependency_unblocked" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "dependency_unblocked")
	}
	if evt.TaskID != "task-3" {
		t.Errorf("TaskID = %q, want %q", evt.TaskID, "task-3")
	}
	if evt.DependencyID != "completed-dep" {
		t.Errorf("DependencyID = %q, want %q", evt.DependencyID, "completed-dep")
	}
}

func TestRecordDependencyCycleRejected(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	st.RecordDependencyCycleRejected("task-4", "would-cycle-dep")

	if st.metrics.DependencyEventCount != 1 {
		t.Fatalf("DependencyEventCount = %d, want 1", st.metrics.DependencyEventCount)
	}

	evt := st.metrics.DependencyEvents[0]
	if evt.EventType != "dependency_cycle_rejected" {
		t.Errorf("EventType = %q, want %q", evt.EventType, "dependency_cycle_rejected")
	}
	if evt.TaskID != "task-4" {
		t.Errorf("TaskID = %q, want %q", evt.TaskID, "task-4")
	}
	if evt.DependencyID != "would-cycle-dep" {
		t.Errorf("DependencyID = %q, want %q", evt.DependencyID, "would-cycle-dep")
	}
}

func TestDependencyEventsMultiple(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	st.RecordDependencyAdded("t1", "d1")
	st.RecordDependencyRemoved("t2", "d2")
	st.RecordDependencyUnblocked("t3", "d3")
	st.RecordDependencyCycleRejected("t4", "d4")

	if st.metrics.DependencyEventCount != 4 {
		t.Fatalf("DependencyEventCount = %d, want 4", st.metrics.DependencyEventCount)
	}
	if len(st.metrics.DependencyEvents) != 4 {
		t.Fatalf("len(DependencyEvents) = %d, want 4", len(st.metrics.DependencyEvents))
	}

	expectedTypes := []string{
		"dependency_added",
		"dependency_removed",
		"dependency_unblocked",
		"dependency_cycle_rejected",
	}
	for i, want := range expectedTypes {
		if st.metrics.DependencyEvents[i].EventType != want {
			t.Errorf("event[%d].EventType = %q, want %q", i, st.metrics.DependencyEvents[i].EventType, want)
		}
	}
}

func TestDependencyEventsNoneByDefault(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	if st.metrics.DependencyEventCount != 0 {
		t.Errorf("DependencyEventCount = %d, want 0", st.metrics.DependencyEventCount)
	}
	if len(st.metrics.DependencyEvents) != 0 {
		t.Errorf("len(DependencyEvents) = %d, want 0", len(st.metrics.DependencyEvents))
	}
}

func TestDependencyEventJSONSerialization(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	st.RecordDependencyAdded("task-abc", "dep-xyz")
	st.RecordDependencyUnblocked("task-def", "dep-completed")

	metrics := st.Finalize()
	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded SessionMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.DependencyEventCount != 2 {
		t.Errorf("decoded DependencyEventCount = %d, want 2", decoded.DependencyEventCount)
	}
	if len(decoded.DependencyEvents) != 2 {
		t.Fatalf("decoded len(DependencyEvents) = %d, want 2", len(decoded.DependencyEvents))
	}
	if decoded.DependencyEvents[0].EventType != "dependency_added" {
		t.Errorf("event[0].EventType = %q, want %q", decoded.DependencyEvents[0].EventType, "dependency_added")
	}
	if decoded.DependencyEvents[0].TaskID != "task-abc" {
		t.Errorf("event[0].TaskID = %q, want %q", decoded.DependencyEvents[0].TaskID, "task-abc")
	}
	if decoded.DependencyEvents[0].DependencyID != "dep-xyz" {
		t.Errorf("event[0].DependencyID = %q, want %q", decoded.DependencyEvents[0].DependencyID, "dep-xyz")
	}
	if decoded.DependencyEvents[1].EventType != "dependency_unblocked" {
		t.Errorf("event[1].EventType = %q, want %q", decoded.DependencyEvents[1].EventType, "dependency_unblocked")
	}
}

func TestDependencyEventsOmittedWhenEmpty(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	metrics := st.Finalize()
	data, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if _, exists := raw["dependency_events"]; exists {
		t.Error("dependency_events should be omitted from JSON when empty")
	}
}

func TestDependencyEventTimestampsAreUTC(t *testing.T) {
	t.Parallel()
	st := NewSessionTracker()

	st.RecordDependencyAdded("t1", "d1")

	evt := st.metrics.DependencyEvents[0]
	if evt.Timestamp.Location().String() != "UTC" {
		t.Errorf("Timestamp location = %q, want UTC", evt.Timestamp.Location().String())
	}
}
