package tui

import (
	"fmt"
	"strings"

	"github.com/arcavenae/ThreeDoors/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

// NextStepOption represents a single selectable option in the next-steps view.
type NextStepOption struct {
	Label  string
	Action string // action identifier dispatched via NextStepSelectedMsg
}

// NextStepsView displays contextual next-step options after key actions.
type NextStepsView struct {
	context string // "completed" or "added"
	header  string
	options []NextStepOption
	width   int
	height  int
}

// NewNextStepsView creates a new NextStepsView with context-aware options.
func NewNextStepsView(context string, pool *core.TaskPool, cc *core.CompletionCounter) *NextStepsView {
	nv := &NextStepsView{context: context}
	nv.generateOptions(context, pool, cc)
	return nv
}

func (nv *NextStepsView) generateOptions(context string, pool *core.TaskPool, cc *core.CompletionCounter) {
	switch context {
	case "completed":
		nv.header = pickCompletionHeader(cc)
		nv.options = generateCompletionOptions(pool)
	case "added":
		nv.header = "Task captured! What's next?"
		nv.options = generateAddedOptions(pool)
	default:
		nv.header = "What would you like to do next?"
		nv.options = defaultOptions()
	}
}

func pickCompletionHeader(cc *core.CompletionCounter) string {
	if cc != nil {
		today := cc.GetTodayCount()
		if today >= 5 {
			return "You're on fire! 5+ tasks today. What's next?"
		}
		if today >= 3 {
			return "Great momentum! Keep it rolling?"
		}
	}
	return "Nice work! What would you like to do next?"
}

func generateCompletionOptions(pool *core.TaskPool) []NextStepOption {
	options := []NextStepOption{
		{Label: "Open another door", Action: "doors"},
	}

	if pool.Count() > 0 {
		blocked := pool.GetTasksByStatus(core.StatusBlocked)
		if len(blocked) > 0 {
			options = append(options, NextStepOption{
				Label:  fmt.Sprintf("Review blocked tasks (%d)", len(blocked)),
				Action: "search",
			})
		}
	}

	options = append(options, NextStepOption{Label: "Add a new task", Action: "add"})
	options = append(options, NextStepOption{Label: "Check stats", Action: "stats"})
	options = append(options, NextStepOption{Label: "Log your mood", Action: "mood"})

	return options
}

func generateAddedOptions(pool *core.TaskPool) []NextStepOption {
	options := []NextStepOption{
		{Label: "Back to doors", Action: "doors"},
		{Label: "Add another task", Action: "add"},
	}

	if pool.Count() > 0 {
		options = append(options, NextStepOption{Label: "Search tasks", Action: "search"})
	}

	return options
}

func defaultOptions() []NextStepOption {
	return []NextStepOption{
		{Label: "Back to doors", Action: "doors"},
		{Label: "Add a new task", Action: "add"},
		{Label: "Log your mood", Action: "mood"},
	}
}

// SetWidth sets the terminal width.
func (nv *NextStepsView) SetWidth(w int) {
	nv.width = w
}

// SetHeight sets the terminal height for layout decisions.
func (nv *NextStepsView) SetHeight(h int) {
	nv.height = h
}

// Update handles key input for next-step selection.
func (nv *NextStepsView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return func() tea.Msg { return NextStepDismissedMsg{} }
		default:
			// Handle number key selection (1-based)
			if len(msg.String()) == 1 {
				key := msg.String()[0]
				if key >= '1' && key <= '9' {
					idx := int(key - '1')
					if idx < len(nv.options) {
						action := nv.options[idx].Action
						return func() tea.Msg { return NextStepSelectedMsg{Action: action} }
					}
				}
			}
		}
	}
	return nil
}

// View renders the next-steps view.
func (nv *NextStepsView) View() string {
	s := strings.Builder{}
	s.WriteString(nextStepsHeaderStyle.Render(nv.header))
	s.WriteString("\n\n")

	for i, opt := range nv.options {
		line := fmt.Sprintf("  %d. %s", i+1, opt.Label)
		s.WriteString(nextStepsOptionStyle.Render(line))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render(fmt.Sprintf("Press 1-%d to select | Esc to return to doors", len(nv.options))))

	w := nv.width - 6
	if w < 40 {
		w = 40
	}
	return detailBorder.Width(w).Render(s.String())
}
