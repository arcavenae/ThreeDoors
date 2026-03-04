package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/intelligence"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetailViewMode tracks the sub-view within the detail view.
type DetailViewMode int

const (
	DetailModeView DetailViewMode = iota
	DetailModeBlockerInput
	DetailModeExpandInput
	DetailModeLinkSelect
	DetailModeLinkBrowse
)

// DetailView displays full task details and status action menu.
type DetailView struct {
	task              *core.Task
	mode              DetailViewMode
	blockerInput      string
	expandInput       string
	width             int
	tracker           *core.SessionTracker
	enrichDB          *enrichment.DB
	pool              *core.TaskPool
	agentService      *intelligence.AgentService
	crossRefs         []enrichment.CrossReference
	linkCandidates    []*core.Task
	linkSelectedIndex int
	linkBrowseIndex   int
}

// NewDetailView creates a detail view for the given task.
func NewDetailView(task *core.Task, tracker *core.SessionTracker, edb *enrichment.DB, pool *core.TaskPool) *DetailView {
	if tracker != nil {
		tracker.RecordDetailView()
	}
	dv := &DetailView{
		task:     task,
		mode:     DetailModeView,
		tracker:  tracker,
		enrichDB: edb,
		pool:     pool,
	}
	dv.loadCrossRefs()
	return dv
}

// loadCrossRefs fetches cross-references for the current task from the enrichment DB.
func (dv *DetailView) loadCrossRefs() {
	if dv.enrichDB == nil || dv.task == nil {
		return
	}
	refs, err := dv.enrichDB.GetCrossReferences(dv.task.ID)
	if err != nil {
		return
	}
	dv.crossRefs = refs
}

// SetAgentService sets the agent service for LLM task decomposition.
func (dv *DetailView) SetAgentService(svc *intelligence.AgentService) {
	dv.agentService = svc
}

// SetWidth sets the terminal width.
func (dv *DetailView) SetWidth(w int) {
	dv.width = w
}

// Update handles key input in the detail view.
func (dv *DetailView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch dv.mode {
		case DetailModeBlockerInput:
			return dv.handleBlockerInput(msg)
		case DetailModeExpandInput:
			return dv.handleExpandInput(msg)
		case DetailModeLinkSelect:
			return dv.handleLinkSelect(msg)
		case DetailModeLinkBrowse:
			return dv.handleLinkBrowse(msg)
		default:
			return dv.handleDetailKeys(msg)
		}
	}
	return nil
}

func (dv *DetailView) handleDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "c", "C":
		if err := dv.task.UpdateStatus(core.StatusComplete); err != nil {
			return func() tea.Msg { return FlashMsg{Text: "Cannot complete: " + err.Error()} }
		}
		if dv.tracker != nil {
			dv.tracker.RecordStatusChange()
			dv.tracker.RecordTaskCompleted()
		}
		return func() tea.Msg { return TaskCompletedMsg{Task: dv.task} }
	case "b", "B":
		if core.IsValidTransition(dv.task.Status, core.StatusBlocked) {
			dv.mode = DetailModeBlockerInput
			dv.blockerInput = ""
		}
	case "i", "I":
		if err := dv.task.UpdateStatus(core.StatusInProgress); err != nil {
			return func() tea.Msg { return FlashMsg{Text: "Cannot change status: " + err.Error()} }
		}
		if dv.tracker != nil {
			dv.tracker.RecordStatusChange()
		}
		return func() tea.Msg { return TaskUpdatedMsg{Task: dv.task} }
	case "e", "E":
		dv.mode = DetailModeExpandInput
		dv.expandInput = ""
	case "f", "F":
		forked := core.NewTask(dv.task.Text)
		return func() tea.Msg { return TaskAddedMsg{Task: forked} }
	case "p", "P":
		// Procrastinate: just return to doors (task stays in pool)
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "r", "R":
		// Rework: keep in pool, just return
		return func() tea.Msg { return ReturnToDoorsMsg{} }
	case "m", "M":
		return func() tea.Msg { return ShowMoodMsg{} }
	case "l", "L":
		if dv.enrichDB == nil || dv.pool == nil {
			return func() tea.Msg { return FlashMsg{Text: "Linking not available"} }
		}
		dv.linkCandidates = dv.buildLinkCandidates()
		if len(dv.linkCandidates) == 0 {
			return func() tea.Msg { return FlashMsg{Text: "No tasks available to link"} }
		}
		dv.linkSelectedIndex = 0
		dv.mode = DetailModeLinkSelect
	case "x", "X":
		if len(dv.crossRefs) > 0 {
			dv.linkBrowseIndex = 0
			dv.mode = DetailModeLinkBrowse
		}
	case "g", "G":
		if dv.agentService == nil {
			return func() tea.Msg { return FlashMsg{Text: "LLM not configured"} }
		}
		desc := strings.TrimSpace(dv.task.Text)
		if desc == "" {
			return func() tea.Msg { return FlashMsg{Text: "Task has no description to decompose"} }
		}
		return func() tea.Msg {
			return DecomposeStartMsg{TaskID: dv.task.ID, TaskDescription: desc}
		}
	}
	return nil
}

