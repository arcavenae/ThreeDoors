package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// PatternReport contains the results of session pattern analysis.
type PatternReport struct {
	GeneratedAt       time.Time                    `json:"generated_at"`
	SessionCount      int                          `json:"session_count"`
	DoorPositionBias  DoorPositionStats            `json:"door_position_bias"`
	TaskTypeStats     map[string]TypeSelectionRate `json:"task_type_stats"`
	TimeOfDayPatterns []TimeOfDayPattern           `json:"time_of_day_patterns"`
	MoodCorrelations  []MoodCorrelation            `json:"mood_correlations"`
	AvoidanceList     []AvoidanceEntry             `json:"avoidance_list"`
}

// DoorPositionStats tracks door position selection frequency.
type DoorPositionStats struct {
	LeftCount         int    `json:"left_count"`
	CenterCount       int    `json:"center_count"`
	RightCount        int    `json:"right_count"`
	TotalSelections   int    `json:"total_selections"`
	PreferredPosition string `json:"preferred_position"`
}

// TypeSelectionRate tracks how often a task type is selected vs bypassed.
type TypeSelectionRate struct {
	TimesShown    int     `json:"times_shown"`
	TimesSelected int     `json:"times_selected"`
	TimesBypassed int     `json:"times_bypassed"`
	SelectionRate float64 `json:"selection_rate"`
}

// TimeOfDayPattern tracks session patterns by time of day.
type TimeOfDayPattern struct {
	Period            string  `json:"period"`
	SessionCount      int     `json:"session_count"`
	AvgTasksCompleted float64 `json:"avg_tasks_completed"`
	AvgDuration       float64 `json:"avg_duration_minutes"`
}

// MoodCorrelation tracks correlation between mood and task selection.
type MoodCorrelation struct {
	Mood              string  `json:"mood"`
	SessionCount      int     `json:"session_count"`
	PreferredType     string  `json:"preferred_type"`
	PreferredEffort   string  `json:"preferred_effort"`
	AvoidedType       string  `json:"avoided_type,omitempty"`
	AvgTasksCompleted float64 `json:"avg_tasks_completed"`
}

// AvoidanceEntry tracks tasks that are repeatedly bypassed.
type AvoidanceEntry struct {
	TaskText      string `json:"task_text"`
	TimesBypassed int    `json:"times_bypassed"`
	TimesShown    int    `json:"times_shown"`
	NeverSelected bool   `json:"never_selected"`
}

// TaskCategoryInfo holds categorization for a task, keyed by task text.
type TaskCategoryInfo struct {
	Type   TaskType
	Effort TaskEffort
}

// DoorPreferences holds door position selection percentages and bias detection.
type DoorPreferences struct {
	LeftPercent     float64
	CenterPercent   float64
	RightPercent    float64
	TotalSelections int
	HasBias         bool
	BiasPosition    string // "left", "center", or "right" (lowercase)
}

// WeekComparison holds week-over-week completion comparison data.
type WeekComparison struct {
	ThisWeekTotal int
	LastWeekTotal int
	PercentChange float64
	Direction     string // "up", "down", "same"
}

// PatternAnalyzer analyzes session metrics for user behavior patterns.
type PatternAnalyzer struct {
	taskCategories map[string]TaskCategoryInfo
	sessions       []SessionMetrics
	nowFunc        func() time.Time
}

// NewPatternAnalyzer creates a new PatternAnalyzer.
func NewPatternAnalyzer() *PatternAnalyzer {
	return &PatternAnalyzer{
		nowFunc: time.Now,
	}
}

// NewPatternAnalyzerWithNow creates a new PatternAnalyzer with a custom time function for testing.
func NewPatternAnalyzerWithNow(nowFunc func() time.Time) *PatternAnalyzer {
	return &PatternAnalyzer{
		nowFunc: nowFunc,
	}
}

// SetTaskCategories sets the task categorization lookup table.
// This allows the analyzer to populate PreferredType and PreferredEffort
// in MoodCorrelation results by mapping task text to categories.
func (pa *PatternAnalyzer) SetTaskCategories(categories map[string]TaskCategoryInfo) {
	pa.taskCategories = categories
}

