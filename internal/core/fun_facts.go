package core

import (
	"fmt"
	"sort"
	"time"
)

// DefaultFunFact is shown when no data-driven facts can be generated.
const DefaultFunFact = "Every door you open is a step forward!"

// FunFactGenerator produces rotating fun facts from real session data.
// Facts follow D-089 content rules: observe (don't prescribe), celebrate totals
// (not rates), frame gaps as potential, and never use decline comparisons.
type FunFactGenerator struct {
	analyzer *PatternAnalyzer
	counter  *CompletionCounter
	nowFunc  func() time.Time
}

// NewFunFactGenerator creates a FunFactGenerator using the analyzer's time function.
// Safe to call with nil analyzer or counter — will produce only the default fact.
func NewFunFactGenerator(analyzer *PatternAnalyzer, counter *CompletionCounter) *FunFactGenerator {
	nowFunc := time.Now
	if analyzer != nil {
		nowFunc = analyzer.nowFunc
	}
	return &FunFactGenerator{
		analyzer: analyzer,
		counter:  counter,
		nowFunc:  nowFunc,
	}
}

// Generate returns a single fun fact for today using deterministic daily rotation.
// Same UTC day always produces the same fact. Empty pool returns DefaultFunFact.
func (g *FunFactGenerator) Generate() string {
	pool := g.BuildFactPool()
	if len(pool) == 0 {
		return DefaultFunFact
	}

	now := g.nowFunc().UTC()
	day := now.YearDay() + now.Year()*366
	idx := day % len(pool)
	return pool[idx]
}

// BuildFactPool assembles all available facts from current data.
// Facts are sorted for deterministic ordering (important for daily rotation).
func (g *FunFactGenerator) BuildFactPool() []string {
	if g.analyzer == nil {
		return nil
	}

	var facts []string

	facts = append(facts, g.totalTasksFact()...)
	facts = append(facts, g.totalDoorsFact()...)
	facts = append(facts, g.totalSessionsFact()...)
	facts = append(facts, g.streakFact()...)
	facts = append(facts, g.bestDayFact()...)
	facts = append(facts, g.doorPreferenceFact()...)
	facts = append(facts, g.moodFacts()...)
	facts = append(facts, g.sessionCountFact()...)

	// Sort for deterministic pool ordering
	sort.Strings(facts)
	return facts
}

// totalTasksFact generates a fact about total tasks completed.
func (g *FunFactGenerator) totalTasksFact() []string {
	total := g.analyzer.GetTotalCompleted()
	if total == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("You've completed %d tasks in total — every one counts!", total),
	}
}

// totalDoorsFact generates a fact about total doors opened.
func (g *FunFactGenerator) totalDoorsFact() []string {
	var totalDoors int
	for _, s := range g.analyzer.sessions {
		totalDoors += len(s.DoorSelections)
	}
	if totalDoors == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("You've opened %d doors since your first session!", totalDoors),
	}
}

// totalSessionsFact generates a fact about the total number of sessions.
func (g *FunFactGenerator) totalSessionsFact() []string {
	count := len(g.analyzer.sessions)
	if count < 2 {
		return nil
	}
	return []string{
		fmt.Sprintf("You've had %d ThreeDoors sessions — each one a fresh start!", count),
	}
}

// streakFact generates a fact about the current streak.
func (g *FunFactGenerator) streakFact() []string {
	if g.counter == nil {
		return nil
	}
	streak := g.counter.GetStreak()
	if streak < 2 {
		return nil
	}
	return []string{
		fmt.Sprintf("You're on a %d-day streak — keep the momentum going!", streak),
	}
}

// bestDayFact generates a fact about which day of the week is most productive.
func (g *FunFactGenerator) bestDayFact() []string {
	daily := g.analyzer.GetDailyCompletions(30) // look at last 30 days
	if len(daily) == 0 {
		return nil
	}

	// Aggregate by day-of-week
	dayCounts := make(map[time.Weekday]int)
	for dateStr, count := range daily {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		dayCounts[t.Weekday()] += count
	}

	var bestDay time.Weekday
	bestCount := 0
	for day, count := range dayCounts {
		if count > bestCount {
			bestCount = count
			bestDay = day
		}
	}
	if bestCount == 0 {
		return nil
	}

	return []string{
		fmt.Sprintf("You complete the most tasks on %ss!", bestDay.String()),
	}
}

// doorPreferenceFact generates a fact about door position preferences.
func (g *FunFactGenerator) doorPreferenceFact() []string {
	prefs := g.analyzer.GetDoorPositionPreferences()
	if prefs.TotalSelections == 0 {
		return nil
	}

	// Find the most-picked position
	type posPct struct {
		name string
		pct  float64
	}
	positions := []posPct{
		{"left", prefs.LeftPercent},
		{"center", prefs.CenterPercent},
		{"right", prefs.RightPercent},
	}
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].pct > positions[j].pct
	})

	top := positions[0]
	return []string{
		fmt.Sprintf("You pick the %s door %.0f%% of the time!", top.name, top.pct),
	}
}

// moodFacts generates facts about mood correlations.
func (g *FunFactGenerator) moodFacts() []string {
	var facts []string

	// Most productive mood
	productive := g.analyzer.GetMostProductiveMood()
	if productive != "" {
		corrs := g.analyzer.GetMoodCorrelations()
		for _, c := range corrs {
			if c.Mood == productive {
				facts = append(facts, fmt.Sprintf("%s is your power mood — avg %.1f tasks!", productive, c.AvgTasksCompleted))
				break
			}
		}
	}

	// Most frequent mood
	frequent := g.analyzer.GetMostFrequentMood()
	if frequent != "" && frequent != productive {
		facts = append(facts, fmt.Sprintf("%s is your go-to mood — you bring it to most sessions!", frequent))
	}

	return facts
}

// sessionCountFact generates a fact about session activity.
func (g *FunFactGenerator) sessionCountFact() []string {
	sessions := g.analyzer.sessions
	if len(sessions) < 3 {
		return nil
	}

	// Count unique days with sessions
	days := make(map[string]bool)
	for _, s := range sessions {
		days[s.StartTime.UTC().Format("2006-01-02")] = true
	}

	if len(days) >= 3 {
		return []string{
			fmt.Sprintf("You've been active on %d different days — building a great habit!", len(days)),
		}
	}
	return nil
}
