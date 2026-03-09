package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- Test Helpers (Story 4.5 / Insights Dashboard) ---

var baseSessionTime = time.Date(2025, 11, 10, 9, 0, 0, 0, time.UTC)

// makeTestSession creates a SessionMetrics with controllable parameters for testing.
func makeTestSession(startTime time.Time, completed int, moods []string, doorPositions []int) SessionMetrics {
	entries := make([]MoodEntry, len(moods))
	for i, m := range moods {
		entries[i] = MoodEntry{Mood: m, Timestamp: startTime}
	}
	selections := make([]DoorSelectionRecord, len(doorPositions))
	for i, p := range doorPositions {
		selections[i] = DoorSelectionRecord{DoorPosition: p, TaskText: fmt.Sprintf("task-%d", i), Timestamp: startTime}
	}
	return SessionMetrics{
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

// writeSessionsFile creates a sessions.jsonl file with the given sessions.
func writeSessionsFile(t *testing.T, dir string, sessions []SessionMetrics) string {
	t.Helper()
	path := filepath.Join(dir, "sessions.jsonl")
	var buf bytes.Buffer
	for _, s := range sessions {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("writeSessionsFile: marshal error: %v", err)
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("writeSessionsFile: write error: %v", err)
	}
	return path
}

// frozenTimePA returns a nowFunc that always returns the given time.
func frozenTimePA(year int, month time.Month, day, hour int) func() time.Time {
	return func() time.Time {
		return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	}
}

// --- Test Helpers (Story 4.2 / Pattern Report) ---

// makeReportTestSession creates a SessionMetrics with full control for pattern report tests.
func makeReportTestSession(id string, start time.Time, completed int, selections []DoorSelectionRecord, bypasses [][]string, moods []MoodEntry) SessionMetrics {
	end := start.Add(15 * time.Minute)
	return SessionMetrics{
		SessionID:       id,
		StartTime:       start,
		EndTime:         end,
		DurationSeconds: end.Sub(start).Seconds(),
		TasksCompleted:  completed,
		DoorsViewed:     len(selections) + len(bypasses),
		DoorSelections:  selections,
		TaskBypasses:    bypasses,
		MoodEntries:     moods,
		MoodEntryCount:  len(moods),
	}
}

func writeReportSessionsFile(t *testing.T, dir string, sessions []SessionMetrics) string {
	t.Helper()
	path := filepath.Join(dir, "sessions.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create sessions file: %v", err)
	}
	defer func() { _ = f.Close() }()
	for _, s := range sessions {
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("failed to marshal session: %v", err)
		}
		_, _ = f.Write(data)
		_, _ = f.Write([]byte("\n"))
	}
	return path
}

func makeFiveMinimalSessions() []SessionMetrics {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		sessions[i] = makeReportTestSession(
			"sess-"+string(rune('a'+i)),
			baseSessionTime.Add(time.Duration(i)*24*time.Hour),
			1,
			[]DoorSelectionRecord{{
				Timestamp:    baseSessionTime.Add(time.Duration(i)*24*time.Hour + 5*time.Minute),
				DoorPosition: i % 3,
				TaskText:     "Task " + string(rune('A'+i)),
			}},
			nil,
			nil,
		)
	}
	return sessions
}

// ============================================================
// Story 4.5 — Insights Dashboard Tests
// ============================================================

// --- Constructor Tests ---

func TestNewPatternAnalyzer(t *testing.T) {
	pa := NewPatternAnalyzer()
	if pa == nil {
		t.Fatal("NewPatternAnalyzer() returned nil")
	}
	if pa.HasSufficientData() {
		t.Error("new analyzer should not have sufficient data")
	}
}

func TestNewPatternAnalyzerWithNow(t *testing.T) {
	frozen := frozenTimePA(2026, 3, 2, 14)
	pa := NewPatternAnalyzerWithNow(frozen)
	if pa == nil {
		t.Fatal("NewPatternAnalyzerWithNow() returned nil")
	}
}

// --- LoadSessions Tests ---

func TestLoadSessions_BasicLoad(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 3, []string{"Focused"}, []int{0, 1}),
		makeTestSession(now.Add(-1*time.Hour), 2, []string{"Tired"}, []int{2}),
		makeTestSession(now, 4, []string{"Energized"}, []int{1, 0, 2}),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzerWithNow(frozenTimePA(2026, 3, 2, 14))
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}
	if !pa.HasSufficientData() {
		t.Error("3 sessions should be sufficient data")
	}
}

func TestLoadSessions_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.jsonl")

	pa := NewPatternAnalyzer()
	err := pa.LoadSessions(path)
	if err != nil {
		t.Errorf("LoadSessions() should return nil for non-existent file, got: %v", err)
	}
	if pa.HasSufficientData() {
		t.Error("non-existent file should mean no sufficient data")
	}
}

func TestLoadSessions_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}

	pa := NewPatternAnalyzer()
	err := pa.LoadSessions(path)
	if err != nil {
		t.Errorf("LoadSessions() should not error on empty file, got: %v", err)
	}
	if pa.HasSufficientData() {
		t.Error("empty file should mean no sufficient data")
	}
}

func TestLoadSessions_MalformedLines(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	validSession := makeTestSession(now, 3, []string{"Focused"}, []int{0})
	validJSON, _ := json.Marshal(validSession)

	content := fmt.Sprintf("%s\nthis is garbage\n{bad json}\n%s\n%s\n",
		string(validJSON), string(validJSON), string(validJSON))
	path := filepath.Join(dir, "sessions.jsonl")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}

	pa := NewPatternAnalyzer()
	err := pa.LoadSessions(path)
	if err != nil {
		t.Errorf("LoadSessions() should not error on malformed lines, got: %v", err)
	}
	// 3 valid sessions should be loaded (malformed skipped)
	if !pa.HasSufficientData() {
		t.Error("3 valid sessions should be sufficient data")
	}
}

func TestLoadSessions_PermissionError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")
	if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Skipf("cannot change permissions on this OS: %v", err)
	}
	defer os.Chmod(path, 0o644) //nolint:errcheck

	pa := NewPatternAnalyzer()
	err := pa.LoadSessions(path)
	if err == nil {
		t.Error("LoadSessions() should return error for permission-denied file")
	}
}

// --- HasSufficientData / GetSessionsNeeded Tests ---

func TestHasSufficientData_BelowThreshold(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now, 1, nil, nil),
		makeTestSession(now.Add(time.Hour), 2, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}
	if pa.HasSufficientData() {
		t.Error("2 sessions should NOT be sufficient (threshold is 3)")
	}
	if got := pa.GetSessionsNeeded(); got != 1 {
		t.Errorf("GetSessionsNeeded() = %d, want 1", got)
	}
}

func TestHasSufficientData_AtThreshold(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now, 1, nil, nil),
		makeTestSession(now.Add(time.Hour), 2, nil, nil),
		makeTestSession(now.Add(2*time.Hour), 3, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}
	if !pa.HasSufficientData() {
		t.Error("3 sessions should be sufficient")
	}
	if got := pa.GetSessionsNeeded(); got != 0 {
		t.Errorf("GetSessionsNeeded() = %d, want 0", got)
	}
}

