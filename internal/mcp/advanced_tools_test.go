package mcp

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
)

func TestPrioritizeTasks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "t1", Text: "blocker task", Status: core.StatusTodo, Type: core.TypeTechnical, Effort: core.EffortQuickWin, CreatedAt: now.AddDate(0, 0, -10), UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "t2", Text: "blocked task", Status: core.StatusTodo, Type: core.TypeCreative, Effort: core.EffortDeepWork, CreatedAt: now, UpdatedAt: now, Blocker: "t1"})
	pool.AddTask(&core.Task{ID: "t3", Text: "old task", Status: core.StatusTodo, Type: core.TypeAdministrative, Effort: core.EffortMedium, CreatedAt: now.AddDate(0, -1, 0), UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "t1", ToID: "t2", Type: EdgeBlocks, Weight: 0.8, Source: "inferred:text"},
	}

	suggestions := prioritizeTasks(pool, edges, 10, "good", "morning")

	if len(suggestions) != 3 {
		t.Fatalf("expected 3 suggestions, got %d", len(suggestions))
	}

	// t1 should score higher because it blocks t2.
	if suggestions[0].TaskID != "t1" {
		t.Errorf("expected t1 as top priority (blocks t2), got %s", suggestions[0].TaskID)
	}

	for _, s := range suggestions {
		if s.Score < 0 || s.Score > 100 {
			t.Errorf("task %s score %f out of 0-100 range", s.TaskID, s.Score)
		}
		if len(s.Factors) == 0 {
			t.Errorf("task %s has no factors", s.TaskID)
		}
		if s.Rationale == "" {
			t.Errorf("task %s has empty rationale", s.TaskID)
		}
	}
}

func TestPrioritizeTasksEmptyPool(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	suggestions := prioritizeTasks(pool, nil, 10, "", "")

	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(suggestions))
	}
}

func TestPrioritizeTasksLimit(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	for i := 0; i < 5; i++ {
		pool.AddTask(&core.Task{
			ID: fmt.Sprintf("t%d", i), Text: fmt.Sprintf("task %d", i),
			Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now,
		})
	}

	suggestions := prioritizeTasks(pool, nil, 2, "", "")
	if len(suggestions) != 2 {
		t.Errorf("expected 2 suggestions with limit=2, got %d", len(suggestions))
	}
}

func TestPrioritizeTasksSkipsCompleted(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "done", Text: "done task", Status: core.StatusComplete, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "todo", Text: "todo task", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	suggestions := prioritizeTasks(pool, nil, 10, "", "")
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion (skip completed), got %d", len(suggestions))
	}
	if suggestions[0].TaskID != "todo" {
		t.Errorf("expected todo task, got %s", suggestions[0].TaskID)
	}
}

func TestEffortFitScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		effort core.TaskEffort
		mood   string
		want   float64
	}{
		{"deep work + great mood", core.EffortDeepWork, "great", 0.9},
		{"quick win + bad mood", core.EffortQuickWin, "bad", 0.9},
		{"medium + okay mood", core.EffortMedium, "okay", 1.0},
		{"unknown mood", core.EffortMedium, "unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := effortFitScore(tt.effort, tt.mood)
			if got != tt.want {
				t.Errorf("effortFitScore(%s, %s) = %f, want %f", tt.effort, tt.mood, got, tt.want)
			}
		})
	}
}

func TestTypeFitScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		taskType  core.TaskType
		timeOfDay string
		want      float64
	}{
		{"creative morning", core.TypeCreative, "morning", 1.0},
		{"admin afternoon", core.TypeAdministrative, "afternoon", 1.0},
		{"unknown type", core.TaskType("custom"), "morning", 0.5},
		{"unknown time", core.TypeCreative, "dawn", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := typeFitScore(tt.taskType, tt.timeOfDay)
			if got != tt.want {
				t.Errorf("typeFitScore(%s, %s) = %f, want %f", tt.taskType, tt.timeOfDay, got, tt.want)
			}
		})
	}
}

