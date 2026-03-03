package core

import (
	"fmt"
	"strings"
)

// FormatInsights returns a formatted string summary of pattern analysis.
// Returns an encouragement message if report is nil.
func FormatInsights(report *PatternReport) string {
	if report == nil {
		return "Keep going! After 5 sessions, you'll start seeing patterns."
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Session Insights (%d sessions analyzed)\n\n", report.SessionCount)

	// Door position preference
	bias := report.DoorPositionBias
	if bias.TotalSelections > 0 && bias.PreferredPosition != "none" {
		pct := 0
		switch bias.PreferredPosition {
		case "left":
			pct = bias.LeftCount * 100 / bias.TotalSelections
		case "center":
			pct = bias.CenterCount * 100 / bias.TotalSelections
		case "right":
			pct = bias.RightCount * 100 / bias.TotalSelections
		}
		fmt.Fprintf(&b, "Door Preference: You tend to pick the %s door (%d%%)\n\n", bias.PreferredPosition, pct)
	} else {
		b.WriteString("Door Preference: No strong door preference detected\n\n")
	}

	// Task type stats — find most selected and most bypassed
	if len(report.TaskTypeStats) > 0 {
		var mostSelected, mostBypassed string
		maxSel, maxByp := 0, 0
		for text, stats := range report.TaskTypeStats {
			if stats.TimesSelected > maxSel {
				maxSel = stats.TimesSelected
				mostSelected = text
			}
			if stats.TimesBypassed > maxByp {
				maxByp = stats.TimesBypassed
				mostBypassed = text
			}
		}
		if mostSelected != "" || mostBypassed != "" {
			parts := []string{}
			if mostSelected != "" {
				parts = append(parts, fmt.Sprintf("Most selected: %q (%d times)", truncate(mostSelected, 30), maxSel))
			}
			if mostBypassed != "" {
				parts = append(parts, fmt.Sprintf("Most bypassed: %q (%d times)", truncate(mostBypassed, 30), maxByp))
			}
			fmt.Fprintf(&b, "Task Types: %s\n\n", strings.Join(parts, " | "))
		}
	}

	// Time of day
	if len(report.TimeOfDayPatterns) > 0 {
		best := report.TimeOfDayPatterns[0]
		for _, p := range report.TimeOfDayPatterns[1:] {
			if p.AvgTasksCompleted > best.AvgTasksCompleted {
				best = p
			}
		}
		fmt.Fprintf(&b, "Best Time: You complete the most tasks in the %s (avg %.1f/session)\n\n", best.Period, best.AvgTasksCompleted)
	}

	// Mood correlations
	if len(report.MoodCorrelations) > 0 {
		b.WriteString("Mood Patterns:\n")
		for _, mc := range report.MoodCorrelations {
			detail := fmt.Sprintf("avg %.1f completed", mc.AvgTasksCompleted)
			if mc.PreferredType != "" {
				detail = fmt.Sprintf("prefer %s tasks, %s", mc.PreferredType, detail)
			}
			fmt.Fprintf(&b, "  - When %s: %s\n", mc.Mood, detail)
		}
		b.WriteString("\n")
	}

	// Avoidance
	avoidance5Plus := []AvoidanceEntry{}
	for _, a := range report.AvoidanceList {
		if a.TimesBypassed >= 5 {
			avoidance5Plus = append(avoidance5Plus, a)
		}
	}
	if len(avoidance5Plus) > 0 {
		fmt.Fprintf(&b, "Avoidance Alert: %d tasks bypassed 5+ times\n", len(avoidance5Plus))
		for _, a := range avoidance5Plus {
			fmt.Fprintf(&b, "  - %q (bypassed %d times)\n", truncate(a.TaskText, 40), a.TimesBypassed)
		}
	}

	return b.String()
}

// FormatMoodInsights returns a focused mood-correlation summary.
// Returns appropriate messages for nil report or empty correlations.
func FormatMoodInsights(report *PatternReport) string {
	if report == nil {
		return "Not enough data yet — need at least 5 sessions for insights."
	}
	if len(report.MoodCorrelations) == 0 {
		return "No mood correlation data yet. Log moods during sessions (press M) to build patterns. Need at least 3 sessions with the same mood."
	}

	var b strings.Builder

	// Count sessions with mood data
	totalMoodSessions := 0
	for _, mc := range report.MoodCorrelations {
		totalMoodSessions += mc.SessionCount
	}

	fmt.Fprintf(&b, "Mood Correlation Analysis (%d sessions with mood data)\n\n", totalMoodSessions)

	// Table header
	fmt.Fprintf(&b, "%-12s | %-8s | %-16s | %-16s | %s\n",
		"Mood", "Sessions", "Preferred Type", "Preferred Effort", "Avg Completed")
	fmt.Fprintf(&b, "%-12s-+-%-8s-+-%-16s-+-%-16s-+-%s\n",
		"------------", "--------", "----------------", "----------------", "-------------")

	// Find most productive mood
	var mostProductiveMood string
	bestAvg := 0.0

	for _, mc := range report.MoodCorrelations {
		typeStr := mc.PreferredType
		if typeStr == "" {
			typeStr = "-"
		}
		effortStr := mc.PreferredEffort
		if effortStr == "" {
			effortStr = "-"
		}
		fmt.Fprintf(&b, "%-12s | %-8d | %-16s | %-16s | %.1f\n",
			mc.Mood, mc.SessionCount, typeStr, effortStr, mc.AvgTasksCompleted)

		if mc.AvgTasksCompleted > bestAvg {
			bestAvg = mc.AvgTasksCompleted
			mostProductiveMood = mc.Mood
		}
	}

	// Insights section
	b.WriteString("\nInsights:\n")
	if mostProductiveMood != "" {
		fmt.Fprintf(&b, "  - Your most productive mood is %q (avg %.1f tasks/session)\n", mostProductiveMood, bestAvg)
	}
	b.WriteString("  - Mood data improves door selection — keep logging moods!\n")

	return b.String()
}

// FormatAvoidanceInsights returns a focused avoidance-pattern summary.
func FormatAvoidanceInsights(report *PatternReport) string {
	if report == nil {
		return "Not enough data yet — need at least 5 sessions for insights."
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Avoidance Analysis (%d sessions analyzed)\n\n", report.SessionCount)

	// Tasks bypassed 5+ times
	avoidance5Plus := []AvoidanceEntry{}
	for _, a := range report.AvoidanceList {
		if a.TimesBypassed >= 5 {
			avoidance5Plus = append(avoidance5Plus, a)
		}
	}

	if len(avoidance5Plus) == 0 {
		b.WriteString("No significant avoidance patterns detected yet.\n")
		b.WriteString("Tasks need to be bypassed 5+ times to appear here.\n")
		return b.String()
	}

	fmt.Fprintf(&b, "Tasks bypassed 5+ times: %d\n\n", len(avoidance5Plus))
	for _, a := range avoidance5Plus {
		label := "seen"
		if a.NeverSelected {
			label = "never selected"
		}
		fmt.Fprintf(&b, "  - %q (bypassed %d/%d times, %s)\n",
			truncate(a.TaskText, 40), a.TimesBypassed, a.TimesShown, label)
	}
	b.WriteString("\n")

	// Mood-avoidance correlation
	if len(report.MoodCorrelations) > 0 {
		hasAvoidedTypes := false
		for _, mc := range report.MoodCorrelations {
			if mc.AvoidedType != "" {
				hasAvoidedTypes = true
				break
			}
		}
		if hasAvoidedTypes {
			b.WriteString("Mood-Avoidance Patterns:\n")
			for _, mc := range report.MoodCorrelations {
				if mc.AvoidedType != "" {
					fmt.Fprintf(&b, "  - When %s, you tend to bypass %s tasks\n", mc.Mood, mc.AvoidedType)
				}
			}
			b.WriteString("\n")
		}
	}

	// Persistent avoidance (10+)
	persistent := []AvoidanceEntry{}
	for _, a := range report.AvoidanceList {
		if a.TimesBypassed >= 10 {
			persistent = append(persistent, a)
		}
	}
	if len(persistent) > 0 {
		fmt.Fprintf(&b, "Persistent avoidance (10+ bypasses): %d tasks\n", len(persistent))
		b.WriteString("These tasks will prompt you with options when they appear in doors.\n")
	}

	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