// BuildTaskCategoryMap creates a lookup table from a TaskPool.
// Includes ALL tasks, even those with zero-value Type/Effort.
// Returns empty map (not nil) for nil pool or pool with no tasks.
func BuildTaskCategoryMap(pool *TaskPool) map[string]TaskCategoryInfo {
	m := make(map[string]TaskCategoryInfo)
	if pool == nil {
		return m
	}
	for _, t := range pool.GetAllTasks() {
		m[t.Text] = TaskCategoryInfo{
			Type:   t.Type,
			Effort: t.Effort,
		}
	}
	return m
}

// ReadSessions reads session metrics from a JSON Lines file.
// Returns empty slice and nil error for missing or empty files.
func (pa *PatternAnalyzer) ReadSessions(path string) ([]SessionMetrics, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("opening sessions file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var sessions []SessionMetrics
	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var sm SessionMetrics
		if err := json.Unmarshal([]byte(line), &sm); err != nil {
			// Skip malformed lines
			continue
		}
		sessions = append(sessions, sm)
	}
	if err := scanner.Err(); err != nil {
		return sessions, fmt.Errorf("reading sessions file: %w", err)
	}
	return sessions, nil
}

// LoadSessions reads and parses sessions.jsonl into internal storage.
// Returns nil for non-existent files. Malformed lines are skipped.
// Returns error for permission issues.
func (pa *PatternAnalyzer) LoadSessions(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()

	pa.sessions = nil
	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var sm SessionMetrics
		if err := json.Unmarshal(line, &sm); err != nil {
			continue // skip malformed lines
		}
		pa.sessions = append(pa.sessions, sm)
	}
	return scanner.Err()
}

// HasSufficientData returns true if at least 3 sessions exist.
func (pa *PatternAnalyzer) HasSufficientData() bool {
	return len(pa.sessions) >= 3
}

// GetSessionCount returns the total number of sessions loaded.
func (pa *PatternAnalyzer) GetSessionCount() int {
	return len(pa.sessions)
}

// GetSessionsNeeded returns how many more sessions are needed for insights.
func (pa *PatternAnalyzer) GetSessionsNeeded() int {
	needed := 3 - len(pa.sessions)
	if needed < 0 {
		return 0
	}
	return needed
}

// Analyze processes session metrics and returns a PatternReport.
// Returns nil report (no error) if fewer than 5 sessions (cold start guard).
func (pa *PatternAnalyzer) Analyze(sessions []SessionMetrics) (*PatternReport, error) {
	if len(sessions) < 5 {
		return nil, nil
	}

	report := &PatternReport{
		GeneratedAt:  time.Now().UTC(),
		SessionCount: len(sessions),
	}

	report.DoorPositionBias = pa.analyzeDoorPositionBias(sessions)
	report.TaskTypeStats = pa.analyzeTaskTypeStats(sessions)
	report.TimeOfDayPatterns = pa.analyzeTimeOfDay(sessions)
	report.MoodCorrelations = pa.analyzeMoodCorrelations(sessions)
	report.AvoidanceList = pa.analyzeAvoidance(sessions)

	return report, nil
}

// analyzeDoorPositionBias calculates door position selection frequency.
func (pa *PatternAnalyzer) analyzeDoorPositionBias(sessions []SessionMetrics) DoorPositionStats {
	stats := DoorPositionStats{}
	for _, s := range sessions {
		for _, sel := range s.DoorSelections {
			stats.TotalSelections++
			switch sel.DoorPosition {
			case 0:
				stats.LeftCount++
			case 1:
				stats.CenterCount++
			case 2:
				stats.RightCount++
			}
		}
	}

	if stats.TotalSelections == 0 {
		stats.PreferredPosition = "none"
		return stats
	}

	// >40% threshold for bias
	threshold := float64(stats.TotalSelections) * 0.4
	if float64(stats.LeftCount) > threshold {
		stats.PreferredPosition = "left"
	} else if float64(stats.CenterCount) > threshold {
		stats.PreferredPosition = "center"
	} else if float64(stats.RightCount) > threshold {
		stats.PreferredPosition = "right"
	} else {
		stats.PreferredPosition = "none"
	}

	return stats
}

