package tui

import (
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/core/connection"
	"github.com/arcaven/ThreeDoors/internal/tui/themes"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
)

// doorHintKeys maps door index to its selection key for inline hints.
var doorHintKeys = [3]string{"a", "w", "d"}

// typeIcon returns the emoji icon for a task type.
func typeIcon(t core.TaskType) string {
	switch t {
	case core.TypeCreative:
		return "🎨"
	case core.TypeAdministrative:
		return "📋"
	case core.TypeTechnical:
		return "🔧"
	case core.TypePhysical:
		return "💪"
	default:
		return ""
	}
}

// categoryBadge builds a compact badge string for a task's categories.
func categoryBadge(task *core.Task) string {
	var parts []string
	if icon := typeIcon(task.Type); icon != "" {
		parts = append(parts, icon)
	}
	if task.Effort != "" {
		parts = append(parts, string(task.Effort))
	}
	if task.Location != "" {
		parts = append(parts, string(task.Location))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}

// DoorsView renders the three doors interface.
type DoorsView struct {
	pool              *core.TaskPool
	currentDoors      []*core.Task
	selectedDoorIndex int
	completedCount    int
	width             int
	height            int
	tracker           *core.SessionTracker
	greeting          string
	footerMessage     string
	avoidanceMap      map[string]int // task text → bypass count (TimesBypassed)
	avoidanceShown    map[string]int // task text → shown count (TimesShown)
	patternAnalyzer   *core.PatternAnalyzer
	completionCounter *core.CompletionCounter
	syncTracker       *core.SyncStatusTracker
	syncSpinner       *SyncSpinner
	timeContext       *core.TimeContext
	pendingConflicts  int
	pendingProposals  int
	theme             *themes.DoorTheme
	themeRegistry     *themes.Registry
	duplicateTaskIDs  map[string]bool
	doorAnimation     *DoorAnimation
	planningTimestamp *time.Time
	showKeyHints      bool
	baseThemeName     string // user's configured theme for fallback
	seasonalEnabled   bool   // whether seasonal auto-switch is active
	connMgr           *connection.ConnectionManager
}

// NewDoorsView creates a new DoorsView.
func NewDoorsView(pool *core.TaskPool, tracker *core.SessionTracker) *DoorsView {
	dv := &DoorsView{
		pool:              pool,
		selectedDoorIndex: -1,
		tracker:           tracker,
		greeting:          pickGreeting(-1),
		footerMessage:     pickFooterMessage(-1),
		avoidanceMap:      make(map[string]int),
		avoidanceShown:    make(map[string]int),
		duplicateTaskIDs:  make(map[string]bool),
		doorAnimation:     NewDoorAnimation(),
	}
	dv.RefreshDoors()
	return dv
}

// SetAvoidanceData populates the avoidance map from a pattern report.
func (dv *DoorsView) SetAvoidanceData(report *core.PatternReport) {
	dv.avoidanceMap = make(map[string]int)
	dv.avoidanceShown = make(map[string]int)
	if report == nil {
		return
	}
	for _, entry := range report.AvoidanceList {
		dv.avoidanceMap[entry.TaskText] = entry.TimesBypassed
		dv.avoidanceShown[entry.TaskText] = entry.TimesShown
	}
}

// SetInsightsData sets the pattern analyzer and completion counter for the multi-dimensional greeting.
func (dv *DoorsView) SetInsightsData(pa *core.PatternAnalyzer, cc *core.CompletionCounter) {
	dv.patternAnalyzer = pa
	dv.completionCounter = cc
}

// SetSyncTracker sets the sync status tracker for displaying provider sync state.
func (dv *DoorsView) SetSyncTracker(tracker *core.SyncStatusTracker) {
	dv.syncTracker = tracker
}

// SetSyncSpinner sets the spinner used during provider sync operations.
func (dv *DoorsView) SetSyncSpinner(sp *SyncSpinner) {
	dv.syncSpinner = sp
}

// SetTimeContext sets the calendar time context for time-aware door selection and display.
func (dv *DoorsView) SetTimeContext(tc *core.TimeContext) {
	dv.timeContext = tc
}

// TimeContext returns the current time context (for testing).
func (dv *DoorsView) TimeContext() *core.TimeContext {
	return dv.timeContext
}

// SetDuplicateTaskIDs sets the set of task IDs flagged as potential duplicates.
func (dv *DoorsView) SetDuplicateTaskIDs(ids map[string]bool) {
	dv.duplicateTaskIDs = ids
}

// SetPlanningTimestamp sets the planning session timestamp for focus boost.
func (dv *DoorsView) SetPlanningTimestamp(t *time.Time) {
	dv.planningTimestamp = t
}

// SetShowKeyHints sets whether key hints are visible on doors.
func (dv *DoorsView) SetShowKeyHints(show bool) {
	dv.showKeyHints = show
}

// SetPendingConflicts sets the number of unresolved sync conflicts.
func (dv *DoorsView) SetPendingConflicts(count int) {
	dv.pendingConflicts = count
}

// SetPendingProposals sets the number of pending LLM proposals.
func (dv *DoorsView) SetPendingProposals(count int) {
	dv.pendingProposals = count
}

// SetThemeByName looks up the named theme in the registry and sets it as active.
// Falls back to DefaultThemeName if the name is not found, and logs a warning.
func (dv *DoorsView) SetThemeByName(name string) {
	if dv.themeRegistry == nil {
		dv.themeRegistry = themes.NewDefaultRegistry()
	}
	if name == "" {
		name = themes.DefaultThemeName
	}
	theme, ok := dv.themeRegistry.Get(name)
	if !ok {
		log.Printf("theme %q not found, falling back to %q", name, themes.DefaultThemeName)
		theme, _ = dv.themeRegistry.Get(themes.DefaultThemeName)
	}
	dv.theme = theme
}

// Theme returns the currently active door theme.
func (dv *DoorsView) Theme() *themes.DoorTheme {
	return dv.theme
}

// SetThemeRegistry sets a custom theme registry (useful for testing).
func (dv *DoorsView) SetThemeRegistry(r *themes.Registry) {
	dv.themeRegistry = r
}

// SetBaseThemeName stores the user's configured theme name and activates it.
// This name is used as the fallback when a seasonal theme's MinWidth is not met.
func (dv *DoorsView) SetBaseThemeName(name string) {
	dv.baseThemeName = name
	dv.SetThemeByName(name)
}

// SetConnectionManager sets the connection manager for health alert rendering.
func (dv *DoorsView) SetConnectionManager(cm *connection.ConnectionManager) {
	dv.connMgr = cm
}

// SetSeasonalEnabled enables or disables automatic seasonal theme switching.
func (dv *DoorsView) SetSeasonalEnabled(enabled bool) {
	dv.seasonalEnabled = enabled
}

// ResolveSeasonalTheme checks if a seasonal theme should override the base
// theme for the current date. Called at construction time and on planning
// session start. When no matching seasonal theme exists in the registry,
// the base theme is left unchanged.
func (dv *DoorsView) ResolveSeasonalTheme(now time.Time) {
	if !dv.seasonalEnabled {
		return
	}
	if dv.themeRegistry == nil {
		dv.themeRegistry = themes.NewDefaultRegistry()
	}
	season := themes.ResolveSeason(now, themes.DefaultSeasonRanges)
	if season == "" {
		return
	}
	if seasonalTheme, ok := dv.themeRegistry.GetBySeason(season); ok {
		dv.theme = seasonalTheme
	}
}

// pickGreeting selects a random greeting, avoiding lastIdx to prevent consecutive repeats.
func pickGreeting(lastIdx int) string {
	if len(greetingMessages) <= 1 {
		return greetingMessages[0]
	}
	idx := rand.IntN(len(greetingMessages))
	for idx == lastIdx {
		idx = rand.IntN(len(greetingMessages))
	}
	return greetingMessages[idx]
}

// Greeting returns the current startup greeting message.
func (dv *DoorsView) Greeting() string {
	return dv.greeting
}

// pickFooterMessage selects a random footer message from the greeting pool,
// avoiding lastIdx to prevent consecutive repeats.
func pickFooterMessage(lastIdx int) string {
	if len(greetingMessages) <= 1 {
		return greetingMessages[0]
	}
	idx := rand.IntN(len(greetingMessages))
	for idx == lastIdx {
		idx = rand.IntN(len(greetingMessages))
	}
	return greetingMessages[idx]
}

// RotateFooterMessage picks a new footer message (called on refresh/return).
func (dv *DoorsView) RotateFooterMessage() {
	dv.footerMessage = pickFooterMessage(-1)
}

// RefreshDoors selects new random doors from the pool.
// Uses time-contextual selection when calendar data is available,
// and applies focus boost when a valid planning timestamp exists.
func (dv *DoorsView) RefreshDoors() {
	if dv.timeContext != nil && dv.timeContext.HasCalendar {
		dv.currentDoors = core.SelectDoorsWithTimeContext(dv.pool, 3, dv.timeContext)
	} else if dv.planningTimestamp != nil && !core.IsFocusExpired(*dv.planningTimestamp) {
		dv.currentDoors = core.SelectDoorsWithFocus(dv.pool, 3, dv.planningTimestamp)
	} else {
		dv.currentDoors = core.SelectDoors(dv.pool, 3)
	}
	dv.selectedDoorIndex = -1
}

// GetCurrentDoorTexts returns the text of currently displayed doors.
func (dv *DoorsView) GetCurrentDoorTexts() []string {
	var texts []string
	for _, t := range dv.currentDoors {
		texts = append(texts, t.Text)
	}
	return texts
}

// IncrementCompleted increments the session completion count.
func (dv *DoorsView) IncrementCompleted() {
	dv.completedCount++
}

// SetWidth sets the terminal width for rendering.
func (dv *DoorsView) SetWidth(w int) {
	dv.width = w
}

// SetHeight sets the terminal height for rendering.
func (dv *DoorsView) SetHeight(h int) {
	dv.height = h
}

// RenderHeader returns the header content: title, greeting, multi-dimensional
// greeting, and time context badge. This is used by MainModel to compose the
// header zone independently of the doors.
func (dv *DoorsView) RenderHeader() string {
	var header strings.Builder
	header.WriteString(headerStyle.Render("ThreeDoors - Technical Demo"))
	header.WriteString("\n")
	header.WriteString(greetingStyle.Render(dv.greeting))
	if dv.patternAnalyzer != nil && dv.completionCounter != nil {
		multiGreeting := core.FormatMultiDimensionalGreeting(dv.patternAnalyzer, dv.completionCounter)
		if multiGreeting != "" {
			header.WriteString("\n")
			header.WriteString(greetingStyle.Render(multiGreeting))
		}
	}
	if timeStr := core.FormatTimeContext(dv.timeContext); timeStr != "" {
		header.WriteString("\n")
		header.WriteString(badgeStyle.Render(timeStr))
	}
	return header.String()
}

// RenderCompactHeader returns a single-line header for compact breakpoints.
func (dv *DoorsView) RenderCompactHeader() string {
	return headerStyle.Render("ThreeDoors")
}

// RenderDoors returns only the rendered three doors (no header, footer, or
// status section). Returns an empty-state message when no doors are available.
func (dv *DoorsView) RenderDoors() string {
	if len(dv.currentDoors) == 0 {
		var s strings.Builder
		s.WriteString(flashStyle.Render("All tasks done! Great work!"))
		s.WriteString("\n\nPress 'q' to quit.\n")
		return s.String()
	}

	hintsEnabled := dv.showKeyHints

	doorWidth := dv.doorWidth()
	doorHeight := dv.doorHeight()

	activeTheme := dv.resolveActiveTheme(doorWidth, doorHeight)
	usePerDoorColors := dv.width >= 60
	hasSelection := dv.selectedDoorIndex >= 0

	var renderedDoors []string
	for i, task := range dv.currentDoors {
		parts := []string{task.Text}

		if srcBadge := SourceBadge(task.SourceProvider); srcBadge != "" {
			parts = append(parts, srcBadge)
		}
		if dv.duplicateTaskIDs[task.ID] {
			parts = append(parts, DuplicateIndicator())
		}
		if prBadge := DevDispatchBadge(task); prBadge != "" {
			parts = append(parts, prBadge)
		}
		if core.HasFocusTag(task) && dv.planningTimestamp != nil && !core.IsFocusExpired(*dv.planningTimestamp) {
			parts = append(parts, focusBadgeStyle.Render("focus"))
		}
		if badge := categoryBadge(task); badge != "" {
			parts = append(parts, badgeStyle.Render(badge))
		}
		if bypassCount, ok := dv.avoidanceMap[task.Text]; ok && bypassCount >= 5 {
			shownCount := dv.avoidanceShown[task.Text]
			if shownCount == 0 {
				shownCount = bypassCount
			}
			avoidStyle := lipgloss.NewStyle().Faint(true)
			parts = append(parts, avoidStyle.Render(fmt.Sprintf("Seen %d times", shownCount)))
		}

		content := lipgloss.JoinVertical(lipgloss.Left, parts...)

		statusIndicator := lipgloss.NewStyle().
			Foreground(StatusColor(string(task.Status))).
			Render(fmt.Sprintf("[%s]", task.Status))
		devBadge := ""
		if task.DevDispatch != nil && task.DevDispatch.Queued {
			devBadge = " " + DevDispatchBadge(task)
		}
		content = statusIndicator + devBadge + "\n\n" + content

		isSelected := i == dv.selectedDoorIndex
		animating := dv.doorAnimation != nil && dv.doorAnimation.Active()

		if animating {
			// During animation, skip content styling — border color conveys emphasis
		} else if hasSelection {
			if isSelected {
				content = selectedContentStyle.Render(content)
			} else {
				content = unselectedContentStyle.Render(content)
			}
		}

		hint := ""
		if hintsEnabled && i < len(doorHintKeys) {
			hint = renderDoorHint(doorHintKeys[i], true, isSelected, hasSelection)
		}

		doorEmphasis := 0.0
		if dv.doorAnimation != nil {
			doorEmphasis = dv.doorAnimation.Emphasis(i)
		} else if isSelected {
			doorEmphasis = 1.0
		}

		if activeTheme != nil {
			renderedDoors = append(renderedDoors, activeTheme.Render(content, doorWidth, doorHeight, isSelected, hint, doorEmphasis))
		} else if animating {
			emphasis := dv.doorAnimation.Emphasis(i)
			style := animatedDoorStyle(i, emphasis, doorWidth, doorHeight, usePerDoorColors)
			renderedDoors = append(renderedDoors, style.Render(content))
		} else {
			var style lipgloss.Style
			if isSelected {
				style = selectedDoorStyle.Width(doorWidth).Height(doorHeight).AlignVertical(lipgloss.Center)
			} else if hasSelection {
				style = unselectedDoorStyle.Width(doorWidth).Height(doorHeight).AlignVertical(lipgloss.Center)
			} else if usePerDoorColors && i < len(doorColors) {
				style = doorStyle.BorderForeground(doorColors[i]).Width(doorWidth).Height(doorHeight).AlignVertical(lipgloss.Center)
			} else {
				style = doorStyle.Width(doorWidth).Height(doorHeight).AlignVertical(lipgloss.Center)
			}
			renderedDoors = append(renderedDoors, style.Render(content))
		}
	}

	var doorSection strings.Builder
	doorRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedDoors...)
	doorSection.WriteString(doorRow)

	// Continuous threshold/floor line beneath all doors (Story 48.2).
	if firstNewline := strings.IndexByte(doorRow, '\n'); firstNewline > 0 {
		rowWidth := ansi.StringWidth(doorRow[:firstNewline])
		thresholdLine := strings.Repeat("▔", rowWidth)
		thresholdStyle := separatorStyle
		if activeTheme != nil {
			thresholdStyle = lipgloss.NewStyle().Foreground(activeTheme.Colors.Frame)
		}
		fmt.Fprintf(&doorSection, "\n%s", thresholdStyle.Render(thresholdLine))
	}

	return doorSection.String()
}

