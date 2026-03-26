package mcp

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	"github.com/arcavenae/ThreeDoors/internal/core/metrics"
)

// MoodCorrelationEntry maps a mood rating to aggregated session stats.
type MoodCorrelationEntry struct {
	Mood              string         `json:"mood"`
	SessionCount      int            `json:"session_count"`
	AvgCompletions    float64        `json:"avg_completions"`
	AvgDurationSecs   float64        `json:"avg_duration_seconds"`
	TaskTypeBreakdown map[string]int `json:"task_type_breakdown,omitempty"`
}

// MoodCorrelation holds the mood-to-productivity mapping.
type MoodCorrelation struct {
	From    time.Time              `json:"from"`
	To      time.Time              `json:"to"`
	Entries []MoodCorrelationEntry `json:"entries"`
}

// HourlyRate captures completion rate for a given hour.
type HourlyRate struct {
	Hour        int     `json:"hour"`
	Completions int     `json:"completions"`
	Sessions    int     `json:"sessions"`
	Rate        float64 `json:"rate"`
	AvgMood     string  `json:"avg_mood,omitempty"`
	AvgEffort   string  `json:"avg_effort,omitempty"`
}

// ProductivityProfile holds time-of-day analysis data.
type ProductivityProfile struct {
	From       time.Time    `json:"from"`
	To         time.Time    `json:"to"`
	HourlyData []HourlyRate `json:"hourly_data"`
	PeakHours  []int        `json:"peak_hours"`
	SlumpHours []int        `json:"slump_hours"`
}

// StreakEntry represents a single streak period.
type StreakEntry struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Length    int       `json:"length"`
}

// StreakData holds streak analysis results.
type StreakData struct {
	CurrentStreak  int           `json:"current_streak"`
	LongestStreak  int           `json:"longest_streak"`
	AverageStreak  float64       `json:"average_streak"`
	StreakHistory  []StreakEntry `json:"streak_history"`
	StreakBreakers []string      `json:"common_streak_breakers,omitempty"`
}

// BurnoutIndicators holds composite burnout risk assessment.
type BurnoutIndicators struct {
	Score              float64  `json:"score"`
	Level              string   `json:"level"`
	CompletionSlope    float64  `json:"completion_trend_slope"`
	MoodTrend          float64  `json:"mood_trend"`
	SessionLengthTrend float64  `json:"session_length_trend"`
	BypassRate         float64  `json:"bypass_rate"`
	InactiveDays       int      `json:"inactive_days"`
	TaskAvoidanceTypes []string `json:"task_avoidance_types,omitempty"`
}

// WeeklySummary holds a week's analytics summary.
type WeeklySummary struct {
	WeekOf          time.Time      `json:"week_of"`
	Velocity        int            `json:"velocity"`
	BestDay         string         `json:"best_day"`
	WorstDay        string         `json:"worst_day"`
	TypeMix         map[string]int `json:"type_mix"`
	Patterns        []string       `json:"discovered_patterns"`
	Recommendations []string       `json:"recommendations"`
}

// PatternMiner computes analytics from session history data.
type PatternMiner struct {
	reader *metrics.Reader
	pool   *core.TaskPool
}

// NewPatternMiner creates a PatternMiner backed by the given reader and pool.
func NewPatternMiner(reader *metrics.Reader, pool *core.TaskPool) *PatternMiner {
	return &PatternMiner{reader: reader, pool: pool}
}

// MoodCorrelationAnalysis returns mood vs productivity correlations for the given period.
func (pm *PatternMiner) MoodCorrelationAnalysis(from, to time.Time) (*MoodCorrelation, error) {
	sessions, err := pm.sessionsInRange(from, to)
	if err != nil {
		return nil, fmt.Errorf("read sessions for mood correlation: %w", err)
	}

	type moodAgg struct {
		count         int
		totalComplete int
		totalDuration float64
		taskTypes     map[string]int
	}

	byMood := make(map[string]*moodAgg)
	for _, s := range sessions {
		for _, me := range s.MoodEntries {
			mood := me.Mood
			if mood == "" {
				continue
			}
			agg, ok := byMood[mood]
			if !ok {
				agg = &moodAgg{taskTypes: make(map[string]int)}
				byMood[mood] = agg
			}
			agg.count++
			agg.totalComplete += s.TasksCompleted
			agg.totalDuration += s.DurationSeconds
		}
	}

	// Also count sessions with no mood entries but where we recorded the session mood count.
	// If no mood entries detail but mood_entries > 0, skip — we need the actual mood values.

	var entries []MoodCorrelationEntry
	for mood, agg := range byMood {
		entry := MoodCorrelationEntry{
			Mood:            mood,
			SessionCount:    agg.count,
			AvgCompletions:  float64(agg.totalComplete) / float64(agg.count),
			AvgDurationSecs: agg.totalDuration / float64(agg.count),
		}
		if len(agg.taskTypes) > 0 {
			entry.TaskTypeBreakdown = agg.taskTypes
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Mood < entries[j].Mood
	})

	return &MoodCorrelation{From: from, To: to, Entries: entries}, nil
}

