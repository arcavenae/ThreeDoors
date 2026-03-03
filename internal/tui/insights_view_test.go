package tui

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// makeInsightsTestSession creates a SessionMetrics for insights view tests.
func makeInsightsTestSession(startTime time.Time, completed int, moods []string, doorPositions []int) tasks.SessionMetrics {
	entries := make([]tasks.MoodEntry, len(moods))
	for i, m := range moods {
		entries[i] = tasks.MoodEntry{Mood: m, Timestamp: startTime}
	}
	selections := make([]tasks.DoorSelectionRecord, len(doorPositions))
	for i, p := range doorPositions {
		selections[i] = tasks.DoorSelectionRecord{DoorPosition: p, TaskText: "task", Timestamp: startTime}
	}
	return tasks.SessionMetrics{
		SessionID:       uuid.New().String(),
		StartTime:       startTime,
		EndTime:         startTime.Add(30 * time.Minute),
		DurationSeconds: 1800,
		TasksCompleted:  completed,
		MoodEntries:     entries,
		MoodEntryCount:  len(moods),
		DoorSelections:  selections,
	}
}

// writeInsightsSessionsFile creates a sessions.jsonl for insights tests.
func writeInsightsSessionsFile(t *testing.T, dir string, sessions []tasks.SessionMetrics) string {
	t.Helper()
	path := filepath.Join(dir, "sessions.jsonl")
	var buf bytes.Buffer
	for _, s := range sessions {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	return path
}

func setupInsightsView(t *testing.T) *InsightsView {
	t.Helper()
	dir := t.TempDir()
	now := time.Date(2026, 3, 7, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	sessions := []tasks.SessionMetrics{
		makeInsightsTestSession(time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC), 3, []string{"Focused"}, []int{0, 1}),
		makeInsightsTestSession(time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC), 5, []string{"Tired"}, []int{1, 1, 2}),
		makeInsightsTestSession(time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC), 4, []string{"Focused", "Energized"}, []int{0, 2}),
	}
	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := tasks.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := tasks.NewCompletionCounterWithNow(frozen)

	iv := NewInsightsView(pa, cc)
	iv.SetWidth(80)
	return iv
}

func TestNewInsightsView(t *testing.T) {
	iv := setupInsightsView(t)
	if iv == nil {
		t.Fatal("NewInsightsView() returned nil")
	}
}

func TestInsightsView_View_ContainsSections(t *testing.T) {
	iv := setupInsightsView(t)
	output := iv.View()

	expectedSections := []string{
		"Your Insights Dashboard",
		"COMPLETION TRENDS",
		"STREAKS",
		"MOOD & PRODUCTIVITY",
		"DOOR POSITION PREFERENCES",
		"Press Esc to return",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("View() output missing section %q", section)
		}
	}
}

func TestInsightsView_View_ColdStart(t *testing.T) {
	// Only 1 session — below threshold
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	frozen := func() time.Time { return now }

	sessions := []tasks.SessionMetrics{
		makeInsightsTestSession(now, 2, []string{"Focused"}, []int{1}),
	}
	paPath := writeInsightsSessionsFile(t, dir, sessions)
	pa := tasks.NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	cc := tasks.NewCompletionCounterWithNow(frozen)
	iv := NewInsightsView(pa, cc)
	iv.SetWidth(80)

	output := iv.View()
	if !strings.Contains(output, "Keep using ThreeDoors to unlock insights!") {
		t.Errorf("cold start view should contain unlock message, got: %q", output)
	}
}

func TestInsightsView_Update_EscReturns(t *testing.T) {
	iv := setupInsightsView(t)

	cmd := iv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc should produce a command")
	}

	// Execute the command and check if it produces ReturnToDoorsMsg
	msg := cmd()
	if _, ok := msg.(ReturnToDoorsMsg); !ok {
		t.Errorf("Esc should produce ReturnToDoorsMsg, got %T", msg)
	}
}

func TestInsightsView_SetWidth(t *testing.T) {
	iv := setupInsightsView(t)
	iv.SetWidth(120)

	// Should not panic; view should render at new width
	output := iv.View()
	if output == "" {
		t.Error("View() should not return empty string after SetWidth")
	}
}