// RenderStatusSection returns the status indicators below the doors:
// completion count, conflict notifications, proposal badges, sync bar,
// and connection health alerts.
func (dv *DoorsView) RenderStatusSection() string {
	var s strings.Builder

	if dv.completedCount > 0 {
		fmt.Fprintf(&s, "Completed this session: %d", dv.completedCount)
	}

	if dv.pendingConflicts > 0 {
		if s.Len() > 0 {
			s.WriteString("\n\n")
		}
		s.WriteString(conflictHeaderStyle.Render(fmt.Sprintf("⚠ %d sync conflict(s) detected — press C to resolve", dv.pendingConflicts)))
	}

	if dv.pendingProposals > 0 {
		if s.Len() > 0 {
			s.WriteString("\n\n")
		}
		s.WriteString(proposalBadgeStyle.Render(fmt.Sprintf("[%d suggestions] — press S to review", dv.pendingProposals)))
	}

	if syncBar := RenderSyncStatusBarWithSpinner(dv.syncTracker, dv.syncSpinner); syncBar != "" {
		if s.Len() > 0 {
			s.WriteString("\n\n")
		}
		s.WriteString(syncBar)
	}

	if alert := dv.connectionHealthAlert(); alert != "" {
		if s.Len() > 0 {
			s.WriteString("\n\n")
		}
		s.WriteString(alert)
	}

	return s.String()
}

