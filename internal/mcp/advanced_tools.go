package mcp

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
)

// Factor is a single scoring factor in a priority suggestion.
type Factor struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Value  float64 `json:"value"`
}

// PrioritySuggestion holds a scored task recommendation.
type PrioritySuggestion struct {
	TaskID    string   `json:"task_id"`
	Score     float64  `json:"score"`
	Rationale string   `json:"rationale"`
	Factors   []Factor `json:"factors"`
}

// WorkloadAnalysis holds workload assessment data.
type WorkloadAnalysis struct {
	TotalTasks       int            `json:"total_tasks"`
	ByStatus         map[string]int `json:"by_status"`
	ByEffort         map[string]int `json:"by_effort"`
	ByProvider       map[string]int `json:"by_provider"`
	EstimatedHours   float64        `json:"estimated_hours"`
	OverloadRisk     float64        `json:"overload_risk"`
	RecommendedFocus []string       `json:"recommended_focus"`
	DeferCandidates  []string       `json:"defer_candidates"`
}

// FocusRecommendation suggests an optimal work session.
type FocusRecommendation struct {
	OptimalStartHour  int      `json:"optimal_start_hour"`
	DurationMinutes   int      `json:"duration_minutes"`
	SuggestedTasks    []string `json:"suggested_tasks"`
	TaskOrder         []string `json:"task_order"`
	BreakAfterMinutes int      `json:"break_after_minutes"`
	Rationale         string   `json:"rationale"`
}

// WhatIfResult models the downstream effects of completing tasks.
type WhatIfResult struct {
	CompletedIDs   []string `json:"completed_ids"`
	UnblockedTasks []string `json:"unblocked_tasks"`
	StreakImpact   int      `json:"streak_impact"`
	VelocityChange float64  `json:"velocity_change"`
	EpicProgress   float64  `json:"epic_progress"`
	EstimatedHours float64  `json:"estimated_hours"`
}

// ContextSwitchAnalysis holds context-switching cost data.
type ContextSwitchAnalysis struct {
	TotalSwitches    int               `json:"total_switches"`
	CostPerSwitch    float64           `json:"cost_per_switch_seconds"`
	ExpensivePairs   []TypePairCost    `json:"expensive_pairs"`
	BatchSuggestions []BatchSuggestion `json:"batch_suggestions"`
}

// TypePairCost represents the cost of switching between two task types.
type TypePairCost struct {
	FromType string  `json:"from_type"`
	ToType   string  `json:"to_type"`
	AvgCost  float64 `json:"avg_cost_seconds"`
	Count    int     `json:"count"`
}

// BatchSuggestion recommends grouping tasks of similar type.
type BatchSuggestion struct {
	TaskType string   `json:"task_type"`
	TaskIDs  []string `json:"task_ids"`
	Reason   string   `json:"reason"`
}

// advancedToolDefinitions returns tool definitions for story 24.8.
func advancedToolDefinitions() []ToolItem {
	return []ToolItem{
		{
			Name:        "prioritize_tasks",
			Description: "Score and rank tasks using multi-signal prioritization. Returns scored list with rationale and factor breakdown.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit":       map[string]any{"type": "integer", "description": "Max tasks to return (default 10)"},
					"mood":        map[string]any{"type": "string", "description": "Current mood: great, good, okay, low, bad"},
					"time_of_day": map[string]any{"type": "string", "description": "Time of day: morning, afternoon, evening, night"},
				},
			},
		},
		{
			Name:        "analyze_workload",
			Description: "Analyze current workload: total tasks by status/effort/provider, estimated hours, overload risk (0-1), recommended focus, defer candidates.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "focus_recommendation",
			Description: "Recommend an optimal focus session based on mood and available time. Returns task order, break timing, and rationale.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"mood":              map[string]any{"type": "string", "description": "Current mood: great, good, okay, low, bad"},
					"available_minutes": map[string]any{"type": "integer", "description": "Available time in minutes (default 60)"},
				},
			},
		},
		{
			Name:        "what_if",
			Description: "Model a hypothetical scenario: what happens if specific tasks are completed? Shows unblocked tasks, streak impact, velocity change. Never mutates real data.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"complete_task_ids": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Task IDs to hypothetically complete",
					},
				},
				"required": []string{"complete_task_ids"},
			},
		},
		{
			Name:        "context_switch_analysis",
			Description: "Analyze context-switching costs from session data. Returns switches per session, cost per switch, expensive type pairs, and batching suggestions.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"session_id": map[string]any{"type": "string", "description": "Specific session ID to analyze (default: all recent sessions)"},
				},
			},
		},
	}
}