// analyzeTaskTypeStats tracks task type selection and bypass rates.
func (pa *PatternAnalyzer) analyzeTaskTypeStats(sessions []SessionMetrics) map[string]TypeSelectionRate {
	stats := map[string]TypeSelectionRate{}

	// Count selections by task text
	selectedTexts := map[string]int{}
	bypassedTexts := map[string]int{}

	for _, s := range sessions {
		for _, sel := range s.DoorSelections {
			selectedTexts[sel.TaskText]++
		}
		for _, bypass := range s.TaskBypasses {
			for _, text := range bypass {
				bypassedTexts[text]++
			}
		}
	}

	// Aggregate by text (we don't have type info in session data,
	// so we track by task text for now)
	allTexts := map[string]bool{}
	for text := range selectedTexts {
		allTexts[text] = true
	}
	for text := range bypassedTexts {
		allTexts[text] = true
	}

	for text := range allTexts {
		selected := selectedTexts[text]
		bypassed := bypassedTexts[text]
		shown := selected + bypassed
		rate := 0.0
		if shown > 0 {
			rate = float64(selected) / float64(shown)
		}
		stats[text] = TypeSelectionRate{
			TimesShown:    shown,
			TimesSelected: selected,
			TimesBypassed: bypassed,
			SelectionRate: math.Round(rate*100) / 100,
		}
	}

	return stats
}

// analyzeTimeOfDay groups sessions by time period.
func (pa *PatternAnalyzer) analyzeTimeOfDay(sessions []SessionMetrics) []TimeOfDayPattern {
	type periodAcc struct {
		count     int
		totalComp int
		totalDur  float64
	}
	periods := map[string]*periodAcc{
		"morning":   {},
		"afternoon": {},
		"evening":   {},
		"night":     {},
	}

	for _, s := range sessions {
		period := hourToPeriod(s.StartTime.Hour())
		acc := periods[period]
		acc.count++
		acc.totalComp += s.TasksCompleted
		acc.totalDur += s.DurationSeconds / 60.0
	}

	var patterns []TimeOfDayPattern
	for _, period := range []string{"morning", "afternoon", "evening", "night"} {
		acc := periods[period]
		if acc.count == 0 {
			continue
		}
		patterns = append(patterns, TimeOfDayPattern{
			Period:            period,
			SessionCount:      acc.count,
			AvgTasksCompleted: math.Round(float64(acc.totalComp)/float64(acc.count)*10) / 10,
			AvgDuration:       math.Round(acc.totalDur/float64(acc.count)*10) / 10,
		})
	}

	return patterns
}

func hourToPeriod(hour int) string {
	switch {
	case hour >= 5 && hour <= 11:
		return "morning"
	case hour >= 12 && hour <= 16:
		return "afternoon"
	case hour >= 17 && hour <= 20:
		return "evening"
	default:
		return "night"
	}
}

