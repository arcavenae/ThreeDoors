package tui

import (
	"fmt"
	"strings"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/enrichment"
	"github.com/arcaven/ThreeDoors/internal/intelligence"
	"github.com/arcaven/ThreeDoors/internal/intelligence/services"
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
	DetailModeDispatchConfirm
	DetailModeDepBrowse
	DetailModeDepAdd
	DetailModeEnrichLoading
	DetailModeEnrichResult
)

// DetailView displays full task details and status action menu.
type DetailView struct {
	task                *core.Task
	mode                DetailViewMode
	blockerInput        string
	expandInput         string
	width               int
	height              int
	tracker             *core.SessionTracker
	enrichDB            *enrichment.DB
	pool                *core.TaskPool
	agentService        *intelligence.AgentService
	crossRefs           []enrichment.CrossReference
	linkCandidates      []*core.Task
	linkSelectedIndex   int
	linkBrowseIndex     int
	isDuplicate         bool
	dedupStore          *core.DedupStore
	duplicatePair       *core.DuplicatePair
	devDispatchEnabled  bool
	dispatcherAvailable bool
	depBrowseIndex      int
	depAddCandidates    []*core.Task
	depAddSelectedIndex int
	hintEnabled         bool
	enricher            *services.TaskEnricher
	enrichResult        *services.EnrichedTask
	expandSubtaskCount  int
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

// SetDuplicateInfo sets the duplicate detection state for the current task.
func (dv *DetailView) SetDuplicateInfo(isDup bool, store *core.DedupStore, pair *core.DuplicatePair) {
	dv.isDuplicate = isDup
	dv.dedupStore = store
	dv.duplicatePair = pair
}

// SetDevDispatchInfo sets whether dev dispatch is enabled and available.
func (dv *DetailView) SetDevDispatchInfo(enabled, available bool) {
	dv.devDispatchEnabled = enabled
	dv.dispatcherAvailable = available
}

// SetEnricher sets the task enricher for LLM task enrichment.
func (dv *DetailView) SetEnricher(enricher *services.TaskEnricher) {
	dv.enricher = enricher
}

// SetInlineHints sets the inline hint display state.
func (dv *DetailView) SetInlineHints(enabled bool) {
	dv.hintEnabled = enabled
}

// SetWidth sets the terminal width.
func (dv *DetailView) SetWidth(w int) {
	dv.width = w
}

// SetHeight sets the terminal height for layout decisions.
func (dv *DetailView) SetHeight(h int) {
	dv.height = h
}

// Update handles key input in the detail view.
func (dv *DetailView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case EnrichResultMsg:
		if msg.Err != nil {
			dv.mode = DetailModeView
			errText := msg.Err.Error()
			return func() tea.Msg { return FlashMsg{Text: "Enrich failed: " + errText} }
		}
		dv.enrichResult = msg.Result
		dv.mode = DetailModeEnrichResult
		return nil
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
		case DetailModeDispatchConfirm:
			return dv.handleDispatchConfirm(msg)
		case DetailModeDepBrowse:
			return dv.handleDepBrowse(msg)
		case DetailModeDepAdd:
			return dv.handleDepAdd(msg)
		case DetailModeEnrichLoading:
			return nil // loading state, ignore keys
		case DetailModeEnrichResult:
			return dv.handleEnrichResult(msg)
		default:
			return dv.handleDetailKeys(msg)
		}
	}
	return nil
}