// RenderFooter returns the footer content: action hints and help text.
// Used by MainModel to compose the footer zone independently.
func (dv *DoorsView) RenderFooter() string {
	hintsEnabled := dv.showKeyHints
	hasSelection := dv.selectedDoorIndex >= 0

	var footer strings.Builder
	if hintsEnabled {
		var actionParts []string
		actionParts = append(actionParts, renderInlineHint("s", true)+" re-roll")
		actionParts = append(actionParts, renderInlineHint("n", true)+" add task")
		if hasSelection {
			actionParts = append(actionParts, renderInlineHint("enter", true)+" confirm")
		}
		footer.WriteString(helpStyle.Render(strings.Join(actionParts, "  ")))
		footer.WriteString("\n\n")
		footer.WriteString(helpStyle.Render("/ search | m mood | : command | ? help | q quit"))
		footer.WriteString("\n")
		footer.WriteString(greetingStyle.Render("hints: on — h to hide"))
	} else {
		footer.WriteString(helpStyle.Render("a/left, w/up, d/right to select (again to deselect) | s/down to re-roll | Enter/Space to open | N feedback | / search | M mood | q quit"))
		footer.WriteString("\n")
		footer.WriteString(greetingStyle.Render(dv.footerMessage))
	}
	return footer.String()
}

