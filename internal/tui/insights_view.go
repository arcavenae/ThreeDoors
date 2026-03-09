package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	"github.com/charmbracelet/bubbles/viewport"
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

// Tab constants for the insights view.
const (
	tabOverview = 0
	tabDetail   = 1
	tabCount    = 2
)

// Animation constants for counter reveals.
const (
	animationTickInterval = 30 * time.Millisecond
	animationDuration     = 500 * time.Millisecond
	animationStep         = float64(animationTickInterval) / float64(animationDuration)
)

// InsightsView displays the user insights dashboard.
type InsightsView struct {
	analyzer   *core.PatternAnalyzer
	counter    *core.CompletionCounter
	funFactGen *core.FunFactGenerator
	theme      *themes.DoorTheme
	width      int
	height     int
	cachedView string
	cacheValid bool
	lastWidth  int
	activeTab  int
	viewport   viewport.Model

	// Animation state for counter reveals (Story 40.7).
	animating              bool
	animationProgress      float64
	hasAnimatedThisSession bool

	// Milestone celebration banner
	milestoneChecker *core.MilestoneChecker
	activeMilestone  *core.Milestone
	bannerActive     bool
}

// NewInsightsView creates a new InsightsView.
// theme may be nil, in which case the independent fallback palette is used.
// milestoneChecker may be nil to disable milestone celebrations.
func NewInsightsView(analyzer *core.PatternAnalyzer, counter *core.CompletionCounter, theme *themes.DoorTheme, milestoneChecker *core.MilestoneChecker) *InsightsView {
	return &InsightsView{
		analyzer:         analyzer,
		counter:          counter,
		funFactGen:       core.NewFunFactGenerator(analyzer, counter),
		theme:            theme,
		milestoneChecker: milestoneChecker,
	}
}

// SetWidth sets the terminal width for rendering.
func (iv *InsightsView) SetWidth(w int) {
	if w != iv.width {
		iv.cacheValid = false
	}
	iv.width = w
}

// SetHeight sets the terminal height and updates the viewport.
func (iv *InsightsView) SetHeight(h int) {
	iv.height = h
	// Reserve lines for header + tab indicator + footer
	vpHeight := h - 4
	if vpHeight < 1 {
		vpHeight = 1
	}
	iv.viewport.Height = vpHeight
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

// StartAnimation begins the counter reveal animation if it hasn't played yet this session.
// Returns a tea.Cmd to start the tick loop, or nil if animation was already shown.
func (iv *InsightsView) StartAnimation() tea.Cmd {
	if iv.hasAnimatedThisSession {
		return nil
	}
	iv.animating = true
	iv.animationProgress = 0.0
	iv.invalidateCache()
	return statsAnimationTickCmd()
}

// statsAnimationTickCmd returns a command that sends a StatsAnimationTickMsg after the tick interval.
func statsAnimationTickCmd() tea.Cmd {
	return tea.Tick(animationTickInterval, func(_ time.Time) tea.Msg {
		return StatsAnimationTickMsg{}
	})
}

// statsAccentColor returns the theme's stats accent or the fallback gold.
func (iv *InsightsView) statsAccentColor() string {
	if iv.theme != nil && iv.theme.Colors.StatsAccent != "" {
		return iv.theme.Colors.StatsAccent
	}
	return "#FCD34D" // fallback gold (Phase 1 independent palette)
}

// statsGradientColors returns the sparkline gradient endpoints from the theme,
// or the Phase 1 fallback (blue→yellow).
func (iv *InsightsView) statsGradientColors() (start, end string) {
	if iv.theme != nil && iv.theme.Colors.StatsGradientStart != "" && iv.theme.Colors.StatsGradientEnd != "" {
		return iv.theme.Colors.StatsGradientStart, iv.theme.Colors.StatsGradientEnd
	}
	return sparkColorStart.Dark, sparkColorEnd.Dark
}

// statsPanelBorderColor returns the theme's stats accent as a lipgloss color,
// or the default gray border.
func (iv *InsightsView) statsPanelBorderColor() lipgloss.AdaptiveColor {
	if iv.theme != nil && iv.theme.Colors.StatsAccent != "" {
		return lipgloss.AdaptiveColor{Light: iv.theme.Colors.StatsAccent, Dark: iv.theme.Colors.StatsAccent}
	}
	return lipgloss.AdaptiveColor{Light: "#555555", Dark: "#555555"}
}

// CheckAndShowMilestone checks for an unshown milestone and activates the banner.
// Returns a tea.Cmd for the auto-dismiss timer, or nil if no milestone.
func (iv *InsightsView) CheckAndShowMilestone(totalTasks, currentStreak, sessionCount int) tea.Cmd {
	if iv.milestoneChecker == nil {
		return nil
	}

	m := iv.milestoneChecker.CheckMilestones(totalTasks, currentStreak, sessionCount)
	if m == nil {
		return nil
	}

	iv.activeMilestone = m
	iv.bannerActive = true
	iv.invalidateCache()

	// Mark as shown immediately so it won't appear again
	_ = iv.milestoneChecker.MarkShown(m.ID)

	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return MilestoneDismissMsg{}
	})
}

