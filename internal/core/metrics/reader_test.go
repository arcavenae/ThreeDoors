package metrics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// writeJSONL is a test helper that writes SessionMetrics as JSONL to a temp file.
func writeJSONL(t *testing.T, dir string, sessions []core.SessionMetrics) string {
	t.Helper()
	path := filepath.Join(dir, "sessions.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating test file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	for _, s := range sessions {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshaling session: %v", err)
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			t.Fatalf("writing session: %v", err)
		}
	}
	return path
}

// writeRaw is a test helper that writes raw lines to a JSONL file.
func writeRaw(t *testing.T, dir string, lines []string) string {
	t.Helper()
	path := filepath.Join(dir, "sessions.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating test file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	for _, line := range lines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			t.Fatalf("writing line: %v", err)
		}
	}
	return path
}

func makeSession(id string, start time.Time, completed int) core.SessionMetrics {
	return core.SessionMetrics{
		SessionID:       id,
		StartTime:       start,
		EndTime:         start.Add(5 * time.Minute),
		DurationSeconds: 300,
		TasksCompleted:  completed,
	}
}

func TestReadAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string) string
		wantCount int
		wantErr   bool
		wantIDs   []string
	}{
		{
			name: "empty file",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return writeJSONL(t, dir, nil)
			},
			wantCount: 0,
		},
		{
			name: "single session",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return writeJSONL(t, dir, []core.SessionMetrics{
					makeSession("s1", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), 3),
				})
			},
			wantCount: 1,
			wantIDs:   []string{"s1"},
		},
		{
			name: "multiple sessions",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return writeJSONL(t, dir, []core.SessionMetrics{
					makeSession("s1", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), 2),
					makeSession("s2", time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC), 5),
					makeSession("s3", time.Date(2025, 1, 3, 10, 0, 0, 0, time.UTC), 1),
				})
			},
			wantCount: 3,
			wantIDs:   []string{"s1", "s2", "s3"},
		},
		{
			name: "corrupted lines skipped",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				good := makeSession("s1", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), 2)
				goodJSON, _ := json.Marshal(good)
				return writeRaw(t, dir, []string{
					string(goodJSON),
					"not valid json at all",
					"{invalid json}}}",
					"",
				})
			},
			wantCount: 1,
			wantIDs:   []string{"s1"},
		},
		{
			name: "missing file returns empty",
			setup: func(t *testing.T, dir string) string {
				t.Helper()
				return filepath.Join(dir, "nonexistent.jsonl")
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := tt.setup(t, dir)

			r := NewReader(path)
			sessions, err := r.ReadAll()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(sessions) != tt.wantCount {
				t.Errorf("got %d sessions, want %d", len(sessions), tt.wantCount)
			}
			for i, wantID := range tt.wantIDs {
				if i >= len(sessions) {
					break
				}
				if sessions[i].SessionID != wantID {
					t.Errorf("session[%d].SessionID = %q, want %q", i, sessions[i].SessionID, wantID)
				}
			}
		})
	}
}

func TestReadSince(t *testing.T) {
	t.Parallel()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	sessions := []core.SessionMetrics{
		makeSession("s1", base, 1),
		makeSession("s2", base.Add(24*time.Hour), 2),
		makeSession("s3", base.Add(48*time.Hour), 3),
		makeSession("s4", base.Add(72*time.Hour), 4),
	}

	tests := []struct {
		name      string
		since     time.Time
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "all sessions after epoch",
			since:     time.Time{},
			wantCount: 4,
			wantIDs:   []string{"s1", "s2", "s3", "s4"},
		},
		{
			name:      "sessions after day 2",
			since:     base.Add(36 * time.Hour),
			wantCount: 2,
			wantIDs:   []string{"s3", "s4"},
		},
		{
			name:      "sessions after last one",
			since:     base.Add(100 * time.Hour),
			wantCount: 0,
		},
		{
			name:      "exact boundary includes matching session",
			since:     base.Add(48 * time.Hour),
			wantCount: 2,
			wantIDs:   []string{"s3", "s4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := writeJSONL(t, dir, sessions)

			r := NewReader(path)
			result, err := r.ReadSince(tt.since)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.wantCount {
				t.Errorf("got %d sessions, want %d", len(result), tt.wantCount)
			}
			for i, wantID := range tt.wantIDs {
				if i >= len(result) {
					break
				}
				if result[i].SessionID != wantID {
					t.Errorf("session[%d].SessionID = %q, want %q", i, result[i].SessionID, wantID)
				}
			}
		})
	}
}