func (dv *DetailView) handleDetailKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q", " ", "enter":
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
		dv.expandSubtaskCount = 0
	case "f", "F":
		original := dv.task
		variant := core.ForkTask(original)
		return func() tea.Msg { return TaskForkedMsg{Original: original, Variant: variant} }
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
		if dv.devDispatchEnabled && dv.dispatcherAvailable {
			if dv.task.DevDispatch != nil && dv.task.DevDispatch.Queued {
				return func() tea.Msg { return FlashMsg{Text: "Task already dispatched"} }
			}
			dv.mode = DetailModeDispatchConfirm
			return nil
		}
		if len(dv.crossRefs) > 0 {
			dv.linkBrowseIndex = 0
			dv.mode = DetailModeLinkBrowse
		}
	case "z", "Z":
		task := dv.task
		return func() tea.Msg { return ShowSnoozeMsg{Task: task} }
	case "g", "G":
		desc := strings.TrimSpace(dv.task.Text)
		if desc == "" {
			return func() tea.Msg { return FlashMsg{Text: "Task has no description to break down"} }
		}
		task := dv.task
		return func() tea.Msg {
			return BreakdownStartMsg{Task: task}
		}
	case "+":
		if dv.pool == nil {
			return func() tea.Msg { return FlashMsg{Text: "Dependencies not available"} }
		}
		dv.depAddCandidates = dv.buildDepAddCandidates()
		if len(dv.depAddCandidates) == 0 {
			return func() tea.Msg { return FlashMsg{Text: "No tasks available to add as dependency"} }
		}
		dv.depAddSelectedIndex = 0
		dv.mode = DetailModeDepAdd
	case "-":
		deps := dv.getDependencies()
		if len(deps) == 0 {
			return func() tea.Msg { return FlashMsg{Text: "No dependencies to remove"} }
		}
		dv.depBrowseIndex = 0
		dv.mode = DetailModeDepBrowse
	case "u", "U":
		if dv.task.Status == core.StatusComplete {
			completedAt := dv.task.CompletedAt
			if err := dv.task.UpdateStatus(core.StatusTodo); err != nil {
				return func() tea.Msg { return FlashMsg{Text: "Cannot undo: " + err.Error()} }
			}
			if dv.tracker != nil {
				dv.tracker.RecordStatusChange()
				if completedAt != nil {
					dv.tracker.RecordUndoComplete(dv.task.ID, *completedAt)
				}
			}
			return func() tea.Msg { return TaskUndoneMsg{Task: dv.task} }
		}
	case "n", "N":
		if dv.enricher == nil {
			return func() tea.Msg { return FlashMsg{Text: "LLM not configured — enrichment unavailable"} }
		}
		desc := strings.TrimSpace(dv.task.Text)
		if desc == "" {
			return func() tea.Msg { return FlashMsg{Text: "Task has no text to enrich"} }
		}
		dv.mode = DetailModeEnrichLoading
		taskID := dv.task.ID
		return func() tea.Msg {
			return EnrichStartMsg{TaskID: taskID, TaskText: desc}
		}
	case "d", "D":
		if dv.isDuplicate && dv.dedupStore != nil && dv.duplicatePair != nil {
			_ = dv.dedupStore.RecordDecision(dv.duplicatePair.TaskA.ID, dv.duplicatePair.TaskB.ID, core.DecisionDistinct)
			dv.isDuplicate = false
			return func() tea.Msg { return DuplicateDismissedMsg{Task: dv.task} }
		}
	case "y", "Y":
		if dv.isDuplicate && dv.dedupStore != nil && dv.duplicatePair != nil {
			_ = dv.dedupStore.RecordDecision(dv.duplicatePair.TaskA.ID, dv.duplicatePair.TaskB.ID, core.DecisionDuplicate)
			// Remove the other task (the one that's not currently being viewed)
			otherTask := dv.duplicatePair.TaskB
			if dv.task.ID == dv.duplicatePair.TaskB.ID {
				otherTask = dv.duplicatePair.TaskA
			}
			dv.isDuplicate = false
			return func() tea.Msg { return DuplicateMergedMsg{Task: dv.task, RemovedTask: otherTask} }
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
			return nil
		}
		parentTask := dv.task
		dv.expandInput = ""
		dv.expandSubtaskCount++
		return func() tea.Msg {
			return ExpandTaskMsg{ParentTask: parentTask, NewTaskText: text}
		}
	case "esc":
		dv.mode = DetailModeView
		dv.expandInput = ""
		dv.expandSubtaskCount = 0
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