// RenderCompactFooter returns a single-line footer for compact breakpoints.
func (dv *DoorsView) RenderCompactFooter() string {
	return helpStyle.Render("a/w/d select | s re-roll | q quit")
}

// HasDoors reports whether there are doors to display.
func (dv *DoorsView) HasDoors() bool {
	return len(dv.currentDoors) > 0
}

// doorWidth computes the door width from the terminal width.
func (dv *DoorsView) doorWidth() int {
	doorWidth := 30
	if dv.width > 20 {
		doorWidth = (dv.width - 6) / 3
		if doorWidth < 15 {
			doorWidth = 15
		}
	}
	return doorWidth
}

// doorHeight computes the door height from the terminal height.
func (dv *DoorsView) doorHeight() int {
	doorHeight := 0
	if dv.height > 0 {
		doorHeight = int(float64(dv.height) * 0.5)
		if doorHeight < 10 {
			doorHeight = 10
		}
		if doorHeight > 25 {
			doorHeight = 25
		}
		doorHeight-- // reserve 1 row for threshold floor line (Story 48.2)
	}
	return doorHeight
}

// resolveActiveTheme resolves the theme for this render cycle, falling back
// to base or classic theme when dimensions are too small.
func (dv *DoorsView) resolveActiveTheme(doorWidth, doorHeight int) *themes.DoorTheme {
	activeTheme := dv.theme
	if activeTheme != nil && (doorWidth < activeTheme.MinWidth || (activeTheme.MinHeight > 0 && doorHeight < activeTheme.MinHeight)) {
		if activeTheme.Season != "" && dv.baseThemeName != "" && dv.themeRegistry != nil {
			if baseTheme, ok := dv.themeRegistry.Get(dv.baseThemeName); ok {
				activeTheme = baseTheme
			}
		} else if dv.themeRegistry != nil {
			if classic, ok := dv.themeRegistry.Get("classic"); ok {
				activeTheme = classic
			}
		}
	}
	return activeTheme
}