// analyzeMoodCorrelations correlates mood entries with task selections.
func (pa *PatternAnalyzer) analyzeMoodCorrelations(sessions []SessionMetrics) []MoodCorrelation {
	type moodAcc struct {
		count       int
		totalComp   int
		taskTexts   []string
		bypassTexts []string
	}
	moods := map[string]*moodAcc{}

	for _, s := range sessions {
		if len(s.MoodEntries) == 0 {
			continue
		}
		// Use first mood entry as session mood
		mood := strings.ToLower(strings.TrimSpace(s.MoodEntries[0].Mood))
		if mood == "" {
			continue
		}
		if moods[mood] == nil {
			moods[mood] = &moodAcc{}
		}
		acc := moods[mood]
		acc.count++
		acc.totalComp += s.TasksCompleted
		for _, sel := range s.DoorSelections {
			acc.taskTexts = append(acc.taskTexts, sel.TaskText)
		}
		for _, bypass := range s.TaskBypasses {
			acc.bypassTexts = append(acc.bypassTexts, bypass...)
		}
	}

	var correlations []MoodCorrelation
	for mood, acc := range moods {
		// Minimum 3 sessions for a correlation
		if acc.count < 3 {
			continue
		}

		// Determine preferred type and effort from category map
		preferredType := ""
		preferredEffort := ""
		if len(pa.taskCategories) > 0 {
			typeCounts := map[TaskType]int{}
			effortCounts := map[TaskEffort]int{}
			for _, text := range acc.taskTexts {
				info, ok := pa.taskCategories[text]
				if !ok {
					continue // Task not in category map (deleted/completed) — skip
				}
				if info.Type != "" {
					typeCounts[info.Type]++
				}
				if info.Effort != "" {
					effortCounts[info.Effort]++
				}
			}
			// Find most frequent type
			maxTypeCount := 0
			for t, c := range typeCounts {
				if c > maxTypeCount {
					maxTypeCount = c
					preferredType = string(t)
				}
			}
			// Find most frequent effort
			maxEffortCount := 0
			for e, c := range effortCounts {
				if c > maxEffortCount {
					maxEffortCount = c
					preferredEffort = string(e)
				}
			}
		}

		// Determine avoided type from bypassed tasks
		avoidedType := ""
		if len(pa.taskCategories) > 0 && len(acc.bypassTexts) > 0 {
			bypassTypeCounts := map[TaskType]int{}
			for _, text := range acc.bypassTexts {
				info, ok := pa.taskCategories[text]
				if !ok {
					continue
				}
				if info.Type != "" {
					bypassTypeCounts[info.Type]++
				}
			}
			maxBypassCount := 0
			for t, c := range bypassTypeCounts {
				if c > maxBypassCount {
					maxBypassCount = c
					avoidedType = string(t)
				}
			}
		}

		correlations = append(correlations, MoodCorrelation{
			Mood:              mood,
			SessionCount:      acc.count,
			PreferredType:     preferredType,
			PreferredEffort:   preferredEffort,
			AvoidedType:       avoidedType,
			AvgTasksCompleted: math.Round(float64(acc.totalComp)/float64(acc.count)*10) / 10,
		})
	}

	// Sort for deterministic output
	sort.Slice(correlations, func(i, j int) bool {
		return correlations[i].Mood < correlations[j].Mood
	})

	return correlations
}

// analyzeAvoidance identifies tasks that are repeatedly bypassed.
func (pa *PatternAnalyzer) analyzeAvoidance(sessions []SessionMetrics) []AvoidanceEntry {
	bypassCounts := map[string]int{}
	shownCounts := map[string]int{}
	selectedTexts := map[string]bool{}

	for _, s := range sessions {
		for _, sel := range s.DoorSelections {
			selectedTexts[sel.TaskText] = true
			shownCounts[sel.TaskText]++
		}
		for _, bypass := range s.TaskBypasses {
			for _, text := range bypass {
				bypassCounts[text]++
				shownCounts[text]++
			}
		}
	}

	var entries []AvoidanceEntry
	for text, count := range bypassCounts {
		if count < 3 {
			continue
		}
		entries = append(entries, AvoidanceEntry{
			TaskText:      text,
			TimesBypassed: count,
			TimesShown:    shownCounts[text],
			NeverSelected: !selectedTexts[text],
		})
	}

	// Sort by bypass count descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TimesBypassed > entries[j].TimesBypassed
	})

	return entries
}

// SavePatterns writes a PatternReport to a JSON file atomically.
func (pa *PatternAnalyzer) SavePatterns(report *PatternReport, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling patterns: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating patterns directory: %w", err)
	}
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("writing patterns temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("renaming patterns file: %w", err)
	}
	return nil
}

// LoadPatterns reads a PatternReport from a JSON file.
// Returns nil report and nil error for missing files.
func (pa *PatternAnalyzer) LoadPatterns(path string) (*PatternReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading patterns file: %w", err)
	}

	var report PatternReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing patterns file: %w", err)
	}
	return &report, nil
}

// NeedsReanalysis returns true if the cached report is stale.
func (pa *PatternAnalyzer) NeedsReanalysis(cached *PatternReport, sessions []SessionMetrics) bool {
	if cached == nil {
		return true
	}
	if len(sessions) > cached.SessionCount {
		return true
	}
	// Check if latest session is newer than cache
	if len(sessions) > 0 {
		latest := sessions[len(sessions)-1]
		if latest.EndTime.After(cached.GeneratedAt) {
			return true
		}
	}
	return false
}

// --- Insights Dashboard methods (Story 4.5) ---

