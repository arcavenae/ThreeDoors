package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Layout mode constants for responsive breakpoints.
const (
	layoutCompact  = "compact"  // < 60 columns
	layoutNarrow   = "narrow"   // 60-79 columns
	layoutStandard = "standard" // 80-119 columns
	layoutWide     = "wide"     // 120+ columns
)

// InsightsView displays the user insights dashboard.
type InsightsView struct {
	analyzer   *core.PatternAnalyzer
	counter    *core.CompletionCounter
	funFactGen *core.FunFactGenerator
	width      int
	height     int
	cachedView string
	cacheValid bool
	lastWidth  int
}

// NewInsightsView creates a new InsightsView.
func NewInsightsView(analyzer *core.PatternAnalyzer, counter *core.CompletionCounter) *InsightsView {
	return &InsightsView{
		analyzer:   analyzer,
		counter:    counter,
		funFactGen: core.NewFunFactGenerator(analyzer, counter),
	}
}

// SetWidth sets the terminal width for rendering.
func (iv *InsightsView) SetWidth(w int) {
	if w != iv.width {
		iv.cacheValid = false
	}
	iv.width = w
}

// SetHeight sets the terminal height for future viewport support.
func (iv *InsightsView) SetHeight(h int) {
	iv.height = h
}

// layoutMode returns the layout mode based on the current terminal width.
func (iv *InsightsView) layoutMode() string {
	switch {
	case iv.width < 60:
		return layoutCompact
	case iv.width < 80:
		return layoutNarrow
	case iv.width < 120:
		return layoutStandard
	default:
		return layoutWide
	}
}

// invalidateCache marks the render cache as stale.
func (iv *InsightsView) invalidateCache() {
	iv.cacheValid = false
}

// Update handles messages for the insights view.
func (iv *InsightsView) Update(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.Type == tea.KeyEscape || msg.String() == "q" {
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		}
	}
	return nil
}

// View renders the insights dashboard.
func (iv *InsightsView) View() string {
	if iv.cacheValid && iv.lastWidth == iv.width {
		return iv.cachedView
	}

	var s strings.Builder

	// Dashboard header
	iv.renderDashboardHeader(&s)

	if !iv.analyzer.HasSufficientData() {
		iv.renderColdStart(&s)
		result := s.String()
		iv.cachedView = result
		iv.cacheValid = true
		iv.lastWidth = iv.width
		return result
	}

	// Hero number
	iv.renderHeroNumber(&s)

	// Content panels arranged by layout mode
	iv.renderPanelLayout(&s)

	// Fun fact at the bottom
	iv.renderFunFact(&s)

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press Esc to return"))

	result := s.String()
	iv.cachedView = result
	iv.cacheValid = true
	iv.lastWidth = iv.width
	return result
}

// renderDashboardHeader renders the styled "YOUR INSIGHTS DASHBOARD" header.
func (iv *InsightsView) renderDashboardHeader(s *strings.Builder) {
	// statsDashboardHeaderStyle has a border, so subtract 2 for border chars.
	headerContentWidth := iv.contentWidth() - 2
	if headerContentWidth < 1 {
		headerContentWidth = 1
	}
	header := statsDashboardHeaderStyle.Width(headerContentWidth).Render("YOUR INSIGHTS DASHBOARD")
	fmt.Fprintf(s, "%s\n", header)
}

// renderColdStart renders the insufficient-data message with styling.
func (iv *InsightsView) renderColdStart(s *strings.Builder) {
	needed := iv.analyzer.GetSessionsNeeded()
	msg := fmt.Sprintf("Keep using ThreeDoors to unlock insights! (%d more sessions needed)", needed)
	panelContentWidth := iv.contentWidth() - 2 // subtract border chars
	if panelContentWidth < 1 {
		panelContentWidth = 1
	}
	panel := statsPanelStyle.Width(panelContentWidth).Padding(1, 2).Render(msg)
	fmt.Fprintf(s, "\n%s\n\n", panel)
	s.WriteString(helpStyle.Render("Press Esc to return"))
}

// renderHeroNumber renders the gold-styled total tasks completed count.
func (iv *InsightsView) renderHeroNumber(s *strings.Builder) {
	total := iv.analyzer.GetTotalCompleted()
	heroText := fmt.Sprintf("★ %d tasks completed ★", total)
	hero := statsHeroStyle.Width(iv.contentWidth()).Render(heroText)
	fmt.Fprintf(s, "\n%s\n\n", hero)
}

// renderFunFact renders the daily fun fact with gold star styling.
func (iv *InsightsView) renderFunFact(s *strings.Builder) {
	fact := iv.funFactGen.Generate()
	styled := statsHeroStyle.Width(iv.contentWidth()).Render("★ " + fact)
	fmt.Fprintf(s, "\n%s\n", styled)
}