// Update handles messages for the insights view.
func (iv *InsightsView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case StatsAnimationTickMsg:
		if !iv.animating {
			return nil
		}
		iv.animationProgress += animationStep
		iv.invalidateCache()
		if iv.animationProgress >= 1.0 {
			iv.animationProgress = 1.0
			iv.animating = false
			iv.hasAnimatedThisSession = true
			return nil
		}
		return statsAnimationTickCmd()

	case MilestoneDismissMsg:
		iv.bannerActive = false
		iv.invalidateCache()
		return nil

	case tea.KeyMsg:
		// If banner is active, any key dismisses it without passing through
		if iv.bannerActive {
			iv.bannerActive = false
			iv.invalidateCache()
			return nil
		}

		switch {
		case msg.Type == tea.KeyEscape || msg.String() == "q":
			iv.animating = false
			return func() tea.Msg { return ReturnToDoorsMsg{} }
		case msg.Type == tea.KeyTab:
			iv.activeTab = (iv.activeTab + 1) % tabCount
			iv.invalidateCache()
			return nil
		case msg.Type == tea.KeyShiftTab:
			iv.activeTab = (iv.activeTab + tabCount - 1) % tabCount
			iv.invalidateCache()
			return nil
		default:
			// Forward scroll keys to viewport when on Detail tab
			if iv.activeTab == tabDetail {
				var cmd tea.Cmd
				iv.viewport, cmd = iv.viewport.Update(msg)
				if cmd != nil {
					iv.invalidateCache()
				}
				return cmd
			}
		}
	}
	return nil
}

// View renders the insights dashboard.
func (iv *InsightsView) View() string {
	// Disable cache during animation — every tick must re-render.
	if !iv.animating && iv.cacheValid && iv.lastWidth == iv.width {
		return iv.cachedView
	}

	var s strings.Builder

	// Milestone celebration banner (rendered above everything)
	if iv.bannerActive && iv.activeMilestone != nil {
		iv.renderMilestoneBanner(&s)
	}

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

	// Render tab indicator
	iv.renderTabIndicator(&s)

	switch iv.activeTab {
	case tabOverview:
		// Hero number
		iv.renderHeroNumber(&s)
		// Content panels arranged by layout mode
		iv.renderPanelLayout(&s)
	case tabDetail:
		iv.renderDetailTab(&s)
	}

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

// renderTabIndicator renders the tab bar showing which tab is active.
func (iv *InsightsView) renderTabIndicator(s *strings.Builder) {
	activeStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	inactiveStyle := lipgloss.NewStyle()

	var overview, detail string
	if iv.activeTab == tabOverview {
		overview = activeStyle.Render("[Overview]")
		detail = inactiveStyle.Render(" Detail")
	} else {
		overview = inactiveStyle.Render("Overview ")
		detail = activeStyle.Render("[Detail]")
	}

	tabLine := overview + " " + detail
	fmt.Fprintf(s, "%s\n", tabLine)
}

// renderDetailTab renders the Detail tab content with a viewport.
func (iv *InsightsView) renderDetailTab(s *strings.Builder) {
	var content strings.Builder
	iv.renderHeatmap(&content)
	iv.viewport.Width = iv.contentWidth()
	iv.viewport.SetContent(content.String())
	fmt.Fprintf(s, "\n%s\n", iv.viewport.View())
}

// heatmapLevel maps a daily completion count to a 0-4 intensity level.
// Level 0: 0 completions, 1: 1-2, 2: 3-4, 3: 5-6, 4: 7+
func heatmapLevel(count int) int {
	switch {
	case count == 0:
		return 0
	case count <= 2:
		return 1
	case count <= 4:
		return 2
	case count <= 6:
		return 3
	default:
		return 4
	}
}

// heatmapChars maps intensity levels to Unicode shade characters.
var heatmapChars = [5]rune{' ', '░', '▒', '▓', '█'}

// Heatmap gradient colors (green tones, GitHub-style).
var (
	heatmapColorStart = "#2D4A22" // dim green
	heatmapColorEnd   = "#39D353" // bright green
)

// heatmapGrid builds a 7-row (Mon-Sun) by N-column (weeks) grid from daily completion data.
// Returns the grid and the number of weeks (columns). Weeks are ISO weeks (Mon-Sun).
// Most recent week is rightmost. Only weeks with data are included.
func heatmapGrid(daily map[string]int) ([7][]int, int) {
	if len(daily) == 0 {
		return [7][]int{}, 0
	}

	// Parse all dates and find the range
	type dateCount struct {
		date  time.Time
		count int
	}
	var entries []dateCount
	for dateStr, count := range daily {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		entries = append(entries, dateCount{t, count})
	}
	if len(entries) == 0 {
		return [7][]int{}, 0
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].date.Before(entries[j].date)
	})

	earliest := entries[0].date
	latest := entries[len(entries)-1].date

	// Find Monday of earliest week and Monday of latest week
	earliestMonday := mondayOf(earliest)
	latestMonday := mondayOf(latest)

	numWeeks := int(latestMonday.Sub(earliestMonday).Hours()/24)/7 + 1
	if numWeeks > 8 {
		// Only show last 8 weeks
		earliestMonday = latestMonday.AddDate(0, 0, -7*7)
		numWeeks = 8
	}

	var grid [7][]int
	for d := 0; d < 7; d++ {
		grid[d] = make([]int, numWeeks)
	}

	// Build lookup
	dateCounts := make(map[string]int, len(entries))
	for _, e := range entries {
		dateCounts[e.date.Format("2006-01-02")] = e.count
	}

	// Fill grid
	for w := 0; w < numWeeks; w++ {
		weekStart := earliestMonday.AddDate(0, 0, w*7)
		for d := 0; d < 7; d++ {
			day := weekStart.AddDate(0, 0, d)
			key := day.Format("2006-01-02")
			grid[d][w] = dateCounts[key]
		}
	}

	return grid, numWeeks
}