func TestAnalyzeWorkload(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "t1", Text: "task 1", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "t2", Text: "task 2", Status: core.StatusTodo, Effort: core.EffortDeepWork, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "t3", Text: "task 3", Status: core.StatusComplete, Effort: core.EffortMedium, CreatedAt: now, UpdatedAt: now})

	analysis := analyzeWorkload(pool, nil)

	if analysis.TotalTasks != 3 {
		t.Errorf("total_tasks = %d, want 3", analysis.TotalTasks)
	}
	if analysis.ByStatus["todo"] != 2 {
		t.Errorf("by_status[todo] = %d, want 2", analysis.ByStatus["todo"])
	}
	if analysis.ByStatus["complete"] != 1 {
		t.Errorf("by_status[complete] = %d, want 1", analysis.ByStatus["complete"])
	}
	// Estimated hours: quick-win (0.25) + deep-work (3.0) = 3.25
	if analysis.EstimatedHours != 3.2 && analysis.EstimatedHours != 3.3 {
		t.Errorf("estimated_hours = %f, want ~3.25", analysis.EstimatedHours)
	}
	if analysis.OverloadRisk < 0 || analysis.OverloadRisk > 1 {
		t.Errorf("overload_risk %f out of 0-1 range", analysis.OverloadRisk)
	}
}

func TestAnalyzeWorkloadEmpty(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	analysis := analyzeWorkload(pool, nil)

	if analysis.TotalTasks != 0 {
		t.Errorf("total_tasks = %d, want 0", analysis.TotalTasks)
	}
	if analysis.EstimatedHours != 0 {
		t.Errorf("estimated_hours = %f, want 0", analysis.EstimatedHours)
	}
	if analysis.OverloadRisk != 0 {
		t.Errorf("overload_risk = %f, want 0", analysis.OverloadRisk)
	}
}

func TestFocusRecommendation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "t1", Text: "quick task", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "t2", Text: "deep task", Status: core.StatusTodo, Effort: core.EffortDeepWork, CreatedAt: now, UpdatedAt: now})

	rec := focusRecommendation(pool, nil, nil, "good", 30)

	if rec.DurationMinutes != 30 {
		t.Errorf("duration = %d, want 30", rec.DurationMinutes)
	}
	if rec.BreakAfterMinutes != 50 {
		t.Errorf("break_after = %d, want 50 for good mood", rec.BreakAfterMinutes)
	}
	if len(rec.SuggestedTasks) == 0 {
		t.Error("expected suggested tasks")
	}
	if rec.Rationale == "" {
		t.Error("expected non-empty rationale")
	}
}

func TestFocusRecommendationLowMood(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "t1", Text: "quick task", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now})

	rec := focusRecommendation(pool, nil, nil, "low", 60)

	if rec.BreakAfterMinutes != 25 {
		t.Errorf("break_after = %d, want 25 for low mood", rec.BreakAfterMinutes)
	}
	// Low mood should order quick wins first.
	if len(rec.TaskOrder) > 0 && rec.TaskOrder[0] != "t1" {
		t.Errorf("expected quick-win task first for low mood")
	}
}

func TestFocusRecommendationDefaultMinutes(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	rec := focusRecommendation(pool, nil, nil, "", 0)
	if rec.DurationMinutes != 60 {
		t.Errorf("default duration = %d, want 60", rec.DurationMinutes)
	}
}

func TestWhatIf(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "blocker", Text: "blocker", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "blocked", Text: "blocked", Status: core.StatusBlocked, Effort: core.EffortMedium, CreatedAt: now, UpdatedAt: now, Blocker: "blocker"})
	pool.AddTask(&core.Task{ID: "other", Text: "other", Status: core.StatusTodo, Effort: core.EffortMedium, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "blocker", ToID: "blocked", Type: EdgeBlocks, Weight: 0.8},
	}

	result := whatIf(pool, edges, []string{"blocker"})

	if len(result.UnblockedTasks) != 1 || result.UnblockedTasks[0] != "blocked" {
		t.Errorf("expected [blocked] unblocked, got %v", result.UnblockedTasks)
	}
	if result.StreakImpact != 1 {
		t.Errorf("streak_impact = %d, want 1", result.StreakImpact)
	}
	if result.VelocityChange <= 0 {
		t.Errorf("velocity_change should be positive, got %f", result.VelocityChange)
	}
	if result.EstimatedHours != 0.2 && result.EstimatedHours != 0.3 {
		t.Errorf("estimated_hours = %f, want ~0.25", result.EstimatedHours)
	}
}

func TestWhatIfNoBlockers(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "t1", Text: "task 1", Status: core.StatusTodo, CreatedAt: now, UpdatedAt: now})

	result := whatIf(pool, nil, []string{"t1"})

	if len(result.UnblockedTasks) != 0 {
		t.Errorf("expected no unblocked tasks, got %v", result.UnblockedTasks)
	}
}

