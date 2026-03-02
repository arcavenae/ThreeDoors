package tasks

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// testSessionMetrics returns a fully populated SessionMetrics for reuse across tests.
// Uses canonical test values from Story 1.5 spec.
func testSessionMetrics() *SessionMetrics {
	now := time.Now().UTC()
	return &SessionMetrics{
		SessionID:           "test-session-aaa",
		StartTime:           now.Add(-5 * time.Minute),
		EndTime:             now,
		DurationSeconds:     300,
		TasksCompleted:      2,
		DoorsViewed:         5,
		RefreshesUsed:       3,
		DetailViews:         2,
		NotesAdded:          0,
		StatusChanges:       2,
		MoodEntryCount:      1,
		TimeToFirstDoorSecs: 1.5,
		DoorSelections: []DoorSelectionRecord{
			{Timestamp: now.Add(-4 * time.Minute), DoorPosition: 0, TaskText: "Task A"},
			{Timestamp: now.Add(-2 * time.Minute), DoorPosition: 2, TaskText: "Task B"},
		},
		TaskBypasses: [][]string{
			{"task1", "task2", "task3"},
		},
		MoodEntries: []MoodEntry{
			{Timestamp: now.Add(-3 * time.Minute), Mood: "Focused", CustomText: ""},
		},
	}
}

func TestMetricsWriter_AppendSession_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewMetricsWriter(tmpDir)

	metrics := testSessionMetrics()
	if err := writer.AppendSession(metrics); err != nil {
		t.Fatalf("AppendSession() error = %v", err)
	}

	path := filepath.Join(tmpDir, "sessions.jsonl")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Expected sessions.jsonl to be created")
	}
}

func TestMetricsWriter_AppendSession_AppendsMultipleSessions(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewMetricsWriter(tmpDir)

	// Write two sessions
	metrics1 := testSessionMetrics()
	metrics1.SessionID = "session-1"
	metrics2 := testSessionMetrics()
	metrics2.SessionID = "session-2"

	if err := writer.AppendSession(metrics1); err != nil {
		t.Fatalf("AppendSession(1) error = %v", err)
	}
	if err := writer.AppendSession(metrics2); err != nil {
		t.Fatalf("AppendSession(2) error = %v", err)
	}

	// Count lines
	path := filepath.Join(tmpDir, "sessions.jsonl")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer f.Close()

	lineCount := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineCount++
	}
	if lineCount != 2 {
		t.Errorf("Expected 2 lines, got %d", lineCount)
	}
}