func TestReadLast(t *testing.T) {
	t.Parallel()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	sessions := []core.SessionMetrics{
		makeSession("s1", base, 1),
		makeSession("s2", base.Add(24*time.Hour), 2),
		makeSession("s3", base.Add(48*time.Hour), 3),
		makeSession("s4", base.Add(72*time.Hour), 4),
	}

	tests := []struct {
		name      string
		n         int
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "last 1",
			n:         1,
			wantCount: 1,
			wantIDs:   []string{"s4"},
		},
		{
			name:      "last 2",
			n:         2,
			wantCount: 2,
			wantIDs:   []string{"s3", "s4"},
		},
		{
			name:      "last exceeds total",
			n:         10,
			wantCount: 4,
			wantIDs:   []string{"s1", "s2", "s3", "s4"},
		},
		{
			name:      "last 0 returns empty",
			n:         0,
			wantCount: 0,
		},
		{
			name:      "negative n returns empty",
			n:         -1,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := writeJSONL(t, dir, sessions)

			r := NewReader(path)
			result, err := r.ReadLast(tt.n)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.wantCount {
				t.Errorf("got %d sessions, want %d", len(result), tt.wantCount)
			}
			for i, wantID := range tt.wantIDs {
				if i >= len(result) {
					break
				}
				if result[i].SessionID != wantID {
					t.Errorf("session[%d].SessionID = %q, want %q", i, result[i].SessionID, wantID)
				}
			}
		})
	}
}

func TestReadLastEmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := writeJSONL(t, dir, nil)

	r := NewReader(path)
	result, err := r.ReadLast(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %d sessions from empty file, want 0", len(result))
	}
}

func TestReadSinceMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")

	r := NewReader(path)
	result, err := r.ReadSince(time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %d sessions from missing file, want 0", len(result))
	}
}

func TestReadLastMissingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")

	r := NewReader(path)
	result, err := r.ReadLast(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("got %d sessions from missing file, want 0", len(result))
	}
}

func TestCorruptedLinesPreserveValidSessions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	s1 := makeSession("valid-1", time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), 2)
	s2 := makeSession("valid-2", time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC), 5)
	s1JSON, _ := json.Marshal(s1)
	s2JSON, _ := json.Marshal(s2)

	path := writeRaw(t, dir, []string{
		string(s1JSON),
		"corrupted line here",
		string(s2JSON),
		"{\"session_id\": incomplete",
		"",
	})

	r := NewReader(path)

	// ReadAll should return both valid sessions
	all, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("ReadAll: got %d sessions, want 2", len(all))
	}
	if len(all) >= 2 {
		if all[0].SessionID != "valid-1" {
			t.Errorf("ReadAll: session[0].SessionID = %q, want %q", all[0].SessionID, "valid-1")
		}
		if all[1].SessionID != "valid-2" {
			t.Errorf("ReadAll: session[1].SessionID = %q, want %q", all[1].SessionID, "valid-2")
		}
		if all[0].TasksCompleted != 2 {
			t.Errorf("ReadAll: session[0].TasksCompleted = %d, want 2", all[0].TasksCompleted)
		}
		if all[1].TasksCompleted != 5 {
			t.Errorf("ReadAll: session[1].TasksCompleted = %d, want 5", all[1].TasksCompleted)
		}
	}

	// ReadSince should also skip corrupted lines
	since, err := r.ReadSince(time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ReadSince: unexpected error: %v", err)
	}
	if len(since) != 1 {
		t.Errorf("ReadSince: got %d sessions, want 1", len(since))
	}

	// ReadLast should also skip corrupted lines
	last, err := r.ReadLast(1)
	if err != nil {
		t.Fatalf("ReadLast: unexpected error: %v", err)
	}
	if len(last) != 1 {
		t.Errorf("ReadLast: got %d sessions, want 1", len(last))
	}
	if len(last) > 0 && last[0].SessionID != "valid-2" {
		t.Errorf("ReadLast: session[0].SessionID = %q, want %q", last[0].SessionID, "valid-2")
	}
}

