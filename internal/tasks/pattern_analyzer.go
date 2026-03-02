package tasks

import (
	"bufio"
	"encoding/json"
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
	AvgTasksCompleted float64 `json:"avg_tasks_completed"`
}

// AvoidanceEntry tracks tasks that are repeatedly bypassed.
type AvoidanceEntry struct {
	TaskText      string `json:"task_text"`
	TimesBypassed int    `json:"times_bypassed"`
	TimesShown    int    `json:"times_shown"`
	NeverSelected bool   `json:"never_selected"`
}

// PatternAnalyzer analyzes session metrics for user behavior patterns.
type PatternAnalyzer struct{}

// NewPatternAnalyzer creates a new PatternAnalyzer.
func NewPatternAnalyzer() *PatternAnalyzer {
	return &PatternAnalyzer{}
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
	scanner := bufio.NewScanner(f)
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
		count     int
		totalComp int
		taskTexts []string
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
	}

	var correlations []MoodCorrelation
	for mood, acc := range moods {
		// Minimum 3 sessions for a correlation
		if acc.count < 3 {
			continue
		}
		correlations = append(correlations, MoodCorrelation{
			Mood:              mood,
			SessionCount:      acc.count,
			PreferredType:     "", // Would need task pool lookup for type
			PreferredEffort:   "", // Would need task pool lookup for effort
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating patterns directory: %w", err)
	}
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
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
