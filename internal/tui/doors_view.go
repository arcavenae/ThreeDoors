package tui

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/tasks"
	"github.com/charmbracelet/lipgloss"
)

// typeIcon returns the emoji icon for a task type.
func typeIcon(t tasks.TaskType) string {
	switch t {
	case tasks.TypeCreative:
		return "🎨"
	case tasks.TypeAdministrative:
		return "📋"
	case tasks.TypeTechnical:
		return "🔧"
	case tasks.TypePhysical:
		return "💪"
	default:
		return ""
	}
}

// categoryBadge builds a compact badge string for a task's categories.
func categoryBadge(task *tasks.Task) string {
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
	pool              *tasks.TaskPool
	currentDoors      []*tasks.Task
	selectedDoorIndex int
	completedCount    int
	width             int
	tracker           *tasks.SessionTracker
	greeting          string
	footerMessage     string
	avoidanceMap      map[string]int // task text → bypass count (TimesBypassed)
	avoidanceShown    map[string]int // task text → shown count (TimesShown)
	patternAnalyzer   *tasks.PatternAnalyzer
	completionCounter *tasks.CompletionCounter
	syncTracker       *tasks.SyncStatusTracker
	timeContext       *tasks.TimeContext
}

// NewDoorsView creates a new DoorsView.
func NewDoorsView(pool *tasks.TaskPool, tracker *tasks.SessionTracker) *DoorsView {
	dv := &DoorsView{
		pool:              pool,
		selectedDoorIndex: -1,
		tracker:           tracker,
		greeting:          pickGreeting(-1),
		footerMessage:     pickFooterMessage(-1),
		avoidanceMap:      make(map[string]int),
		avoidanceShown:    make(map[string]int),
	}
	dv.RefreshDoors()
	return dv
}

// SetAvoidanceData populates the avoidance map from a pattern report.
func (dv *DoorsView) SetAvoidanceData(report *tasks.PatternReport) {
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
func (dv *DoorsView) SetInsightsData(pa *tasks.PatternAnalyzer, cc *tasks.CompletionCounter) {
	dv.patternAnalyzer = pa
	dv.completionCounter = cc
}

// SetSyncTracker sets the sync status tracker for displaying provider sync state.
func (dv *DoorsView) SetSyncTracker(tracker *tasks.SyncStatusTracker) {
	dv.syncTracker = tracker
}

// SetTimeContext sets the calendar time context for time-aware door selection and display.
func (dv *DoorsView) SetTimeContext(tc *tasks.TimeContext) {
	dv.timeContext = tc
}

// TimeContext returns the current time context (for testing).
func (dv *DoorsView) TimeContext() *tasks.TimeContext {
	return dv.timeContext
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
// Uses time-contextual selection when calendar data is available.
func (dv *DoorsView) RefreshDoors() {
	if dv.timeContext != nil && dv.timeContext.HasCalendar {
		dv.currentDoors = tasks.SelectDoorsWithTimeContext(dv.pool, 3, dv.timeContext)
	} else {
		dv.currentDoors = tasks.SelectDoors(dv.pool, 3)
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

// View renders the doors view.
func (dv *DoorsView) View() string {
	s := strings.Builder{}
	s.WriteString(headerStyle.Render("ThreeDoors - Technical Demo"))
	s.WriteString("\n")
	s.WriteString(greetingStyle.Render(dv.greeting))
	s.WriteString("\n")
	if dv.patternAnalyzer != nil && dv.completionCounter != nil {
		multiGreeting := tasks.FormatMultiDimensionalGreeting(dv.patternAnalyzer, dv.completionCounter)
		if multiGreeting != "" {
			s.WriteString(greetingStyle.Render(multiGreeting))
			s.WriteString("\n")
		}
	}
	if timeStr := tasks.FormatTimeContext(dv.timeContext); timeStr != "" {
		s.WriteString(badgeStyle.Render(timeStr))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	if len(dv.currentDoors) == 0 {
		s.WriteString(flashStyle.Render("All tasks done! Great work!"))
		s.WriteString("\n\nPress 'q' to quit.\n")
		return s.String()
	}

	usePerDoorColors := dv.width >= 60

	doorWidth := 30
	if dv.width > 20 {
		doorWidth = (dv.width - 6) / 3
		if doorWidth < 15 {
			doorWidth = 15
		}
	}

	var renderedDoors []string
	for i, task := range dv.currentDoors {
		content := task.Text

		// Category badges
		badge := categoryBadge(task)
		if badge != "" {
			content = content + "\n" + badgeStyle.Render(badge)
		}

		// Avoidance indicator — show when bypassed 5+ times, display total times shown
		if bypassCount, ok := dv.avoidanceMap[task.Text]; ok && bypassCount >= 5 {
			shownCount := dv.avoidanceShown[task.Text]
			if shownCount == 0 {
				shownCount = bypassCount
			}
			avoidStyle := lipgloss.NewStyle().Faint(true)
			content = content + "\n" + avoidStyle.Render(fmt.Sprintf("Seen %d times", shownCount))
		}

		statusIndicator := lipgloss.NewStyle().
			Foreground(StatusColor(string(task.Status))).
			Render(fmt.Sprintf("[%s]", task.Status))
		content = statusIndicator + "\n\n" + content

		var style lipgloss.Style
		if i == dv.selectedDoorIndex {
			style = selectedDoorStyle.Width(doorWidth)
		} else if usePerDoorColors && i < len(doorColors) {
			style = doorStyle.BorderForeground(doorColors[i]).Width(doorWidth)
		} else {
			style = doorStyle.Width(doorWidth)
		}
		renderedDoors = append(renderedDoors, style.Render(content))
	}

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedDoors...))

	if dv.completedCount > 0 {
		fmt.Fprintf(&s, "\n\nCompleted this session: %d", dv.completedCount)
	}

	// Sync status bar
	if syncBar := RenderSyncStatusBar(dv.syncTracker); syncBar != "" {
		s.WriteString("\n\n")
		s.WriteString(syncBar)
	}

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("a/left, w/up, d/right to select | s/down to re-roll | Enter to open | N feedback | / search | M mood | q quit"))
	s.WriteString("\n")
	s.WriteString(greetingStyle.Render(dv.footerMessage))

	return s.String()
}