// --- GetDailyCompletions Tests ---

func TestGetDailyCompletions_Last7Days(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 7, 14) // Saturday March 7

	sessions := []SessionMetrics{
		makeTestSession(time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), 2, nil, nil),
		makeTestSession(time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC), 3, nil, nil),
		makeTestSession(time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC), 5, nil, nil),
		makeTestSession(time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC), 1, nil, nil),
		makeTestSession(time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC), 3, nil, nil),
		makeTestSession(time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC), 4, nil, nil),
		// no session on March 7 (today)
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	daily := pa.GetDailyCompletions(7)

	expected := map[string]int{
		"2026-03-01": 2,
		"2026-03-02": 3,
		"2026-03-03": 5,
		"2026-03-04": 1,
		"2026-03-05": 3,
		"2026-03-06": 4,
		"2026-03-07": 0, // today with no sessions should be present with 0
	}

	for date, want := range expected {
		if got, ok := daily[date]; !ok {
			t.Errorf("GetDailyCompletions() missing key %q", date)
		} else if got != want {
			t.Errorf("GetDailyCompletions()[%q] = %d, want %d", date, got, want)
		}
	}
}

func TestGetDailyCompletions_MultipleSessions_SameDay(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	sessions := []SessionMetrics{
		makeTestSession(time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC), 3, nil, nil),
		makeTestSession(time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC), 2, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	daily := pa.GetDailyCompletions(1)
	if got := daily["2026-03-02"]; got != 5 {
		t.Errorf("GetDailyCompletions() same-day sum = %d, want 5", got)
	}
}

// --- GetMoodCorrelations Tests (Insights) ---

func TestGetMoodCorrelations_BasicCorrelation(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-4*time.Hour), 5, []string{"Focused"}, nil),
		makeTestSession(now.Add(-3*time.Hour), 3, []string{"Focused"}, nil),
		makeTestSession(now.Add(-2*time.Hour), 1, []string{"Tired"}, nil),
		makeTestSession(now.Add(-1*time.Hour), 2, []string{"Tired"}, nil),
		makeTestSession(now, 4, []string{"Energized"}, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	corrs := pa.GetMoodCorrelations()

	if len(corrs) < 2 {
		t.Fatalf("GetMoodCorrelations() returned %d items, want >= 2", len(corrs))
	}

	foundFocused := false
	foundTired := false
	for _, c := range corrs {
		if c.Mood == "Focused" {
			foundFocused = true
			if c.SessionCount != 2 {
				t.Errorf("Focused SessionCount = %d, want 2", c.SessionCount)
			}
			if math.Abs(c.AvgTasksCompleted-4.0) > 0.01 {
				t.Errorf("Focused AvgTasksCompleted = %f, want 4.0", c.AvgTasksCompleted)
			}
		}
		if c.Mood == "Tired" {
			foundTired = true
			if c.SessionCount != 2 {
				t.Errorf("Tired SessionCount = %d, want 2", c.SessionCount)
			}
			if math.Abs(c.AvgTasksCompleted-1.5) > 0.01 {
				t.Errorf("Tired AvgTasksCompleted = %f, want 1.5", c.AvgTasksCompleted)
			}
		}
	}
	if !foundFocused {
		t.Error("missing Focused in correlations")
	}
	if !foundTired {
		t.Error("missing Tired in correlations")
	}
}

func TestGetMoodCorrelations_MultiMoodSession(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 5, []string{"Focused", "Tired"}, nil),
		makeTestSession(now.Add(-1*time.Hour), 3, []string{"Focused"}, nil),
		makeTestSession(now, 1, []string{"Tired"}, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	corrs := pa.GetMoodCorrelations()

	for _, c := range corrs {
		if c.Mood == "Focused" {
			if c.SessionCount != 2 {
				t.Errorf("Focused SessionCount = %d, want 2", c.SessionCount)
			}
			if math.Abs(c.AvgTasksCompleted-4.0) > 0.01 {
				t.Errorf("Focused AvgTasksCompleted = %f, want 4.0", c.AvgTasksCompleted)
			}
		}
		if c.Mood == "Tired" {
			if c.SessionCount != 2 {
				t.Errorf("Tired SessionCount = %d, want 2", c.SessionCount)
			}
			if math.Abs(c.AvgTasksCompleted-3.0) > 0.01 {
				t.Errorf("Tired AvgTasksCompleted = %f, want 3.0", c.AvgTasksCompleted)
			}
		}
	}
}

func TestGetMoodCorrelations_SessionsWithNoMoods_Excluded(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 5, nil, nil), // no mood
		makeTestSession(now.Add(-1*time.Hour), 3, []string{"Focused"}, nil),
		makeTestSession(now, 4, []string{"Focused"}, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	corrs := pa.GetMoodCorrelations()
	for _, c := range corrs {
		if c.Mood == "Focused" {
			if c.SessionCount != 2 {
				t.Errorf("Focused SessionCount = %d, want 2 (no-mood sessions excluded)", c.SessionCount)
			}
		}
	}
}

func TestGetMoodCorrelations_BelowMinSample(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 3, []string{"Focused"}, nil),
		makeTestSession(now.Add(-1*time.Hour), 4, []string{"Focused"}, nil),
		makeTestSession(now, 2, []string{"Calm"}, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	corrs := pa.GetMoodCorrelations()
	for _, c := range corrs {
		if c.Mood == "Calm" {
			t.Error("Calm should be excluded — only 1 session (below min sample of 2)")
		}
	}
}

// --- GetDoorPositionPreferences Tests ---

func TestGetDoorPositionPreferences_EvenDistribution(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 1, nil, []int{0, 1, 2}),
		makeTestSession(now.Add(-1*time.Hour), 1, nil, []int{0, 1, 2}),
		makeTestSession(now, 1, nil, []int{0, 1, 2}),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	prefs := pa.GetDoorPositionPreferences()
	if prefs.TotalSelections != 9 {
		t.Errorf("TotalSelections = %d, want 9", prefs.TotalSelections)
	}
	if math.Abs(prefs.LeftPercent-33.3) > 0.5 {
		t.Errorf("LeftPercent = %f, want ~33.3", prefs.LeftPercent)
	}
	if math.Abs(prefs.CenterPercent-33.3) > 0.5 {
		t.Errorf("CenterPercent = %f, want ~33.3", prefs.CenterPercent)
	}
	if math.Abs(prefs.RightPercent-33.3) > 0.5 {
		t.Errorf("RightPercent = %f, want ~33.3", prefs.RightPercent)
	}
	if prefs.HasBias {
		t.Error("even distribution should not have bias")
	}
}

