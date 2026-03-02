package tasks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Test Helpers ---

var baseSessionTime = time.Date(2025, 11, 10, 9, 0, 0, 0, time.UTC)

func makeTestSession(id string, start time.Time, completed int, selections []DoorSelectionRecord, bypasses [][]string, moods []MoodEntry) SessionMetrics {
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

func writeSessionsFile(t *testing.T, dir string, sessions []SessionMetrics) string {
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
		sessions[i] = makeTestSession(
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

// --- ReadSessions Tests ---

func TestReadSessions_ValidFile(t *testing.T) {
	dir := t.TempDir()
	sessions := makeFiveMinimalSessions()
	path := writeSessionsFile(t, dir, sessions)

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
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
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
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
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
			// One session has a left pick to break 100%
			selections[1] = DoorSelectionRecord{DoorPosition: 0, TaskText: "task-c"}
		}
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1, selections, nil, nil)
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
		// 9am, 10am, 11am, 9am, 10am
		hour := 9 + (i % 3)
		start := time.Date(2025, 11, 10+i, hour, 0, 0, 0, time.UTC)
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), start, 2+i, nil, nil, nil)
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
	times := []int{9, 14, 20, 23, 10} // morning, afternoon, evening, night, morning
	sessions := make([]SessionMetrics, 5)
	for i, hour := range times {
		start := time.Date(2025, 11, 10+i, hour, 0, 0, 0, time.UTC)
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), start, 1, nil, nil, nil)
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
			bypasses = nil // Only bypass in first 3 sessions
		}
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
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
	sessions[0] = makeTestSession("s0", baseSessionTime, 1,
		[]DoorSelectionRecord{{DoorPosition: 0, TaskText: "Selected task"}},
		[][]string{{"Rare bypass task"}}, nil)
	for i := 1; i < 5; i++ {
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
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
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
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

// --- Mood Correlation Tests ---

func TestAnalyze_MoodCorrelations_FocusedTechnical(t *testing.T) {
	sessions := make([]SessionMetrics, 5)
	for i := 0; i < 5; i++ {
		moods := []MoodEntry{{
			Timestamp: baseSessionTime.Add(time.Duration(i) * 24 * time.Hour),
			Mood:      "focused",
		}}
		selections := []DoorSelectionRecord{{
			DoorPosition: 1,
			TaskText:     "Fix API bug", // technical task
		}}
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1, selections, nil, moods)
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
	// Only 2 sessions have mood "tired"
	for i := 0; i < 5; i++ {
		var moods []MoodEntry
		if i < 2 {
			moods = []MoodEntry{{Mood: "tired"}}
		}
		sessions[i] = makeTestSession("s"+string(rune('0'+i)), baseSessionTime.Add(time.Duration(i)*24*time.Hour), 1,
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

	// Check key sections are present
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

	// Create realistic sessions
	sessions := make([]SessionMetrics, 7)
	for i := 0; i < 7; i++ {
		start := time.Date(2025, 11, 10+i, 9, 0, 0, 0, time.UTC) // All morning
		sessions[i] = makeTestSession("sess-"+string(rune('a'+i)), start, 2,
			[]DoorSelectionRecord{
				{Timestamp: start.Add(2 * time.Minute), DoorPosition: 0, TaskText: "Fix API bug"},
			},
			[][]string{{"Buy groceries", "Write report"}},
			[]MoodEntry{{Timestamp: start, Mood: "focused"}},
		)
	}

	sessionsPath := writeSessionsFile(t, dir, sessions)
	patternsPath := filepath.Join(dir, "patterns.json")

	analyzer := NewPatternAnalyzer()

	// Read
	loaded, err := analyzer.ReadSessions(sessionsPath)
	if err != nil {
		t.Fatalf("ReadSessions() error = %v", err)
	}
	if len(loaded) != 7 {
		t.Fatalf("ReadSessions() returned %d, want 7", len(loaded))
	}

	// Analyze
	report, err := analyzer.Analyze(loaded)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report == nil {
		t.Fatal("Analyze() returned nil with 7 sessions")
	}

	// Verify key findings
	if report.DoorPositionBias.PreferredPosition != "left" {
		t.Errorf("Expected left bias, got %q", report.DoorPositionBias.PreferredPosition)
	}

	// All sessions in morning
	morningFound := false
	for _, p := range report.TimeOfDayPatterns {
		if p.Period == "morning" && p.SessionCount == 7 {
			morningFound = true
		}
	}
	if !morningFound {
		t.Error("Expected all 7 sessions in morning period")
	}

	// "Buy groceries" bypassed 7 times
	groceriesFound := false
	for _, a := range report.AvoidanceList {
		if a.TaskText == "Buy groceries" && a.TimesBypassed >= 3 {
			groceriesFound = true
		}
	}
	if !groceriesFound {
		t.Error("Expected 'Buy groceries' in avoidance list")
	}

	// Save and reload
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

	// Format insights
	output := FormatInsights(report)
	if output == "" {
		t.Error("FormatInsights() returned empty")
	}
}

// --- Helper ---

func patternContains(s, substr string) bool {
	return strings.Contains(s, substr)
}