// GetDailyCompletions returns date -> completion count for the last N days.
// Keys are formatted as "2006-01-02". Days with zero sessions are present with value 0.
func (pa *PatternAnalyzer) GetDailyCompletions(days int) map[string]int {
	now := pa.nowFunc().UTC()
	result := make(map[string]int, days)

	// Initialize all days in range with zero
	for i := 0; i < days; i++ {
		d := now.AddDate(0, 0, -i)
		result[d.Format("2006-01-02")] = 0
	}

	// Sum completions by StartTime date
	for _, s := range pa.sessions {
		dateKey := s.StartTime.UTC().Format("2006-01-02")
		if _, ok := result[dateKey]; ok {
			result[dateKey] += s.TasksCompleted
		}
	}

	return result
}

// GetMoodCorrelations returns mood -> avg tasks completed per session, sorted by productivity.
// Uses internally loaded sessions. Sessions with no mood entries are excluded.
// Moods with < 2 sessions are excluded. Multi-mood sessions attribute full TasksCompleted to each mood.
func (pa *PatternAnalyzer) GetMoodCorrelations() []MoodCorrelation {
	type moodStats struct {
		totalCompleted int
		sessionCount   int
	}
	stats := make(map[string]*moodStats)

	for _, s := range pa.sessions {
		if len(s.MoodEntries) == 0 {
			continue
		}
		// Track unique moods per session to avoid double-counting
		seenMoods := make(map[string]bool)
		for _, me := range s.MoodEntries {
			if seenMoods[me.Mood] {
				continue
			}
			seenMoods[me.Mood] = true
			ms, ok := stats[me.Mood]
			if !ok {
				ms = &moodStats{}
				stats[me.Mood] = ms
			}
			ms.totalCompleted += s.TasksCompleted
			ms.sessionCount++
		}
	}

	var result []MoodCorrelation
	for mood, ms := range stats {
		if ms.sessionCount < 2 {
			continue // below minimum sample size
		}
		result = append(result, MoodCorrelation{
			Mood:              mood,
			SessionCount:      ms.sessionCount,
			AvgTasksCompleted: float64(ms.totalCompleted) / float64(ms.sessionCount),
		})
	}

	// Sort by productivity (highest avg first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].AvgTasksCompleted > result[j].AvgTasksCompleted
	})

	return result
}

// GetDoorPositionPreferences returns door position selection percentages and bias detection.
// Sessions with no door selections are excluded.
func (pa *PatternAnalyzer) GetDoorPositionPreferences() DoorPreferences {
	var left, center, right, total int

	for _, s := range pa.sessions {
		for _, ds := range s.DoorSelections {
			total++
			switch ds.DoorPosition {
			case 0:
				left++
			case 1:
				center++
			case 2:
				right++
			}
		}
	}

	if total == 0 {
		return DoorPreferences{}
	}

	leftPct := math.Round(float64(left)/float64(total)*1000) / 10
	centerPct := math.Round(float64(center)/float64(total)*1000) / 10
	rightPct := math.Round(float64(right)/float64(total)*1000) / 10

	prefs := DoorPreferences{
		LeftPercent:     leftPct,
		CenterPercent:   centerPct,
		RightPercent:    rightPct,
		TotalSelections: total,
	}

	if leftPct > 50 {
		prefs.HasBias = true
		prefs.BiasPosition = "left"
	} else if centerPct > 50 {
		prefs.HasBias = true
		prefs.BiasPosition = "center"
	} else if rightPct > 50 {
		prefs.HasBias = true
		prefs.BiasPosition = "right"
	}

	return prefs
}

// GetWeekOverWeek returns this week vs last week completion comparison.
// Weeks run Monday-Sunday (ISO 8601).
func (pa *PatternAnalyzer) GetWeekOverWeek() WeekComparison {
	now := pa.nowFunc().UTC()

	// Find Monday of current week
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	daysFromMonday := int(weekday) - int(time.Monday)
	thisMonday := now.AddDate(0, 0, -daysFromMonday)
	thisMonday = time.Date(thisMonday.Year(), thisMonday.Month(), thisMonday.Day(), 0, 0, 0, 0, time.UTC)
	lastMonday := thisMonday.AddDate(0, 0, -7)

	var thisWeek, lastWeek int
	for _, s := range pa.sessions {
		st := s.StartTime.UTC()
		if !st.Before(thisMonday) {
			thisWeek += s.TasksCompleted
		} else if !st.Before(lastMonday) && st.Before(thisMonday) {
			lastWeek += s.TasksCompleted
		}
	}

	wk := WeekComparison{
		ThisWeekTotal: thisWeek,
		LastWeekTotal: lastWeek,
	}

	if lastWeek == 0 && thisWeek == 0 {
		wk.PercentChange = 0.0
		wk.Direction = "same"
	} else if lastWeek == 0 {
		wk.PercentChange = 100.0
		wk.Direction = "up"
	} else {
		wk.PercentChange = math.Round(float64(thisWeek-lastWeek)/float64(lastWeek)*10000) / 100
		if thisWeek > lastWeek {
			wk.Direction = "up"
		} else if thisWeek < lastWeek {
			wk.Direction = "down"
		} else {
			wk.Direction = "same"
		}
	}

	return wk
}