func TestGetDoorPositionPreferences_StrongBias(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 1, nil, []int{1, 1, 1}),
		makeTestSession(now.Add(-1*time.Hour), 1, nil, []int{1, 1}),
		makeTestSession(now, 1, nil, []int{1}),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	prefs := pa.GetDoorPositionPreferences()
	if !prefs.HasBias {
		t.Error("100% center should detect bias")
	}
	if prefs.BiasPosition != "center" {
		t.Errorf("BiasPosition = %q, want %q", prefs.BiasPosition, "center")
	}
	if math.Abs(prefs.CenterPercent-100.0) > 0.01 {
		t.Errorf("CenterPercent = %f, want 100.0", prefs.CenterPercent)
	}
}

func TestGetDoorPositionPreferences_NoSelections(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-1*time.Hour), 1, nil, nil),
		makeTestSession(now, 1, nil, nil),
		makeTestSession(now.Add(time.Hour), 1, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	prefs := pa.GetDoorPositionPreferences()
	if prefs.TotalSelections != 0 {
		t.Errorf("TotalSelections = %d, want 0", prefs.TotalSelections)
	}
	if prefs.HasBias {
		t.Error("no selections should not have bias")
	}
}

// --- GetWeekOverWeek Tests ---

func TestGetWeekOverWeek_BasicComparison(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 4, 14)

	sessions := []SessionMetrics{
		makeTestSession(time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC), 3, nil, nil),
		makeTestSession(time.Date(2026, 2, 26, 10, 0, 0, 0, time.UTC), 4, nil, nil),
		makeTestSession(time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC), 7, nil, nil),
		makeTestSession(time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC), 5, nil, nil),
		makeTestSession(time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC), 8, nil, nil),
		makeTestSession(time.Date(2026, 3, 4, 10, 0, 0, 0, time.UTC), 5, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	wk := pa.GetWeekOverWeek()
	if wk.LastWeekTotal != 14 {
		t.Errorf("LastWeekTotal = %d, want 14", wk.LastWeekTotal)
	}
	if wk.ThisWeekTotal != 18 {
		t.Errorf("ThisWeekTotal = %d, want 18", wk.ThisWeekTotal)
	}
	if wk.Direction != "up" {
		t.Errorf("Direction = %q, want %q", wk.Direction, "up")
	}
	if math.Abs(wk.PercentChange-28.57) > 1.0 {
		t.Errorf("PercentChange = %f, want ~28.57", wk.PercentChange)
	}
}

func TestGetWeekOverWeek_LastWeekZero(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 4, 14)

	sessions := []SessionMetrics{
		makeTestSession(time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC), 5, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	wk := pa.GetWeekOverWeek()
	if wk.LastWeekTotal != 0 {
		t.Errorf("LastWeekTotal = %d, want 0", wk.LastWeekTotal)
	}
	if wk.ThisWeekTotal != 5 {
		t.Errorf("ThisWeekTotal = %d, want 5", wk.ThisWeekTotal)
	}
	if math.Abs(wk.PercentChange-100.0) > 0.01 {
		t.Errorf("PercentChange = %f, want 100.0 (last week zero)", wk.PercentChange)
	}
	if wk.Direction != "up" {
		t.Errorf("Direction = %q, want %q", wk.Direction, "up")
	}
}

func TestGetWeekOverWeek_BothZero(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 4, 14)

	sessions := []SessionMetrics{
		makeTestSession(time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC), 5, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	wk := pa.GetWeekOverWeek()
	if wk.PercentChange != 0.0 {
		t.Errorf("PercentChange = %f, want 0.0 (both weeks zero)", wk.PercentChange)
	}
	if wk.Direction != "same" {
		t.Errorf("Direction = %q, want %q", wk.Direction, "same")
	}
}

// --- GetMostProductiveMood / GetMostFrequentMood Tests ---

func TestGetMostProductiveMood(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-4*time.Hour), 5, []string{"Focused"}, nil),
		makeTestSession(now.Add(-3*time.Hour), 4, []string{"Focused"}, nil),
		makeTestSession(now.Add(-2*time.Hour), 1, []string{"Tired"}, nil),
		makeTestSession(now.Add(-1*time.Hour), 2, []string{"Tired"}, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	if got := pa.GetMostProductiveMood(); got != "Focused" {
		t.Errorf("GetMostProductiveMood() = %q, want %q", got, "Focused")
	}
}

func TestGetMostFrequentMood(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-4*time.Hour), 5, []string{"Focused"}, nil),
		makeTestSession(now.Add(-3*time.Hour), 4, []string{"Tired"}, nil),
		makeTestSession(now.Add(-2*time.Hour), 1, []string{"Tired"}, nil),
		makeTestSession(now.Add(-1*time.Hour), 2, []string{"Tired"}, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	if got := pa.GetMostFrequentMood(); got != "Tired" {
		t.Errorf("GetMostFrequentMood() = %q, want %q", got, "Tired")
	}
}

func TestGetMostRecentMood(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 3, []string{"Tired"}, nil),
		makeTestSession(now.Add(-1*time.Hour), 2, nil, nil),
		makeTestSession(now, 4, []string{"Energized", "Focused"}, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	if got := pa.GetMostRecentMood(); got != "Focused" {
		t.Errorf("GetMostRecentMood() = %q, want %q", got, "Focused")
	}
}

func TestGetMostRecentMood_NoMoodData(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeTestSession(now, 3, nil, nil),
	}
	path := writeSessionsFile(t, dir, sessions)

	pa := NewPatternAnalyzer()
	if err := pa.LoadSessions(path); err != nil {
		t.Fatalf("LoadSessions() error: %v", err)
	}

	if got := pa.GetMostRecentMood(); got != "" {
		t.Errorf("GetMostRecentMood() = %q, want empty string when no mood data", got)
	}
}

// --- Computation on Empty Analyzer ---

func TestComputationMethods_BeforeLoadSessions(t *testing.T) {
	pa := NewPatternAnalyzer()

	daily := pa.GetDailyCompletions(7)
	if len(daily) != 7 {
		t.Errorf("GetDailyCompletions() returned %d entries, want 7 (with zeros)", len(daily))
	}

	corrs := pa.GetMoodCorrelations()
	if len(corrs) != 0 {
		t.Errorf("GetMoodCorrelations() returned %d items, want 0", len(corrs))
	}

	prefs := pa.GetDoorPositionPreferences()
	if prefs.TotalSelections != 0 {
		t.Errorf("GetDoorPositionPreferences().TotalSelections = %d, want 0", prefs.TotalSelections)
	}

	wk := pa.GetWeekOverWeek()
	if wk.ThisWeekTotal != 0 || wk.LastWeekTotal != 0 {
		t.Error("GetWeekOverWeek() should return zero totals before loading")
	}

	if got := pa.GetMostProductiveMood(); got != "" {
		t.Errorf("GetMostProductiveMood() = %q, want empty", got)
	}

	if got := pa.GetMostFrequentMood(); got != "" {
		t.Errorf("GetMostFrequentMood() = %q, want empty", got)
	}

	if got := pa.GetMostRecentMood(); got != "" {
		t.Errorf("GetMostRecentMood() = %q, want empty", got)
	}
}

