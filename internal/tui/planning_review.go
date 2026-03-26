package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ReviewDecision represents the user's choice for a reviewed task.
type ReviewDecision int

const (
	ReviewUndecided ReviewDecision = iota
	ReviewContinue
	ReviewDefer
	ReviewDrop
)

// reviewAutoAdvanceMsg fires after the empty-state pause to auto-complete.
type reviewAutoAdvanceMsg struct{}

// reviewTickMsg is sent every second to update elapsed time.
type reviewTickMsg time.Time

// ReviewCompleteMsg is sent when the review step finishes.
type ReviewCompleteMsg struct {
	Reviewed  int
	Continued int
	Deferred  int
	Dropped   int
}

// ReviewView displays incomplete tasks one at a time for review decisions.
type ReviewView struct {
	tasks         []*core.Task
	decisions     []ReviewDecision
	priorStatuses []core.TaskStatus // saved for undo on drop
	current       int
	width         int
	startTime     time.Time
	elapsed       time.Duration
	showHelp      bool
}

// NewReviewView creates a ReviewView for the given incomplete tasks.
func NewReviewView(tasks []*core.Task) *ReviewView {
	n := len(tasks)
	decisions := make([]ReviewDecision, n)
	priorStatuses := make([]core.TaskStatus, n)
	for i, t := range tasks {
		priorStatuses[i] = t.Status
	}
	return &ReviewView{
		tasks:         tasks,
		decisions:     decisions,
		priorStatuses: priorStatuses,
		startTime:     time.Now().UTC(),
	}
}

// SetWidth sets the terminal width for responsive rendering.
func (rv *ReviewView) SetWidth(w int) {
	rv.width = w
}

// Init starts the elapsed-time ticker. For empty state, returns a delayed auto-advance.
func (rv *ReviewView) Init() tea.Cmd {
	if len(rv.tasks) == 0 {
		return tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return reviewAutoAdvanceMsg{}
		})
	}
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return reviewTickMsg(t)
	})
}

// Update handles key input and timer ticks.
func (rv *ReviewView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case reviewAutoAdvanceMsg:
		return rv.completeCmd()
	case reviewTickMsg:
		rv.elapsed = time.Since(rv.startTime)
		return tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return reviewTickMsg(t)
		})
	case tea.KeyMsg:
		if rv.showHelp {
			return rv.handleHelpInput(msg)
		}
		return rv.handleKeyInput(msg)
	}
	return nil
}

func (rv *ReviewView) handleHelpInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "?", "esc":
		rv.showHelp = false
	}
	return nil
}

func (rv *ReviewView) handleKeyInput(msg tea.KeyMsg) tea.Cmd {
	if len(rv.tasks) == 0 {
		return nil
	}

	switch msg.Type {
	case tea.KeyEscape:
		return rv.completeCmd()
	case tea.KeyEnter:
		return rv.advance()
	case tea.KeyRunes:
		return rv.handleRune(msg.String())
	}
	return nil
}

func (rv *ReviewView) handleRune(key string) tea.Cmd {
	switch strings.ToLower(key) {
	case "c":
		rv.decisions[rv.current] = ReviewContinue
	case "d":
		rv.decisions[rv.current] = ReviewDefer
	case "x":
		rv.decisions[rv.current] = ReviewDrop
		// Transition task to deferred status
		_ = rv.tasks[rv.current].UpdateStatus(core.StatusDeferred)
	case "u":
		rv.undo()
	case "?":
		rv.showHelp = true
	}
	return nil
}

func (rv *ReviewView) undo() {
	if rv.decisions[rv.current] == ReviewUndecided {
		return
	}
	// If the task was dropped, restore its prior status
	if rv.decisions[rv.current] == ReviewDrop {
		rv.tasks[rv.current].Status = rv.priorStatuses[rv.current]
		rv.tasks[rv.current].UpdatedAt = time.Now().UTC()
	}
	rv.decisions[rv.current] = ReviewUndecided
}

func (rv *ReviewView) advance() tea.Cmd {
	if rv.decisions[rv.current] == ReviewUndecided {
		return nil
	}
	if rv.current >= len(rv.tasks)-1 {
		return rv.completeCmd()
	}
	rv.current++
	return nil
}

