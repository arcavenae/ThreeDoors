package tasks

import (
	"fmt"
	"strings"
)

// FormatMultiDimensionalGreeting returns a compact "better than yesterday" greeting line.
// Returns empty string if no data is available. Output is deterministic (no randomness).
//
// Dimensions shown (when data exists):
//   - Tasks completed: day-over-day comparison
//   - Doors opened: day-over-day comparison
//   - Streak: consecutive days with completions
//   - Mood trend: improving/stable/declining (requires 6+ sessions)
//   - Week-over-week: comparison summary
//
// Messaging is encouraging regardless of direction.
func FormatMultiDimensionalGreeting(analyzer *PatternAnalyzer, counter *CompletionCounter) string {
	today := counter.GetTodayCount()
	yesterday := counter.GetYesterdayCount()
	streak := counter.GetStreak()
	recentMood := analyzer.GetMostRecentMood()
	dayStats := analyzer.GetDayOverDay()
	moodTrend := analyzer.GetMoodTrend()
	weekComp := analyzer.GetWeekOverWeek()

	// No data at all — return empty
	if today == 0 && yesterday == 0 && streak == 0 && recentMood == "" &&
		dayStats.TodayDoors == 0 && dayStats.YesterdayDoors == 0 {
		return ""
	}

	var parts []string

	// Task completion comparison with encouraging message
	parts = append(parts, formatTaskComparison(today, yesterday)...)

	// Doors opened comparison
	if dayStats.TodayDoors > 0 || dayStats.YesterdayDoors > 0 {
		parts = append(parts, formatDoorComparison(dayStats))
	}

	// Streak
	if streak > 0 {
		parts = append(parts, fmt.Sprintf("Streak: %d days", streak))
	}

	// Mood + trend
	if recentMood != "" {
		moodPart := fmt.Sprintf("Mood: %s", recentMood)
		if moodTrend != "" {
			moodPart += fmt.Sprintf(" (%s)", moodTrend)
		}
		parts = append(parts, moodPart)
	}

	if len(parts) == 0 {
		return ""
	}

	primary := "📈 " + strings.Join(parts, " | ")

	// Week-over-week detail line (compact, appended when data exists)
	weekLine := formatWeekComparison(weekComp)
	if weekLine != "" {
		return primary + "\n" + weekLine
	}

	return primary
}

// formatTaskComparison returns the task completion parts with encouraging messaging.
func formatTaskComparison(today, yesterday int) []string {
	if today == 0 && yesterday == 0 {
		return nil
	}

	var parts []string
	if today > 0 && yesterday > 0 {
		msg := encouragingComparison(today, yesterday, "tasks")
		parts = append(parts, msg)
	} else if today > 0 {
		parts = append(parts, fmt.Sprintf("Today: %d tasks", today))
	} else if yesterday > 0 {
		parts = append(parts, fmt.Sprintf("Yesterday: %d tasks", yesterday))
	}
	return parts
}

// formatDoorComparison returns a compact doors-opened comparison.
func formatDoorComparison(stats DayOverDayStats) string {
	if stats.TodayDoors > 0 && stats.YesterdayDoors > 0 {
		return encouragingComparison(stats.TodayDoors, stats.YesterdayDoors, "doors")
	}
	if stats.TodayDoors > 0 {
		return fmt.Sprintf("Doors: %d today", stats.TodayDoors)
	}
	return fmt.Sprintf("Doors: %d yesterday", stats.YesterdayDoors)
}

// encouragingComparison formats a today-vs-yesterday comparison that is always positive.
func encouragingComparison(today, yesterday int, dimension string) string {
	if today > yesterday {
		return fmt.Sprintf("%d %s today vs %d yesterday — on a roll!", today, dimension, yesterday)
	}
	if today == yesterday {
		return fmt.Sprintf("%d %s today, matching yesterday — steady momentum!", today, dimension)
	}
	return fmt.Sprintf("%d %s today vs %d yesterday — every one counts!", today, dimension, yesterday)
}

// formatWeekComparison returns a compact week-over-week line, or empty if no data.
func formatWeekComparison(wk WeekComparison) string {
	if wk.ThisWeekTotal == 0 && wk.LastWeekTotal == 0 {
		return ""
	}

	switch wk.Direction {
	case "up":
		return fmt.Sprintf("   Week: %d tasks (last week: %d) — trending up!", wk.ThisWeekTotal, wk.LastWeekTotal)
	case "down":
		return fmt.Sprintf("   Week: %d tasks (last week: %d) — building momentum", wk.ThisWeekTotal, wk.LastWeekTotal)
	default:
		return fmt.Sprintf("   Week: %d tasks (last week: %d) — steady pace", wk.ThisWeekTotal, wk.LastWeekTotal)
	}
}