// ============================================================
// Story 4.2 — Pattern Report Tests
// ============================================================

// --- ReadSessions Tests ---

func TestReadSessions_ValidFile(t *testing.T) {
	dir := t.TempDir()
	sessions := makeFiveMinimalSessions()
	path := writeReportSessionsFile(t, dir, sessions)

	analyzer := NewPatternAnalyzer()
	got, err := analyzer.ReadSessions(path)
	if err != nil {
		t.Fatalf("ReadSessions() error = %v", err)
	}
	if len(got) != 5 {
		t.Errorf("ReadSessions() returned %d sessions, want 5", len(got))
	}
	if got[0].SessionID != "sess-a" {
		t.Errorf("ReadSessions()[0].SessionID = %q, want %q", got[0].SessionID, "sess-a")
	}
}

func TestReadSessions_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")
	_ = os.WriteFile(path, []byte(""), 0o644)

	analyzer := NewPatternAnalyzer()
	got, err := analyzer.ReadSessions(path)
	if err != nil {
		t.Fatalf("ReadSessions() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ReadSessions() returned %d sessions, want 0", len(got))
	}
}

func TestReadSessions_MissingFile(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	got, err := analyzer.ReadSessions("/nonexistent/path/sessions.jsonl")
	if err != nil {
		t.Fatalf("ReadSessions() error = %v, want nil for missing file", err)
	}
	if len(got) != 0 {
		t.Errorf("ReadSessions() returned %d sessions, want 0", len(got))
	}
}

func TestReadSessions_MalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")
	content := `{"session_id":"valid-1","start_time":"2025-11-10T09:00:00Z","end_time":"2025-11-10T09:15:00Z","duration_seconds":900,"tasks_completed":1}
this is not json
{"session_id":"valid-2","start_time":"2025-11-11T09:00:00Z","end_time":"2025-11-11T09:15:00Z","duration_seconds":900,"tasks_completed":2}
`
	_ = os.WriteFile(path, []byte(content), 0o644)

	analyzer := NewPatternAnalyzer()
	got, err := analyzer.ReadSessions(path)
	if err != nil {
		t.Fatalf("ReadSessions() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("ReadSessions() returned %d sessions, want 2 (skipping malformed line)", len(got))
	}
}

// --- Cold Start Guard Tests ---

func TestAnalyze_ColdStartGuard_FourSessions(t *testing.T) {
	sessions := makeFiveMinimalSessions()[:4]
	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report != nil {
		t.Error("Analyze() with 4 sessions should return nil report (cold start guard)")
	}
}

func TestAnalyze_ColdStartGuard_FiveSessions(t *testing.T) {
	sessions := makeFiveMinimalSessions()
	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report == nil {
		t.Fatal("Analyze() with 5 sessions should return a report")
	}
	if report.SessionCount != 5 {
		t.Errorf("report.SessionCount = %d, want 5", report.SessionCount)
	}
}

func TestAnalyze_ColdStartGuard_ZeroSessions(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(nil)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report != nil {
		t.Error("Analyze() with nil sessions should return nil report")
	}
}

// --- Door Position Bias Tests ---

func TestAnalyze_DoorPositionBias_AllLeft(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
			[]DoorSelectionRecord{
				{DoorPosition: 0, TaskText: "task-" + string(rune('a'+i))},
				{DoorPosition: 0, TaskText: "task-" + string(rune('f'+i))},
			}, nil, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.DoorPositionBias.PreferredPosition != "left" {
		t.Errorf("PreferredPosition = %q, want %q", report.DoorPositionBias.PreferredPosition, "left")
	}
	if report.DoorPositionBias.LeftCount != 10 {
		t.Errorf("LeftCount = %d, want 10", report.DoorPositionBias.LeftCount)
	}
}

func TestAnalyze_DoorPositionBias_EvenDistribution(t *testing.T) {
	sessions := make([]SessionMetrics, 6)
	for i := 0; i < 6; i++ {
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
			[]DoorSelectionRecord{
				{DoorPosition: i % 3, TaskText: "task"},
			}, nil, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.DoorPositionBias.PreferredPosition != "none" {
		t.Errorf("PreferredPosition = %q, want %q for even distribution", report.DoorPositionBias.PreferredPosition, "none")
	}
}

func TestAnalyze_DoorPositionBias_RightBias(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		selections := []DoorSelectionRecord{
			{DoorPosition: 2, TaskText: "task-a"},
			{DoorPosition: 2, TaskText: "task-b"},
		}
		if i == 0 {
			selections[1] = DoorSelectionRecord{DoorPosition: 0, TaskText: "task-c"}
		}
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1, selections, nil, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.DoorPositionBias.PreferredPosition != "right" {
		t.Errorf("PreferredPosition = %q, want %q", report.DoorPositionBias.PreferredPosition, "right")
	}
}

// --- Time of Day Pattern Tests ---

func TestAnalyze_TimeOfDayPatterns_AllMorning(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		hour := 9 + (i % 3)
		start := time.Date(2025, 11, 10+i, hour, 0, 0, 0, time.UTC)
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), start, 2+i, nil, nil, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	found := false
	for _, p := range report.TimeOfDayPatterns {
		if p.Period == "morning" {
			found = true
			if p.SessionCount != 5 {
				t.Errorf("morning SessionCount = %d, want 5", p.SessionCount)
			}
		}
	}
	if !found {
		t.Error("expected morning time-of-day pattern, not found")
	}
}

func TestAnalyze_TimeOfDayPatterns_Mixed(t *testing.T) {
	times := []int{9, 14, 20, 23, 10}
	sessions := make([]SessionMetrics, 5)
	for i, hour := range times {
		start := time.Date(2025, 11, 10+i, hour, 0, 0, 0, time.UTC)
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), start, 1, nil, nil, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	periodCounts := map[string]int{}
	for _, p := range report.TimeOfDayPatterns {
		periodCounts[p.Period] = p.SessionCount
	}
	if periodCounts["morning"] != 2 {
		t.Errorf("morning count = %d, want 2", periodCounts["morning"])
	}
	if periodCounts["afternoon"] != 1 {
		t.Errorf("afternoon count = %d, want 1", periodCounts["afternoon"])
	}
	if periodCounts["evening"] != 1 {
		t.Errorf("evening count = %d, want 1", periodCounts["evening"])
	}
	if periodCounts["night"] != 1 {
		t.Errorf("night count = %d, want 1", periodCounts["night"])
	}
}

// --- Avoidance Detection Tests ---

func TestAnalyze_Avoidance_TaskBypassedThreeTimes(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		bypasses := [][]string{{"Buy groceries", "Write report"}}
		if i >= 3 {
			bypasses = nil
		}
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
			[]DoorSelectionRecord{{DoorPosition: 0, TaskText: "Other task"}},
			bypasses, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	found := false
	for _, entry := range report.AvoidanceList {
		if entry.TaskText == "Buy groceries" {
			found = true
			if entry.TimesBypassed != 3 {
				t.Errorf("Buy groceries TimesBypassed = %d, want 3", entry.TimesBypassed)
			}
		}
	}
	if !found {
		t.Error("expected 'Buy groceries' in avoidance list with 3 bypasses")
	}
}