// GetMostProductiveMood returns the mood with the highest average completion rate.
func (pa *PatternAnalyzer) GetMostProductiveMood() string {
	corrs := pa.GetMoodCorrelations()
	if len(corrs) == 0 {
		return ""
	}
	return corrs[0].Mood // already sorted by productivity
}

// GetMostFrequentMood returns the mood that appears most often across all sessions.
func (pa *PatternAnalyzer) GetMostFrequentMood() string {
	counts := make(map[string]int)
	for _, s := range pa.sessions {
		seen := make(map[string]bool)
		for _, me := range s.MoodEntries {
			if !seen[me.Mood] {
				counts[me.Mood]++
				seen[me.Mood] = true
			}
		}
	}

	var best string
	var bestCount int
	for mood, count := range counts {
		if count > bestCount || (count == bestCount && mood < best) {
			best = mood
			bestCount = count
		}
	}
	return best
}

// GetMostRecentMood returns the last mood entry from the most recent session with mood data.
func (pa *PatternAnalyzer) GetMostRecentMood() string {
	// Sessions are in order of insertion; find the last one with mood entries
	for i := len(pa.sessions) - 1; i >= 0; i-- {
		s := pa.sessions[i]
		if len(s.MoodEntries) > 0 {
			return s.MoodEntries[len(s.MoodEntries)-1].Mood
		}
	}
	return ""
}

// GetBypassRate returns the percentage of door refreshes across all sessions.
func (pa *PatternAnalyzer) GetBypassRate() float64 {
	var totalRefreshes, totalViews int
	for _, s := range pa.sessions {
		totalRefreshes += s.RefreshesUsed
		totalViews += s.DoorsViewed
	}
	if totalViews == 0 {
		return 0
	}
	return math.Round(float64(totalRefreshes)/float64(totalViews)*1000) / 10
}

// DayOverDayStats holds day-over-day comparison across multiple dimensions.
type DayOverDayStats struct {
	TodayDoors        int
	YesterdayDoors    int
	TodayBypasses     int
	YesterdayBypasses int
	TodayTasks        int
	YesterdayTasks    int
}

// GetDayOverDay returns day-over-day comparison of doors opened, bypasses, and tasks completed.
// Uses session StartTime (UTC) to group by day.
func (pa *PatternAnalyzer) GetDayOverDay() DayOverDayStats {
	now := pa.nowFunc().UTC()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	var stats DayOverDayStats
	for _, s := range pa.sessions {
		day := s.StartTime.UTC().Format("2006-01-02")
		switch day {
		case today:
			stats.TodayDoors += len(s.DoorSelections)
			stats.TodayBypasses += s.RefreshesUsed
			stats.TodayTasks += s.TasksCompleted
		case yesterday:
			stats.YesterdayDoors += len(s.DoorSelections)
			stats.YesterdayBypasses += s.RefreshesUsed
			stats.YesterdayTasks += s.TasksCompleted
		}
	}
	return stats
}

// GetMoodTrend returns a trend description based on mood-correlated productivity.
// Returns "improving", "stable", "declining", or "" if insufficient data.
// Compares the last 3 sessions' mood-productivity against the preceding 3.
func (pa *PatternAnalyzer) GetMoodTrend() string {
	if len(pa.sessions) < 6 {
		return ""
	}

	// Use the last 6 sessions, split into recent 3 and preceding 3
	n := len(pa.sessions)
	recent := pa.sessions[n-3:]
	preceding := pa.sessions[n-6 : n-3]

	avgRecent := avgCompleted(recent)
	avgPreceding := avgCompleted(preceding)

	if avgPreceding == 0 {
		if avgRecent > 0 {
			return "improving"
		}
		return "stable"
	}

	change := (avgRecent - avgPreceding) / avgPreceding
	switch {
	case change > 0.1:
		return "improving"
	case change < -0.1:
		return "declining"
	default:
		return "stable"
	}
}