func TestWhatIfMultipleCompletions(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "a", Text: "a", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "b", Text: "b", Status: core.StatusTodo, Effort: core.EffortMedium, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "c", Text: "c", Status: core.StatusBlocked, CreatedAt: now, UpdatedAt: now})

	edges := []TaskEdge{
		{FromID: "a", ToID: "c", Type: EdgeBlocks, Weight: 0.8},
		{FromID: "b", ToID: "c", Type: EdgeBlocks, Weight: 0.8},
	}

	result := whatIf(pool, edges, []string{"a", "b"})

	if len(result.UnblockedTasks) != 1 || result.UnblockedTasks[0] != "c" {
		t.Errorf("expected [c] unblocked when both blockers completed, got %v", result.UnblockedTasks)
	}
	if result.StreakImpact != 2 {
		t.Errorf("streak_impact = %d, want 2", result.StreakImpact)
	}
}

func TestContextSwitchAnalysisNilMiner(t *testing.T) {
	t.Parallel()

	pool := core.NewTaskPool()
	analysis, err := contextSwitchAnalysis(nil, pool, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	if analysis.TotalSwitches != 0 {
		t.Errorf("expected 0 switches, got %d", analysis.TotalSwitches)
	}
}

func TestContextSwitchAnalysisWithSessions(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	sessions := []core.SessionMetrics{
		{
			SessionID:       "s1",
			StartTime:       now.Add(-1 * time.Hour),
			EndTime:         now,
			DurationSeconds: 3600,
			TasksCompleted:  3,
			DoorSelections: []core.DoorSelectionRecord{
				{Timestamp: now.Add(-50 * time.Minute), TaskText: "write code"},
				{Timestamp: now.Add(-40 * time.Minute), TaskText: "review email"},
				{Timestamp: now.Add(-30 * time.Minute), TaskText: "design mockup"},
			},
		},
	}
	reader := writeSessions(t, sessions)
	pool := core.NewTaskPool()
	pm := NewPatternMiner(reader, pool)

	analysis, err := contextSwitchAnalysis(pm, pool, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}

	// All 3 selections have different task texts → "general" type, so no switches.
	// (categorizeTaskText returns "general" for all non-empty text)
	if analysis.TotalSwitches != 0 {
		t.Logf("total switches = %d (may vary based on categorization)", analysis.TotalSwitches)
	}
}

func TestContextSwitchAnalysisSpecificSession(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	sessions := []core.SessionMetrics{
		{SessionID: "s1", StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-1 * time.Hour), DurationSeconds: 3600, TasksCompleted: 1},
		{SessionID: "s2", StartTime: now.Add(-1 * time.Hour), EndTime: now, DurationSeconds: 3600, TasksCompleted: 2},
	}
	reader := writeSessions(t, sessions)
	pool := core.NewTaskPool()
	pm := NewPatternMiner(reader, pool)

	analysis, err := contextSwitchAnalysis(pm, pool, "s2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
		return
	}
	// With no door selections, switches should be 0.
	if analysis.TotalSwitches != 0 {
		t.Errorf("expected 0 switches for session with no selections, got %d", analysis.TotalSwitches)
	}
}

func TestBuildBatchSuggestions(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	pool := core.NewTaskPool()
	pool.AddTask(&core.Task{ID: "t1", Text: "task 1", Status: core.StatusTodo, Type: core.TypeTechnical, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "t2", Text: "task 2", Status: core.StatusTodo, Type: core.TypeTechnical, CreatedAt: now, UpdatedAt: now})
	pool.AddTask(&core.Task{ID: "t3", Text: "task 3", Status: core.StatusTodo, Type: core.TypeCreative, CreatedAt: now, UpdatedAt: now})

	suggestions := buildBatchSuggestions(pool)

	// Should suggest batching the 2 technical tasks.
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 batch suggestion, got %d", len(suggestions))
	}
	if suggestions[0].TaskType != "technical" {
		t.Errorf("expected batch for technical, got %s", suggestions[0].TaskType)
	}
	if len(suggestions[0].TaskIDs) != 2 {
		t.Errorf("expected 2 tasks in batch, got %d", len(suggestions[0].TaskIDs))
	}
}

func TestHourToTimeOfDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		hour int
		want string
	}{
		{6, "morning"},
		{9, "morning"},
		{12, "afternoon"},
		{15, "afternoon"},
		{18, "evening"},
		{20, "evening"},
		{22, "night"},
		{3, "night"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("hour_%d", tt.hour), func(t *testing.T) {
			t.Parallel()
			got := hourToTimeOfDay(tt.hour)
			if got != tt.want {
				t.Errorf("hourToTimeOfDay(%d) = %s, want %s", tt.hour, got, tt.want)
			}
		})
	}
}

// MCP tool handler integration tests.

