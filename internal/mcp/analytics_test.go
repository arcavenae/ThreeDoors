package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/core/metrics"
)

// writeSessions creates a temp JSONL file with the given sessions and returns a Reader.
func writeSessions(t *testing.T, sessions []core.SessionMetrics) *metrics.Reader {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sessions.jsonl")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
		return nil
	}
	t.Cleanup(func() { _ = f.Close() })

	enc := json.NewEncoder(f)
	for _, s := range sessions {
		if err := enc.Encode(s); err != nil {
			t.Fatalf("encode session: %v", err)
		}
	}
	_ = f.Close()

	return metrics.NewReader(path)
}

func syntheticSessions(days int) []core.SessionMetrics {
	now := time.Now().UTC()
	var sessions []core.SessionMetrics

	moods := []string{"great", "good", "okay", "low", "bad"}

	for i := 0; i < days; i++ {
		day := now.AddDate(0, 0, -i)
		hour := 9 + (i % 12) // vary start hour

		startTime := time.Date(day.Year(), day.Month(), day.Day(), hour, 0, 0, 0, time.UTC)
		endTime := startTime.Add(time.Duration(30+i*5) * time.Minute)

		completions := (i % 5) + 1
		mood := moods[i%len(moods)]

		s := core.SessionMetrics{
			SessionID:       "sess-" + startTime.Format("20060102"),
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: endTime.Sub(startTime).Seconds(),
			TasksCompleted:  completions,
			DoorsViewed:     completions + 2,
			RefreshesUsed:   i % 3,
			MoodEntryCount:  1,
			MoodEntries: []core.MoodEntry{
				{Timestamp: startTime, Mood: mood},
			},
			DoorSelections: []core.DoorSelectionRecord{
				{Timestamp: startTime, DoorPosition: 0, TaskText: "task-" + mood},
			},
		}
		sessions = append(sessions, s)
	}

	return sessions
}

func TestMoodCorrelationAnalysis(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(14)
	reader := writeSessions(t, sessions)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	now := time.Now().UTC()
	result, err := pm.MoodCorrelationAnalysis(now.AddDate(0, 0, -30), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("expected non-nil result")
		return
	}

	if len(result.Entries) == 0 {
		t.Fatal("expected mood correlation entries")
	}

	// Verify all mood entries have valid data.
	for _, entry := range result.Entries {
		if entry.Mood == "" {
			t.Error("mood entry has empty mood")
		}
		if entry.SessionCount <= 0 {
			t.Errorf("mood %q has zero session count", entry.Mood)
		}
		if entry.AvgCompletions < 0 {
			t.Errorf("mood %q has negative avg completions", entry.Mood)
		}
	}
}

func TestMoodCorrelationNoSessions(t *testing.T) {
	t.Parallel()

	reader := writeSessions(t, nil)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	now := time.Now().UTC()
	result, err := pm.MoodCorrelationAnalysis(now.AddDate(0, 0, -7), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestProductivityProfileAnalysis(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(14)
	reader := writeSessions(t, sessions)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	now := time.Now().UTC()
	result, err := pm.ProductivityProfileAnalysis(now.AddDate(0, 0, -30), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("expected non-nil result")
		return
	}

	if len(result.HourlyData) != 24 {
		t.Errorf("expected 24 hourly entries, got %d", len(result.HourlyData))
	}

	// At least one hour should have sessions.
	var hasActivity bool
	for _, h := range result.HourlyData {
		if h.Sessions > 0 {
			hasActivity = true
			break
		}
	}
	if !hasActivity {
		t.Error("expected at least one hour with activity")
	}

	if len(result.PeakHours) == 0 {
		t.Error("expected peak hours to be identified")
	}
}

func TestStreakAnalysis(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(10)
	reader := writeSessions(t, sessions)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	result, err := pm.StreakAnalysis()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("expected non-nil result")
		return
	}

	if result.LongestStreak <= 0 {
		t.Error("expected positive longest streak")
	}

	if len(result.StreakHistory) == 0 {
		t.Error("expected streak history entries")
	}

	if result.AverageStreak <= 0 {
		t.Error("expected positive average streak")
	}
}

func TestStreakAnalysisNoSessions(t *testing.T) {
	t.Parallel()

	reader := writeSessions(t, nil)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	result, err := pm.StreakAnalysis()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if result.CurrentStreak != 0 {
		t.Errorf("expected 0 current streak, got %d", result.CurrentStreak)
	}
	if result.LongestStreak != 0 {
		t.Errorf("expected 0 longest streak, got %d", result.LongestStreak)
	}
}

func TestBurnoutRisk(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(14)
	reader := writeSessions(t, sessions)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	result, err := pm.BurnoutRisk()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("expected non-nil result")
		return
	}

	if result.Score < 0 || result.Score > 1 {
		t.Errorf("score %f out of 0-1 range", result.Score)
	}

	validLevels := map[string]bool{"low": true, "moderate": true, "warning": true, "unknown": true}
	if !validLevels[result.Level] {
		t.Errorf("unexpected level: %s", result.Level)
	}
}