// GetTotalCompleted returns the total number of tasks completed across all sessions.
func (pa *PatternAnalyzer) GetTotalCompleted() int {
	total := 0
	for _, s := range pa.sessions {
		total += s.TasksCompleted
	}
	return total
}

// SessionHighlights holds computed aggregate metrics from session history.
type SessionHighlights struct {
	TotalDoors         int
	TotalTasks         int
	AvgSessionDuration time.Duration
	FastestFirstDoor   time.Duration
	TotalDetailViews   int
	TotalNotesAdded    int
	LongestStreak      int
	PeakHour           int // 0-23, -1 if no data
}

// GetSessionHighlights computes aggregate metrics from all loaded sessions.
// Returns a SessionHighlights with zero-value fields for metrics that have no data.
// PeakHour is -1 when no sessions exist.
func (pa *PatternAnalyzer) GetSessionHighlights() SessionHighlights {
	h := SessionHighlights{PeakHour: -1}
	if len(pa.sessions) == 0 {
		return h
	}

	var totalDuration float64
	var fastestDoor float64
	hourCounts := make(map[int]int)
	completionDates := make(map[string]bool)

	for _, s := range pa.sessions {
		h.TotalTasks += s.TasksCompleted
		h.TotalDoors += len(s.DoorSelections)
		h.TotalDetailViews += s.DetailViews
		h.TotalNotesAdded += s.NotesAdded
		totalDuration += s.DurationSeconds

		// Fastest first door: minimum positive value
		if s.TimeToFirstDoorSecs > 0 {
			if fastestDoor == 0 || s.TimeToFirstDoorSecs < fastestDoor {
				fastestDoor = s.TimeToFirstDoorSecs
			}
		}

		// Peak hour: bucket by start hour
		hourCounts[s.StartTime.Hour()]++

		// Collect dates with completions for streak calculation
		if s.TasksCompleted > 0 {
			completionDates[s.StartTime.UTC().Format("2006-01-02")] = true
		}
	}

	// Average session duration
	h.AvgSessionDuration = time.Duration(totalDuration/float64(len(pa.sessions))*1000) * time.Millisecond

	// Fastest first door
	if fastestDoor > 0 {
		h.FastestFirstDoor = time.Duration(fastestDoor*1000) * time.Millisecond
	}

	// Peak hour: find the hour with most sessions
	if len(hourCounts) > 0 {
		maxCount := 0
		peakHour := 0
		// Iterate 0-23 for deterministic tie-breaking (earliest hour wins)
		for hour := 0; hour < 24; hour++ {
			if count, ok := hourCounts[hour]; ok && count > maxCount {
				maxCount = count
				peakHour = hour
			}
		}
		h.PeakHour = peakHour
	}

	// Longest streak from completion dates
	h.LongestStreak = longestConsecutiveDays(completionDates)

	return h
}

// longestConsecutiveDays finds the longest run of consecutive days in a set of date strings.
func longestConsecutiveDays(dates map[string]bool) int {
	if len(dates) == 0 {
		return 0
	}

	// Parse all dates and sort
	var parsed []time.Time
	for d := range dates {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		parsed = append(parsed, t)
	}
	if len(parsed) == 0 {
		return 0
	}

	sort.Slice(parsed, func(i, j int) bool {
		return parsed[i].Before(parsed[j])
	})

	longest := 1
	current := 1
	for i := 1; i < len(parsed); i++ {
		diff := parsed[i].Sub(parsed[i-1]).Hours() / 24
		if diff == 1 {
			current++
			if current > longest {
				longest = current
			}
		} else if diff > 1 {
			current = 1
		}
		// diff == 0: same day, skip
	}
	return longest
}

func avgCompleted(sessions []SessionMetrics) float64 {
	if len(sessions) == 0 {
		return 0
	}
	total := 0
	for _, s := range sessions {
		total += s.TasksCompleted
	}
	return float64(total) / float64(len(sessions))
}