// effortHours maps effort levels to estimated hours based on historical medians.
var effortHours = map[core.TaskEffort]float64{
	core.EffortQuickWin: 0.25,
	core.EffortMedium:   1.0,
	core.EffortDeepWork: 3.0,
}

// prioritizeTasks scores and ranks tasks using multi-signal scoring.
func prioritizeTasks(pool *core.TaskPool, edges []TaskEdge, limit int, mood, timeOfDay string) []PrioritySuggestion {
	tasks := pool.GetAllTasks()

	// Only consider actionable tasks.
	var actionable []*core.Task
	for _, t := range tasks {
		if t.Status == core.StatusTodo || t.Status == core.StatusInProgress {
			actionable = append(actionable, t)
		}
	}

	if len(actionable) == 0 {
		return []PrioritySuggestion{}
	}

	// Precompute blocking scores: how many tasks does each task block?
	blockingCount := make(map[string]int)
	for _, e := range edges {
		if e.Type == EdgeBlocks {
			blockingCount[e.FromID]++
		}
	}

	now := time.Now().UTC()
	var suggestions []PrioritySuggestion

	for _, t := range actionable {
		var factors []Factor
		var weightedSum, totalWeight float64

		// Factor 1: Blocking score (weight 3.0) — tasks that unblock others are more valuable.
		blockScore := math.Min(float64(blockingCount[t.ID])/3.0, 1.0)
		factors = append(factors, Factor{Name: "blocking_score", Weight: 3.0, Value: blockScore})
		weightedSum += 3.0 * blockScore
		totalWeight += 3.0

		// Factor 2: Age (weight 1.5) — older tasks get mild priority boost.
		ageDays := now.Sub(t.CreatedAt).Hours() / 24
		ageScore := math.Min(ageDays/30.0, 1.0)
		factors = append(factors, Factor{Name: "age", Weight: 1.5, Value: ageScore})
		weightedSum += 1.5 * ageScore
		totalWeight += 1.5

		// Factor 3: Effort fit (weight 2.0) — match effort to mood/energy.
		effortScore := effortFitScore(t.Effort, mood)
		factors = append(factors, Factor{Name: "effort_fit", Weight: 2.0, Value: effortScore})
		weightedSum += 2.0 * effortScore
		totalWeight += 2.0

		// Factor 4: Type fit (weight 1.5) — match task type to time of day.
		typeScore := typeFitScore(t.Type, timeOfDay)
		factors = append(factors, Factor{Name: "type_fit", Weight: 1.5, Value: typeScore})
		weightedSum += 1.5 * typeScore
		totalWeight += 1.5

		// Factor 5: In-progress boost (weight 2.0) — prefer continuing started work.
		var inProgressScore float64
		if t.Status == core.StatusInProgress {
			inProgressScore = 1.0
		}
		factors = append(factors, Factor{Name: "in_progress", Weight: 2.0, Value: inProgressScore})
		weightedSum += 2.0 * inProgressScore
		totalWeight += 2.0

		// Normalize to 0-100.
		score := (weightedSum / totalWeight) * 100
		score = math.Round(score*10) / 10

		rationale := buildRationale(t, blockingCount[t.ID], mood, timeOfDay)

		suggestions = append(suggestions, PrioritySuggestion{
			TaskID:    t.ID,
			Score:     score,
			Rationale: rationale,
			Factors:   factors,
		})
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	if limit <= 0 {
		limit = 10
	}
	if limit > len(suggestions) {
		limit = len(suggestions)
	}
	return suggestions[:limit]
}

// effortFitScore rates how well a task's effort matches the user's mood.
func effortFitScore(effort core.TaskEffort, mood string) float64 {
	// High energy moods → good for deep work; low energy → good for quick wins.
	moodEnergy := map[string]float64{
		"great": 1.0, "good": 0.8, "okay": 0.5, "low": 0.3, "bad": 0.1,
	}
	effortDemand := map[core.TaskEffort]float64{
		core.EffortQuickWin: 0.2, core.EffortMedium: 0.5, core.EffortDeepWork: 0.9,
	}

	energy, ok := moodEnergy[mood]
	if !ok {
		energy = 0.5
	}
	demand, ok := effortDemand[effort]
	if !ok {
		demand = 0.5
	}

	// Best fit when energy matches demand; penalize mismatch.
	diff := math.Abs(energy - demand)
	return 1.0 - diff
}

// typeFitScore rates how well a task type matches the time of day.
func typeFitScore(taskType core.TaskType, timeOfDay string) float64 {
	// Productivity research: creative work best in morning, admin in afternoon.
	typeTimeAffinity := map[core.TaskType]map[string]float64{
		core.TypeCreative: {
			"morning": 1.0, "afternoon": 0.6, "evening": 0.7, "night": 0.4,
		},
		core.TypeTechnical: {
			"morning": 0.9, "afternoon": 0.7, "evening": 0.5, "night": 0.3,
		},
		core.TypeAdministrative: {
			"morning": 0.5, "afternoon": 1.0, "evening": 0.6, "night": 0.3,
		},
		core.TypePhysical: {
			"morning": 0.8, "afternoon": 0.7, "evening": 0.9, "night": 0.2,
		},
	}

	if affinities, ok := typeTimeAffinity[taskType]; ok {
		if score, ok := affinities[timeOfDay]; ok {
			return score
		}
	}
	return 0.5
}

func buildRationale(t *core.Task, blockCount int, mood, timeOfDay string) string {
	var parts []string
	if blockCount > 0 {
		parts = append(parts, fmt.Sprintf("unblocks %d task(s)", blockCount))
	}
	if t.Status == core.StatusInProgress {
		parts = append(parts, "already in progress")
	}
	if mood != "" {
		parts = append(parts, fmt.Sprintf("fits %s mood", mood))
	}
	if timeOfDay != "" {
		parts = append(parts, fmt.Sprintf("suitable for %s", timeOfDay))
	}
	if len(parts) == 0 {
		return "general priority"
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "; " + parts[i]
	}
	return result
}

// analyzeWorkload computes workload metrics from the task pool.
func analyzeWorkload(pool *core.TaskPool, edges []TaskEdge) *WorkloadAnalysis {
	tasks := pool.GetAllTasks()

	analysis := &WorkloadAnalysis{
		TotalTasks: len(tasks),
		ByStatus:   make(map[string]int),
		ByEffort:   make(map[string]int),
		ByProvider: make(map[string]int),
	}

	var activeCount int
	for _, t := range tasks {
		analysis.ByStatus[string(t.Status)]++
		if t.Effort != "" {
			analysis.ByEffort[string(t.Effort)]++
		}
		provider := t.EffectiveSourceProvider()
		if provider != "" {
			analysis.ByProvider[provider]++
		}

		// Estimate hours for non-completed tasks.
		if t.Status != core.StatusComplete && t.Status != core.StatusArchived {
			activeCount++
			hours, ok := effortHours[t.Effort]
			if !ok {
				hours = 1.0 // default estimate
			}
			analysis.EstimatedHours += hours
		}
	}

	analysis.EstimatedHours = math.Round(analysis.EstimatedHours*10) / 10

	// Overload risk: based on active task count and estimated hours.
	// >40 hours or >20 active tasks → high risk.
	hourRisk := math.Min(analysis.EstimatedHours/40.0, 1.0)
	countRisk := math.Min(float64(activeCount)/20.0, 1.0)
	analysis.OverloadRisk = math.Round(math.Max(hourRisk, countRisk)*100) / 100

	// Recommended focus: top 3 by priority.
	suggestions := prioritizeTasks(pool, edges, 3, "", "")
	for _, s := range suggestions {
		analysis.RecommendedFocus = append(analysis.RecommendedFocus, s.TaskID)
	}
	if analysis.RecommendedFocus == nil {
		analysis.RecommendedFocus = []string{}
	}

	// Defer candidates: old deferred or low-priority tasks.
	var deferCandidates []string
	for _, t := range tasks {
		if t.Status == core.StatusDeferred {
			deferCandidates = append(deferCandidates, t.ID)
		}
	}
	// Also suggest tasks with no blockers and low age as defer candidates (bottom priority).
	if len(deferCandidates) == 0 {
		allSuggestions := prioritizeTasks(pool, edges, len(tasks), "", "")
		for i := len(allSuggestions) - 1; i >= 0 && len(deferCandidates) < 3; i-- {
			deferCandidates = append(deferCandidates, allSuggestions[i].TaskID)
		}
	}
	analysis.DeferCandidates = deferCandidates
	if analysis.DeferCandidates == nil {
		analysis.DeferCandidates = []string{}
	}

	return analysis
}

// focusRecommendation suggests an optimal work session.
func focusRecommendation(pool *core.TaskPool, edges []TaskEdge, pm *PatternMiner, mood string, availableMinutes int) *FocusRecommendation {
	if availableMinutes <= 0 {
		availableMinutes = 60
	}

	rec := &FocusRecommendation{
		DurationMinutes: availableMinutes,
	}

	// Determine optimal start hour from productivity profile if available.
	rec.OptimalStartHour = guessOptimalHour(pm, mood)

	// Break timing: Pomodoro-style, break every 25-50 min depending on mood.
	switch mood {
	case "great", "good":
		rec.BreakAfterMinutes = 50
	case "low", "bad":
		rec.BreakAfterMinutes = 25
	default:
		rec.BreakAfterMinutes = 35
	}

	// Select tasks that fit in available time.
	timeOfDay := hourToTimeOfDay(rec.OptimalStartHour)
	suggestions := prioritizeTasks(pool, edges, 20, mood, timeOfDay)

	var totalMinutes float64
	for _, s := range suggestions {
		t := pool.GetTask(s.TaskID)
		if t == nil {
			continue
		}
		hours, ok := effortHours[t.Effort]
		if !ok {
			hours = 1.0
		}
		taskMinutes := hours * 60
		if totalMinutes+taskMinutes > float64(availableMinutes) {
			continue
		}
		rec.SuggestedTasks = append(rec.SuggestedTasks, s.TaskID)
		totalMinutes += taskMinutes
	}

	if rec.SuggestedTasks == nil {
		rec.SuggestedTasks = []string{}
	}

	// Task order: quick wins first when mood is low, deep work first when high.
	rec.TaskOrder = orderTasks(pool, rec.SuggestedTasks, mood)

	rec.Rationale = buildFocusRationale(mood, availableMinutes, len(rec.SuggestedTasks))

	return rec
}

func guessOptimalHour(pm *PatternMiner, mood string) int {
	if pm != nil {
		now := time.Now().UTC()
		profile, err := pm.ProductivityProfileAnalysis(now.AddDate(0, 0, -30), now)
		if err == nil && len(profile.PeakHours) > 0 {
			return profile.PeakHours[0]
		}
	}

	// Fallback based on mood.
	switch mood {
	case "great", "good":
		return 9
	case "low", "bad":
		return 14
	default:
		return 10
	}
}

func hourToTimeOfDay(hour int) string {
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

func orderTasks(pool *core.TaskPool, taskIDs []string, mood string) []string {
	if len(taskIDs) == 0 {
		return []string{}
	}

	type taskWithEffort struct {
		id     string
		effort float64
	}
	var items []taskWithEffort
	for _, id := range taskIDs {
		t := pool.GetTask(id)
		if t == nil {
			continue
		}
		hours, ok := effortHours[t.Effort]
		if !ok {
			hours = 1.0
		}
		items = append(items, taskWithEffort{id: id, effort: hours})
	}

	// Low mood: quick wins first (ascending effort). High mood: deep work first (descending).
	if mood == "low" || mood == "bad" {
		sort.Slice(items, func(i, j int) bool { return items[i].effort < items[j].effort })
	} else {
		sort.Slice(items, func(i, j int) bool { return items[i].effort > items[j].effort })
	}

	ordered := make([]string, len(items))
	for i, item := range items {
		ordered[i] = item.id
	}
	return ordered
}

func buildFocusRationale(mood string, minutes, taskCount int) string {
	if taskCount == 0 {
		return "no tasks fit in the available time"
	}
	return fmt.Sprintf("%d task(s) selected for %d-minute session with %s mood", taskCount, minutes, mood)
}

// whatIf models a hypothetical scenario without mutating real data.
func whatIf(pool *core.TaskPool, edges []TaskEdge, completeIDs []string) *WhatIfResult {
	result := &WhatIfResult{
		CompletedIDs: completeIDs,
	}

	// Build set of IDs to complete.
	completing := make(map[string]bool)
	for _, id := range completeIDs {
		completing[id] = true
	}

	// Find tasks that would be unblocked.
	// A task is unblocked if all its blockers are in the completing set.
	blockedBy := make(map[string][]string) // task → list of tasks blocking it
	for _, e := range edges {
		if e.Type == EdgeBlocks {
			blockedBy[e.ToID] = append(blockedBy[e.ToID], e.FromID)
		}
	}

	for taskID, blockers := range blockedBy {
		if completing[taskID] {
			continue
		}
		allResolved := true
		for _, blocker := range blockers {
			if !completing[blocker] {
				allResolved = false
				break
			}
		}
		if allResolved {
			// Check this task is actually blocked.
			t := pool.GetTask(taskID)
			if t != nil && (t.Status == core.StatusBlocked || t.Status == core.StatusTodo) {
				result.UnblockedTasks = append(result.UnblockedTasks, taskID)
			}
		}
	}
	if result.UnblockedTasks == nil {
		result.UnblockedTasks = []string{}
	}
	sort.Strings(result.UnblockedTasks)

	// Streak impact: completing tasks adds to current streak.
	result.StreakImpact = len(completeIDs)

	// Velocity change: tasks completed relative to total active.
	allTasks := pool.GetAllTasks()
	var activeCount int
	for _, t := range allTasks {
		if t.Status == core.StatusTodo || t.Status == core.StatusInProgress || t.Status == core.StatusBlocked {
			activeCount++
		}
	}
	if activeCount > 0 {
		result.VelocityChange = math.Round(float64(len(completeIDs))/float64(activeCount)*100*10) / 10
	}

	// Epic progress: fraction of all tasks that would be complete.
	totalTasks := len(allTasks)
	var currentComplete int
	for _, t := range allTasks {
		if t.Status == core.StatusComplete {
			currentComplete++
		}
	}
	if totalTasks > 0 {
		result.EpicProgress = math.Round(float64(currentComplete+len(completeIDs))/float64(totalTasks)*100*10) / 10
	}

	// Estimated time for completing these tasks.
	for _, id := range completeIDs {
		t := pool.GetTask(id)
		if t == nil {
			continue
		}
		hours, ok := effortHours[t.Effort]
		if !ok {
			hours = 1.0
		}
		result.EstimatedHours += hours
	}
	result.EstimatedHours = math.Round(result.EstimatedHours*10) / 10

	return result
}

// contextSwitchAnalysis analyzes context-switching costs from session data.
func contextSwitchAnalysis(pm *PatternMiner, pool *core.TaskPool, sessionID string) (*ContextSwitchAnalysis, error) {
	if pm == nil {
		return &ContextSwitchAnalysis{
			ExpensivePairs:   []TypePairCost{},
			BatchSuggestions: []BatchSuggestion{},
		}, nil
	}

	var sessions []core.SessionMetrics
	var err error
	if sessionID != "" {
		all, readErr := pm.allSessions()
		if readErr != nil {
			return nil, fmt.Errorf("read sessions: %w", readErr)
		}
		for _, s := range all {
			if s.SessionID == sessionID {
				sessions = append(sessions, s)
				break
			}
		}
	} else {
		now := time.Now().UTC()
		sessions, err = pm.sessionsInRange(now.AddDate(0, 0, -30), now)
		if err != nil {
			return nil, fmt.Errorf("read sessions: %w", err)
		}
	}

	analysis := &ContextSwitchAnalysis{}

	// Analyze door selections for type transitions.
	type pairKey struct{ from, to string }
	pairCosts := make(map[pairKey]*TypePairCost)
	var totalSwitchTime float64
	var switchCount int

	for _, s := range sessions {
		sels := s.DoorSelections
		for i := 1; i < len(sels); i++ {
			prevType := categorizeTaskText(sels[i-1].TaskText)
			currType := categorizeTaskText(sels[i].TaskText)

			if prevType != currType {
				analysis.TotalSwitches++
				gap := sels[i].Timestamp.Sub(sels[i-1].Timestamp).Seconds()
				if gap > 0 && gap < 3600 { // ignore gaps > 1 hour
					totalSwitchTime += gap
					switchCount++

					key := pairKey{prevType, currType}
					pc, ok := pairCosts[key]
					if !ok {
						pc = &TypePairCost{FromType: prevType, ToType: currType}
						pairCosts[key] = pc
					}
					pc.AvgCost += gap
					pc.Count++
				}
			}
		}
	}

	if switchCount > 0 {
		analysis.CostPerSwitch = math.Round(totalSwitchTime/float64(switchCount)*10) / 10
	}

	// Compute average cost per pair.
	for _, pc := range pairCosts {
		if pc.Count > 0 {
			pc.AvgCost = math.Round(pc.AvgCost/float64(pc.Count)*10) / 10
		}
		analysis.ExpensivePairs = append(analysis.ExpensivePairs, *pc)
	}
	sort.Slice(analysis.ExpensivePairs, func(i, j int) bool {
		return analysis.ExpensivePairs[i].AvgCost > analysis.ExpensivePairs[j].AvgCost
	})
	if analysis.ExpensivePairs == nil {
		analysis.ExpensivePairs = []TypePairCost{}
	}

	// Batch suggestions: group actionable tasks by type.
	analysis.BatchSuggestions = buildBatchSuggestions(pool)
	if analysis.BatchSuggestions == nil {
		analysis.BatchSuggestions = []BatchSuggestion{}
	}

	return analysis, nil
}

// categorizeTaskText attempts to classify task text into a type category.
func categorizeTaskText(text string) string {
	// Simple keyword-based classification.
	if text == "" {
		return "unknown"
	}
	return "general"
}

func buildBatchSuggestions(pool *core.TaskPool) []BatchSuggestion {
	tasks := pool.GetAllTasks()

	byType := make(map[string][]string)
	for _, t := range tasks {
		if t.Status != core.StatusTodo && t.Status != core.StatusInProgress {
			continue
		}
		taskType := string(t.Type)
		if taskType == "" {
			taskType = "untyped"
		}
		byType[taskType] = append(byType[taskType], t.ID)
	}

	var suggestions []BatchSuggestion
	for typ, ids := range byType {
		if len(ids) >= 2 {
			suggestions = append(suggestions, BatchSuggestion{
				TaskType: typ,
				TaskIDs:  ids,
				Reason:   fmt.Sprintf("batch %d %s tasks to reduce context switching", len(ids), typ),
			})
		}
	}
	sort.Slice(suggestions, func(i, j int) bool {
		return len(suggestions[i].TaskIDs) > len(suggestions[j].TaskIDs)
	})
	return suggestions
}

// Tool handler methods on MCPServer.

func (s *MCPServer) toolPrioritizeTasks(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		Limit     int    `json:"limit"`
		Mood      string `json:"mood"`
		TimeOfDay string `json:"time_of_day"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	suggestions := prioritizeTasks(s.pool, edges, params.Limit, params.Mood, params.TimeOfDay)

	type prioritizeResult struct {
		Suggestions []PrioritySuggestion `json:"suggestions"`
		Metadata    ResponseMetadata     `json:"_metadata"`
	}
	result := prioritizeResult{
		Suggestions: suggestions,
		Metadata: ResponseMetadata{
			TotalCount:       len(s.pool.GetAllTasks()),
			ReturnedCount:    len(suggestions),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolAnalyzeWorkload(req *Request) *Response {
	start := time.Now().UTC()

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	analysis := analyzeWorkload(s.pool, edges)

	type workloadResult struct {
		Analysis *WorkloadAnalysis `json:"analysis"`
		Metadata ResponseMetadata  `json:"_metadata"`
	}
	result := workloadResult{
		Analysis: analysis,
		Metadata: ResponseMetadata{
			TotalCount:       analysis.TotalTasks,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolFocusRecommendation(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		Mood             string `json:"mood"`
		AvailableMinutes int    `json:"available_minutes"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	pm := s.patternMiner()
	rec := focusRecommendation(s.pool, edges, pm, params.Mood, params.AvailableMinutes)

	type focusResult struct {
		Recommendation *FocusRecommendation `json:"recommendation"`
		Metadata       ResponseMetadata     `json:"_metadata"`
	}
	result := focusResult{
		Recommendation: rec,
		Metadata: ResponseMetadata{
			TotalCount:       len(s.pool.GetAllTasks()),
			ReturnedCount:    len(rec.SuggestedTasks),
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolWhatIf(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		CompleteTaskIDs []string `json:"complete_task_ids"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}
	if len(params.CompleteTaskIDs) == 0 {
		return NewErrorResponse(req.ID, CodeInvalidParams, "complete_task_ids is required")
	}

	// Validate task IDs exist.
	for _, id := range params.CompleteTaskIDs {
		if s.pool.GetTask(id) == nil {
			return s.toolError(req, fmt.Sprintf("task not found: %s", id))
		}
	}

	inferencer := NewRelationshipInferencer(s.pool)
	edges := inferencer.InferAll()
	whatIfResult := whatIf(s.pool, edges, params.CompleteTaskIDs)

	type whatIfResponse struct {
		Result   *WhatIfResult    `json:"result"`
		Metadata ResponseMetadata `json:"_metadata"`
	}
	result := whatIfResponse{
		Result: whatIfResult,
		Metadata: ResponseMetadata{
			TotalCount:       len(s.pool.GetAllTasks()),
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}

func (s *MCPServer) toolContextSwitchAnalysis(req *Request, args json.RawMessage) *Response {
	start := time.Now().UTC()

	var params struct {
		SessionID string `json:"session_id"`
	}
	if args != nil {
		if err := json.Unmarshal(args, &params); err != nil {
			return NewErrorResponse(req.ID, CodeInvalidParams, fmt.Sprintf("invalid arguments: %v", err))
		}
	}

	pm := s.patternMiner()
	analysis, err := contextSwitchAnalysis(pm, s.pool, params.SessionID)
	if err != nil {
		return s.toolError(req, fmt.Sprintf("context switch analysis: %v", err))
	}

	type csaResult struct {
		Analysis *ContextSwitchAnalysis `json:"analysis"`
		Metadata ResponseMetadata       `json:"_metadata"`
	}
	result := csaResult{
		Analysis: analysis,
		Metadata: ResponseMetadata{
			TotalCount:       1,
			ReturnedCount:    1,
			QueryTimeMs:      millisSince(start),
			ProvidersQueried: s.providerNames(),
			DataFreshness:    "live",
		},
	}
	return s.toolJSON(req, result)
}