// mondayOf returns the Monday of the ISO week containing t.
func mondayOf(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	offset := int(weekday) - int(time.Monday)
	monday := t.AddDate(0, 0, -offset)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
}

// renderHeatmap renders the GitHub-style activity heatmap.
func (iv *InsightsView) renderHeatmap(s *strings.Builder) {
	if iv.layoutMode() == layoutCompact {
		fmt.Fprintf(s, "Widen terminal to see activity heatmap\n")
		return
	}

	daily := iv.analyzer.GetDailyCompletions(56)
	grid, numWeeks := heatmapGrid(daily)
	if numWeeks == 0 {
		fmt.Fprintf(s, "Not enough data for activity heatmap.\n")
		return
	}

	dayLabels := [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}

	// Column headers: W1..WN
	fmt.Fprintf(s, "ACTIVITY HEATMAP (8 weeks)\n\n")
	fmt.Fprintf(s, "    ") // row label spacer
	for w := 0; w < numWeeks; w++ {
		fmt.Fprintf(s, " W%-3d", w+1)
	}
	s.WriteString("\n")

	// Grid rows
	for d := 0; d < 7; d++ {
		fmt.Fprintf(s, " %s ", dayLabels[d])
		for w := 0; w < numWeeks; w++ {
			count := grid[d][w]
			level := heatmapLevel(count)
			ch := string(heatmapChars[level])

			if level == 0 {
				// No color for empty cells — use dim style
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D4D4D4", Dark: "#333333"})
				fmt.Fprintf(s, " %s   ", dimStyle.Render("·"))
			} else {
				t := float64(level-1) / 3.0
				color := blendHexColors(heatmapColorStart, heatmapColorEnd, t)
				style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				fmt.Fprintf(s, " %s   ", style.Render(ch))
			}
		}
		s.WriteString("\n")
	}

	// Legend
	s.WriteString("\n")
	s.WriteString(" Less ")
	legendLevels := [5]struct {
		ch    string
		label string
	}{
		{"·", "0"},
		{string(heatmapChars[1]), "1-2"},
		{string(heatmapChars[2]), "3-4"},
		{string(heatmapChars[3]), "5-6"},
		{string(heatmapChars[4]), "7+"},
	}
	for i, l := range legendLevels {
		if i == 0 {
			dimStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D4D4D4", Dark: "#333333"})
			fmt.Fprintf(s, "%s ", dimStyle.Render(l.ch))
		} else {
			t := float64(i-1) / 3.0
			color := blendHexColors(heatmapColorStart, heatmapColorEnd, t)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			fmt.Fprintf(s, "%s ", style.Render(l.ch))
		}
	}
	s.WriteString("More\n")
}

