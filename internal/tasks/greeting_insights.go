package tasks

import (
	"fmt"
	"strings"
)

// FormatMultiDimensionalGreeting returns a compact "better than yesterday" greeting line.
// Returns empty string if no data is available. Output is deterministic (no randomness).
func FormatMultiDimensionalGreeting(analyzer *PatternAnalyzer, counter *CompletionCounter) string {
	today := counter.GetTodayCount()
	yesterday := counter.GetYesterdayCount()
	streak := counter.GetStreak()
	recentMood := analyzer.GetMostRecentMood()

	// Case 4: No data at all
	if today == 0 && yesterday == 0 && streak == 0 && recentMood == "" {
		return ""
	}

	var parts []string

	if today > 0 {
		// Case 1: Today has completions
		parts = append(parts, fmt.Sprintf("Today: %d", today))
		if yesterday > 0 {
			parts = append(parts, fmt.Sprintf("Yesterday: %d", yesterday))
		}
	} else if yesterday > 0 {
		// Case 2: No completions today, yesterday data exists
		parts = append(parts, fmt.Sprintf("Yesterday: %d", yesterday))
	}

	// Case 5: Streak > 0
	if streak > 0 {
		parts = append(parts, fmt.Sprintf("Streak: %d days", streak))
	}

	// Case 3: Mood data available
	if recentMood != "" {
		parts = append(parts, fmt.Sprintf("Mood: %s", recentMood))
	}

	if len(parts) == 0 {
		return ""
	}

	return "📈 " + strings.Join(parts, " | ")
}
