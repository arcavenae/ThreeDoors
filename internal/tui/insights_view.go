package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	tea "github.com/charmbracelet/bubbletea"
)

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// InsightsView displays the user insights dashboard.
type InsightsView struct {
	analyzer *tasks.PatternAnalyzer
	counter  *tasks.CompletionCounter
	width    int
}

// NewInsightsView creates a new InsightsView.
func NewInsightsView(analyzer *tasks.PatternAnalyzer, counter *tasks.CompletionCounter) *InsightsView {
	return &InsightsView{
		analyzer: analyzer,
		counter:  counter,
	}
}

// SetWidth sets the terminal width for rendering.
func (iv *InsightsView) SetWidth(w int) {
	iv.width = w
}

// Update handles messages for the insights view.
func (iv *InsightsView) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.Type == tea.KeyEscape {
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}
	return nil
}

// View renders the insights dashboard.
func (iv *InsightsView) View() string {
	s := strings.Builder{}

	s.WriteString(headerStyle.Render("Your Insights Dashboard"))
	s.WriteString("\n\n")

	if !iv.analyzer.HasSufficientData() {
		needed := iv.analyzer.GetSessionsNeeded()
		fmt.Fprintf(&s, "  Keep using ThreeDoors to unlock insights! (%d more sessions needed)\n\n", needed)
		s.WriteString(helpStyle.Render("Press Esc to return"))
		return s.String()
	}

	// Completion Trends
	iv.renderCompletionTrends(&s)

	// Streaks
	iv.renderStreaks(&s)

	// Mood & Productivity
	iv.renderMoodCorrelations(&s)

	// Door Position Preferences
	iv.renderDoorPreferences(&s)

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press Esc to return"))

	return s.String()
}

func (iv *InsightsView) renderCompletionTrends(s *strings.Builder) {
	s.WriteString("  COMPLETION TRENDS (Last 7 Days)\n")

	daily := iv.analyzer.GetDailyCompletions(7)

	// Sort keys chronologically
	type dayEntry struct {
		date  string
		count int
	}
	var entries []dayEntry
	for date, count := range daily {
		entries = append(entries, dayEntry{date, count})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].date < entries[j].date
	})

	// Day labels and counts
	var labels []string
	var counts []int
	for _, e := range entries {
		t, _ := time.Parse("2006-01-02", e.date)
		labels = append(labels, t.Format("Mon"))
		counts = append(counts, e.count)
	}

	// Sparkline
	spark := sparkline(counts)
	s.WriteString("  ")
	for _, label := range labels {
		fmt.Fprintf(s, "%-5s", label)
	}
	s.WriteString("\n  ")
	for _, ch := range spark {
		fmt.Fprintf(s, "%-5s", string(ch))
	}
	s.WriteString("\n  ")
	for _, c := range counts {
		fmt.Fprintf(s, "%-5d", c)
	}
	s.WriteString("\n\n")

	// Week over week
	wk := iv.analyzer.GetWeekOverWeek()
	var arrow string
	switch wk.Direction {
	case "up":
		arrow = "↑"
	case "down":
		arrow = "↓"
	case "same":
		arrow = "→"
	}
	pct := math.Abs(wk.PercentChange)
	fmt.Fprintf(s, "  This week: %d  |  Last week: %d  |  %s %.0f%%\n\n", wk.ThisWeekTotal, wk.LastWeekTotal, arrow, pct)
}

func (iv *InsightsView) renderStreaks(s *strings.Builder) {
	s.WriteString("  STREAKS\n")
	streak := iv.counter.GetStreak()
	if streak > 0 {
		fmt.Fprintf(s, "  Current streak: %d days\n\n", streak)
	} else {
		s.WriteString("  No active streak — complete a task to start one!\n\n")
	}
}

func (iv *InsightsView) renderMoodCorrelations(s *strings.Builder) {
	s.WriteString("  MOOD & PRODUCTIVITY\n")

	corrs := iv.analyzer.GetMoodCorrelations()
	if len(corrs) == 0 {
		s.WriteString("  Not enough mood data yet. Try logging moods with :mood\n\n")
		return
	}

	for _, c := range corrs {
		fmt.Fprintf(s, "  %-12s avg %.1f tasks/session (%d sessions)\n", c.Mood+":", c.AvgTasksCompleted, c.SessionCount)
	}

	// corrs is already sorted by productivity (highest first)
	mostProductive := corrs[0].Mood
	mostFrequent := iv.analyzer.GetMostFrequentMood()
	fmt.Fprintf(s, "  Most productive mood: %s\n", mostProductive)
	if mostFrequent != "" {
		fmt.Fprintf(s, "  Most frequent mood: %s\n", mostFrequent)
	}
	s.WriteString("\n")
}

func (iv *InsightsView) renderDoorPreferences(s *strings.Builder) {
	s.WriteString("  DOOR POSITION PREFERENCES\n")

	prefs := iv.analyzer.GetDoorPositionPreferences()
	if prefs.TotalSelections == 0 {
		s.WriteString("  No door selection data yet.\n")
		return
	}

	fmt.Fprintf(s, "  Left: %.0f%%  |  Center: %.0f%%  |  Right: %.0f%%\n", prefs.LeftPercent, prefs.CenterPercent, prefs.RightPercent)

	if prefs.HasBias {
		fmt.Fprintf(s, "  You tend to pick the %s door — try mixing it up!\n", prefs.BiasPosition)
	}
}

// sparkline renders a text sparkline using Unicode block characters.
func sparkline(values []int) string {
	if len(values) == 0 {
		return ""
	}
	maxVal := 0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		return strings.Repeat(string(sparkChars[0]), len(values))
	}
	var result strings.Builder
	for _, v := range values {
		idx := int(float64(v) / float64(maxVal) * float64(len(sparkChars)-1))
		if idx >= len(sparkChars) {
			idx = len(sparkChars) - 1
		}
		result.WriteRune(sparkChars[idx])
	}
	return result.String()
}