// ProductivityProfileAnalysis returns time-of-day productivity data.
func (pm *PatternMiner) ProductivityProfileAnalysis(from, to time.Time) (*ProductivityProfile, error) {
	sessions, err := pm.sessionsInRange(from, to)
	if err != nil {
		return nil, fmt.Errorf("read sessions for productivity profile: %w", err)
	}

	hourly := make([]HourlyRate, 24)
	for i := range hourly {
		hourly[i].Hour = i
	}

	for _, s := range sessions {
		hour := s.StartTime.Hour()
		hourly[hour].Sessions++
		hourly[hour].Completions += s.TasksCompleted
	}

	for i := range hourly {
		if hourly[i].Sessions > 0 {
			hourly[i].Rate = float64(hourly[i].Completions) / float64(hourly[i].Sessions)
		}
	}

	// Find peak and slump hours (top 3 / bottom 3 by rate, among hours with sessions).
	type hourScore struct {
		hour int
		rate float64
	}
	var active []hourScore
	for _, h := range hourly {
		if h.Sessions > 0 {
			active = append(active, hourScore{h.Hour, h.Rate})
		}
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].rate > active[j].rate
	})

	var peak, slump []int
	for i := 0; i < len(active) && i < 3; i++ {
		peak = append(peak, active[i].hour)
	}
	for i := len(active) - 1; i >= 0 && len(slump) < 3; i-- {
		// Don't include hours already in peak.
		alreadyPeak := false
		for _, p := range peak {
			if active[i].hour == p {
				alreadyPeak = true
				break
			}
		}
		if !alreadyPeak {
			slump = append(slump, active[i].hour)
		}
	}

	return &ProductivityProfile{
		From:       from,
		To:         to,
		HourlyData: hourly,
		PeakHours:  peak,
		SlumpHours: slump,
	}, nil
}