// View renders the doors view. Uses the sub-renderers and composes them
// with breakpoint-aware vertical padding.
func (dv *DoorsView) View() string {
	headerStr := dv.RenderHeader()

	if !dv.HasDoors() {
		var s strings.Builder
		s.WriteString(headerStr)
		s.WriteString("\n\n")
		s.WriteString(flashStyle.Render("All tasks done! Great work!"))
		s.WriteString("\n\nPress 'q' to quit.\n")
		return s.String()
	}

	// Only apply breakpoint degradation when height is known (> 0).
	// Height 0 means no WindowSizeMsg received yet — show full UI.
	bp := BreakpointStandard
	if dv.height > 0 {
		bp = layoutBreakpoint(dv.height)
	}

	// Compose header based on breakpoint.
	switch bp {
	case BreakpointMinimal:
		headerStr = "" // no header
	case BreakpointCompact:
		headerStr = dv.RenderCompactHeader()
	default:
		// full header already set
	}

	doorStr := dv.RenderDoors()
	statusStr := dv.RenderStatusSection()

	// Compose footer based on breakpoint.
	var footerStr string
	switch bp {
	case BreakpointMinimal:
		footerStr = "" // no footer
	case BreakpointCompact:
		footerStr = dv.RenderCompactFooter()
	default:
		footerStr = dv.RenderFooter()
	}

	// Merge doors + status into the content zone.
	content := doorStr
	if statusStr != "" {
		content += "\n\n" + statusStr
	}

	if dv.height <= 0 {
		return joinNonEmpty(headerStr, content, footerStr)
	}

	// Distribute vertical padding: 40% above doors, 60% below (perceptual centering).
	headerLines := strings.Count(headerStr, "\n") + 1
	if headerStr == "" {
		headerLines = 0
	}
	contentLines := strings.Count(content, "\n") + 1
	footerLines := strings.Count(footerStr, "\n") + 1
	if footerStr == "" {
		footerLines = 0
	}
	usedLines := headerLines + contentLines + footerLines
	remaining := dv.height - usedLines

	if remaining <= 0 {
		return joinNonEmpty(headerStr, content, footerStr)
	}

	topPad := remaining * 2 / 5     // 40%
	bottomPad := remaining - topPad // 60%

	var out strings.Builder
	if headerStr != "" {
		out.WriteString(headerStr)
	}
	if topPad > 0 {
		out.WriteString(strings.Repeat("\n", topPad))
	} else if headerStr != "" {
		out.WriteString("\n")
	}
	out.WriteString(content)
	if bottomPad > 0 {
		out.WriteString(strings.Repeat("\n", bottomPad))
	} else if footerStr != "" {
		out.WriteString("\n")
	}
	if footerStr != "" {
		out.WriteString(footerStr)
	}
	return out.String()
}