func TestUndoCompleteEventsParsedCorrectly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	start := time.Date(2025, 3, 15, 14, 30, 0, 0, time.UTC)
	completedAt := start.Add(-1 * time.Hour)
	session := core.SessionMetrics{
		SessionID:       "undo-session",
		StartTime:       start,
		EndTime:         start.Add(5 * time.Minute),
		DurationSeconds: 300,
		TasksCompleted:  1,
		UndoCompletes: []core.UndoCompleteEntry{
			{
				Timestamp:           start.Add(2 * time.Minute),
				TaskID:              "task-undo-1",
				OriginalCompletedAt: completedAt,
				ElapsedSeconds:      3720,
			},
		},
		UndoCompleteCount: 1,
	}
	path := writeJSONL(t, dir, []core.SessionMetrics{session})

	r := NewReader(path)
	result, err := r.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d sessions, want 1", len(result))
	}

	got := result[0]
	if got.UndoCompleteCount != 1 {
		t.Errorf("UndoCompleteCount = %d, want 1", got.UndoCompleteCount)
	}
	if len(got.UndoCompletes) != 1 {
		t.Fatalf("UndoCompletes length = %d, want 1", len(got.UndoCompletes))
	}
	if got.UndoCompletes[0].TaskID != "task-undo-1" {
		t.Errorf("UndoCompletes[0].TaskID = %q, want %q", got.UndoCompletes[0].TaskID, "task-undo-1")
	}
	if got.UndoCompletes[0].ElapsedSeconds != 3720 {
		t.Errorf("UndoCompletes[0].ElapsedSeconds = %f, want 3720", got.UndoCompletes[0].ElapsedSeconds)
	}
	if !got.UndoCompletes[0].OriginalCompletedAt.Equal(completedAt) {
		t.Errorf("UndoCompletes[0].OriginalCompletedAt = %v, want %v", got.UndoCompletes[0].OriginalCompletedAt, completedAt)
	}
}

func TestSessionFieldsParsedCorrectly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	start := time.Date(2025, 3, 15, 14, 30, 0, 0, time.UTC)
	session := core.SessionMetrics{
		SessionID:           "full-session",
		StartTime:           start,
		EndTime:             start.Add(10 * time.Minute),
		DurationSeconds:     600,
		TasksCompleted:      3,
		DoorsViewed:         8,
		RefreshesUsed:       2,
		DetailViews:         4,
		NotesAdded:          1,
		StatusChanges:       5,
		MoodEntryCount:      1,
		TimeToFirstDoorSecs: 1.5,
		DoorSelections: []core.DoorSelectionRecord{
			{Timestamp: start.Add(time.Minute), DoorPosition: 0, TaskText: "Task A"},
		},
		MoodEntries: []core.MoodEntry{
			{Timestamp: start.Add(2 * time.Minute), Mood: "Focused"},
		},
		DoorFeedback: []core.DoorFeedbackEntry{
			{Timestamp: start.Add(3 * time.Minute), TaskID: "t1", FeedbackType: "not-now"},
		},
		DoorFeedbackCount: 1,
	}
	path := writeJSONL(t, dir, []core.SessionMetrics{session})

	r := NewReader(path)
	result, err := r.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d sessions, want 1", len(result))
	}

	got := result[0]
	if got.SessionID != "full-session" {
		t.Errorf("SessionID = %q, want %q", got.SessionID, "full-session")
	}
	if got.TasksCompleted != 3 {
		t.Errorf("TasksCompleted = %d, want 3", got.TasksCompleted)
	}
	if got.DoorsViewed != 8 {
		t.Errorf("DoorsViewed = %d, want 8", got.DoorsViewed)
	}
	if got.DurationSeconds != 600 {
		t.Errorf("DurationSeconds = %f, want 600", got.DurationSeconds)
	}
	if len(got.DoorSelections) != 1 {
		t.Fatalf("DoorSelections length = %d, want 1", len(got.DoorSelections))
	}
	if got.DoorSelections[0].TaskText != "Task A" {
		t.Errorf("DoorSelections[0].TaskText = %q, want %q", got.DoorSelections[0].TaskText, "Task A")
	}
	if len(got.MoodEntries) != 1 {
		t.Fatalf("MoodEntries length = %d, want 1", len(got.MoodEntries))
	}
	if got.MoodEntries[0].Mood != "Focused" {
		t.Errorf("MoodEntries[0].Mood = %q, want %q", got.MoodEntries[0].Mood, "Focused")
	}
	if len(got.DoorFeedback) != 1 {
		t.Fatalf("DoorFeedback length = %d, want 1", len(got.DoorFeedback))
	}
	if got.DoorFeedback[0].FeedbackType != "not-now" {
		t.Errorf("DoorFeedback[0].FeedbackType = %q, want %q", got.DoorFeedback[0].FeedbackType, "not-now")
	}
}