func (dv *DetailView) handleDepBrowse(msg tea.KeyMsg) tea.Cmd {
	deps := dv.getDependencies()
	switch msg.String() {
	case "esc":
		dv.mode = DetailModeView
	case "up", "k":
		if dv.depBrowseIndex > 0 {
			dv.depBrowseIndex--
		}
	case "down", "j":
		if dv.depBrowseIndex < len(deps)-1 {
			dv.depBrowseIndex++
		}
	case "enter":
		if dv.depBrowseIndex >= 0 && dv.depBrowseIndex < len(deps) {
			dep := deps[dv.depBrowseIndex]
			if dv.pool != nil {
				if target := dv.pool.GetTask(dep.ID); target != nil {
					return func() tea.Msg { return NavigateToLinkedMsg{Task: target} }
				}
			}
			return func() tea.Msg { return FlashMsg{Text: "Dependency task not found in pool"} }
		}
	case "-", "backspace", "delete":
		if dv.depBrowseIndex >= 0 && dv.depBrowseIndex < len(dv.task.DependsOn) {
			removedID := dv.task.DependsOn[dv.depBrowseIndex]
			dv.task.DependsOn = append(dv.task.DependsOn[:dv.depBrowseIndex], dv.task.DependsOn[dv.depBrowseIndex+1:]...)
			if len(dv.task.DependsOn) == 0 {
				dv.task.DependsOn = nil
				dv.mode = DetailModeView
			} else if dv.depBrowseIndex >= len(dv.task.DependsOn) {
				dv.depBrowseIndex = len(dv.task.DependsOn) - 1
			}
			task := dv.task
			return func() tea.Msg { return DependencyRemovedMsg{Task: task, DependencyID: removedID} }
		}
	}
	return nil
}

func (dv *DetailView) handleDepAdd(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		dv.mode = DetailModeView
		dv.depAddCandidates = nil
	case "up", "k":
		if dv.depAddSelectedIndex > 0 {
			dv.depAddSelectedIndex--
		}
	case "down", "j":
		if dv.depAddSelectedIndex < len(dv.depAddCandidates)-1 {
			dv.depAddSelectedIndex++
		}
	case "enter":
		if dv.depAddSelectedIndex >= 0 && dv.depAddSelectedIndex < len(dv.depAddCandidates) {
			candidate := dv.depAddCandidates[dv.depAddSelectedIndex]
			if core.WouldCreateCycle(dv.task.ID, candidate.ID, dv.pool) {
				return func() tea.Msg {
					return FlashMsg{Text: "Cannot add dependency: would create a circular chain"}
				}
			}
			dv.task.DependsOn = append(dv.task.DependsOn, candidate.ID)
			dv.mode = DetailModeView
			dv.depAddCandidates = nil
			task := dv.task
			depID := candidate.ID
			return func() tea.Msg { return DependencyAddedMsg{Task: task, DependencyID: depID} }
		}
	}
	return nil
}

func (dv *DetailView) handleEnrichResult(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "a", "A":
		result := dv.enrichResult
		taskID := dv.task.ID
		dv.mode = DetailModeView
		dv.enrichResult = nil
		return func() tea.Msg {
			return EnrichAcceptMsg{
				TaskID:       taskID,
				EnrichedText: result.EnrichedText,
				Tags:         result.Tags,
				Effort:       result.Effort,
				Context:      result.Context,
			}
		}
	case "d", "D", "esc":
		dv.mode = DetailModeView
		dv.enrichResult = nil
		return func() tea.Msg { return FlashMsg{Text: "Enrichment discarded"} }
	}
	return nil
}

func (dv *DetailView) handleDispatchConfirm(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		task := dv.task
		dv.mode = DetailModeView
		return func() tea.Msg { return DevDispatchRequestMsg{Task: task} }
	case "n", "N", "esc":
		dv.mode = DetailModeView
	}
	return nil
}

// FormatBlockedBy renders a "Blocked by: [text] (+N more)" indicator.
// The first blocker's text is truncated to maxLen characters.
func FormatBlockedBy(blockers []*core.Task, maxLen int) string {
	if len(blockers) == 0 {
		return ""
	}
	text := blockers[0].Text
	if len(text) > maxLen {
		text = text[:maxLen-3] + "..."
	}
	result := "Blocked by: " + text
	if len(blockers) > 1 {
		result += fmt.Sprintf(" (+%d more)", len(blockers)-1)
	}
	return result
}