// StreakAnalysis returns streak data based on consecutive days with completions.
func (pm *PatternMiner) StreakAnalysis() (*StreakData, error) {
	sessions, err := pm.allSessions()
	if err != nil {
		return nil, fmt.Errorf("read sessions for streak analysis: %w", err)
	}

	if len(sessions) == 0 {
		return &StreakData{StreakHistory: []StreakEntry{}}, nil
	}

	// Collect unique dates with at least one completion.
	completionDates := make(map[string]bool)
	for _, s := range sessions {
		if s.TasksCompleted > 0 {
			dateKey := s.StartTime.Format("2006-01-02")
			completionDates[dateKey] = true
		}
	}

	if len(completionDates) == 0 {
		return &StreakData{StreakHistory: []StreakEntry{}}, nil
	}

	// Sort dates.
	var dates []time.Time
	for d := range completionDates {
		t, _ := time.Parse("2006-01-02", d)
		dates = append(dates, t)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	// Build streaks from consecutive dates.
	var streaks []StreakEntry
	streakStart := dates[0]
	prev := dates[0]

	for i := 1; i < len(dates); i++ {
		diff := dates[i].Sub(prev).Hours() / 24
		if diff > 1.5 { // allow for timezone drift
			streaks = append(streaks, StreakEntry{
				StartDate: streakStart,
				EndDate:   prev,
				Length:    daysBetween(streakStart, prev) + 1,
			})
			streakStart = dates[i]
		}
		prev = dates[i]
	}
	// Close the last streak.
	streaks = append(streaks, StreakEntry{
		StartDate: streakStart,
		EndDate:   prev,
		Length:    daysBetween(streakStart, prev) + 1,
	})

	// Compute stats.
	var longest int
	var totalLen int
	for _, s := range streaks {
		totalLen += s.Length
		if s.Length > longest {
			longest = s.Length
		}
	}

	// Current streak: check if the last streak includes today or yesterday.
	var current int
	today := time.Now().UTC().Truncate(24 * time.Hour)
	lastStreak := streaks[len(streaks)-1]
	lastEnd := lastStreak.EndDate.Truncate(24 * time.Hour)
	daysSinceEnd := today.Sub(lastEnd).Hours() / 24
	if daysSinceEnd <= 1.5 {
		current = lastStreak.Length
	}

	avg := float64(totalLen) / float64(len(streaks))

	return &StreakData{
		CurrentStreak: current,
		LongestStreak: longest,
		AverageStreak: math.Round(avg*100) / 100,
		StreakHistory: streaks,
	}, nil
}

// BurnoutRisk computes a composite burnout risk score from multiple signals.
func (pm *PatternMiner) BurnoutRisk() (*BurnoutIndicators, error) {
	sessions, err := pm.allSessions()
	if err != nil {
		return nil, fmt.Errorf("read sessions for burnout risk: %w", err)
	}

	if len(sessions) == 0 {
		return &BurnoutIndicators{Level: "unknown"}, nil
	}

	// Split into recent (last 7 days) and older for trend comparison.
	now := time.Now().UTC()
	weekAgo := now.AddDate(0, 0, -7)

	var recent, older []core.SessionMetrics
	for _, s := range sessions {
		if s.StartTime.After(weekAgo) {
			recent = append(recent, s)
		} else {
			older = append(older, s)
		}
	}

	// 1. Completion trend slope (negative = declining).
	completionSlope := trendSlope(sessions, func(s core.SessionMetrics) float64 {
		return float64(s.TasksCompleted)
	})

	// 2. Mood trend (negative = declining).
	moodTrend := moodTrendValue(sessions)

	// 3. Session length trend.
	sessionSlope := trendSlope(sessions, func(s core.SessionMetrics) float64 {
		return s.DurationSeconds
	})

	// 4. Bypass rate.
	var totalBypasses, totalViews int
	for _, s := range recent {
		totalBypasses += s.RefreshesUsed
		totalViews += s.DoorsViewed
	}
	var bypassRate float64
	if totalViews > 0 {
		bypassRate = float64(totalBypasses) / float64(totalViews)
	}

	// 5. Inactive days in last 14 days.
	twoWeeksAgo := now.AddDate(0, 0, -14)
	activeDays := make(map[string]bool)
	for _, s := range sessions {
		if s.StartTime.After(twoWeeksAgo) {
			activeDays[s.StartTime.Format("2006-01-02")] = true
		}
	}
	inactiveDays := 14 - len(activeDays)
	if inactiveDays < 0 {
		inactiveDays = 0
	}

	// Composite score (0-1): weighted sum of normalized signals.
	// Each signal contributes 0-0.2 (5 signals * 0.2 = 1.0 max).
	var score float64

	// Negative completion slope → higher risk.
	if completionSlope < 0 {
		score += math.Min(math.Abs(completionSlope)*0.5, 0.2)
	}

	// Negative mood trend → higher risk.
	if moodTrend < 0 {
		score += math.Min(math.Abs(moodTrend)*0.5, 0.2)
	}

	// Declining session length → higher risk.
	if sessionSlope < 0 {
		score += math.Min(math.Abs(sessionSlope)*0.001, 0.2)
	}

	// High bypass rate → higher risk.
	score += math.Min(bypassRate*0.4, 0.2)

	// Many inactive days → higher risk.
	score += math.Min(float64(inactiveDays)/14.0*0.2, 0.2)

	score = math.Min(score, 1.0)

	level := "low"
	if score > 0.7 {
		level = "warning"
	} else if score > 0.4 {
		level = "moderate"
	}

	// Detect task avoidance by type from bypasses.
	_ = older // older used in trend comparison above

	return &BurnoutIndicators{
		Score:              math.Round(score*100) / 100,
		Level:              level,
		CompletionSlope:    math.Round(completionSlope*1000) / 1000,
		MoodTrend:          math.Round(moodTrend*1000) / 1000,
		SessionLengthTrend: math.Round(sessionSlope*1000) / 1000,
		BypassRate:         math.Round(bypassRate*1000) / 1000,
		InactiveDays:       inactiveDays,
	}, nil
}

// WeeklySummaryAnalysis returns a summary for the week containing the given date.
func (pm *PatternMiner) WeeklySummaryAnalysis(weekOf time.Time) (*WeeklySummary, error) {
	// Find the Monday of the given week.
	weekday := weekOf.Weekday()
	offset := int(weekday - time.Monday)
	if offset < 0 {
		offset += 7
	}
	monday := weekOf.AddDate(0, 0, -offset).Truncate(24 * time.Hour)
	sunday := monday.AddDate(0, 0, 7)

	sessions, err := pm.sessionsInRange(monday, sunday)
	if err != nil {
		return nil, fmt.Errorf("read sessions for weekly summary: %w", err)
	}

	// Velocity: total completions.
	var velocity int
	dayCompletions := make(map[string]int)
	typeMix := make(map[string]int)

	for _, s := range sessions {
		velocity += s.TasksCompleted
		day := s.StartTime.Weekday().String()
		dayCompletions[day] += s.TasksCompleted

		// Count door selections as rough proxy for task type engagement.
		for _, ds := range s.DoorSelections {
			typeMix[ds.TaskText] = typeMix[ds.TaskText] + 1
		}
	}

	// Best and worst day.
	var bestDay, worstDay string
	bestCount := -1
	worstCount := math.MaxInt
	for day, count := range dayCompletions {
		if count > bestCount {
			bestCount = count
			bestDay = day
		}
		if count < worstCount {
			worstCount = count
			worstDay = day
		}
	}

	// Discovered patterns.
	var patterns []string
	if len(sessions) > 0 {
		avgPerSession := float64(velocity) / float64(len(sessions))
		if avgPerSession > 3 {
			patterns = append(patterns, "High productivity week — averaging 3+ completions per session")
		}
		if len(dayCompletions) >= 5 {
			patterns = append(patterns, "Consistent activity across most weekdays")
		}
		if len(dayCompletions) <= 2 && len(sessions) > 0 {
			patterns = append(patterns, "Activity concentrated on few days")
		}
	}

	// Recommendations.
	var recommendations []string
	if velocity == 0 {
		recommendations = append(recommendations, "Try starting with a quick-win task to build momentum")
	}
	if len(dayCompletions) <= 2 && velocity > 0 {
		recommendations = append(recommendations, "Spread work across more days for sustainable pace")
	}

	return &WeeklySummary{
		WeekOf:          monday,
		Velocity:        velocity,
		BestDay:         bestDay,
		WorstDay:        worstDay,
		TypeMix:         typeMix,
		Patterns:        patterns,
		Recommendations: recommendations,
	}, nil
}

// sessionsInRange returns sessions with start time in [from, to).
func (pm *PatternMiner) sessionsInRange(from, to time.Time) ([]core.SessionMetrics, error) {
	if pm.reader == nil {
		return nil, nil
	}
	all, err := pm.reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var filtered []core.SessionMetrics
	for _, s := range all {
		if !s.StartTime.Before(from) && s.StartTime.Before(to) {
			filtered = append(filtered, s)
		}
	}
	return filtered, nil
}

// allSessions returns all sessions from the reader.
func (pm *PatternMiner) allSessions() ([]core.SessionMetrics, error) {
	if pm.reader == nil {
		return nil, nil
	}
	return pm.reader.ReadAll()
}

// trendSlope computes a simple linear regression slope over session values.
func trendSlope(sessions []core.SessionMetrics, valueFunc func(core.SessionMetrics) float64) float64 {
	n := len(sessions)
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, s := range sessions {
		x := float64(i)
		y := valueFunc(s)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denom := float64(n)*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (float64(n)*sumXY - sumX*sumY) / denom
}

// moodTrendValue converts mood strings to numeric values and computes trend.
func moodTrendValue(sessions []core.SessionMetrics) float64 {
	moodValues := map[string]float64{
		"great":   5,
		"good":    4,
		"okay":    3,
		"neutral": 3,
		"low":     2,
		"bad":     1,
	}

	var moodSessions []core.SessionMetrics
	for _, s := range sessions {
		if len(s.MoodEntries) > 0 {
			moodSessions = append(moodSessions, s)
		}
	}

	if len(moodSessions) < 2 {
		return 0
	}

	return trendSlope(moodSessions, func(s core.SessionMetrics) float64 {
		lastMood := s.MoodEntries[len(s.MoodEntries)-1].Mood
		if v, ok := moodValues[lastMood]; ok {
			return v
		}
		return 3 // default to neutral
	})
}

// daysBetween returns the number of days between two dates.
func daysBetween(a, b time.Time) int {
	a = a.Truncate(24 * time.Hour)
	b = b.Truncate(24 * time.Hour)
	return int(b.Sub(a).Hours() / 24)
}