func TestAnalyze_Avoidance_TaskBypassedOnce_NotInList(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	sessions[0] = makeReportTestSession("s0", baseSessionTime, 1,
		[]DoorSelectionRecord{{DoorPosition: 0, TaskText: "Selected task"}},
		[][]string{{"Rare bypass task"}}, nil)
	for i := 1; i < 5; i++ {
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
			[]DoorSelectionRecord{{DoorPosition: 0, TaskText: "Selected task"}},
			nil, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	for _, entry := range report.AvoidanceList {
		if entry.TaskText == "Rare bypass task" {
			t.Error("task bypassed only once should NOT appear in avoidance list")
		}
	}
}

func TestAnalyze_Avoidance_TaskBypassedAndSelected(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
			[]DoorSelectionRecord{{DoorPosition: 0, TaskText: "Mixed task"}},
			[][]string{{"Mixed task", "Other"}}, nil)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	for _, entry := range report.AvoidanceList {
		if entry.TaskText == "Mixed task" {
			if entry.NeverSelected {
				t.Error("Mixed task was selected — NeverSelected should be false")
			}
		}
	}
}

// --- Mood Correlation Tests (Report) ---

func TestAnalyze_MoodCorrelations_FocusedTechnical(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		moods := []MoodEntry{{
			Timestamp: baseSessionTime.Add(time.Duration(i) * 24 * time.Hour),
			Mood:      "focused",
		}}
		selections := []DoorSelectionRecord{{
			DoorPosition: 1,
			TaskText:     "Fix API bug",
		}}
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1, selections, nil, moods)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	found := false
	for _, mc := range report.MoodCorrelations {
		if mc.Mood == "focused" {
			found = true
			if mc.SessionCount != 5 {
				t.Errorf("focused SessionCount = %d, want 5", mc.SessionCount)
			}
		}
	}
	if !found {
		t.Error("expected mood correlation for 'focused'")
	}
}

func TestAnalyze_MoodCorrelations_TooFewSessions(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		var moods []MoodEntry
		if i < 2 {
			moods = []MoodEntry{{Mood: "tired"}}
		}
		sessions[i] = makeReportTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
			[]DoorSelectionRecord{{DoorPosition: 0, TaskText: "Task"}},
			nil, moods)
	}

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	for _, mc := range report.MoodCorrelations {
		if mc.Mood == "tired" {
			t.Error("mood 'tired' with only 2 sessions should not appear in correlations (minimum 3)")
		}
	}
}

// --- Patterns Cache Persistence Tests ---

func TestSaveAndLoadPatterns_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "patterns.json")

	report := &PatternReport{
		GeneratedAt:  time.Date(2025, 11, 15, 12, 0, 0, 0, time.UTC),
		SessionCount: 10,
		DoorPositionBias: DoorPositionStats{
			LeftCount:         5,
			CenterCount:       3,
			RightCount:        2,
			TotalSelections:   10,
			PreferredPosition: "left",
		},
		TaskTypeStats: map[string]TypeSelectionRate{
			"technical": {TimesShown: 20, TimesSelected: 15, TimesBypassed: 5, SelectionRate: 0.75},
		},
		TimeOfDayPatterns: []TimeOfDayPattern{
			{Period: "morning", SessionCount: 6, AvgTasksCompleted: 3.5, AvgDuration: 12.0},
		},
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", SessionCount: 4, PreferredType: "technical", PreferredEffort: "deep-work", AvgTasksCompleted: 4.0},
		},
		AvoidanceList: []AvoidanceEntry{
			{TaskText: "Buy groceries", TimesBypassed: 7, TimesShown: 10, NeverSelected: false},
		},
	}

	analyzer := NewPatternAnalyzer()
	if err := analyzer.SavePatterns(report, path); err != nil {
		t.Fatalf("SavePatterns() error = %v", err)
	}

	loaded, err := analyzer.LoadPatterns(path)
	if err != nil {
		t.Fatalf("LoadPatterns() error = %v", err)
	}
	if loaded == nil {
		t.Fatal("LoadPatterns() returned nil")
	}
	if loaded.SessionCount != 10 {
		t.Errorf("loaded.SessionCount = %d, want 10", loaded.SessionCount)
	}
	if loaded.DoorPositionBias.PreferredPosition != "left" {
		t.Errorf("loaded.DoorPositionBias.PreferredPosition = %q, want %q", loaded.DoorPositionBias.PreferredPosition, "left")
	}
	if len(loaded.AvoidanceList) != 1 {
		t.Errorf("loaded.AvoidanceList length = %d, want 1", len(loaded.AvoidanceList))
	}
}

func TestLoadPatterns_MissingFile(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	report, err := analyzer.LoadPatterns("/nonexistent/patterns.json")
	if err != nil {
		t.Fatalf("LoadPatterns() error = %v, want nil for missing file", err)
	}
	if report != nil {
		t.Error("LoadPatterns() should return nil for missing file")
	}
}

func TestLoadPatterns_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "patterns.json")
	_ = os.WriteFile(path, []byte("not valid json{{{"), 0o644)

	analyzer := NewPatternAnalyzer()
	report, err := analyzer.LoadPatterns(path)
	if err == nil {
		t.Error("LoadPatterns() should return error for corrupt file")
	}
	if report != nil {
		t.Error("LoadPatterns() should return nil report for corrupt file")
	}
}

// --- NeedsReanalysis Tests ---

func TestNeedsReanalysis_NilCached(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	sessions := makeFiveMinimalSessions()
	if !analyzer.NeedsReanalysis(nil, sessions) {
		t.Error("NeedsReanalysis(nil, sessions) should return true")
	}
}

func TestNeedsReanalysis_MoreSessions(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	cached := &PatternReport{SessionCount: 5, GeneratedAt: time.Now()}
	sessions := make([]SessionMetrics, 7)
	if !analyzer.NeedsReanalysis(cached, sessions) {
		t.Error("NeedsReanalysis should return true when more sessions exist than cached")
	}
}

func TestNeedsReanalysis_SameCount(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	cached := &PatternReport{SessionCount: 5, GeneratedAt: time.Now()}
	sessions := make([]SessionMetrics, 5)
	if analyzer.NeedsReanalysis(cached, sessions) {
		t.Error("NeedsReanalysis should return false when session count matches")
	}
}

func TestNeedsReanalysis_NewerSession(t *testing.T) {
	analyzer := NewPatternAnalyzer()
	cached := &PatternReport{SessionCount: 5, GeneratedAt: time.Date(2025, 11, 10, 12, 0, 0, 0, time.UTC)}
	sessions := make([]SessionMetrics, 5)
	sessions[4] = SessionMetrics{EndTime: time.Date(2025, 11, 15, 12, 0, 0, 0, time.UTC)}
	if !analyzer.NeedsReanalysis(cached, sessions) {
		t.Error("NeedsReanalysis should return true when latest session is newer than GeneratedAt")
	}
}