// renderPanelLayout arranges content panels based on the current layout mode.
func (iv *InsightsView) renderPanelLayout(s *strings.Builder) {
	mode := iv.layoutMode()

	trendsContent := iv.buildCompletionTrends()
	streaksContent := iv.buildStreaks()
	moodContent := iv.buildMoodCorrelations()
	doorContent := iv.buildDoorPreferences()

	switch mode {
	case layoutCompact:
		// Single column: all panels stacked vertically
		iv.renderSingleColumn(s, trendsContent, streaksContent, moodContent, doorContent)
	case layoutNarrow:
		// Two small panels side by side where possible
		iv.renderTwoColumn(s, trendsContent, streaksContent, moodContent, doorContent)
	case layoutStandard:
		// 2-column layout: [trends | streaks], [mood | door picks]
		iv.renderTwoColumn(s, trendsContent, streaksContent, moodContent, doorContent)
	case layoutWide:
		// 3-column top row: [trends | streaks | door picks], mood full width below
		iv.renderWideLayout(s, trendsContent, streaksContent, moodContent, doorContent)
	}
}

// contentWidth returns the usable content width.
func (iv *InsightsView) contentWidth() int {
	w := iv.width
	if w <= 0 {
		w = 80
	}
	if w > 2 {
		return w - 2 // small margin
	}
	return w
}

// panelWidth returns the width for a panel in multi-column layout.
// Accounts for 1-char gaps between columns and border overhead.
func (iv *InsightsView) panelWidth(columns int) int {
	cw := iv.contentWidth()
	gaps := columns - 1 // 1-char gap between columns
	return (cw - gaps) / columns
}

// makePanel wraps content in a styled bordered panel with a section header.
// width is the total rendered width including borders.
// Lipgloss Width() sets content width (excluding border), so we subtract 2 for L+R borders.
func makePanel(title, content string, width int) string {
	header := statsSectionHeaderStyle.Render(title)
	body := fmt.Sprintf("%s\n%s", header, content)
	contentWidth := width - 2 // subtract left + right border chars
	if contentWidth < 1 {
		contentWidth = 1
	}
	return statsPanelStyle.Width(contentWidth).Padding(0, 1).Render(body)
}

// renderSingleColumn renders all panels in a single vertical column.
func (iv *InsightsView) renderSingleColumn(s *strings.Builder, trends, streaks, mood, doors string) {
	w := iv.contentWidth()
	panels := []struct {
		title   string
		content string
	}{
		{"COMPLETION TRENDS", trends},
		{"STREAKS", streaks},
		{"MOOD & PRODUCTIVITY", mood},
		{"DOOR PICKS", doors},
	}
	for _, p := range panels {
		fmt.Fprintf(s, "%s\n", makePanel(p.title, p.content, w))
	}
}

// renderTwoColumn renders panels in a 2-column layout.
func (iv *InsightsView) renderTwoColumn(s *strings.Builder, trends, streaks, mood, doors string) {
	w := iv.panelWidth(2)

	row1Left := makePanel("COMPLETION TRENDS", trends, w)
	row1Right := makePanel("STREAKS", streaks, w)
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, row1Left, " ", row1Right)
	fmt.Fprintf(s, "%s\n", row1)

	row2Left := makePanel("MOOD & PRODUCTIVITY", mood, w)
	row2Right := makePanel("DOOR PICKS", doors, w)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, row2Left, " ", row2Right)
	fmt.Fprintf(s, "%s\n", row2)
}

// renderWideLayout renders a 3-column top row with mood full-width below.
func (iv *InsightsView) renderWideLayout(s *strings.Builder, trends, streaks, mood, doors string) {
	w3 := iv.panelWidth(3)

	col1 := makePanel("COMPLETION TRENDS", trends, w3)
	col2 := makePanel("STREAKS", streaks, w3)
	col3 := makePanel("DOOR PICKS", doors, w3)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, col1, " ", col2, " ", col3)
	fmt.Fprintf(s, "%s\n", topRow)

	// Full-width mood panel — use total content width for the rendered panel width.
	moodPanel := makePanel("MOOD & PRODUCTIVITY", mood, iv.contentWidth())
	fmt.Fprintf(s, "%s\n", moodPanel)
}

// buildCompletionTrends builds the completion trends panel content (no border).
func (iv *InsightsView) buildCompletionTrends() string {
	var s strings.Builder

	daily := iv.analyzer.GetDailyCompletions(7)

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

	var labels []string
	var counts []int
	for _, e := range entries {
		t, _ := time.Parse("2006-01-02", e.date)
		labels = append(labels, t.Format("Mon"))
		counts = append(counts, e.count)
	}

	styledChars := styledSparklineChars(counts)
	for _, label := range labels {
		fmt.Fprintf(&s, "%-5s", label)
	}
	s.WriteString("\n")
	for _, ch := range styledChars {
		fmt.Fprintf(&s, "%-5s", ch)
	}
	s.WriteString("\n")
	for _, c := range counts {
		fmt.Fprintf(&s, "%-5d", c)
	}

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
	fmt.Fprintf(&s, "\nThis week: %d | Last week: %d | %s %.0f%%", wk.ThisWeekTotal, wk.LastWeekTotal, arrow, pct)

	return s.String()
}

// buildStreaks builds the streaks panel content (no border).
func (iv *InsightsView) buildStreaks() string {
	streak := iv.counter.GetStreak()
	if streak > 0 {
		return fmt.Sprintf("Current streak: %d days", streak)
	}
	return "No active streak — complete a task to start one!"
}