func (dv *DetailView) handleBlockerInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		if err := dv.task.UpdateStatus(core.StatusBlocked); err == nil {
			if dv.blockerInput != "" {
				_ = dv.task.SetBlocker(dv.blockerInput)
			}
			if dv.tracker != nil {
				dv.tracker.RecordStatusChange()
			}
			dv.mode = DetailModeView
			return func() tea.Msg { return TaskUpdatedMsg{Task: dv.task} }
		}
		dv.mode = DetailModeView
	case "esc":
		dv.mode = DetailModeView
		dv.blockerInput = ""
	case "backspace":
		if len(dv.blockerInput) > 0 {
			dv.blockerInput = dv.blockerInput[:len(dv.blockerInput)-1]
		}
	default:
		if len(msg.String()) == 1 && len(dv.blockerInput) < 200 {
			dv.blockerInput += msg.String()
		}
	}
	return nil
}

func (dv *DetailView) handleExpandInput(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(dv.expandInput)
		if text == "" {
			return func() tea.Msg { return FlashMsg{Text: "Subtask text cannot be empty"} }
		}
		parentTask := dv.task
		dv.mode = DetailModeView
		dv.expandInput = ""
		return func() tea.Msg {
			return ExpandTaskMsg{ParentTask: parentTask, NewTaskText: text}
		}
	case "esc":
		dv.mode = DetailModeView
		dv.expandInput = ""
	case "backspace":
		if len(dv.expandInput) > 0 {
			dv.expandInput = dv.expandInput[:len(dv.expandInput)-1]
		}
	default:
		if len(msg.String()) == 1 && len(dv.expandInput) < 500 {
			dv.expandInput += msg.String()
		}
	}
	return nil
}

// buildLinkCandidates returns tasks that can be linked to (excluding current task and already-linked tasks).
func (dv *DetailView) buildLinkCandidates() []*core.Task {
	linkedIDs := make(map[string]bool)
	linkedIDs[dv.task.ID] = true
	for _, ref := range dv.crossRefs {
		linkedIDs[ref.SourceTaskID] = true
		linkedIDs[ref.TargetTaskID] = true
	}

	var candidates []*core.Task
	for _, t := range dv.pool.GetAllTasks() {
		if !linkedIDs[t.ID] {
			candidates = append(candidates, t)
		}
	}
	return candidates
}

func (dv *DetailView) handleLinkSelect(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		dv.mode = DetailModeView
		dv.linkCandidates = nil
	case "up", "k":
		if dv.linkSelectedIndex > 0 {
			dv.linkSelectedIndex--
		}
	case "down", "j":
		if dv.linkSelectedIndex < len(dv.linkCandidates)-1 {
			dv.linkSelectedIndex++
		}
	case "enter":
		if dv.linkSelectedIndex >= 0 && dv.linkSelectedIndex < len(dv.linkCandidates) {
			target := dv.linkCandidates[dv.linkSelectedIndex]
			ref := &enrichment.CrossReference{
				SourceTaskID: dv.task.ID,
				TargetTaskID: target.ID,
				SourceSystem: "local",
				Relationship: "related",
			}
			if err := dv.enrichDB.AddCrossReference(ref); err != nil {
				dv.mode = DetailModeView
				dv.linkCandidates = nil
				return func() tea.Msg { return FlashMsg{Text: "Link failed: " + err.Error()} }
			}
			dv.loadCrossRefs()
			dv.mode = DetailModeView
			dv.linkCandidates = nil
			return func() tea.Msg { return FlashMsg{Text: "Linked!"} }
		}
	}
	return nil
}

func (dv *DetailView) handleLinkBrowse(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		dv.mode = DetailModeView
	case "up", "k":
		if dv.linkBrowseIndex > 0 {
			dv.linkBrowseIndex--
		}
	case "down", "j":
		if dv.linkBrowseIndex < len(dv.crossRefs)-1 {
			dv.linkBrowseIndex++
		}
	case "enter":
		if dv.linkBrowseIndex >= 0 && dv.linkBrowseIndex < len(dv.crossRefs) {
			ref := dv.crossRefs[dv.linkBrowseIndex]
			targetID := ref.TargetTaskID
			if targetID == dv.task.ID {
				targetID = ref.SourceTaskID
			}
			if dv.pool != nil {
				if target := dv.pool.GetTask(targetID); target != nil {
					return func() tea.Msg { return NavigateToLinkedMsg{Task: target} }
				}
			}
			return func() tea.Msg { return FlashMsg{Text: "Linked task not found in pool"} }
		}
	case "u", "U":
		if dv.linkBrowseIndex >= 0 && dv.linkBrowseIndex < len(dv.crossRefs) {
			ref := dv.crossRefs[dv.linkBrowseIndex]
			if err := dv.enrichDB.DeleteCrossReference(ref.ID); err != nil {
				return func() tea.Msg { return FlashMsg{Text: "Unlink failed: " + err.Error()} }
			}
			dv.loadCrossRefs()
			if len(dv.crossRefs) == 0 {
				dv.mode = DetailModeView
			} else if dv.linkBrowseIndex >= len(dv.crossRefs) {
				dv.linkBrowseIndex = len(dv.crossRefs) - 1
			}
			return func() tea.Msg { return FlashMsg{Text: "Unlinked"} }
		}
	}
	return nil
}