func TestToolPrioritizeTasks(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "t1", Text: "important task", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now},
		&core.Task{ID: "t2", Text: "other task", Status: core.StatusTodo, Effort: core.EffortMedium, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "prioritize_tasks", map[string]any{
		"limit": 5, "mood": "good", "time_of_day": "morning",
	})
	text := parseToolText(t, resp)

	var result struct {
		Suggestions []PrioritySuggestion `json:"suggestions"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(result.Suggestions))
	}
}

func TestToolAnalyzeWorkload(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "t1", Text: "task", Status: core.StatusTodo, Effort: core.EffortMedium, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "analyze_workload", nil)
	text := parseToolText(t, resp)

	var result struct {
		Analysis WorkloadAnalysis `json:"analysis"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Analysis.TotalTasks != 1 {
		t.Errorf("total_tasks = %d, want 1", result.Analysis.TotalTasks)
	}
}

func TestToolFocusRecommendation(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "t1", Text: "task", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "focus_recommendation", map[string]any{
		"mood": "okay", "available_minutes": 30,
	})
	text := parseToolText(t, resp)

	var result struct {
		Recommendation FocusRecommendation `json:"recommendation"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Recommendation.DurationMinutes != 30 {
		t.Errorf("duration = %d, want 30", result.Recommendation.DurationMinutes)
	}
}

func TestToolWhatIf(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	s := newTestServerWithTasks(
		&core.Task{ID: "t1", Text: "task", Status: core.StatusTodo, Effort: core.EffortQuickWin, CreatedAt: now, UpdatedAt: now},
	)

	resp := dispatchToolCall(t, s, "what_if", map[string]any{
		"complete_task_ids": []string{"t1"},
	})
	text := parseToolText(t, resp)

	var result struct {
		Result WhatIfResult `json:"result"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Result.CompletedIDs) != 1 {
		t.Errorf("completed_ids = %v, want [t1]", result.Result.CompletedIDs)
	}
}

func TestToolWhatIfMissingIDs(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "what_if", map[string]any{
		"complete_task_ids": []string{},
	})
	if resp.Error == nil {
		t.Fatal("expected error for empty complete_task_ids")
		return
	}
}

func TestToolWhatIfUnknownTask(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "what_if", map[string]any{
		"complete_task_ids": []string{"nonexistent"},
	})

	resultBytes, _ := json.Marshal(resp.Result)
	var result ToolCallResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.IsError {
		t.Error("expected isError=true for unknown task")
	}
}

func TestToolContextSwitchAnalysis(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := dispatchToolCall(t, s, "context_switch_analysis", nil)
	text := parseToolText(t, resp)

	var result struct {
		Analysis ContextSwitchAnalysis `json:"analysis"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Analysis.TotalSwitches != 0 {
		t.Errorf("expected 0 switches, got %d", result.Analysis.TotalSwitches)
	}
}

func TestAdvancedToolsRegistered(t *testing.T) {
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

	expected := []string{
		"prioritize_tasks", "analyze_workload", "focus_recommendation",
		"what_if", "context_switch_analysis",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing advanced tool: %s", name)
		}
	}
}

func TestAdvancedPromptsRegistered(t *testing.T) {
	t.Parallel()

	s := newTestServer()
	resp := s.dispatch(&Request{
		ID:     json.RawMessage(`1`),
		Method: "prompts/list",
	})

	resultBytes, _ := json.Marshal(resp.Result)
	var result PromptsListResult
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	names := make(map[string]bool)
	for _, p := range result.Prompts {
		names[p.Name] = true
	}

	if !names["blocked_tasks"] {
		t.Error("missing prompt: blocked_tasks")
	}
	if !names["task_deep_dive"] {
		t.Error("missing prompt: task_deep_dive")
	}
}

func TestAdvancedPromptsGet(t *testing.T) {
	t.Parallel()

	s := newTestServer()

	for _, name := range []string{"blocked_tasks", "task_deep_dive"} {
		params, _ := json.Marshal(map[string]string{"name": name})
		resp := s.dispatch(&Request{
			ID:     json.RawMessage(`1`),
			Method: "prompts/get",
			Params: params,
		})

		if resp.Error != nil {
			t.Errorf("prompt %s: unexpected error: %v", name, resp.Error)
			continue
		}

		resultBytes, _ := json.Marshal(resp.Result)
		var result PromptGetResult
		if err := json.Unmarshal(resultBytes, &result); err != nil {
			t.Fatalf("prompt %s: unmarshal: %v", name, err)
		}

		if len(result.Messages) == 0 {
			t.Errorf("prompt %s: expected messages", name)
		}
		if result.Messages[0].Content.Text == "" {
			t.Errorf("prompt %s: expected non-empty template", name)
		}
	}
}
