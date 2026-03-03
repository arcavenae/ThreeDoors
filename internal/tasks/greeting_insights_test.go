package tasks

import (
	"strings"
	"testing"
	"time"
)

func TestFormatMultiDimensionalGreeting_FullData(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	// Setup CompletionCounter with today=3, yesterday=5
	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b", "c"},
		"2026-03-01": {"d", "e", "f", "g", "h"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("CompletionCounter LoadFromFile() error: %v", err)
	}

	// Setup PatternAnalyzer with mood data and door selections
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeSessionWithDoors(yesterday, 5, []string{"Focused"}, 4, 2),
		makeSessionWithDoors(now.Add(-2*time.Hour), 1, []string{"Focused"}, 2, 1),
		makeSessionWithDoors(now.Add(-1*time.Hour), 1, []string{"Energized"}, 1, 0),
		makeSessionWithDoors(now, 1, []string{"Focused"}, 3, 1),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("PatternAnalyzer LoadSessions() error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Should have task comparison with encouraging message
	if !strings.Contains(greeting, "tasks today vs") {
		t.Errorf("greeting should contain task comparison, got: %q", greeting)
	}
	// Should have encouraging message (declining: 3 vs 5)
	if !strings.Contains(greeting, "every one counts!") {
		t.Errorf("greeting should contain encouraging message for decline, got: %q", greeting)
	}
	// Should have doors opened
	if !strings.Contains(greeting, "doors") {
		t.Errorf("greeting should contain doors comparison, got: %q", greeting)
	}
	// Should have streak
	if !strings.Contains(greeting, "Streak:") {
		t.Errorf("greeting should contain 'Streak:', got: %q", greeting)
	}
	// Should have mood
	if !strings.Contains(greeting, "Mood: Focused") {
		t.Errorf("greeting should contain 'Mood: Focused' (most recent), got: %q", greeting)
	}
	if !strings.HasPrefix(greeting, "📈") {
		t.Errorf("greeting should start with 📈, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_TodayBetterThanYesterday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b", "c", "d", "e"},
		"2026-03-01": {"f", "g", "h"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeSessionWithDoors(now.Add(-24*time.Hour), 3, nil, 2, 1),
		makeSessionWithDoors(now, 5, nil, 4, 0),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Should have "on a roll!" for beating yesterday
	if !strings.Contains(greeting, "on a roll!") {
		t.Errorf("greeting should contain 'on a roll!' when today > yesterday, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_EqualDays(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b", "c"},
		"2026-03-01": {"d", "e", "f"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeSessionWithDoors(now.Add(-24*time.Hour), 3, nil, 2, 0),
		makeSessionWithDoors(now, 3, nil, 2, 0),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	if !strings.Contains(greeting, "steady momentum!") {
		t.Errorf("greeting should contain 'steady momentum!' for equal days, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_NoCompletionsToday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	// Setup CompletionCounter with today=0, yesterday=4
	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-01": {"a", "b", "c", "d"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("CompletionCounter LoadFromFile() error: %v", err)
	}

	// Setup PatternAnalyzer with mood data
	now := time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now, 4, []string{"Energized"}, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("PatternAnalyzer LoadSessions() error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Case: Today=0, yesterday exists — show "Yesterday: X tasks"
	if !strings.Contains(greeting, "Yesterday: 4 tasks") {
		t.Errorf("greeting should contain 'Yesterday: 4 tasks' when today=0, got: %q", greeting)
	}
	if strings.Contains(greeting, "vs") {
		t.Errorf("greeting should NOT contain comparison when today=0, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_NoMoodData(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b"},
		"2026-03-01": {"c"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("CompletionCounter LoadFromFile() error: %v", err)
	}

	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now, 2, nil, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("PatternAnalyzer LoadSessions() error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// No mood data — Mood segment omitted
	if strings.Contains(greeting, "Mood:") {
		t.Errorf("greeting should NOT contain 'Mood:' when no mood data, got: %q", greeting)
	}
	if !strings.Contains(greeting, "tasks") {
		t.Errorf("greeting should contain task info, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_NoPreviousData(t *testing.T) {
	cc := NewCompletionCounter()
	pa := NewPatternAnalyzer()

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// No data at all — empty string
	if greeting != "" {
		t.Errorf("greeting should be empty when no data, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_ZeroStreak(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 5, 14)

	// Completions from 3 days ago (streak = 0)
	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 5, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("CompletionCounter LoadFromFile() error: %v", err)
	}

	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now, 2, []string{"Focused"}, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("PatternAnalyzer LoadSessions() error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Streak = 0 — "Streak" segment omitted
	if strings.Contains(greeting, "Streak:") {
		t.Errorf("greeting should NOT contain 'Streak:' when streak=0, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_Deterministic(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a"},
		"2026-03-01": {"b"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now, 1, []string{"Calm"}, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	// Call multiple times — should return identical results (deterministic, no randomness)
	g1 := FormatMultiDimensionalGreeting(pa, cc)
	g2 := FormatMultiDimensionalGreeting(pa, cc)
	if g1 != g2 {
		t.Errorf("greeting should be deterministic: first=%q, second=%q", g1, g2)
	}
}

func TestFormatMultiDimensionalGreeting_WithMoodTrend(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b", "c", "d", "e"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	// 6 sessions needed for mood trend — preceding 3 low, recent 3 high = "improving"
	base := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(base, 1, []string{"Tired"}, nil),
		makeTestSession(base.Add(1*time.Hour), 1, []string{"Tired"}, nil),
		makeTestSession(base.Add(2*time.Hour), 1, []string{"Tired"}, nil),
		makeTestSession(base.Add(20*time.Hour), 5, []string{"Focused"}, nil),
		makeTestSession(base.Add(21*time.Hour), 5, []string{"Focused"}, nil),
		makeTestSession(base.Add(22*time.Hour), 5, []string{"Focused"}, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	if !strings.Contains(greeting, "Mood: Focused (improving)") {
		t.Errorf("greeting should contain 'Mood: Focused (improving)', got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_WeekOverWeek(t *testing.T) {
	dir := t.TempDir()
	// 2026-03-02 is a Monday
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	// This week: 5 tasks, last week: 3 tasks
	thisWeek := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	lastWeek := time.Date(2026, 2, 25, 10, 0, 0, 0, time.UTC) // previous Monday
	sessions := []SessionMetrics{
		makeTestSession(lastWeek, 3, nil, nil),
		makeTestSession(thisWeek, 5, nil, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Should have week-over-week line
	if !strings.Contains(greeting, "Week:") {
		t.Errorf("greeting should contain week-over-week line, got: %q", greeting)
	}
	if !strings.Contains(greeting, "trending up!") {
		t.Errorf("greeting should contain 'trending up!' for improving week, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_WeekOverWeekDecline(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	thisWeek := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	lastWeek := time.Date(2026, 2, 25, 10, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(lastWeek, 10, nil, nil),
		makeTestSession(thisWeek, 2, nil, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Declining week should still be encouraging
	if !strings.Contains(greeting, "building momentum") {
		t.Errorf("greeting should contain 'building momentum' for declining week, got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_DoorsOnlyToday(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeSessionWithDoors(now, 1, nil, 3, 1),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Doors today only, no yesterday comparison
	if !strings.Contains(greeting, "Doors: 3 today") {
		t.Errorf("greeting should contain 'Doors: 3 today' for today-only doors, got: %q", greeting)
	}
}

// makeSessionWithDoors creates a SessionMetrics with door selections and refresh counts.
func makeSessionWithDoors(startTime time.Time, completed int, moods []string, doorCount int, refreshes int) SessionMetrics {
	s := makeTestSession(startTime, completed, moods, nil)
	selections := make([]DoorSelectionRecord, doorCount)
	for i := range doorCount {
		selections[i] = DoorSelectionRecord{
			DoorPosition: i % 3,
			TaskText:     "task",
			Timestamp:    startTime,
		}
	}
	s.DoorSelections = selections
	s.DoorsViewed = doorCount
	s.RefreshesUsed = refreshes
	return s
}

func TestGetDayOverDay(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	yesterday := time.Date(2026, 3, 1, 14, 0, 0, 0, time.UTC)

	sessions := []SessionMetrics{
		makeSessionWithDoors(yesterday, 3, nil, 5, 2),
		makeSessionWithDoors(now, 2, nil, 4, 1),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	stats := pa.GetDayOverDay()

	tests := []struct {
		name string
		got  int
		want int
	}{
		{"TodayDoors", stats.TodayDoors, 4},
		{"YesterdayDoors", stats.YesterdayDoors, 5},
		{"TodayBypasses", stats.TodayBypasses, 1},
		{"YesterdayBypasses", stats.YesterdayBypasses, 2},
		{"TodayTasks", stats.TodayTasks, 2},
		{"YesterdayTasks", stats.YesterdayTasks, 3},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("DayOverDayStats.%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestGetDayOverDay_NoSessions(t *testing.T) {
	pa := NewPatternAnalyzer()
	stats := pa.GetDayOverDay()

	if stats.TodayDoors != 0 || stats.YesterdayDoors != 0 {
		t.Errorf("empty analyzer should return zero stats, got: %+v", stats)
	}
}

func TestGetMoodTrend(t *testing.T) {
	tests := []struct {
		name        string
		completions []int // 6 sessions' completion counts
		want        string
	}{
		{
			name:        "improving — recent higher than preceding",
			completions: []int{1, 1, 1, 5, 5, 5},
			want:        "improving",
		},
		{
			name:        "declining — recent lower than preceding",
			completions: []int{5, 5, 5, 1, 1, 1},
			want:        "declining",
		},
		{
			name:        "stable — similar averages",
			completions: []int{3, 3, 3, 3, 3, 3},
			want:        "stable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			base := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
			var sessions []SessionMetrics
			for i, c := range tt.completions {
				sessions = append(sessions, makeTestSession(base.Add(time.Duration(i)*time.Hour), c, nil, nil))
			}
			paPath := writeSessionsFile(t, dir, sessions)
			pa := NewPatternAnalyzerWithNow(frozenTimePA(2026, 3, 2, 14))
			if err := pa.LoadSessions(paPath); err != nil {
				t.Fatalf("error: %v", err)
			}

			got := pa.GetMoodTrend()
			if got != tt.want {
				t.Errorf("GetMoodTrend() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetMoodTrend_InsufficientData(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2026, 3, 2, 10, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(base, 1, nil, nil),
		makeTestSession(base.Add(time.Hour), 2, nil, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozenTimePA(2026, 3, 2, 14))
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("error: %v", err)
	}

	got := pa.GetMoodTrend()
	if got != "" {
		t.Errorf("GetMoodTrend() with <6 sessions = %q, want empty", got)
	}
}

func TestEncouragingComparison(t *testing.T) {
	tests := []struct {
		name      string
		today     int
		yesterday int
		dimension string
		wantSub   string
	}{
		{"better", 5, 3, "tasks", "on a roll!"},
		{"equal", 3, 3, "tasks", "steady momentum!"},
		{"worse", 2, 5, "tasks", "every one counts!"},
		{"doors better", 8, 4, "doors", "on a roll!"},
		{"doors worse", 2, 6, "doors", "every one counts!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encouragingComparison(tt.today, tt.yesterday, tt.dimension)
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("encouragingComparison(%d, %d, %q) = %q, want substring %q",
					tt.today, tt.yesterday, tt.dimension, got, tt.wantSub)
			}
			// Should always contain the numbers
			if !strings.Contains(got, tt.dimension) {
				t.Errorf("result should contain dimension %q, got: %q", tt.dimension, got)
			}
		})
	}
}

func TestFormatWeekComparison(t *testing.T) {
	tests := []struct {
		name      string
		wk        WeekComparison
		wantSub   string
		wantEmpty bool
	}{
		{
			name:    "up",
			wk:      WeekComparison{ThisWeekTotal: 10, LastWeekTotal: 5, Direction: "up"},
			wantSub: "trending up!",
		},
		{
			name:    "down",
			wk:      WeekComparison{ThisWeekTotal: 3, LastWeekTotal: 10, Direction: "down"},
			wantSub: "building momentum",
		},
		{
			name:    "same",
			wk:      WeekComparison{ThisWeekTotal: 5, LastWeekTotal: 5, Direction: "same"},
			wantSub: "steady pace",
		},
		{
			name:      "no data",
			wk:        WeekComparison{},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatWeekComparison(tt.wk)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("formatWeekComparison() = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantSub) {
				t.Errorf("formatWeekComparison() = %q, want substring %q", got, tt.wantSub)
			}
		})
	}
}
