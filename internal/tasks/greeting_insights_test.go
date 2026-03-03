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

	// Setup PatternAnalyzer with mood data
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now.Add(-2*time.Hour), 3, []string{"Focused"}, nil),
		makeTestSession(now.Add(-1*time.Hour), 2, []string{"Energized"}, nil),
		makeTestSession(now, 4, []string{"Focused"}, nil),
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("PatternAnalyzer LoadSessions() error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Case 1: Today has completions AND yesterday data exists
	// Format: 📈 Today: {today} | Yesterday: {yesterday} | Streak: {streak} days | Mood: {lastMood}
	if !strings.Contains(greeting, "Today: 3") {
		t.Errorf("greeting should contain 'Today: 3', got: %q", greeting)
	}
	if !strings.Contains(greeting, "Yesterday: 5") {
		t.Errorf("greeting should contain 'Yesterday: 5', got: %q", greeting)
	}
	if !strings.Contains(greeting, "Streak:") {
		t.Errorf("greeting should contain 'Streak:', got: %q", greeting)
	}
	if !strings.Contains(greeting, "Mood: Focused") {
		t.Errorf("greeting should contain 'Mood: Focused' (most recent), got: %q", greeting)
	}
	if !strings.HasPrefix(greeting, "📈") {
		t.Errorf("greeting should start with 📈, got: %q", greeting)
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

	// Case 2: Today=0, yesterday exists
	// Format: 📈 Yesterday: {yesterday} | Streak: {streak} days | Mood: {lastMood}
	if strings.Contains(greeting, "Today:") {
		t.Errorf("greeting should NOT contain 'Today:' when today=0, got: %q", greeting)
	}
	if !strings.Contains(greeting, "Yesterday: 4") {
		t.Errorf("greeting should contain 'Yesterday: 4', got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_NoMoodData(t *testing.T) {
	dir := t.TempDir()
	frozen := frozenTimePA(2026, 3, 2, 14)

	// Setup CompletionCounter with today=2, yesterday=1
	ccPath := writeCompletedFile(t, dir, map[string][]string{
		"2026-03-02": {"a", "b"},
		"2026-03-01": {"c"},
	})
	cc := NewCompletionCounterWithNow(frozenTime(2026, 3, 2, 14))
	if err := cc.LoadFromFile(ccPath); err != nil {
		t.Fatalf("CompletionCounter LoadFromFile() error: %v", err)
	}

	// Setup PatternAnalyzer with NO mood data
	now := time.Date(2026, 3, 2, 14, 0, 0, 0, time.UTC)
	sessions := []SessionMetrics{
		makeTestSession(now, 2, nil, nil), // no moods
	}
	paPath := writeSessionsFile(t, dir, sessions)
	pa := NewPatternAnalyzerWithNow(frozen)
	if err := pa.LoadSessions(paPath); err != nil {
		t.Fatalf("PatternAnalyzer LoadSessions() error: %v", err)
	}

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Case 3: No mood data — Mood segment omitted
	if strings.Contains(greeting, "Mood:") {
		t.Errorf("greeting should NOT contain 'Mood:' when no mood data, got: %q", greeting)
	}
	if !strings.Contains(greeting, "Today: 2") {
		t.Errorf("greeting should contain 'Today: 2', got: %q", greeting)
	}
}

func TestFormatMultiDimensionalGreeting_NoPreviousData(t *testing.T) {
	// Empty analyzer and counter
	cc := NewCompletionCounter()
	pa := NewPatternAnalyzer()

	greeting := FormatMultiDimensionalGreeting(pa, cc)

	// Case 4: No data at all — empty string
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

	// Case 5: Streak = 0 — "Streak" segment omitted
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