// getDependencies returns all dependencies for the current task with their status.
func (dv *DetailView) getDependencies() []*core.Task {
	if dv.pool == nil || len(dv.task.DependsOn) == 0 {
		return nil
	}
	var deps []*core.Task
	for _, depID := range dv.task.DependsOn {
		dep := dv.pool.GetTask(depID)
		if dep == nil {
			deps = append(deps, &core.Task{
				ID:   depID,
				Text: "[deleted task]",
			})
		} else {
			deps = append(deps, dep)
		}
	}
	return deps
}

// buildDepAddCandidates returns tasks that can be added as dependencies
// (excluding current task and existing dependencies).
func (dv *DetailView) buildDepAddCandidates() []*core.Task {
	if dv.pool == nil {
		return nil
	}
	existing := make(map[string]bool)
	existing[dv.task.ID] = true
	for _, depID := range dv.task.DependsOn {
		existing[depID] = true
	}
	var candidates []*core.Task
	for _, t := range dv.pool.GetAllTasks() {
		if !existing[t.ID] {
			candidates = append(candidates, t)
		}
	}
	return candidates
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

	if dv.task.DevDispatch != nil && dv.task.DevDispatch.Queued {
		fmt.Fprintf(&s, "%s\n", DevDispatchBadge(dv.task))
	}

	if dv.isDuplicate {
		fmt.Fprintf(&s, "%s\n", DuplicateIndicator())
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

	// Show dependencies
	deps := dv.getDependencies()
	if len(deps) > 0 {
		depHeaderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		fmt.Fprintf(&s, "\n%s (%d):\n", depHeaderStyle.Render("Dependencies"), len(deps))
		for i, dep := range deps {
			checkbox := "[ ]"
			if dep.Status == core.StatusComplete {
				checkbox = "[x]"
			}
			prefix := "  "
			if dv.mode == DetailModeDepBrowse && i == dv.depBrowseIndex {
				prefix = "> "
			}
			fmt.Fprintf(&s, "%s%s %s\n", prefix, checkbox, dep.Text)
		}

		// Show blocked-by summary if there are blocking deps
		blockers := core.GetBlockingDependencies(dv.task, dv.pool)
		if len(blockers) > 0 {
			blockedStyle := lipgloss.NewStyle().Foreground(colorBlocked)
			s.WriteString(blockedStyle.Render(FormatBlockedBy(blockers, 40)))
			s.WriteString("\n")
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
		if dv.expandSubtaskCount > 0 {
			fmt.Fprintf(&s, "Subtask %d added. Next subtask (Esc to finish):\n", dv.expandSubtaskCount)
		} else {
			s.WriteString("New subtask (Enter to add, Esc to finish):\n")
		}
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
	case DetailModeDepBrowse:
		s.WriteString(helpStyle.Render("[Enter] Navigate [- / Del] Remove [Esc] Back"))
	case DetailModeDepAdd:
		s.WriteString("Select task to add as dependency (Enter to add, Esc to cancel):\n\n")
		for i, t := range dv.depAddCandidates {
			prefix := "  "
			if i == dv.depAddSelectedIndex {
				prefix = "> "
			}
			fmt.Fprintf(&s, "%s%s\n", prefix, t.Text)
		}
	case DetailModeEnrichLoading:
		s.WriteString("Enriching task...\n")
	case DetailModeEnrichResult:
		if dv.enrichResult != nil {
			origStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
			enrichStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
			labelStyle := lipgloss.NewStyle().Bold(true)

			fmt.Fprintf(&s, "%s\n", labelStyle.Render("Original:"))
			fmt.Fprintf(&s, "  %s\n\n", origStyle.Render(dv.enrichResult.OriginalText))
			fmt.Fprintf(&s, "%s\n", labelStyle.Render("Enriched:"))
			fmt.Fprintf(&s, "  %s\n", enrichStyle.Render(dv.enrichResult.EnrichedText))
			if len(dv.enrichResult.Tags) > 0 {
				fmt.Fprintf(&s, "  Tags: %s\n", strings.Join(dv.enrichResult.Tags, ", "))
			}
			if dv.enrichResult.Effort > 0 {
				fmt.Fprintf(&s, "  Effort: %d/5\n", dv.enrichResult.Effort)
			}
			if dv.enrichResult.Context != "" {
				fmt.Fprintf(&s, "  Context: %s\n", dv.enrichResult.Context)
			}
			s.WriteString("\n")
			s.WriteString(helpStyle.Render("[A]ccept [D]iscard [Esc] Cancel"))
		}
	case DetailModeDispatchConfirm:
		truncated := dv.task.Text
		if len(truncated) > 50 {
			truncated = truncated[:47] + "..."
		}
		fmt.Fprintf(&s, "Dispatch '%s' to dev queue? [y/n]\n", truncated)
	default:
		if dv.hintEnabled {
			h := func(key string) string { return renderInlineHint(key, dv.hintEnabled) }
			var parts []string
			parts = append(parts, h("esc")+" Back")
			parts = append(parts, h("c")+" Complete")
			parts = append(parts, h("b")+" Blocked")
			parts = append(parts, h("i")+" In-progress")
			parts = append(parts, h("e")+" Expand")
			parts = append(parts, h("f")+" Fork")
			parts = append(parts, h("p")+" Procrastinate")
			parts = append(parts, h("r")+" Rework")
			parts = append(parts, h("m")+" Mood")
			parts = append(parts, h("z")+" Snooze")
			if dv.task.Status == core.StatusComplete {
				parts = append(parts, h("u")+" Undo")
			}
			if dv.enrichDB != nil {
				parts = append(parts, h("l")+" Link")
			}
			if len(dv.crossRefs) > 0 {
				parts = append(parts, h("x")+" Xrefs")
			}
			parts = append(parts, h("g")+" Breakdown")
			if dv.agentService != nil {
				parts = append(parts, h("g")+" Decompose")
			}
			if dv.enricher != nil {
				parts = append(parts, h("n")+" Enrich")
			}
			if dv.isDuplicate && dv.dedupStore != nil {
				parts = append(parts, h("d")+" Dismiss-dup")
				parts = append(parts, h("y")+" Merge-dup")
			}
			if dv.devDispatchEnabled && dv.dispatcherAvailable {
				parts = append(parts, h("x")+" Dispatch")
			}
			if dv.pool != nil {
				parts = append(parts, h("+")+" Dep")
				parts = append(parts, h("-")+" Dep")
			}
			s.WriteString(helpStyle.Render(strings.Join(parts, " ")))
		} else {
			linkHint := ""
			if dv.enrichDB != nil {
				linkHint = " [L]ink"
			}
			browseHint := ""
			if len(dv.crossRefs) > 0 {
				browseHint = " [X]refs"
			}
			breakdownHint := " [G]breakdown"
			decomposeHint := ""
			if dv.agentService != nil {
				decomposeHint = " [G]enerate stories"
			}
			enrichHint := ""
			if dv.enricher != nil {
				enrichHint = " [N]enrich"
			}
			dupHint := ""
			if dv.isDuplicate && dv.dedupStore != nil {
				dupHint = " [D]ismiss-dup [Y]es-merge"
			}
			dispatchHint := ""
			if dv.devDispatchEnabled && dv.dispatcherAvailable {
				dispatchHint = " [X]dispatch"
			}
			undoHint := ""
			if dv.task.Status == core.StatusComplete {
				undoHint = " [U]ndo"
			}
			depHint := ""
			if dv.pool != nil {
				depHint = " [+]dep [-]dep"
			}
			s.WriteString(helpStyle.Render("[C]omplete [B]locked [I]n-progress [E]xpand [F]ork [P]rocrastinate [R]ework [M]ood [Z]Snooze" + undoHint + linkHint + browseHint + breakdownHint + dupHint + dispatchHint + depHint + " [Esc]Back"))
			s.WriteString(helpStyle.Render("[C]omplete [B]locked [I]n-progress [E]xpand [F]ork [P]rocrastinate [R]ework [M]ood [Z]Snooze" + undoHint + linkHint + browseHint + decomposeHint + enrichHint + dupHint + dispatchHint + depHint + " [Esc]Back"))
		}
	}

	return detailBorder.Width(w).Render(s.String())
}