func TestMetricsWriter_AppendSession_ValidJSONLines(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewMetricsWriter(tmpDir)

	metrics := testSessionMetrics()
	if err := writer.AppendSession(metrics); err != nil {
		t.Fatalf("AppendSession() error = %v", err)
	}

	path := filepath.Join(tmpDir, "sessions.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Verify each line is valid JSON
	var parsed SessionMetrics
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if parsed.SessionID != "test-session-aaa" {
		t.Errorf("SessionID = %q, want %q", parsed.SessionID, "test-session-aaa")
	}
	if parsed.TasksCompleted != 2 {
		t.Errorf("TasksCompleted = %d, want 2", parsed.TasksCompleted)
	}
	if parsed.DurationSeconds != 300 {
		t.Errorf("DurationSeconds = %f, want 300", parsed.DurationSeconds)
	}
	if len(parsed.DoorSelections) != 2 {
		t.Errorf("DoorSelections count = %d, want 2", len(parsed.DoorSelections))
	}
	if len(parsed.MoodEntries) != 1 {
		t.Errorf("MoodEntries count = %d, want 1", len(parsed.MoodEntries))
	}
}

func TestMetricsWriter_AppendSession_ErrorOnInvalidPath(t *testing.T) {
	writer := NewMetricsWriter("/nonexistent/directory/that/does/not/exist")

	metrics := testSessionMetrics()
	if err := writer.AppendSession(metrics); err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestMetricsWriter_AppendSession_PreservesAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	writer := NewMetricsWriter(tmpDir)

	metrics := testSessionMetrics()
	if err := writer.AppendSession(metrics); err != nil {
		t.Fatalf("AppendSession() error = %v", err)
	}

	path := filepath.Join(tmpDir, "sessions.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Parse into generic map to check all fields present
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("JSON unmarshal to map error = %v", err)
	}

	requiredFields := []string{
		"session_id", "start_time", "end_time", "duration_seconds",
		"tasks_completed", "doors_viewed", "refreshes_used",
		"detail_views", "notes_added", "status_changes",
		"mood_entries", "time_to_first_door_seconds",
	}
	for _, field := range requiredFields {
		if _, ok := raw[field]; !ok {
			t.Errorf("Missing required field %q in JSON output", field)
		}
	}
}

// Integration test: SessionTracker -> MetricsWriter full pipeline
func TestIntegration_SessionTrackerToMetricsWriter(t *testing.T) {
	// Create tracker and simulate a session
	tracker := NewSessionTracker()

	// Simulate door selection
	tracker.RecordDoorSelection(1, "Write architecture document")

	// Simulate refresh with bypassed tasks
	tracker.RecordRefresh([]string{"Task A", "Task B", "Task C"})

	// Simulate detail view and status change
	tracker.RecordDetailView()
	tracker.RecordStatusChange()
	tracker.RecordTaskCompleted()

	// Simulate mood
	tracker.RecordMood("Focused", "")

	// Simulate note
	tracker.RecordNoteAdded()

	// Finalize
	metrics := tracker.Finalize()

	// Write via MetricsWriter
	tmpDir := t.TempDir()
	writer := NewMetricsWriter(tmpDir)
	if err := writer.AppendSession(metrics); err != nil {
		t.Fatalf("AppendSession() error = %v", err)
	}

	// Read back and verify
	path := filepath.Join(tmpDir, "sessions.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var parsed SessionMetrics
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	// Verify all counters
	if parsed.SessionID == "" {
		t.Error("Expected non-empty session ID")
	}
	if parsed.TasksCompleted != 1 {
		t.Errorf("TasksCompleted = %d, want 1", parsed.TasksCompleted)
	}
	if parsed.DoorsViewed != 1 {
		t.Errorf("DoorsViewed = %d, want 1", parsed.DoorsViewed)
	}
	if parsed.RefreshesUsed != 1 {
		t.Errorf("RefreshesUsed = %d, want 1", parsed.RefreshesUsed)
	}
	if parsed.DetailViews != 1 {
		t.Errorf("DetailViews = %d, want 1", parsed.DetailViews)
	}
	if parsed.NotesAdded != 1 {
		t.Errorf("NotesAdded = %d, want 1", parsed.NotesAdded)
	}
	if parsed.StatusChanges != 1 {
		t.Errorf("StatusChanges = %d, want 1", parsed.StatusChanges)
	}
	if parsed.MoodEntryCount != 1 {
		t.Errorf("MoodEntryCount = %d, want 1", parsed.MoodEntryCount)
	}
	if parsed.DurationSeconds <= 0 {
		if parsed.DurationSeconds < 0 {
			t.Error("Expected non-negative duration")
		}
	}
	if parsed.TimeToFirstDoorSecs < 0 {
		// Should have been set by RecordDoorSelection
		if parsed.TimeToFirstDoorSecs == -1 {
			t.Error("Expected time-to-first-door to be set after door selection")
		}
	}

	// Verify nested data
	if len(parsed.DoorSelections) != 1 {
		t.Errorf("DoorSelections count = %d, want 1", len(parsed.DoorSelections))
	} else if parsed.DoorSelections[0].DoorPosition != 1 {
		t.Errorf("DoorPosition = %d, want 1", parsed.DoorSelections[0].DoorPosition)
	}
	if len(parsed.TaskBypasses) != 1 {
		t.Errorf("TaskBypasses count = %d, want 1", len(parsed.TaskBypasses))
	} else if len(parsed.TaskBypasses[0]) != 3 {
		t.Errorf("Bypassed tasks = %d, want 3", len(parsed.TaskBypasses[0]))
	}
	if len(parsed.MoodEntries) != 1 {
		t.Errorf("MoodEntries count = %d, want 1", len(parsed.MoodEntries))
	} else if parsed.MoodEntries[0].Mood != "Focused" {
		t.Errorf("Mood = %q, want %q", parsed.MoodEntries[0].Mood, "Focused")
	}
}