// renderDashboardHeader renders the styled "YOUR INSIGHTS DASHBOARD" header.
func (iv *InsightsView) renderDashboardHeader(s *strings.Builder) {
	// statsDashboardHeaderStyle has a border, so subtract 2 for border chars.
	headerContentWidth := iv.contentWidth() - 2
	if headerContentWidth < 1 {
		headerContentWidth = 1
	}
	style := statsDashboardHeaderStyle.BorderForeground(iv.statsPanelBorderColor())
	header := style.Width(headerContentWidth).Render("YOUR INSIGHTS DASHBOARD")
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
	style := statsPanelStyle.BorderForeground(iv.statsPanelBorderColor())
	panel := style.Width(panelContentWidth).Padding(1, 2).Render(msg)
	fmt.Fprintf(s, "\n%s\n\n", panel)
	s.WriteString(helpStyle.Render("Press Esc to return"))
}

// renderMilestoneBanner renders a celebration banner at the top of the insights view.
func (iv *InsightsView) renderMilestoneBanner(s *strings.Builder) {
	accentColor := iv.statsAccentColor()
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1A1A1A")).
		Background(lipgloss.Color(accentColor)).
		Align(lipgloss.Center).
		Padding(0, 2)

	contentWidth := iv.contentWidth()
	if contentWidth < 1 {
		contentWidth = 1
	}

	banner := bannerStyle.Width(contentWidth).Render("★ " + iv.activeMilestone.Message + " ★")
	fmt.Fprintf(s, "%s\n\n", banner)
}

// renderHeroNumber renders the themed total tasks completed count.
// During animation, the displayed value interpolates from 0 to the final value.
func (iv *InsightsView) renderHeroNumber(s *strings.Builder) {
	total := iv.analyzer.GetTotalCompleted()
	displayed := iv.animatedValue(total)
	heroText := fmt.Sprintf("★ %d tasks completed ★", displayed)
	style := statsHeroStyle.Foreground(lipgloss.Color(iv.statsAccentColor()))
	hero := style.Width(iv.contentWidth()).Render(heroText)
	fmt.Fprintf(s, "\n%s\n\n", hero)
}

// animatedValue returns the interpolated value based on animation progress.
// When not animating (or animation complete), returns the final value unchanged.
func (iv *InsightsView) animatedValue(finalValue int) int {
	if !iv.animating {
		return finalValue
	}
	return int(float64(finalValue) * iv.animationProgress)
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
	highlightsContent := iv.buildSessionHighlights()

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

	// Session highlights panel — added as full-width panel after existing layout
	if highlightsContent != "" {
		fmt.Fprintf(s, "%s\n", makePanel("SESSION HIGHLIGHTS", highlightsContent, iv.contentWidth(), iv.statsPanelBorderColor()))
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
func makePanel(title, content string, width int, borderColor lipgloss.AdaptiveColor) string {
	header := statsSectionHeaderStyle.Render(title)
	body := fmt.Sprintf("%s\n%s", header, content)
	contentWidth := width - 2 // subtract left + right border chars
	if contentWidth < 1 {
		contentWidth = 1
	}
	style := statsPanelStyle.BorderForeground(borderColor)
	return style.Width(contentWidth).Padding(0, 1).Render(body)
}

// renderSingleColumn renders all panels in a single vertical column.
func (iv *InsightsView) renderSingleColumn(s *strings.Builder, trends, streaks, mood, doors string) {
	w := iv.contentWidth()
	bc := iv.statsPanelBorderColor()
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
		fmt.Fprintf(s, "%s\n", makePanel(p.title, p.content, w, bc))
	}
}

// renderTwoColumn renders panels in a 2-column layout.
func (iv *InsightsView) renderTwoColumn(s *strings.Builder, trends, streaks, mood, doors string) {
	w := iv.panelWidth(2)
	bc := iv.statsPanelBorderColor()

	row1Left := makePanel("COMPLETION TRENDS", trends, w, bc)
	row1Right := makePanel("STREAKS", streaks, w, bc)
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, row1Left, " ", row1Right)
	fmt.Fprintf(s, "%s\n", row1)

	row2Left := makePanel("MOOD & PRODUCTIVITY", mood, w, bc)
	row2Right := makePanel("DOOR PICKS", doors, w, bc)
	row2 := lipgloss.JoinHorizontal(lipgloss.Top, row2Left, " ", row2Right)
	fmt.Fprintf(s, "%s\n", row2)
}