// moodBarWidth returns the available width for bar charts based on layout mode.
func (iv *InsightsView) moodBarWidth() int {
	mode := iv.layoutMode()
	switch mode {
	case layoutCompact:
		return 10
	case layoutNarrow:
		return 15
	default:
		return 20
	}
}

// barChart renders a horizontal bar using █ (filled) and ░ (empty) characters.
// ratio is clamped to [0.0, 1.0], width must be > 0.
func barChart(ratio float64, width int, color lipgloss.AdaptiveColor) string {
	if width <= 0 {
		return ""
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	filled := int(math.Round(ratio * float64(width)))
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	style := lipgloss.NewStyle().Foreground(color)
	return style.Render(bar)
}

// buildMoodCorrelations builds the mood panel content (no border).
func (iv *InsightsView) buildMoodCorrelations() string {
	corrs := iv.analyzer.GetMoodCorrelations()
	if len(corrs) == 0 {
		return "Not enough mood data yet. Try logging moods with :mood"
	}

	// Find the maximum avg for proportional scaling.
	maxAvg := 0.0
	for _, c := range corrs {
		if c.AvgTasksCompleted > maxAvg {
			maxAvg = c.AvgTasksCompleted
		}
	}

	barWidth := iv.moodBarWidth()
	var s strings.Builder
	for _, c := range corrs {
		ratio := 0.0
		if maxAvg > 0 {
			ratio = c.AvgTasksCompleted / maxAvg
		}

		color, ok := moodColors[c.Mood]
		if !ok {
			color = defaultMoodColor
		}

		bar := barChart(ratio, barWidth, color)
		label := c.Mood
		if iv.layoutMode() == layoutCompact && len(label) > 6 {
			label = label[:6]
		}
		fmt.Fprintf(&s, "%-10s %s %.1f (%d)\n", label, bar, c.AvgTasksCompleted, c.SessionCount)
	}

	mostProductive := corrs[0].Mood
	mostFrequent := iv.analyzer.GetMostFrequentMood()
	fmt.Fprintf(&s, "Most productive mood: %s", mostProductive)
	if mostFrequent != "" {
		fmt.Fprintf(&s, "\nMost frequent mood: %s", mostFrequent)
	}

	return s.String()
}

// buildDoorPreferences builds the door preferences panel content (no border).
func (iv *InsightsView) buildDoorPreferences() string {
	prefs := iv.analyzer.GetDoorPositionPreferences()
	if prefs.TotalSelections == 0 {
		return "No door selection data yet."
	}

	var s strings.Builder
	fmt.Fprintf(&s, "Left: %.0f%%  |  Center: %.0f%%  |  Right: %.0f%%", prefs.LeftPercent, prefs.CenterPercent, prefs.RightPercent)

	if prefs.HasBias {
		fmt.Fprintf(&s, "\nYou tend to pick the %s door — try mixing it up!", prefs.BiasPosition)
	}

	return s.String()
}

// Gradient endpoint colors for the sparkline (color-blind safe: blue→yellow).
var (
	sparkColorStart = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#3B82F6"} // blue
	sparkColorEnd   = lipgloss.AdaptiveColor{Light: "#CA8A04", Dark: "#EAB308"} // yellow
)

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

// styledSparklineChars returns individually styled sparkline characters
// with a gradient from blue (low) to yellow (high).
func styledSparklineChars(values []int) []string {
	if len(values) == 0 {
		return nil
	}
	maxVal := 0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	chars := make([]string, len(values))
	for i, v := range values {
		var idx int
		var t float64
		if maxVal > 0 {
			t = float64(v) / float64(maxVal)
			idx = int(t * float64(len(sparkChars)-1))
			if idx >= len(sparkChars) {
				idx = len(sparkChars) - 1
			}
		}
		ch := string(sparkChars[idx])
		blended := blendHexColors(sparkColorStart.Dark, sparkColorEnd.Dark, t)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(blended))
		chars[i] = style.Render(ch)
	}
	return chars
}

// styledSparkline renders a gradient-colored sparkline as a single string.
func styledSparkline(values []int) string {
	chars := styledSparklineChars(values)
	return strings.Join(chars, "")
}

// blendHexColors linearly interpolates between two hex colors.
// t is clamped to [0, 1]. Colors must be "#RRGGBB" format.
func blendHexColors(from, to string, t float64) string {
	if t <= 0 {
		return from
	}
	if t >= 1 {
		return to
	}
	r1, g1, b1 := parseHex(from)
	r2, g2, b2 := parseHex(to)
	r := uint8(float64(r1) + t*(float64(r2)-float64(r1)))
	g := uint8(float64(g1) + t*(float64(g2)-float64(g1)))
	b := uint8(float64(b1) + t*(float64(b2)-float64(b1)))
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// parseHex extracts RGB components from a "#RRGGBB" hex string.
func parseHex(hex string) (r, g, b uint8) {
	if len(hex) == 7 && hex[0] == '#' {
		_, _ = fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	}
	return
}
