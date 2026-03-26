package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmCompleteMsg is sent when the user confirms the planning session.
type ConfirmCompleteMsg struct {
	FocusTasks     []*core.Task
	EnergyLevel    core.EnergyLevel
	EnergyOverride bool
}

// ConfirmCancelMsg is sent when the user goes back to select step.
type ConfirmCancelMsg struct{}

// ConfirmView shows the final planning session summary before confirmation.
type ConfirmView struct {
	focusTasks     []*core.Task
	reviewMetrics  ReviewCompleteMsg
	energyLevel    core.EnergyLevel
	energyOverride bool
	startTime      time.Time
	elapsed        time.Duration
	width          int
	height         int
	nudgeShown     bool
}

// NewConfirmView creates a ConfirmView with the session results.
func NewConfirmView(
	focusTasks []*core.Task,
	reviewMetrics ReviewCompleteMsg,
	energyLevel core.EnergyLevel,
	energyOverride bool,
	sessionStart time.Time,
) *ConfirmView {
	return &ConfirmView{
		focusTasks:     focusTasks,
		reviewMetrics:  reviewMetrics,
		energyLevel:    energyLevel,
		energyOverride: energyOverride,
		startTime:      sessionStart,
	}
}

// SetWidth sets the terminal width.
func (cv *ConfirmView) SetWidth(w int) {
	cv.width = w
}

// SetHeight sets the terminal height.
func (cv *ConfirmView) SetHeight(h int) {
	cv.height = h
}

// confirmTickMsg fires every second to update elapsed time.
type confirmTickMsg time.Time

// confirmNudgeMsg fires after 10 minutes as a soft time nudge.
type confirmNudgeMsg struct{}

// Init starts the elapsed-time ticker and the 10-minute nudge timer.
func (cv *ConfirmView) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return confirmTickMsg(t)
		}),
		tea.Tick(10*time.Minute, func(_ time.Time) tea.Msg {
			return confirmNudgeMsg{}
		}),
	)
}

// Update handles key input and timer ticks.
func (cv *ConfirmView) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case confirmTickMsg:
		cv.elapsed = time.Since(cv.startTime)
		return tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return confirmTickMsg(t)
		})
	case confirmNudgeMsg:
		cv.nudgeShown = true
		return nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.Type {
	case tea.KeyEnter:
		return cv.confirmCmd()
	case tea.KeyEscape:
		return func() tea.Msg { return ConfirmCancelMsg{} }
	}
	return nil
}

func (cv *ConfirmView) confirmCmd() tea.Cmd {
	tasks := cv.focusTasks
	energy := cv.energyLevel
	override := cv.energyOverride
	return func() tea.Msg {
		return ConfirmCompleteMsg{
			FocusTasks:     tasks,
			EnergyLevel:    energy,
			EnergyOverride: override,
		}
	}
}

// View renders the confirm step.
func (cv *ConfirmView) View() string {
	var s strings.Builder

	cv.renderHeader(&s)
	cv.renderFocusTasks(&s)
	cv.renderSessionSummary(&s)
	cv.renderNudge(&s)
	cv.renderFooter(&s)

	w := cv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}

func (cv *ConfirmView) renderHeader(s *strings.Builder) {
	elapsed := cv.formatElapsed()
	s.WriteString(confirmHeaderStyle.Render("Confirm Focus"))
	fmt.Fprintf(s, "  %s\n\n", helpStyle.Render(elapsed))
}

func (cv *ConfirmView) renderFocusTasks(s *strings.Builder) {
	if len(cv.focusTasks) == 0 {
		s.WriteString("No focus tasks selected.\n\n")
		return
	}

	s.WriteString(confirmSubheaderStyle.Render("Today's Focus"))
	s.WriteString("\n")
	for i, t := range cv.focusTasks {
		displayText := core.RemoveFocusTagFromText(t.Text)
		fmt.Fprintf(s, "  %d. %s\n", i+1, displayText)
	}
	s.WriteString("\n")
}

func (cv *ConfirmView) renderSessionSummary(s *strings.Builder) {
	s.WriteString(confirmSubheaderStyle.Render("Session Summary"))
	s.WriteString("\n")

	rm := cv.reviewMetrics
	fmt.Fprintf(s, "  Tasks reviewed: %d\n", rm.Reviewed)
	if rm.Continued > 0 {
		fmt.Fprintf(s, "    Continued: %d\n", rm.Continued)
	}
	if rm.Deferred > 0 {
		fmt.Fprintf(s, "    Deferred:  %d\n", rm.Deferred)
	}
	if rm.Dropped > 0 {
		fmt.Fprintf(s, "    Dropped:   %d\n", rm.Dropped)
	}

	energyStr := core.EnergyDisplayString(cv.energyLevel, time.Now().UTC(), cv.energyOverride)
	fmt.Fprintf(s, "  %s\n", energyStr)

	fmt.Fprintf(s, "  Focus tasks: %d\n", len(cv.focusTasks))
	fmt.Fprintf(s, "  Elapsed: %s\n\n", cv.formatElapsed())
}

func (cv *ConfirmView) renderNudge(s *strings.Builder) {
	if cv.nudgeShown {
		s.WriteString(confirmNudgeStyle.Render("Planning is taking a while — consider wrapping up!"))
		s.WriteString("\n\n")
	}
}

func (cv *ConfirmView) renderFooter(s *strings.Builder) {
	s.WriteString(helpStyle.Render("[Enter]Confirm  [Esc]Back to Select"))
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Step 3/3 — Confirm"))
}

func (cv *ConfirmView) formatElapsed() string {
	m := int(cv.elapsed.Minutes())
	sec := int(cv.elapsed.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, sec)
}

// Styles for the confirm view
var (
	confirmHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	confirmSubheaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("255"))

	confirmNudgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Italic(true)
)