// --- Insights Formatter Tests ---

func TestFormatInsights_FullReport(t *testing.T) {
	report := &PatternReport{
		SessionCount: 15,
		DoorPositionBias: DoorPositionStats{
			LeftCount:         8,
			CenterCount:       4,
			RightCount:        3,
			TotalSelections:   15,
			PreferredPosition: "left",
		},
		TaskTypeStats: map[string]TypeSelectionRate{
			"technical":      {TimesSelected: 10, TimesBypassed: 2, SelectionRate: 0.83},
			"administrative": {TimesSelected: 3, TimesBypassed: 8, SelectionRate: 0.27},
		},
		TimeOfDayPatterns: []TimeOfDayPattern{
			{Period: "morning", SessionCount: 8, AvgTasksCompleted: 4.0},
			{Period: "evening", SessionCount: 7, AvgTasksCompleted: 2.0},
		},
		MoodCorrelations: []MoodCorrelation{
			{Mood: "focused", PreferredType: "technical", PreferredEffort: "deep-work", AvgTasksCompleted: 5.0},
		},
		AvoidanceList: []AvoidanceEntry{
			{TaskText: "Buy groceries", TimesBypassed: 7},
		},
	}

	output := FormatInsights(report)
	if output == "" {
		t.Fatal("FormatInsights() returned empty string")
	}

	checks := []string{
		"15 sessions",
		"left",
		"morning",
		"focused",
		"Buy groceries",
	}
	for _, check := range checks {
		if !patternContains(output, check) {
			t.Errorf("FormatInsights() output missing expected content: %q", check)
		}
	}
}

func TestFormatInsights_NilReport(t *testing.T) {
	output := FormatInsights(nil)
	if output == "" {
		t.Fatal("FormatInsights(nil) should return encouragement message, not empty string")
	}
	if !patternContains(output, "5 sessions") {
		t.Error("FormatInsights(nil) should mention needing 5 sessions")
	}
}

func TestFormatInsights_NoMoodData(t *testing.T) {
	report := &PatternReport{
		SessionCount: 5,
		DoorPositionBias: DoorPositionStats{
			PreferredPosition: "none",
			TotalSelections:   5,
		},
		TimeOfDayPatterns: []TimeOfDayPattern{
			{Period: "morning", SessionCount: 5, AvgTasksCompleted: 2.0},
		},
	}

	output := FormatInsights(report)
	if patternContains(output, "Mood") {
		t.Error("FormatInsights() should skip mood section when no mood data")
	}
}

func TestFormatInsights_NoAvoidance(t *testing.T) {
	report := &PatternReport{
		SessionCount: 5,
		DoorPositionBias: DoorPositionStats{
			PreferredPosition: "center",
			TotalSelections:   5,
		},
	}

	output := FormatInsights(report)
	if patternContains(output, "Avoidance") {
		t.Error("FormatInsights() should skip avoidance section when no avoidance data")
	}
}

// --- End-to-End Integration Test ---

func TestPatternAnalyzer_EndToEnd(t *testing.T) {
	dir := t.TempDir()

	sessions := make([]SessionMetrics, 7)
	for i := 0; i < 7; i++ {
		start := time.Date(2025, 11, 10+i, 9, 0, 0, 0, time.UTC)
		sessions[i] = makeReportTestSession("sess-"+string(rune('a'+i)), start, 2,
			[]DoorSelectionRecord{
				{Timestamp: start.Add(2 * time.Minute), DoorPosition: 0, TaskText: "Fix API bug"},
			},
			[][]string{{"Buy groceries", "Write report"}},
			[]MoodEntry{{Timestamp: start, Mood: "focused"}},
		)
	}

	sessionsPath := writeReportSessionsFile(t, dir, sessions)
	patternsPath := filepath.Join(dir, "patterns.json")

	analyzer := NewPatternAnalyzer()

	loaded, err := analyzer.ReadSessions(sessionsPath)
	if err != nil {
		t.Fatalf("ReadSessions() error = %v", err)
	}
	if len(loaded) != 7 {
		t.Fatalf("ReadSessions() returned %d, want 7", len(loaded))
	}

	report, err := analyzer.Analyze(loaded)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report == nil {
		t.Fatal("Analyze() returned nil with 7 sessions")
	}

	if report.DoorPositionBias.PreferredPosition != "left" {
		t.Errorf("Expected left bias, got %q", report.DoorPositionBias.PreferredPosition)
	}

	morningFound := false
	for _, p := range report.TimeOfDayPatterns {
		if p.Period == "morning" && p.SessionCount == 7 {
			morningFound = true
		}
	}
	if !morningFound {
		t.Error("Expected all 7 sessions in morning period")
	}

	groceriesFound := false
	for _, a := range report.AvoidanceList {
		if a.TaskText == "Buy groceries" && a.TimesBypassed >= 3 {
			groceriesFound = true
		}
	}
	if !groceriesFound {
		t.Error("Expected 'Buy groceries' in avoidance list")
	}

	if err := analyzer.SavePatterns(report, patternsPath); err != nil {
		t.Fatalf("SavePatterns() error = %v", err)
	}
	reloaded, err := analyzer.LoadPatterns(patternsPath)
	if err != nil {
		t.Fatalf("LoadPatterns() error = %v", err)
	}
	if reloaded.SessionCount != report.SessionCount {
		t.Errorf("Reloaded SessionCount = %d, want %d", reloaded.SessionCount, report.SessionCount)
	}

	output := FormatInsights(report)
	if output == "" {
		t.Error("FormatInsights() returned empty")
	}
}

// --- Mood Correlation with Category Data Tests (Story 4.3) ---