func (rv *ReviewView) completeCmd() tea.Cmd {
	var continued, deferred, dropped, reviewed int
	for _, d := range rv.decisions {
		switch d {
		case ReviewContinue:
			continued++
			reviewed++
		case ReviewDefer:
			deferred++
			reviewed++
		case ReviewDrop:
			dropped++
			reviewed++
		}
	}
	return func() tea.Msg {
		return ReviewCompleteMsg{
			Reviewed:  reviewed,
			Continued: continued,
			Deferred:  deferred,
			Dropped:   dropped,
		}
	}
}

// View renders the review step.
func (rv *ReviewView) View() string {
	var s strings.Builder

	if len(rv.tasks) == 0 {
		return rv.renderEmptyState()
	}

	if rv.showHelp {
		return rv.renderHelp()
	}

	rv.renderHeader(&s)
	rv.renderTaskPanel(&s)
	rv.renderActionKeys(&s)
	rv.renderFooter(&s)

	w := rv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}

func (rv *ReviewView) renderEmptyState() string {
	var s strings.Builder
	s.WriteString(reviewHeaderStyle.Render("Review Incomplete Tasks"))
	s.WriteString("\n\n")
	s.WriteString("No incomplete tasks from yesterday")
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Auto-advancing..."))

	w := rv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}

func (rv *ReviewView) renderHeader(s *strings.Builder) {
	// Elapsed time on the right
	elapsed := rv.formatElapsed()

	s.WriteString(reviewHeaderStyle.Render("Review Incomplete Tasks"))
	fmt.Fprintf(s, "  %s", helpStyle.Render(elapsed))
	s.WriteString("\n")
	fmt.Fprintf(s, "Task %d/%d\n\n", rv.current+1, len(rv.tasks))
}

func (rv *ReviewView) renderTaskPanel(s *strings.Builder) {
	task := rv.tasks[rv.current]

	// Task text
	s.WriteString(reviewTaskTextStyle.Render(task.Text))
	s.WriteString("\n")

	// Status
	statusColor := StatusColor(string(task.Status))
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	fmt.Fprintf(s, "Status: %s", statusStyle.Render(string(task.Status)))

	// Tags
	var tags []string
	if task.Type != "" {
		tags = append(tags, string(task.Type))
	}
	if task.Effort != "" {
		tags = append(tags, string(task.Effort))
	}
	if task.Location != "" {
		tags = append(tags, string(task.Location))
	}
	if len(tags) > 0 {
		fmt.Fprintf(s, "  %s", badgeStyle.Render(strings.Join(tags, " · ")))
	}
	s.WriteString("\n")

	// Decision indicator
	if rv.decisions[rv.current] != ReviewUndecided {
		s.WriteString("\n")
		fmt.Fprintf(s, "Decision: %s", rv.decisionLabel(rv.decisions[rv.current]))
	}
	s.WriteString("\n")
}

func (rv *ReviewView) renderActionKeys(s *strings.Builder) {
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("[C]ontinue  [D]efer  [X]Drop  [U]ndo  [?]Help"))
	s.WriteString("\n")
}

func (rv *ReviewView) renderFooter(s *strings.Builder) {
	s.WriteString(helpStyle.Render("Step 1/3 — Review  |  Esc to skip"))
}

func (rv *ReviewView) renderHelp() string {
	var s strings.Builder
	s.WriteString(reviewHeaderStyle.Render("Review Help"))
	s.WriteString("\n\n")
	s.WriteString("  C — Continue: keep task for focus consideration\n")
	s.WriteString("  D — Defer: leave task in pool without special treatment\n")
	s.WriteString("  X — Drop: move task to deferred status\n")
	s.WriteString("  U — Undo: undo current decision before advancing\n")
	s.WriteString("  Enter — advance to next task\n")
	s.WriteString("  Esc — skip remaining reviews\n")
	s.WriteString("  ? — toggle this help\n")
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press ? or Esc to close"))

	w := rv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}

func (rv *ReviewView) formatElapsed() string {
	m := int(rv.elapsed.Minutes())
	sec := int(rv.elapsed.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, sec)
}

func (rv *ReviewView) decisionLabel(d ReviewDecision) string {
	switch d {
	case ReviewContinue:
		return reviewContinueStyle.Render("Continue")
	case ReviewDefer:
		return reviewDeferStyle.Render("Defer")
	case ReviewDrop:
		return reviewDropStyle.Render("Drop")
	default:
		return ""
	}
}

// Styles for the review view
var (
	reviewHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	reviewTaskTextStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("255"))

	reviewContinueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")).
				Bold(true)

	reviewDeferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)

	reviewDropStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)