func TestBurnoutRiskNoSessions(t *testing.T) {
	t.Parallel()

	reader := writeSessions(t, nil)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	result, err := pm.BurnoutRisk()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if result.Level != "unknown" {
		t.Errorf("expected unknown level, got %s", result.Level)
	}
}

func TestWeeklySummaryAnalysis(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(7)
	reader := writeSessions(t, sessions)
	pm := NewPatternMiner(reader, core.NewTaskPool())

	result, err := pm.WeeklySummaryAnalysis(time.Now().UTC())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("expected non-nil result")
		return
	}

	if result.Velocity < 0 {
		t.Errorf("expected non-negative velocity, got %d", result.Velocity)
	}
}

func TestTrendSlope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		values   []float64
		wantSign int // -1, 0, 1
	}{
		{"increasing", []float64{1, 2, 3, 4, 5}, 1},
		{"decreasing", []float64{5, 4, 3, 2, 1}, -1},
		{"flat", []float64{3, 3, 3, 3}, 0},
		{"single", []float64{5}, 0},
		{"empty", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var sessions []core.SessionMetrics
			for _, v := range tt.values {
				sessions = append(sessions, core.SessionMetrics{TasksCompleted: int(v)})
			}

			slope := trendSlope(sessions, func(s core.SessionMetrics) float64 {
				return float64(s.TasksCompleted)
			})

			switch tt.wantSign {
			case 1:
				if slope <= 0 {
					t.Errorf("expected positive slope, got %f", slope)
				}
			case -1:
				if slope >= 0 {
					t.Errorf("expected negative slope, got %f", slope)
				}
			case 0:
				if slope != 0 {
					t.Errorf("expected zero slope, got %f", slope)
				}
			}
		})
	}
}

func TestDaysBetween(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b time.Time
		want int
	}{
		{"same day", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 23, 0, 0, 0, time.UTC), 0},
		{"one day", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), 1},
		{"five days", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 6, 0, 0, 0, 0, time.UTC), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := daysBetween(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("daysBetween = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestAnalyticsResourcesRegistered verifies that analytics resources appear in resources/list.
func TestAnalyticsResourcesRegistered(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "resources/list",
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result ResourcesListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	uris := make(map[string]bool)
	for _, r := range result.Resources {
		uris[r.URI] = true
	}

	expectedAnalytics := []string{
		"threedoors://analytics/mood-correlation",
		"threedoors://analytics/time-of-day",
		"threedoors://analytics/streaks",
		"threedoors://analytics/burnout-risk",
		"threedoors://analytics/task-preferences",
		"threedoors://analytics/weekly-summary",
	}
	for _, uri := range expectedAnalytics {
		if !uris[uri] {
			t.Errorf("missing analytics resource URI: %s", uri)
		}
	}
}

// TestAnalyticsToolsRegistered verifies that analytics tools appear in tools/list.
func TestAnalyticsToolsRegistered(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "tools/list",
	})

	resultBytes, _ := json.Marshal(resp.Result)
	var result ToolsListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	names := make(map[string]bool)
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}

	expectedTools := []string{
		"get_mood_correlation",
		"get_productivity_profile",
		"burnout_risk",
		"get_completions",
	}
	for _, name := range expectedTools {
		if !names[name] {
			t.Errorf("missing analytics tool: %s", name)
		}
	}
}

// TestPromptsRegistered verifies that prompt templates appear in prompts/list.
func TestPromptsRegistered(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "prompts/list",
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result PromptsListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Prompts) != 4 {
		t.Errorf("expected 4 prompts, got %d", len(result.Prompts))
	}

	names := make(map[string]bool)
	for _, p := range result.Prompts {
		names[p.Name] = true
	}

	if !names["daily_summary"] {
		t.Error("missing prompt: daily_summary")
	}
	if !names["weekly_retrospective"] {
		t.Error("missing prompt: weekly_retrospective")
	}
}