// resolveTaskText looks up task text by ID from the pool.
func (dv *DetailView) resolveTaskText(taskID string) string {
	if dv.pool == nil {
		return taskID
	}
	if t := dv.pool.GetTask(taskID); t != nil {
		return t.Text
	}
	return taskID[:8] + "..."
}

// View renders the detail view.
func (dv *DetailView) View() string {
	s := strings.Builder{}

	w := dv.width - 6
	if w < 40 {
		w = 40
	}

	s.WriteString(headerStyle.Render("TASK DETAILS"))
	s.WriteString("\n\n")

	statusColor := StatusColor(string(dv.task.Status))
	statusStyle := lipgloss.NewStyle().Foreground(statusColor).Bold(true)
	fmt.Fprintf(&s, "Status: %s\n", statusStyle.Render(string(dv.task.Status)))
	fmt.Fprintf(&s, "Created: %s\n", dv.task.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Fprintf(&s, "Updated: %s\n", dv.task.UpdatedAt.Format("2006-01-02 15:04"))

	if dv.task.SourceProvider != "" {
		fmt.Fprintf(&s, "Source: %s\n", SourceBadge(dv.task.SourceProvider))
	}

	if dv.task.Blocker != "" {
		blockerStyle := lipgloss.NewStyle().Foreground(colorBlocked)
		fmt.Fprintf(&s, "Blocker: %s\n", blockerStyle.Render(dv.task.Blocker))
	}

	s.WriteString("\n")
	s.WriteString(dv.task.Text)
	s.WriteString("\n")

	if dv.task.Context != "" {
		contextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Italic(true)
		fmt.Fprintf(&s, "\nWhy: %s\n", contextStyle.Render(dv.task.Context))
	}

	if len(dv.task.Notes) > 0 {
		s.WriteString("\nNotes:\n")
		for _, note := range dv.task.Notes {
			fmt.Fprintf(&s, "  [%s] %s\n", note.Timestamp.Format("15:04"), note.Text)
		}
	}

	// Show cross-references (linked tasks)
	if len(dv.crossRefs) > 0 {
		linkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
		fmt.Fprintf(&s, "\n%s (%d):\n", linkStyle.Render("Linked"), len(dv.crossRefs))
		for i, ref := range dv.crossRefs {
			linkedID := ref.TargetTaskID
			if linkedID == dv.task.ID {
				linkedID = ref.SourceTaskID
			}
			text := dv.resolveTaskText(linkedID)
			prefix := "  "
			if dv.mode == DetailModeLinkBrowse && i == dv.linkBrowseIndex {
				prefix = "> "
			}
			fmt.Fprintf(&s, "%s[%s] %s\n", prefix, ref.Relationship, text)
		}
	}

	s.WriteString("\n")
	s.WriteString(separatorStyle.Render("─────────────────────────────────"))
	s.WriteString("\n\n")

	switch dv.mode {
	case DetailModeBlockerInput:
		s.WriteString("Blocker reason (Enter to submit, Esc to cancel):\n")
		s.WriteString("> " + dv.blockerInput + "_\n")
	case DetailModeExpandInput:
		s.WriteString("New subtask text (Enter to submit, Esc to cancel):\n")
		s.WriteString("> " + dv.expandInput + "_\n")
	case DetailModeLinkSelect:
		s.WriteString("Select task to link (Enter to link, Esc to cancel):\n\n")
		for i, t := range dv.linkCandidates {
			prefix := "  "
			if i == dv.linkSelectedIndex {
				prefix = "> "
			}
			fmt.Fprintf(&s, "%s%s\n", prefix, t.Text)
		}
	case DetailModeLinkBrowse:
		s.WriteString(helpStyle.Render("[Enter] Navigate [U]nlink [Esc] Back"))
	default:
		linkHint := ""
		if dv.enrichDB != nil {
			linkHint = " [L]ink"
		}
		browseHint := ""
		if len(dv.crossRefs) > 0 {
			browseHint = " [X]refs"
		}
		decomposeHint := ""
		if dv.agentService != nil {
			decomposeHint = " [G]enerate stories"
		}
		s.WriteString(helpStyle.Render("[C]omplete [B]locked [I]n-progress [E]xpand [F]ork [P]rocrastinate [R]ework [M]ood" + linkHint + browseHint + decomposeHint + " [Esc]Back"))
	}

	return detailBorder.Width(w).Render(s.String())
}