func TestAnalyzeMoodCorrelations_PopulatesPreferredType(t *testing.T) {
	sessions := []SessionMetrics{
		makeReportTestSession("s1", baseSessionTime, 2,
			[]DoorSelectionRecord{{TaskText: "Fix login bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s2", baseSessionTime.Add(1*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Write unit tests", DoorPosition: 1}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s3", baseSessionTime.Add(2*time.Hour), 2,
			[]DoorSelectionRecord{{TaskText: "Fix login bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s4", baseSessionTime.Add(3*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Write unit tests", DoorPosition: 2}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s5", baseSessionTime.Add(4*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix login bug", DoorPosition: 1}},
			nil, []MoodEntry{{Mood: "focused"}}),
	}

	analyzer := NewPatternAnalyzer()
	analyzer.SetTaskCategories(map[string]TaskCategoryInfo{
		"Fix login bug":    {Type: TypeTechnical, Effort: EffortDeepWork},
		"Write unit tests": {Type: TypeTechnical, Effort: EffortMedium},
	})

	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report == nil {
		t.Fatal("Analyze() returned nil with 5 sessions")
	}

	var focusedCorrelation *MoodCorrelation
	for i := range report.MoodCorrelations {
		if report.MoodCorrelations[i].Mood == "focused" {
			focusedCorrelation = &report.MoodCorrelations[i]
			break
		}
	}
	if focusedCorrelation == nil {
		t.Fatal("Expected mood correlation for 'focused'")
	}
	if focusedCorrelation.PreferredType != string(TypeTechnical) {
		t.Errorf("Expected PreferredType 'technical', got %q", focusedCorrelation.PreferredType)
	}
}

func TestAnalyzeMoodCorrelations_MixedTypes_MostFrequentWins(t *testing.T) {
	sessions := []SessionMetrics{
		makeReportTestSession("s1", baseSessionTime, 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "calm"}}),
		makeReportTestSession("s2", baseSessionTime.Add(1*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "calm"}}),
		makeReportTestSession("s3", baseSessionTime.Add(2*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Reply emails", DoorPosition: 1}},
			nil, []MoodEntry{{Mood: "calm"}}),
		makeReportTestSession("s4", baseSessionTime.Add(3*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "calm"}}),
		makeReportTestSession("s5", baseSessionTime.Add(4*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "calm"}}),
	}

	analyzer := NewPatternAnalyzer()
	analyzer.SetTaskCategories(map[string]TaskCategoryInfo{
		"Fix bug":      {Type: TypeTechnical, Effort: EffortDeepWork},
		"Reply emails": {Type: TypeAdministrative, Effort: EffortQuickWin},
	})

	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	var calmCorrelation *MoodCorrelation
	for i := range report.MoodCorrelations {
		if report.MoodCorrelations[i].Mood == "calm" {
			calmCorrelation = &report.MoodCorrelations[i]
			break
		}
	}
	if calmCorrelation == nil {
		t.Fatal("Expected mood correlation for 'calm'")
	}
	if calmCorrelation.PreferredType != string(TypeTechnical) {
		t.Errorf("Expected PreferredType 'technical' (most frequent), got %q", calmCorrelation.PreferredType)
	}
}

func TestAnalyzeMoodCorrelations_NoCategoryMap_EmptyPreferredType(t *testing.T) {
	sessions := []SessionMetrics{
		makeReportTestSession("s1", baseSessionTime, 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s2", baseSessionTime.Add(1*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s3", baseSessionTime.Add(2*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s4", baseSessionTime.Add(3*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s5", baseSessionTime.Add(4*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
	}

	analyzer := NewPatternAnalyzer()

	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	var focusedCorrelation *MoodCorrelation
	for i := range report.MoodCorrelations {
		if report.MoodCorrelations[i].Mood == "focused" {
			focusedCorrelation = &report.MoodCorrelations[i]
			break
		}
	}
	if focusedCorrelation == nil {
		t.Fatal("Expected mood correlation for 'focused'")
	}
	if focusedCorrelation.PreferredType != "" {
		t.Errorf("Expected empty PreferredType without category map, got %q", focusedCorrelation.PreferredType)
	}
}

func TestAnalyzeMoodCorrelations_TaskNotInMap_Skipped(t *testing.T) {
	sessions := []SessionMetrics{
		makeReportTestSession("s1", baseSessionTime, 1,
			[]DoorSelectionRecord{{TaskText: "Deleted task", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s2", baseSessionTime.Add(1*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s3", baseSessionTime.Add(2*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 1}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s4", baseSessionTime.Add(3*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Deleted task", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
		makeReportTestSession("s5", baseSessionTime.Add(4*time.Hour), 1,
			[]DoorSelectionRecord{{TaskText: "Fix bug", DoorPosition: 0}},
			nil, []MoodEntry{{Mood: "focused"}}),
	}

	analyzer := NewPatternAnalyzer()
	analyzer.SetTaskCategories(map[string]TaskCategoryInfo{
		"Fix bug": {Type: TypeTechnical, Effort: EffortDeepWork},
	})

	report, err := analyzer.Analyze(sessions)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	var focusedCorrelation *MoodCorrelation
	for i := range report.MoodCorrelations {
		if report.MoodCorrelations[i].Mood == "focused" {
			focusedCorrelation = &report.MoodCorrelations[i]
			break
		}
	}
	if focusedCorrelation == nil {
		t.Fatal("Expected mood correlation for 'focused'")
	}
	if focusedCorrelation.PreferredType != string(TypeTechnical) {
		t.Errorf("Expected PreferredType 'technical' (skipping unknown tasks), got %q", focusedCorrelation.PreferredType)
	}
}

// ============================================================
// Story 40.6 — Session Highlights Tests
// ============================================================

func TestGetSessionHighlights_NoSessions(t *testing.T) {
	pa := NewPatternAnalyzer()
	h := pa.GetSessionHighlights()

	if h.TotalDoors != 0 {
		t.Errorf("TotalDoors = %d, want 0", h.TotalDoors)
	}
	if h.TotalTasks != 0 {
		t.Errorf("TotalTasks = %d, want 0", h.TotalTasks)
	}
	if h.PeakHour != -1 {
		t.Errorf("PeakHour = %d, want -1", h.PeakHour)
	}
	if h.LongestStreak != 0 {
		t.Errorf("LongestStreak = %d, want 0", h.LongestStreak)
	}
}

func TestGetSessionHighlights_SingleSession(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		{
			SessionID:           "s1",
			StartTime:           time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
			EndTime:             time.Date(2026, 3, 1, 10, 30, 0, 0, time.UTC),
			DurationSeconds:     1800,
			TasksCompleted:      3,
			DetailViews:         2,
			NotesAdded:          1,
			TimeToFirstDoorSecs: 5.5,
			DoorSelections: []DoorSelectionRecord{
				{DoorPosition: 0, TaskText: "t1"},
				{DoorPosition: 1, TaskText: "t2"},
			},
		},
	}

	h := pa.GetSessionHighlights()

	if h.TotalDoors != 2 {
		t.Errorf("TotalDoors = %d, want 2", h.TotalDoors)
	}
	if h.TotalTasks != 3 {
		t.Errorf("TotalTasks = %d, want 3", h.TotalTasks)
	}
	if h.AvgSessionDuration != 30*time.Minute {
		t.Errorf("AvgSessionDuration = %v, want 30m", h.AvgSessionDuration)
	}
	if h.FastestFirstDoor != 5500*time.Millisecond {
		t.Errorf("FastestFirstDoor = %v, want 5.5s", h.FastestFirstDoor)
	}
	if h.TotalDetailViews != 2 {
		t.Errorf("TotalDetailViews = %d, want 2", h.TotalDetailViews)
	}
	if h.TotalNotesAdded != 1 {
		t.Errorf("TotalNotesAdded = %d, want 1", h.TotalNotesAdded)
	}
	if h.PeakHour != 10 {
		t.Errorf("PeakHour = %d, want 10", h.PeakHour)
	}
	if h.LongestStreak != 1 {
		t.Errorf("LongestStreak = %d, want 1", h.LongestStreak)
	}
}

func TestGetSessionHighlights_MultipleSessions(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		{
			SessionID:           "s1",
			StartTime:           time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
			DurationSeconds:     600,
			TasksCompleted:      2,
			DetailViews:         1,
			NotesAdded:          0,
			TimeToFirstDoorSecs: 8.0,
			DoorSelections:      []DoorSelectionRecord{{DoorPosition: 0}},
		},
		{
			SessionID:           "s2",
			StartTime:           time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC),
			DurationSeconds:     1200,
			TasksCompleted:      5,
			DetailViews:         3,
			NotesAdded:          2,
			TimeToFirstDoorSecs: 3.0,
			DoorSelections:      []DoorSelectionRecord{{DoorPosition: 1}, {DoorPosition: 2}},
		},
		{
			SessionID:           "s3",
			StartTime:           time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC),
			DurationSeconds:     900,
			TasksCompleted:      1,
			DetailViews:         0,
			NotesAdded:          1,
			TimeToFirstDoorSecs: 0, // no first door recorded
			DoorSelections:      []DoorSelectionRecord{{DoorPosition: 0}},
		},
	}

	h := pa.GetSessionHighlights()

	if h.TotalDoors != 4 {
		t.Errorf("TotalDoors = %d, want 4", h.TotalDoors)
	}
	if h.TotalTasks != 8 {
		t.Errorf("TotalTasks = %d, want 8", h.TotalTasks)
	}
	// Avg duration: (600+1200+900)/3 = 900s = 15m
	if h.AvgSessionDuration != 15*time.Minute {
		t.Errorf("AvgSessionDuration = %v, want 15m", h.AvgSessionDuration)
	}
	// Fastest first door: min(8.0, 3.0) = 3.0s
	if h.FastestFirstDoor != 3*time.Second {
		t.Errorf("FastestFirstDoor = %v, want 3s", h.FastestFirstDoor)
	}
	if h.TotalDetailViews != 4 {
		t.Errorf("TotalDetailViews = %d, want 4", h.TotalDetailViews)
	}
	if h.TotalNotesAdded != 3 {
		t.Errorf("TotalNotesAdded = %d, want 3", h.TotalNotesAdded)
	}
	// Peak hour: 10 appears twice (s1, s3), 14 appears once => 10
	if h.PeakHour != 10 {
		t.Errorf("PeakHour = %d, want 10", h.PeakHour)
	}
	// Three consecutive days: Mar 1, 2, 3
	if h.LongestStreak != 3 {
		t.Errorf("LongestStreak = %d, want 3", h.LongestStreak)
	}
}

func TestGetSessionHighlights_PeakHourTieBreaking(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		{SessionID: "s1", StartTime: time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
		{SessionID: "s2", StartTime: time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
		{SessionID: "s3", StartTime: time.Date(2026, 3, 3, 14, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
		{SessionID: "s4", StartTime: time.Date(2026, 3, 4, 9, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
	}

	h := pa.GetSessionHighlights()
	// Both 9 and 14 have 2 sessions each; earliest hour (9) wins
	if h.PeakHour != 9 {
		t.Errorf("PeakHour = %d, want 9 (earliest wins on tie)", h.PeakHour)
	}
}

func TestGetSessionHighlights_FastestFirstDoor_IgnoresZero(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		{SessionID: "s1", TimeToFirstDoorSecs: 0, DurationSeconds: 600},
		{SessionID: "s2", TimeToFirstDoorSecs: -1, DurationSeconds: 600},
		{SessionID: "s3", TimeToFirstDoorSecs: 12.5, DurationSeconds: 600},
	}

	h := pa.GetSessionHighlights()
	if h.FastestFirstDoor != 12500*time.Millisecond {
		t.Errorf("FastestFirstDoor = %v, want 12.5s", h.FastestFirstDoor)
	}
}

func TestGetSessionHighlights_NoFirstDoorData(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		{SessionID: "s1", TimeToFirstDoorSecs: 0, DurationSeconds: 600},
		{SessionID: "s2", TimeToFirstDoorSecs: -1, DurationSeconds: 600},
	}

	h := pa.GetSessionHighlights()
	if h.FastestFirstDoor != 0 {
		t.Errorf("FastestFirstDoor = %v, want 0 (no valid data)", h.FastestFirstDoor)
	}
}

func TestGetSessionHighlights_StreakWithGaps(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		// Streak 1: Mar 1-3 (3 days)
		{SessionID: "s1", StartTime: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
		{SessionID: "s2", StartTime: time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC), TasksCompleted: 2, DurationSeconds: 600},
		{SessionID: "s3", StartTime: time.Date(2026, 3, 3, 10, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
		// Gap: Mar 4
		// Streak 2: Mar 5-9 (5 days)
		{SessionID: "s4", StartTime: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
		{SessionID: "s5", StartTime: time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
		{SessionID: "s6", StartTime: time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC), TasksCompleted: 3, DurationSeconds: 600},
		{SessionID: "s7", StartTime: time.Date(2026, 3, 8, 10, 0, 0, 0, time.UTC), TasksCompleted: 2, DurationSeconds: 600},
		{SessionID: "s8", StartTime: time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC), TasksCompleted: 1, DurationSeconds: 600},
	}

	h := pa.GetSessionHighlights()
	if h.LongestStreak != 5 {
		t.Errorf("LongestStreak = %d, want 5", h.LongestStreak)
	}
}

func TestGetSessionHighlights_ZeroCompletions_NoStreak(t *testing.T) {
	pa := NewPatternAnalyzer()
	pa.sessions = []SessionMetrics{
		{SessionID: "s1", StartTime: time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC), TasksCompleted: 0, DurationSeconds: 600},
		{SessionID: "s2", StartTime: time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC), TasksCompleted: 0, DurationSeconds: 600},
	}

	h := pa.GetSessionHighlights()
	if h.LongestStreak != 0 {
		t.Errorf("LongestStreak = %d, want 0 (no completions)", h.LongestStreak)
	}
}

func TestLongestConsecutiveDays(t *testing.T) {
	tests := []struct {
		name  string
		dates map[string]bool
		want  int
	}{
		{"empty", map[string]bool{}, 0},
		{"single day", map[string]bool{"2026-03-01": true}, 1},
		{"two consecutive", map[string]bool{"2026-03-01": true, "2026-03-02": true}, 2},
		{"gap breaks streak", map[string]bool{"2026-03-01": true, "2026-03-03": true}, 1},
		{"three consecutive", map[string]bool{"2026-03-01": true, "2026-03-02": true, "2026-03-03": true}, 3},
		{"two streaks picks longest", map[string]bool{
			"2026-03-01": true, "2026-03-02": true, // streak of 2
			"2026-03-05": true, "2026-03-06": true, "2026-03-07": true, // streak of 3
		}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestConsecutiveDays(tt.dates)
			if got != tt.want {
				t.Errorf("longestConsecutiveDays() = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- Helper ---

func patternContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