// TestPromptsGet verifies prompts/get returns prompt content.
func TestPromptsGet(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	params, _ := json.Marshal(map[string]string{"name": "daily_summary"})
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "prompts/get",
		Params: params,
	})

	if resp.Error != nil {
		t.Fatalf("unexpected error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
		return
	}

	resultBytes, _ := json.Marshal(resp.Result)
	var result PromptGetResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Messages) == 0 {
		t.Fatal("expected prompt messages")
	}

	if result.Messages[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", result.Messages[0].Role)
	}

	if result.Messages[0].Content.Text == "" {
		t.Error("expected non-empty prompt text")
	}
}

// TestPromptsGetUnknown verifies prompts/get returns error for unknown prompt.
func TestPromptsGetUnknown(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	params, _ := json.Marshal(map[string]string{"name": "nonexistent"})
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "prompts/get",
		Params: params,
	})

	if resp.Error == nil {
		t.Fatal("expected error for unknown prompt")
		return
	}
}

// TestAnalyticsToolCallMoodCorrelation tests the tool call end-to-end.
func TestAnalyticsToolCallMoodCorrelation(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(7)
	reader := writeSessions(t, sessions)

	s := newTestServer()
	s.SetSessionsReader(reader)

	resp := dispatchToolCall(t, s, "get_mood_correlation", nil)
	text := parseToolText(t, resp)

	var result MoodCorrelation
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Entries) == 0 {
		t.Error("expected mood correlation entries from tool call")
	}
}

// TestAnalyticsToolCallBurnoutRisk tests the burnout_risk tool.
func TestAnalyticsToolCallBurnoutRisk(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(14)
	reader := writeSessions(t, sessions)

	s := newTestServer()
	s.SetSessionsReader(reader)

	resp := dispatchToolCall(t, s, "burnout_risk", nil)
	text := parseToolText(t, resp)

	var result BurnoutIndicators
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result.Score < 0 || result.Score > 1 {
		t.Errorf("score out of range: %f", result.Score)
	}
}

// TestAnalyticsToolCallNoReader tests analytics tools when no reader is set.
func TestAnalyticsToolCallNoReader(t *testing.T) {
	t.Parallel()

	s := newTestServer()

	tools := []string{"get_mood_correlation", "get_productivity_profile", "burnout_risk"}
	for _, tool := range tools {
		resp := dispatchToolCall(t, s, tool, nil)
		if resp.Error != nil {
			t.Errorf("tool %s: unexpected JSON-RPC error: %v", tool, resp.Error)
			continue
		}
		resultBytes, _ := json.Marshal(resp.Result)
		var result ToolCallResult
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !result.IsError {
			t.Errorf("tool %s: expected isError=true when no reader", tool)
		}
	}
}

// TestAnalyticsToolCallGetCompletions tests the get_completions tool.
func TestAnalyticsToolCallGetCompletions(t *testing.T) {
	t.Parallel()

	sessions := syntheticSessions(7)
	reader := writeSessions(t, sessions)

	s := newTestServer()
	s.SetSessionsReader(reader)

	resp := dispatchToolCall(t, s, "get_completions", map[string]any{
		"group_by": "day",
	})
	text := parseToolText(t, resp)

	var result struct {
		GroupBy     string `json:"group_by"`
		Completions []struct {
			Key         string `json:"key"`
			Completions int    `json:"completions"`
		} `json:"completions"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result.GroupBy != "day" {
		t.Errorf("expected group_by=day, got %s", result.GroupBy)
	}

	if len(result.Completions) == 0 {
		t.Error("expected completion data")
	}
}

// TestPatternMinerNilReader verifies nil reader produces no errors.
func TestPatternMinerNilReader(t *testing.T) {
	t.Parallel()

	pm := NewPatternMiner(nil, core.NewTaskPool())

	now := time.Now().UTC()
	result, err := pm.MoodCorrelationAnalysis(now.AddDate(0, 0, -7), now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries with nil reader, got %d", len(result.Entries))
	}

	streaks, err := pm.StreakAnalysis()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if streaks.CurrentStreak != 0 {
		t.Errorf("expected 0 current streak with nil reader")
	}

	burnout, err := pm.BurnoutRisk()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if burnout.Level != "unknown" {
		t.Errorf("expected unknown level with nil reader, got %s", burnout.Level)
	}
}