// renderWideLayout renders a 3-column top row with mood full-width below.
func (iv *InsightsView) renderWideLayout(s *strings.Builder, trends, streaks, mood, doors string) {
	w3 := iv.panelWidth(3)
	bc := iv.statsPanelBorderColor()

	col1 := makePanel("COMPLETION TRENDS", trends, w3, bc)
	col2 := makePanel("STREAKS", streaks, w3, bc)
	col3 := makePanel("DOOR PICKS", doors, w3, bc)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, col1, " ", col2, " ", col3)
	fmt.Fprintf(s, "%s\n", topRow)

	// Full-width mood panel — use total content width for the rendered panel width.
	moodPanel := makePanel("MOOD & PRODUCTIVITY", mood, iv.contentWidth(), bc)
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

	gradStart, gradEnd := iv.statsGradientColors()
	styledChars := styledSparklineCharsWithGradient(counts, gradStart, gradEnd)
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
// During animation, the streak count interpolates from 0 to its final value.
func (iv *InsightsView) buildStreaks() string {
	streak := iv.counter.GetStreak()
	if streak > 0 {
		displayed := iv.animatedValue(streak)
		return fmt.Sprintf("Current streak: %d days", displayed)
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

// buildSessionHighlights builds the session highlights panel content (no border).
// Metrics with no data are omitted rather than showing "0" or "N/A".
func (iv *InsightsView) buildSessionHighlights() string {
	h := iv.analyzer.GetSessionHighlights()

	type metric struct {
		label string
		value string
	}
	var metrics []metric

	if h.TotalDoors > 0 {
		metrics = append(metrics, metric{"Doors opened", fmt.Sprintf("%d", h.TotalDoors)})
	}
	if h.TotalTasks > 0 {
		metrics = append(metrics, metric{"Tasks completed", fmt.Sprintf("%d", h.TotalTasks)})
	}
	if h.AvgSessionDuration > 0 {
		metrics = append(metrics, metric{"Avg session", formatDuration(h.AvgSessionDuration)})
	}
	if h.FastestFirstDoor > 0 {
		metrics = append(metrics, metric{"Fastest first pick", formatDuration(h.FastestFirstDoor)})
	}
	if h.TotalDetailViews > 0 {
		metrics = append(metrics, metric{"Detail views", fmt.Sprintf("%d", h.TotalDetailViews)})
	}
	if h.TotalNotesAdded > 0 {
		metrics = append(metrics, metric{"Notes added", fmt.Sprintf("%d", h.TotalNotesAdded)})
	}
	if h.LongestStreak > 0 {
		metrics = append(metrics, metric{"Longest streak", fmt.Sprintf("%d days", h.LongestStreak)})
	}
	if h.PeakHour >= 0 {
		metrics = append(metrics, metric{"Peak hour", formatHour(h.PeakHour)})
	}

	if len(metrics) == 0 {
		return ""
	}

	// Find max label width for alignment
	maxLabel := 0
	for _, m := range metrics {
		if len(m.label) > maxLabel {
			maxLabel = len(m.label)
		}
	}

	var s strings.Builder
	for i, m := range metrics {
		if i > 0 {
			s.WriteString("\n")
		}
		fmt.Fprintf(&s, "%-*s  %s", maxLabel, m.label, m.value)
	}
	return s.String()
}

// formatDuration formats a duration for display, using the most appropriate unit.
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		secs := d.Seconds()
		if secs == float64(int(secs)) {
			return fmt.Sprintf("%ds", int(secs))
		}
		return fmt.Sprintf("%.1fs", secs)
	default:
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		if secs == 0 {
			return fmt.Sprintf("%dm", mins)
		}
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
}

// formatHour formats a 0-23 hour as a human-readable time string (e.g., "10am", "2pm").
func formatHour(hour int) string {
	switch {
	case hour == 0:
		return "12am"
	case hour < 12:
		return fmt.Sprintf("%dam", hour)
	case hour == 12:
		return "12pm"
	default:
		return fmt.Sprintf("%dpm", hour-12)
	}
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
// with a gradient from the start color (low) to the end color (high).
// Uses the default blue→yellow when called without theme context.
func styledSparklineChars(values []int) []string {
	return styledSparklineCharsWithGradient(values, sparkColorStart.Dark, sparkColorEnd.Dark)
}

// styledSparklineCharsWithGradient returns individually styled sparkline characters
// using the specified gradient colors.
func styledSparklineCharsWithGradient(values []int, gradientStart, gradientEnd string) []string {
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
		blended := blendHexColors(gradientStart, gradientEnd, t)
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