// connectionHealthAlert renders a warning line when sources need attention.
// Returns empty string when all connections are healthy.
func (dv *DoorsView) connectionHealthAlert() string {
	if dv.connMgr == nil {
		return ""
	}

	var lines []string

	// Existing: connections in Error or AuthExpired state.
	needsAttention := dv.connMgr.NeedsAttention()
	if len(needsAttention) == 1 {
		conn := needsAttention[0]
		issue := "error"
		if conn.State == connection.StateAuthExpired {
			issue = "auth expired"
		}
		lines = append(lines, connectionAlertStyle.Render(fmt.Sprintf("⚠ %s %s — :sources to fix", conn.Label, issue)))
	} else if len(needsAttention) > 1 {
		lines = append(lines, connectionAlertStyle.Render(fmt.Sprintf("⚠ %d sources need attention — :sources to fix", len(needsAttention))))
	}

	// Proactive: predictive health warnings (token expiry, rate limit, error streak).
	warnings := dv.connMgr.HealthWarnings()
	for _, w := range warnings {
		lines = append(lines, connectionAlertStyle.Render(w.String()))
	}

	return strings.Join(lines, "\n")
}

// animatedDoorStyle builds a lipgloss style with border color interpolated
// by spring emphasis (0.0 = unselected/dim, 1.0 = fully selected/bright).
func animatedDoorStyle(doorIndex int, emphasis float64, w, h int, usePerDoorColors bool) lipgloss.Style {
	// Clamp emphasis to [0, 1] for color interpolation (spring can overshoot)
	if emphasis < 0 {
		emphasis = 0
	}
	if emphasis > 1 {
		emphasis = 1
	}

	// Base color: per-door accent or default accent
	var baseTermColor lipgloss.TerminalColor
	if usePerDoorColors && doorIndex < len(doorColors) {
		baseTermColor = doorColors[doorIndex]
	} else {
		baseTermColor = colorAccent
	}

	// Parse colors for interpolation
	baseColor, _ := colorful.MakeColor(baseTermColor)
	dimColor, _ := colorful.MakeColor(lipgloss.Color("240")) // same as unselectedDoorStyle border
	brightColor, _ := colorful.MakeColor(colorDoorBright)

	// Interpolate: dim → base at emphasis 0→0.5, base → bright at 0.5→1.0
	var borderColor colorful.Color
	if emphasis <= 0.5 {
		t := emphasis * 2 // 0→1 over the 0→0.5 range
		borderColor = dimColor.BlendLab(baseColor, t)
	} else {
		t := (emphasis - 0.5) * 2 // 0→1 over the 0.5→1.0 range
		borderColor = baseColor.BlendLab(brightColor, t)
	}

	// Switch border style at midpoint
	border := lipgloss.RoundedBorder()
	if emphasis > 0.5 {
		border = lipgloss.DoubleBorder()
	}

	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color(borderColor.Hex())).
		Padding(1, 2).
		Width(w).
		Height(h).
		AlignVertical(lipgloss.Center)
}
